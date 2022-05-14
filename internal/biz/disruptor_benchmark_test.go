/*
 * @Author: Tperam
 * @Date: 2022-05-11 22:10:14
 * @LastEditTime: 2022-05-11 23:27:51
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\disruptor_benchmark_test.go
 */
package biz_test

import (
	"sync"
	"testing"
	"time"

	"github.com/tperam/uniqueid/internal/biz"
)

var wg = &sync.WaitGroup{}

func BenchmarkDisruptor65536_1024(b *testing.B) {
	rb := biz.NewDisruptor(65536, 1024)

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
func BenchmarkDisruptor262144_1024(b *testing.B) {
	rb := biz.NewDisruptor(262144, 1024)

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
	for i := 0; i < 100*10000; i++ {
		wg.Add(1)
		go func() {
			rb.GetID()
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkDisruptor1024_2(b *testing.B) {
	rb := biz.NewDisruptor(1024, 2)

	go func() {
		startID := 100000
		step := 10000
		ids := make([]uint64, step)
		for {
			time.Sleep(10 * time.Millisecond)

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
		wg.Add(1)
		go func() {
			rb.GetID()
			wg.Done()
		}()
	}
	wg.Wait()
}
