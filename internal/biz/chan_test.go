/*
 * @Author: Tperam
 * @Date: 2022-05-10 17:01:52
 * @LastEditTime: 2022-05-11 01:11:43
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\chan_test.go
 */
package biz_test

import (
	"sync"
	"testing"
	"time"

	"github.com/tperam/uniqueid/internal/biz"
)

func TestChan(t *testing.T) {

	goNum := 10 * 10000
	perGoRange := 1
	arr := make([]uint64, goNum*perGoRange)
	// biz.NewUniqueChanFill()
	ch := make(chan uint64, 65535)
	rb := biz.NewUniqueChan(ch)
	u := biz.NewUniqueChanFill(nil, ch)

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
			u.Fill(uint64(startID), step)

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
