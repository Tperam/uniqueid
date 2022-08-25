package main

import (
	"context"
	"github.com/rs/zerolog/log"
	ginserver "github.com/tperam/uniqueid/internal/adapt/gin"
	"github.com/tperam/uniqueid/internal/biz"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	ctx := context.TODO()
	dsn := "root:1929564872@tcp(192.168.0.30:30306)/unique_id?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	idBuilders := biz.NewIDBuilderBizs(log.Logger, db)

	ga := ginserver.NewGinAdapt(log.Logger, "0.0.0.0:8001", idBuilders)
	ga.NewRouter(ctx)

}
