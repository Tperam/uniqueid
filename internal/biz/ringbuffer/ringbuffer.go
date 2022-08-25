package ringbuffer

import (
	"sync"
	"sync/atomic"
)

type RingBuffer struct {
	buffer        []uint64
	mask          uint64
	consumeCursor uint64
	produceCursor uint64
	fillSign      chan struct{}
	wait          chan struct{}
	waitMu        sync.Mutex
	waitSign      uint64
	filled        bool
	mu            sync.Mutex
	fillMu        sync.Mutex
}

// NewRingBuffer
// @Description: New RingBuffer
// @param bufferSize buffer size
// @param fillSign The fillSign will get sign When need fill buffer
// @param wait  If filling is not timely, the wait will get sign
// @return *RingBuffer
func NewRingBuffer(bufferSize uint64, fillSign, wait chan struct{}) *RingBuffer {
	if bufferSize == 0 {
		bufferSize = 1024
	}

	bufferSize = tableSizeFor(bufferSize)
	return &RingBuffer{
		buffer:   make([]uint64, bufferSize),
		mask:     bufferSize - 1,
		fillSign: fillSign,
		filled:   true,
		wait:     wait,
	}
}

func (rb *RingBuffer) GetID() uint64 {
	rb.mu.Lock()

	if rb.produceCursor-rb.consumeCursor <= uint64(len(rb.buffer)/10) && rb.filled {
		rb.filled = false
		rb.fillSign <- struct{}{}
	}
	// 由于有锁，此处只会出现小于 或等于，我们只处理等于
	if rb.consumeCursor == rb.produceCursor {
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
	rb.consumeCursor++
	rb.mu.Unlock()
	return result
}

func (rb *RingBuffer) Fill(tasks []uint64) int {
	if len(tasks) == 0 {
		return 0
	}
	rb.fillMu.Lock()
	//
	fillable := len(rb.buffer) - int(rb.produceCursor-rb.consumeCursor)
	if rb.produceCursor < rb.consumeCursor {
		fillable = len(rb.buffer)
	}
	if fillable > len(tasks) {
		fillable = len(tasks)
	}

	// 填充
	for i := 0; i < fillable; i++ {
		rb.buffer[(rb.produceCursor+uint64(i))&rb.mask] = tasks[i]

	}
	rb.filled = true

	// 越界
	if rb.produceCursor+uint64(fillable) < rb.produceCursor {
		rb.produceCursor = rb.produceCursor & rb.mask
		rb.consumeCursor = rb.consumeCursor & rb.mask
	}
	rb.produceCursor += uint64(fillable)

	rb.fillMu.Unlock()
	// 通知，取消阻塞
	rb.waitMu.Lock()
	if atomic.CompareAndSwapUint64(&rb.waitSign, 1, 0) {
		rb.wait <- struct{}{}
	}
	rb.waitMu.Unlock()

	return fillable
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
