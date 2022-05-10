/*
 * @Author: Tperam
 * @Date: 2022-05-08 21:08:14
 * @LastEditTime: 2022-05-10 23:53:23
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\normal_test.go
 */
package biz_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestChanClose(t *testing.T) {
	ch := make(chan struct{})

	consumerFunc := func(i int) {
		fmt.Println("等待消费", i)
		<-ch
		fmt.Println("消费结束", i)
	}
	for i := 0; i < 10; i++ {
		go consumerFunc(i)
	}

	time.Sleep(1 * time.Second)
	go func() {
		close(ch)
	}()
	time.Sleep(1 * time.Second)
}

func TestMaxUInt(t *testing.T) {
	var a uint64
	t.Logf("%b, %b \n", a, ^a)

}
func TestTableSizeFor(t *testing.T) {
	t.Logf("%b \n", tableSizeFor(1))
	t.Logf("%b \n", tableSizeFor(88))
	t.Logf("%b \n", tableSizeFor(1<<62-1+8))

}

func tableSizeFor(cap uint64) uint64 {
	n := cap - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	if n < 0 {
		return 1
	} else {
		return n + 1
	}
}
func BenchmarkAdd(b *testing.B) {
	var a uint64
	for i := 0; i < b.N; i++ {
		a++
	}

}
func BenchmarkAtomicAdd(b *testing.B) {
	var a uint64
	for i := 0; i < 100; i++ {
		go func() {
			for i := 0; i < 10000; i++ {
				atomic.AddUint64(&a, 1)
			}
		}()
	}
	for i := 0; i < b.N; i++ {
		atomic.AddUint64(&a, 1)
	}

}

func BenchmarkChan(b *testing.B) {
	ch := make(chan uint64, b.N)
	for i := 0; i < b.N; i++ {
		ch <- uint64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-ch
	}

}
func BenchmarkLockAndRing(b *testing.B) {
	arr := make([]uint64, b.N)
	mu := sync.Mutex{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		_ = arr[i&0b111111111111111111111111111111]
		mu.Unlock()
	}

}
