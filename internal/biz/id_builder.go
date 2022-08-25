package biz

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/tperam/uniqueid/internal/biz/ringbuffer"
	"github.com/tperam/uniqueid/internal/dao"
	"gorm.io/gorm"
	"sync"
	"time"
)

type IDBuilderBiz struct {
	rb       *ringbuffer.RingBuffer
	ud       *dao.UniqueDao
	log      zerolog.Logger
	fillPool *sync.Pool
	bizTag   string
}

func NewIDBuidlerBiz(log zerolog.Logger, db *gorm.DB, bizTag string) *IDBuilderBiz {
	ud := dao.NewUniqueDao(db)
	fillSign := make(chan struct{}, 1)
	wait := make(chan struct{}, 1)
	rb := ringbuffer.NewRingBuffer(10240, fillSign, wait)
	g := &IDBuilderBiz{
		ud:     ud,
		rb:     rb,
		bizTag: bizTag,
		log:    log,
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Err(err.(error)).Str("bizTag", bizTag).Msg("generation recover: err ")
			}
		}()
		// 不需要初始化，内部会初始化
		var task []uint64
		var err error
		for range fillSign {
			task, err = g.fill(task)
			for i := 0; err != nil && i < 3; i++ {
				task, err = g.fill(task)
			}
		}
	}()
	return g
}

func (g *IDBuilderBiz) GetID() uint64 {
	// 可加 select 用于处理超时
	return g.rb.GetID()
}

func (g *IDBuilderBiz) fill(task []uint64) ([]uint64, error) {
	// 开始填充
	if len(task) == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		seq, err := g.ud.GetSequence(ctx, g.bizTag)
		if err != nil {
			return task, err
		}

		if cap(task) < seq.Step {

			task = make([]uint64, 0, seq.Step)

		}

		id := seq.MaxID - uint64(seq.Step)

		for i := len(task); i < cap(task) && i < seq.Step; i++ {
			task = append(task, id+uint64(i))
		}
	}

	index := g.rb.Fill(task)
	copy(task, task[index:])
	return task[:len(task)-index], nil
}
