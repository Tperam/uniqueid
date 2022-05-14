package ringbuffer_test

import (
	"fmt"
	"github.com/tperam/uniqueid/internal/biz/ringbuffer"
	"sync"
	"testing"
	"time"
)

func TestRingBuffer(t *testing.T) {
	bufferSize := 65535
	signCh := make(chan struct{}, 1)
	waitCh := make(chan struct{}, 1)
	rb := ringbuffer.NewRingBuffer(uint64(bufferSize), signCh, waitCh)

	goNum := 10 * 10000
	perGoRange := 100
	arr := make([]uint64, goNum*perGoRange)
	wg := sync.WaitGroup{}
	for i := 0; i < goNum; i++ {
		wg.Add(1)
		go func(index int) {
			for i := 0; i < perGoRange; i++ {
				arr[index*perGoRange+i] = rb.GetID()
			}
			wg.Done()
			// t.Log("finish,", index)
		}(i)
	}

	go func() {
		startID := 100000
		step := 10000
		for {
			time.Sleep(10 * time.Millisecond)
			ids := make([]uint64, step)
			for i := 0; i < step; i++ {
				ids[i] = uint64(startID - step + i)
			}
			rb.Fill(ids)

			startID += step
		}
	}()
	wg.Wait()

	// 验证是否重复
	idMap := make(map[uint64]struct{}, goNum*perGoRange)
	for i := range arr {
		if _, ok := idMap[arr[i]]; ok {
			panic(fmt.Sprint("出现重复", arr[i]))
		}
		idMap[arr[i]] = struct{}{}
	}
	t.Log("成功")

}
