/*
 * @Author: Tperam
 * @Date: 2022-05-08 17:49:33
 * @LastEditTime: 2022-05-09 00:13:12
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\biz\unqiue_ringBuffer.go
 */
package biz

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/tperam/uniqueid/internal/dao"
)

type UniqueRingBuffer struct {
	IDBuffer  []uint64
	Mask      uint64
	Threshold int64

	// Head ~ Tail 是可用范围
	Head   uint64
	Tail   uint64
	lock   *sync.Mutex
	RWLock *sync.RWMutex

	uniqueDao *dao.UniqueDao
	bizTag    string
}

func (ub *UniqueRingBuffer) GetID() uint64 {
	re := atomic.AddUint64(&ub.Head, 1)

	if re >= ub.Tail {
		// 尝试抢锁
		ub.lock.Lock()
		// 判断是否还超出
		if re >= ub.Tail {
			// 调用更新函数
			ub.generate(context.TODO())
		}
		ub.lock.Unlock()
	}

	return ub.IDBuffer[re&ub.Mask]
}

func (ub *UniqueRingBuffer) Fill(startID uint64, num int) uint64 {
	rangeNum := len(ub.IDBuffer) - int(ub.Tail) - int(ub.Head)
	if rangeNum > len(ub.IDBuffer) {
		rangeNum = len(ub.IDBuffer)
	}
	if rangeNum > num {
		rangeNum = num
	}

	for i := ub.Tail; i < uint64(rangeNum); i++ {
		ub.IDBuffer[i&ub.Mask] = startID + i
	}

	return startID + uint64(rangeNum)
}

func (ub *UniqueRingBuffer) generate(ctx context.Context) error {
	r, err := ub.uniqueDao.GetSequence(ctx, ub.bizTag)
	if err != nil {
		return err
	}
	ub.Fill(uint64(r.MaxID-int64(r.Step)), int(r.MaxID))
	return nil
}
