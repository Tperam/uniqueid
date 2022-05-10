/*
 * @Author: Tperam
 * @Date: 2022-05-09 16:50:56
 * @LastEditTime: 2022-05-10 18:52:48
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\ringbuffer_test.go
 */
package biz_test

import (
	"sync"
	"testing"
	"time"

	"github.com/tperam/uniqueid/internal/biz"
)

func TestRingbuffer(t *testing.T) {
	goNum := 100 * 10000
	perGoRange := 1
	arr := make([]uint64, goNum*perGoRange)
	rb := biz.NewRingBuffer()

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

			// t.Log(rb.Fill(uint64(startID), step), time.Now().Nanosecond())
			rb.Fill(uint64(startID), step)

			startID += step
		}
	}()
	wg.Wait()

	// 验证是否重复
	idMap := make(map[uint64]struct{}, goNum*perGoRange)
	for i := range arr {
		if _, ok := idMap[arr[i]]; ok {
			panic("出现重复")
		}
		idMap[arr[i]] = struct{}{}
	}
	t.Log("成功")
}

func BenchmarkRingBuffer(b *testing.B) {

	rb := biz.NewRingBuffer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			startID := 100000
			step := 10000
			for {
				time.Sleep(1 * time.Millisecond)
				rb.Fill(uint64(startID), step)
				startID += step
			}
		}
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rb.GetID()
	}

}
