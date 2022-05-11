/*
 * @Author: Tperam
 * @Date: 2022-05-11 00:43:07
<<<<<<< HEAD
 * @LastEditTime: 2022-05-11 22:05:23
=======
 * @LastEditTime: 2022-05-11 16:19:48
>>>>>>> 115e25d4325947d0f732f012fcd71defdf5e5fe1
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\disruptor_test.go
*/
package biz_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tperam/uniqueid/internal/biz"
)

func TestDisruptor(t *testing.T) {

	goNum := 100 * 10000
	perGoRange := 1
	arr := make([]uint64, goNum*perGoRange)
	// biz.NewUniqueChanFill()
	rb := biz.NewRingBuffer(65535, 1024)

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
