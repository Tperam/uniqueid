package gin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/tperam/uniqueid/internal/biz"
)

type GinAdapt struct {
}

func (ga *GinAdapt) NewRouter(ctx context.Context, addr string, builder *biz.IDBuilderBiz, log zerolog.Logger) {
	r := gin.Default()

	r.POST("/id", ga.PostID(builder))

	if err := r.Run(addr); err != nil {
		log.Panic().Err(err).Msg("start server error!")
	}
}
