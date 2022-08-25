package ginserver

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/tperam/uniqueid/internal/biz"
)

type GinAdapt struct {
	logger  zerolog.Logger
	addr    string
	builder *biz.IDBuilderBizs
}

func NewGinAdapt(logger zerolog.Logger, addr string, builder *biz.IDBuilderBizs) *GinAdapt {
	return &GinAdapt{logger: logger, addr: addr, builder: builder}
}

func (ga *GinAdapt) NewRouter(ctx context.Context) {
	r := gin.Default()

	r.POST("/id", ga.PostID(ga.builder))

	if err := r.Run(ga.addr); err != nil {
		ga.logger.Panic().Err(err).Msg("start server error!")
	}
}
