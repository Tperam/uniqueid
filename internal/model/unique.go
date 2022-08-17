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
	BizTag     string    `json:"biz_tag"`
	MaxID      uint64    `json:"max_id"`
	Step       int       `json:"step"`
	Desc       string    `json:"desc"`
	UpdateTime time.Time `json:"update_time"`
}
