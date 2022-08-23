/*
 * @Author: Tperam
 * @Date: 2022-05-08 17:38:02
 * @LastEditTime: 2022-05-08 17:39:09
 * @LastEditors: Tperam
 * @Description:
 * @FilePath: \uniqueid\internal\model\uniqueid.go
 */
package model

import "time"

type UnqiueID struct {
	BizTag     string    `json:"biz_tag" gorm:"column:biz_tag"`
	MaxID      uint64    `json:"max_id" gorm:"column:max_id"`
	Step       int       `json:"step" gorm:"column:step"`
	Desc       string    `json:"desc" gorm:"column:desc"`
	StartTime  time.Time `json:"start_time" gorm:"column:start_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
}
