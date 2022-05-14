package ringbuffer

import (
	"sync"
	"sync/atomic"
)

type ringBuffer struct {
	buffer        []uint64
	mask          uint64
	consumeCursor uint64
	produceCursor uint64
	fillSign      chan struct{}
	wait          chan struct{}
	waitMu        sync.Mutex
	waitSign      uint64

	mu     sync.Mutex
	fillMu sync.Mutex
}

func NewRingBuffer(bufferSize uint64, fillSign, wait chan struct{}) *ringBuffer {
	if bufferSize == 0 {
		bufferSize = 1024
	}

	bufferSize = tableSizeFor(bufferSize)
	return &ringBuffer{
		buffer:   make([]uint64, bufferSize),
		mask:     bufferSize - 1,
		fillSign: fillSign,
		wait:     wait,
	}
}

func (rb *ringBuffer) GetID() uint64 {
	rb.mu.Lock()
	rb.consumeCursor++

	if rb.consumeCursor >= rb.produceCursor {
		// 通知更新
		//rb.fillSign <- struct{}{}

		rb.waitMu.Lock()
		if rb.consumeCursor >= rb.produceCursor {
			rb.waitSign = 1
			rb.waitMu.Unlock()
			<-rb.wait
		} else {
			rb.waitMu.Unlock()
		}
	}
	result := rb.buffer[rb.consumeCursor&rb.mask]
	rb.mu.Unlock()
	return result
}

func (rb *ringBuffer) Fill(tasks []uint64) int {
	if len(tasks) == 0 {
		return 0
	}
	rb.fillMu.Lock()
	//
	fillable := uint64(len(rb.buffer)) - (rb.produceCursor - rb.consumeCursor)
	if rb.produceCursor < rb.consumeCursor {
		fillable = uint64(len(rb.buffer))
	}
	if fillable > uint64(len(tasks)) {
		fillable = uint64(len(tasks))
	}

	// 填充
	for i := uint64(0); i < fillable; i++ {
		rb.buffer[(rb.produceCursor)&rb.mask] = tasks[i]
		atomic.AddUint64(&rb.produceCursor, 1)
	}

	rb.fillMu.Unlock()
	// 通知，取消阻塞
	rb.waitMu.Lock()
	if atomic.CompareAndSwapUint64(&rb.waitSign, 1, 0) {
		rb.wait <- struct{}{}
	}
	rb.waitMu.Unlock()

	return 0
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
