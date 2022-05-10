/*
 * @Author: Tperam
 * @Date: 2022-05-10 17:25:30
 * @LastEditTime: 2022-05-10 17:53:14
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\lock.go
 */
package biz

import (
	"sync"
	"sync/atomic"
)

type ringbufferLock struct {
	buffer []uint64
	mask   uint64
	// 头尾指针
	// head 消费指针
	head uint64
	// tail 总生产队列指针
	tail uint64

	mu        sync.Mutex
	requestMu sync.Mutex
}

func NewRingbufferLock() *ringbufferLock {
	return &ringbufferLock{
		buffer: make([]uint64, 1<<10),
		mask:   1<<10 - 1,
	}
}
func (rb *ringbufferLock) GetID() uint64 {

	rb.mu.Lock()
	for {

		if rb.head < rb.tail {
			result := rb.buffer[rb.head&rb.mask]
			rb.head = rb.head + 1
			rb.mu.Unlock()
			return result
		} else {
			// 处理更新
			rb.requestMu.Lock()
			rb.requestMu.Unlock()
		}
	}
}

func (rb *ringbufferLock) Fill(startID uint64, num int) (endID uint64) {
	head := rb.head
	tail := rb.tail
	fillAmount := uint64(len(rb.buffer)) - (tail - head)
	if fillAmount > uint64(num) {
		fillAmount = uint64(num)
	}
	// 请求
	rb.requestMu.Lock()
	for i := uint64(0); i < fillAmount; i++ {
		rb.buffer[(tail+i)&rb.mask] = startID - uint64(num) + i
		atomic.AddUint64(&rb.tail, 1)
	}
	rb.requestMu.Unlock()

	return startID - uint64(num) + fillAmount
}
