/*
 * @Author: Tperam
 * @Date: 2022-05-09 10:10:23
 * @LastEditTime: 2022-05-10 18:55:21
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\ringbuffer.go
 */
package biz

import (
	"sync"
	"sync/atomic"
)

type ringbuffer struct {
	buffer []uint64
	mask   uint64
	// 头尾指针
	// head 消费指针
	head uint64
	// tail 总生产队列指针
	tail uint64

	notice chan struct{}

	mu sync.Mutex
}

func NewRingBuffer() *ringbuffer {
	return &ringbuffer{
		buffer: make([]uint64, 1<<10),
		mask:   1<<10 - 1,
		notice: make(chan struct{}),
	}
}
func (rb *ringbuffer) GetID() uint64 {
	for {

		head := rb.head
		if head < rb.tail {
			result := rb.buffer[head&rb.mask]
			if atomic.CompareAndSwapUint64(&rb.head, head, head+1) {
				return result
			} else {
				continue
			}
		}
		<-rb.notice

	}
}

func (rb *ringbuffer) Fill(startID uint64, num int) (endID uint64) {

	head := rb.head
	tail := rb.tail
	bufferLen := len(rb.buffer)
	fillAmount := uint64(bufferLen) - (tail - head)
	if fillAmount > uint64(num) {
		fillAmount = uint64(num)
	}
	for i := uint64(0); i < fillAmount; i++ {
		rb.buffer[(tail+i)&rb.mask] = startID - uint64(num) + i
		// atomic.AddUint64(&rb.tail, 1)
	}
	atomic.StoreUint64(&rb.tail, rb.tail+fillAmount)
	for rb.head < rb.tail {
		rb.notice <- struct{}{}
	}
	return startID - uint64(num) + fillAmount
}
