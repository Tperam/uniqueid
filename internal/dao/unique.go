/*
 * @Author: Tperam
 * @Date: 2022-05-08 17:30:18
 * @LastEditTime: 2022-05-08 17:43:36
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\dao\uniqueid.go
 */
package dao

import (
	"context"
	"gorm.io/gorm"
	"time"

	"github.com/tperam/uniqueid/internal/model"
)

type UniqueDao struct {
	db *gorm.DB
}

func NewUniqueDao(db *gorm.DB) *UniqueDao {
	return &UniqueDao{db: db}
}

/**
 * @Author: Tperam
 * @description: 获取数据库序列
 * @param {context.Context} ctx
 * @param {string} bizTag
 * @return {*}
 */
func (ud *UniqueDao) GetSequence(ctx context.Context, bizTag string) (seq *model.UnqiueID, err error) {

	err = ud.db.Transaction(func(tx *gorm.DB) error {
		if err != nil {
			return err
		}
		err = tx.WithContext(ctx).Exec("UPDATE unique_id SET max_id = max_id + step, update_time = ? WHERE biz_tag = ?", time.Now(), bizTag).Error
		if err != nil {
			return err
		}

		return tx.WithContext(ctx).Table("unique_id").Where("biz_tag=?", bizTag).Find(&seq).Error

	})

	return seq, err
}
