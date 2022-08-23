package biz_test

import (
	"github.com/rs/zerolog/log"
	"github.com/tperam/uniqueid/internal/biz"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestIDBuilder(t *testing.T) {

	//ctx := context.TODO()
	dsn := "root:1929564872@tcp(192.168.0.30:30306)/unique_id?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	gb := biz.NewIDBuidlerBiz(log.Logger, db, "test1")

	//length := 100000
	//result := make([]uint64, length)
	//for i := range result {
	//	result[i] = gb.GetID()
	//}
	gb.GetID()
	//for i := 0; i < len(result)-1; i++ {
	//	if result[i]+1 != result[i+1] {
	//		panic("error")
	//	}
	//}
	//t.Log(result)
}

func BenchmarkIDBuilder(b *testing.B) {

	//ctx := context.TODO()
	dsn := "root:1929564872@tcp(192.168.0.30:30306)/unique_id?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	gb := biz.NewIDBuidlerBiz(log.Logger, db, "test1")

	result := make([]uint64, b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result[i] = gb.GetID()
	}
}
