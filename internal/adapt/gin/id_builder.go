package ginserver

import (
	"github.com/gin-gonic/gin"
	"github.com/tperam/uniqueid/internal/biz"
	"net/http"
)

func (ga *GinAdapt) PostID(builder *biz.IDBuilderBizs) func(c *gin.Context) {

	return func(c *gin.Context) {

		biz := c.Query("biz")
		if biz == "" {
			c.JSON(http.StatusBadRequest, 0)
		}
		c.JSON(http.StatusOK, builder.GetID(biz))
	}
}
