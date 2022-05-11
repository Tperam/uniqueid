/*
 * @Author: Tperam
 * @Date: 2022-05-08 23:55:29
 * @LastEditTime: 2022-05-11 22:13:42
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\chan.go
 */
package biz

import (
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

func (ucf *UnqiueChanFill) Fill(ids []uint64) error {
	for i := 0; i < len(ids); i++ {
		ucf.idCh <- ids[i]
	}
	return nil
	// r, err := ucf.ud.GetSequence(context.TODO(), ucf.bizTag)
	// if err != nil {
	// 	return err
	// }
	// for i := r.MaxID - int64(r.Step); i < r.MaxID; i++ {
	// 	ucf.idCh <- uint64(i)
	// }
	// return err
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
				// ucg.ucf.Fill()
				atomic.StoreUint64(&ucg.state, 0)
			}()
		}
	}

	return ucg.uc.GetID()
}
