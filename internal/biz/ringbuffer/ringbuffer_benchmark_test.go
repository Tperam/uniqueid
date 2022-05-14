/*
 * @Author: Tperam
 * @Date: 2022-05-11 22:12:16
 * @LastEditTime: 2022-05-11 23:27:37
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\chan_benchmark_test.go
 */
package ringbuffer_test

import (
	"github.com/tperam/uniqueid/internal/biz/ringbuffer"
	"sync"
	"testing"
)

var wg = &sync.WaitGroup{}

func BenchmarkRingBuffer65536(b *testing.B) {
	bufferSize := 65535
	signCh := make(chan struct{}, 1)
	waitCh := make(chan struct{}, 1)
	rb := ringbuffer.NewRingBuffer(uint64(bufferSize), signCh, waitCh)

	go func() {
		startID := 100000
		step := 10000
		ids := make([]uint64, step)
		for {
			for i := 0; i < step; i++ {
				ids[i] = uint64(startID - step + i)
			}
			rb.Fill(ids)

			startID += step
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			rb.GetID()
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkRingBuffer262144(b *testing.B) {
	bufferSize := 262144
	signCh := make(chan struct{}, 1)
	waitCh := make(chan struct{}, 1)
	rb := ringbuffer.NewRingBuffer(uint64(bufferSize), signCh, waitCh)

	go func() {
		startID := 100000
		step := 10000
		ids := make([]uint64, step)
		for {
			for i := 0; i < step; i++ {
				ids[i] = uint64(startID - step + i)
			}
			rb.Fill(ids)

			startID += step
		}
	}()

	b.ResetTimer()
	wg := &sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		for i := 0; i < 100*10000; i++ {
			wg.Add(1)
			go func() {
				rb.GetID()
				wg.Done()
			}()
		}
	}
	wg.Wait()

}
