package biz

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/tperam/uniqueid/internal/biz/ringbuffer"
	"github.com/tperam/uniqueid/internal/dao"
	"sync"
	"time"
)

type IDBuilderBiz struct {
	rb       *ringbuffer.RingBuffer
	ud       *dao.UniqueDao
	log      *zerolog.Logger
	fillPool *sync.Pool
	bizTag   string
}

func NewGenerationBiz(log *zerolog.Logger, ud *dao.UniqueDao, bizTag string) *IDBuilderBiz {
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
				// printLog
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
	// 如果有 70% 容量， 则不进行更新
	if len(task) != 0 {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		seq, err := g.ud.GetSequence(ctx, g.bizTag)
		if err != nil {
			return task, err
		}

		if cap(task) < seq.Step {

			new := make([]uint64, len(task), seq.Step)
			copy(new, task)
			task = new

		}

		id := seq.MaxID - uint64(seq.Step)
		for i := len(task); i < cap(task) && i < seq.Step; i++ {
			task = append(task, id+uint64(i))
		}
	}
	index := g.rb.Fill(task)
	copy(task[0:], task[index:])
	return task[:len(task)-index], nil
}
