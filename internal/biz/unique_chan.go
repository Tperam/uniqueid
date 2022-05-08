/*
 * @Author: Tperam
 * @Date: 2022-05-08 23:55:29
 * @LastEditTime: 2022-05-09 00:11:37
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\unique_chan.go
 */
package biz

import (
	"context"
	"sync/atomic"

	"github.com/tperam/uniqueid/internal/dao"
)

type UniqueChan struct {
	idCh chan uint64
}

func NewUniqueChan(c chan uint64) *UniqueChan {
	return &UniqueChan{
		idCh: c,
	}
}

func (uc *UniqueChan) GetID() uint64 {
	return <-uc.idCh
}

type UnqiueChanFill struct {
	ud     *dao.UniqueDao
	idCh   chan uint64
	bizTag string
}

func NewUniqueChanFill(ud *dao.UniqueDao, idChan chan uint64) *UnqiueChanFill {
	return &UnqiueChanFill{
		ud:   ud,
		idCh: idChan,
	}
}

func (ucf *UnqiueChanFill) Fill() error {
	r, err := ucf.ud.GetSequence(context.TODO(), ucf.bizTag)
	if err != nil {
		return err
	}
	for i := r.MaxID - int64(r.Step); i < r.MaxID; i++ {
		ucf.idCh <- uint64(i)
	}
	return err
}

type UniqueChanGenerateBiz struct {
	uc            *UniqueChan
	ucf           *UnqiueChanFill
	generateCount uint64
	consumeCount  uint64
	state         uint64
	threshold     uint64
}

func NewUniqueChanGenerateBiz() *UniqueChanGenerateBiz {
	ch := make(chan uint64, 10000)
	uniqueChan := NewUniqueChan(ch)
	return &UniqueChanGenerateBiz{
		uc: uniqueChan,
	}
}

func (ucg *UniqueChanGenerateBiz) GetID() uint64 {
	consumeCount := atomic.AddUint64(&ucg.consumeCount, 1)
	if ucg.generateCount-consumeCount < ucg.threshold {
		if atomic.CompareAndSwapUint64(&ucg.state, 0, 1) {
			go func() {
				ucg.ucf.Fill()
				atomic.StoreUint64(&ucg.state, 0)
			}()
		}
	}

	return ucg.uc.GetID()
}
