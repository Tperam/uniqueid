/*
 * @Author: Tperam
 * @Date: 2022-05-10 23:26:23
 * @LastEditTime: 2022-05-12 00:22:47
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\disruptor.go
 */
package biz

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type consume struct {
	mu   sync.Mutex
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
	incr := atomic.AddUint64(&rb.increment, 1)
	consumerID := incr & rb.consumerMask

	rb.consumers[consumerID].mu.Lock()

	consumeCursor := atomic.AddUint64(&rb.consumeCursor, 1)
	rb.consumerWriter[consumerID] = consumeCursor
	fmt.Println("consumerID ", consumerID, "消费指针", consumeCursor, "生产指针", rb.producerCursor, "incr_id", incr)
	// doslow
	if consumeCursor >= rb.producerCursor {

		rb.waitQueneLock.Lock()
		// 如果真的需要等待
		if consumeCursor >= rb.producerCursor {
			// 添加到等待队列
			rb.waitQuene = append(rb.waitQuene, consumerID)
			rb.waitQueneLock.Unlock()
			// 尝试阻塞
			fmt.Println("consumerID", consumerID, "阻塞", consumeCursor)
			<-rb.consumers[consumerID].sign
			fmt.Println("consumerID", consumerID, "解除阻塞", consumeCursor)
		} else {
			rb.waitQueneLock.Unlock()
		}

	}
	result := rb.buffer[consumeCursor&rb.bufferMask]

	rb.consumers[consumerID].mu.Unlock()
	fmt.Println("consumerID ", consumerID, "消费完成,指针", consumeCursor, "incr_id", incr)
	return result
}

// 返回填充长度
func (rb *ringBuffer) Fill(ids []uint64) uint64 {
	if len(ids) == 0 {
		return 0
	}
	rb.producerMu.Lock()

	// 查找最小指针
	minConsumeCursor := min(rb.consumerWriter)

	// 计算可填充容量
	fillable := uint64(len(rb.buffer)) - (rb.producerCursor - minConsumeCursor)
	if rb.producerCursor < minConsumeCursor {
		fillable = uint64(len(rb.buffer))
	}

	if fillable > uint64(len(ids)) {
		fillable = uint64(len(ids))
	}

	// 如果等于0，则代表最小消费指针为极低，影响到循环填充
	// 此时则等待最小消费指针进行更新
	if fillable == 0 {
		// 重新获取
		tmp := min(rb.consumerWriter)
		// 自旋100次
		for i := 0; i < 100 && tmp == minConsumeCursor; i++ {
			tmp = min(rb.consumerWriter)
		}
		// 尝试睡眠等待
		sleepTime := 1
		for i := 0; tmp == minConsumeCursor; i++ {
			if i == 1000 {
				sleepTime++
			}
			tmp = min(rb.consumerWriter)
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			fmt.Println(tmp)
		}
		minConsumeCursor = tmp
		if rb.producerCursor < minConsumeCursor {
			fillable = uint64(len(rb.buffer))
		}

		if fillable > uint64(len(ids)) {
			fillable = uint64(len(ids))
		}
	}

	// 填充
	for i := uint64(0); i < fillable; i++ {
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

func min(arr []uint64) uint64 {
	if len(arr) == 0 {
		return 0
	}
	minConsumeCursor := arr[0]
	for i := 1; i < len(arr); i++ {
		if minConsumeCursor > arr[i] {
			minConsumeCursor = arr[i]
		}
	}
	return minConsumeCursor
}
