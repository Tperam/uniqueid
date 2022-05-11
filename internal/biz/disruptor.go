/*
 * @Author: Tperam
 * @Date: 2022-05-10 23:26:23
 * @LastEditTime: 2022-05-11 16:22:17
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\disruptor.go
 */
package biz

import (
	"sync"
	"sync/atomic"
	"time"
)

type consume struct {
	mu        sync.Mutex
	consuming uint64
	sign      chan struct{}
}

type ringBuffer struct {
	buffer       []uint64
	bufferMask   uint64
	consumers    []*consume
	consumerMask uint64

	// 自增ID，确认落入哪个consumer
	increment uint64

	consumeCursor  uint64
	producerCursor uint64

	waitQueneLock sync.Mutex
	waitQuene     []uint64 // 处理锁

	producerMu sync.Mutex
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
		buffer:       make([]uint64, bufferSize),
		bufferMask:   bufferSize - 1,
		consumers:    consumers,
		consumerMask: consumerSize - 1,
		waitQuene:    make([]uint64, 0, consumerSize),
	}
}

func (rb *ringBuffer) GetID() uint64 {
	consumerID := atomic.AddUint64(&rb.increment, 1) & rb.consumerMask

	rb.consumers[consumerID].mu.Lock()

	consumeCursor := atomic.AddUint64(&rb.consumeCursor, 1)
	rb.consumers[consumerID].consuming = consumeCursor
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
	rb.consumers[consumerID].consuming = 0
	rb.consumers[consumerID].mu.Unlock()
	return result
}

// 返回填充长度
func (rb *ringBuffer) Fill(ids []uint64) uint64 {
	rb.producerMu.Lock()

	// 等待消耗完毕
	lessConsumers := make([]int, 0, len(rb.consumers))
	for i := range rb.consumers {
		if rb.consumers[i].consuming < rb.producerCursor {
			lessConsumers = append(lessConsumers, i)
		}
	}
	// 自旋等待
	for i := 0; i < 100; i++ {
		for j := 0; j < len(lessConsumers); j++ {
			if rb.consumers[lessConsumers[j]].consuming >= rb.producerCursor {
				lessConsumers = append(lessConsumers[:j], lessConsumers[j+1:]...)
				j--
			}
		}
		if len(lessConsumers) == 0 {
			break
		}
	}
	// 睡眠等待
	for len(lessConsumers) != 0 {
		for j := 0; j < len(lessConsumers); j++ {
			if rb.consumers[lessConsumers[j]].consuming >= rb.producerCursor {
				lessConsumers = append(lessConsumers[:j], lessConsumers[j+1:]...)
				j--
			}
		}
		time.Sleep(1 * time.Millisecond)
	}

	minConsumed := rb.consumers[0].consuming
	for i := range rb.consumers {
		if rb.consumers[i].consuming < minConsumed {
			minConsumed = rb.consumers[i].consuming
		}
	}

	// 确定可填充数值
	fillable := uint64(len(rb.buffer)) - (rb.producerCursor - minConsumed)
	if rb.producerCursor < minConsumed {
		fillable = uint64(len(rb.buffer))
	}

	if fillable > uint64(len(ids)) {
		fillable = uint64(len(ids))
	}
	// fmt.Println(fillable, len(rb.buffer), rb.producerCursor, minConsumed, (rb.producerCursor - minConsumed))
	// 填充
	for i := uint64(0); i < fillable; i++ {
		rb.buffer[(rb.producerCursor+i)&rb.bufferMask] = ids[i]
	}
	atomic.AddUint64(&rb.producerCursor, fillable)

	rb.producerMu.Unlock()
	// 解锁
	rb.waitQueneLock.Lock()
	for i := range rb.waitQuene {
		if rb.consumers[rb.waitQuene[i]].consuming < rb.producerCursor {
			rb.consumers[rb.waitQuene[i]].sign <- struct{}{}
		}
	}
	
	rb.waitQuene = rb.waitQuene[:0]
	rb.waitQueneLock.Unlock()
	return fillable
}
