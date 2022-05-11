/*
 * @Author: Tperam
 * @Date: 2022-05-10 23:26:23
<<<<<<< HEAD
 * @LastEditTime: 2022-05-11 23:18:31
=======
 * @LastEditTime: 2022-05-11 16:22:17
>>>>>>> 115e25d4325947d0f732f012fcd71defdf5e5fe1
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\disruptor.go
*/
package biz

import (
	"sync"
	"sync/atomic"
)

type consume struct {
	mu sync.Mutex

	sign chan struct{}
}

type ringBuffer struct {
	buffer     []uint64
	bufferMask uint64

	consumers           []*consume
	consumerMask        uint64
	consumerWriter      []uint64
	consumerWriterMutex sync.RWMutex

	// 自增ID，确认落入哪个consumer
	increment uint64

	consumeCursor  uint64
	producerCursor uint64

	waitQueneLock sync.Mutex
	waitQuene     []uint64 // 处理锁

	producerMu sync.Mutex
	ltProducer []uint64
}

func tableSizeFor(cap uint64) uint64 {
	n := cap - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	if n == 0 {
		return 1
	} else {
		return n + 1
	}
}

func NewRingBuffer(bufferSize, consumerSize uint64) *ringBuffer {
	bufferSize = tableSizeFor(bufferSize)
	consumerSize = tableSizeFor(consumerSize)
	if consumerSize > bufferSize {
		panic("错误值")
	}
	tmpConsumer := make([]consume, consumerSize)
	consumers := make([]*consume, consumerSize)
	for i := range tmpConsumer {
		consumers[i] = &tmpConsumer[i]
		consumers[i].sign = make(chan struct{}, 1)
	}
	return &ringBuffer{
		buffer:         make([]uint64, bufferSize),
		bufferMask:     bufferSize - 1,
		consumers:      consumers,
		consumerMask:   consumerSize - 1,
		consumerWriter: make([]uint64, consumerSize),
		waitQuene:      make([]uint64, 0, consumerSize),
		ltProducer:     make([]uint64, 0, consumerSize),
	}
}

func (rb *ringBuffer) GetID() uint64 {
	consumerID := atomic.AddUint64(&rb.increment, 1) & rb.consumerMask

	rb.consumers[consumerID].mu.Lock()

	rb.consumerWriterMutex.RLock()
	consumeCursor := atomic.AddUint64(&rb.consumeCursor, 1)
	rb.consumerWriter[consumerID] = consumeCursor
	rb.consumerWriterMutex.RUnlock()
	// doslow
	if consumeCursor >= rb.producerCursor {

		rb.waitQueneLock.Lock()
		// 如果真的需要等待
		if consumeCursor >= rb.producerCursor {
			// 添加到等待队列
			rb.waitQuene = append(rb.waitQuene, consumerID)
			rb.waitQueneLock.Unlock()
			// 尝试阻塞
			<-rb.consumers[consumerID].sign
		} else {
			rb.waitQueneLock.Unlock()
		}

	}
	result := rb.buffer[consumeCursor&rb.bufferMask]

	rb.consumerWriterMutex.RLock()
	rb.consumerWriter[consumerID] = 0
	rb.consumerWriterMutex.RUnlock()
	rb.consumers[consumerID].mu.Unlock()
	return result
}

// 返回填充长度
func (rb *ringBuffer) Fill(ids []uint64) uint64 {
	rb.producerMu.Lock()

	// 查看占用指针
	rb.consumerWriterMutex.Lock()
	consumeCursor := rb.consumeCursor
	// lt produce arr
	ltProduce := rb.ltProducer[:0]
	for i := range rb.consumerWriter {
		if rb.consumerWriter[i] == 0 {
			continue
		}
		if rb.consumerWriter[i] < rb.producerCursor {
			rb.ltProducer = append(ltProduce, rb.consumerWriter[i])
			continue
		}
		if rb.consumerWriter[i] < rb.producerCursor {
			ltProduce = append(ltProduce, rb.consumerWriter[i])
			continue
		}
	}
	rb.consumerWriterMutex.Unlock()

	// 计算可填充容量
	fillable := uint64(len(rb.buffer)) - (rb.producerCursor - consumeCursor) - uint64(len(ltProduce))
	if rb.producerCursor < consumeCursor {
		fillable = uint64(len(rb.buffer) - len(ltProduce))
	}

	if fillable > uint64(len(ids)) {
		fillable = uint64(len(ids))
	}

	// 填充
	for i := uint64(0); i < fillable; i++ {
		for j := range ltProduce {
			if ltProduce[j]&rb.bufferMask == (rb.producerCursor+i)&rb.bufferMask {
				ltProduce = append(ltProduce[:j], ltProduce[j+1:]...)
				i++
				break
			}
		}
		rb.buffer[(rb.producerCursor+i)&rb.bufferMask] = ids[i]
	}
	// 更新生产指针
	atomic.AddUint64(&rb.producerCursor, fillable)
	rb.producerMu.Unlock()

	// 唤醒等待
	rb.waitQueneLock.Lock()
	for i := range rb.waitQuene {
		if rb.consumerWriter[rb.waitQuene[i]] < rb.producerCursor {
			rb.consumers[rb.waitQuene[i]].sign <- struct{}{}
		}
	}
	rb.waitQuene = rb.waitQuene[:0]
	rb.waitQueneLock.Unlock()

	return fillable
}
