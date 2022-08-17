package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/tperam/uniqueid/internal/biz"
	"net/http"
)

func (ga *GinAdapt) PostID(builder *biz.IDBuilderBiz) func(c *gin.Context) {

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, builder.GetID())
	}
}
