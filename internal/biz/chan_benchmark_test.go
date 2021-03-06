/*
 * @Author: Tperam
 * @Date: 2022-05-11 22:12:16
 * @LastEditTime: 2022-05-11 23:27:37
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\chan_benchmark_test.go
 */
package biz_test

import (
	"sync"
	"testing"

	"github.com/tperam/uniqueid/internal/biz"
)

func BenchmarkChan65536(b *testing.B) {
	ch := make(chan uint64, 65536)
	c := biz.NewUniqueChan(ch)
	cf := biz.NewUniqueChanFill(nil, ch)

	go func() {
		startID := 100000
		step := 10000
		ids := make([]uint64, step)
		for {
			for i := 0; i < step; i++ {
				ids[i] = uint64(startID - step + i)
			}
			cf.Fill(ids)

			startID += step
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			c.GetID()
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkChan262144(b *testing.B) {
	ch := make(chan uint64, 262144)
	c := biz.NewUniqueChan(ch)
	cf := biz.NewUniqueChanFill(nil, ch)

	go func() {
		startID := 100000
		step := 10000
		ids := make([]uint64, step)
		for {
			for i := 0; i < step; i++ {
				ids[i] = uint64(startID - step + i)
			}
			cf.Fill(ids)

			startID += step
		}
	}()

	b.ResetTimer()
	wg := &sync.WaitGroup{}
	for i := 0; i < 100*10000; i++ {
		wg.Add(1)
		go func() {
			c.GetID()
			wg.Done()
		}()
	}
	wg.Wait()
}
