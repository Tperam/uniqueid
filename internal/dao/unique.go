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
	"database/sql"
	"time"

	"github.com/tperam/uniqueid/internal/model"
)

type UniqueDao struct {
	db *sql.DB
}

/**
 * @Author: Tperam
 * @description: 获取数据库序列
 * @param {context.Context} ctx
 * @param {string} bizTag
 * @return {*}
 */
func (ud *UniqueDao) GetSequence(ctx context.Context, bizTag string) (seq *model.UnqiueID, err error) {
	tx, err := ud.db.Begin()
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, "UPDATE unique_id SET max_id = max_id + step AND update_time = ? WHERE biz_tag = ?", time.Now(), bizTag)
	if err != nil {
		return nil, err
	}

	row := ud.db.QueryRowContext(ctx, "SELECT biz_tag,max_id,step,desc,update_time FROM unique_id WHERE biz_tag = ?", bizTag)

	seq = &model.UnqiueID{}
	err = row.Scan(&seq.BizTag, &seq.MaxID, &seq.Step, &seq.Desc, &seq.UpdateTime)

	return seq, err
}
