/*
 * @Author: Tperam
 * @Date: 2022-05-08 21:08:14
 * @LastEditTime: 2022-05-08 21:46:45
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\normal_test.go
 */
package biz_test

import (
	"fmt"
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
