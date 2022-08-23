package dao_test

import (
	"context"
	"github.com/tperam/uniqueid/internal/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestUniqueDao(t *testing.T) {

	ctx := context.TODO()
	dsn := "root:1929564872@tcp(192.168.0.30:30306)/unique_id?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	ud := dao.NewUniqueDao(db)
	r, err := ud.GetSequence(ctx, "test1")
	t.Log(r, err)
}
