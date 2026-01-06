package data

import (
	"time"
)

// Picture 图片实体
type Picture struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	URL          string    `gorm:"column:url;type:varchar(512);not null" json:"url"`
	Name         string    `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Introduction string    `gorm:"column:introduction;type:varchar(512)" json:"introduction"`
	Category     string    `gorm:"column:category;type:varchar(64)" json:"category"`
	Tags         string    `gorm:"column:tags;type:varchar(512)" json:"tags"` // JSON 数组
	PicSize      int64     `gorm:"column:picSize" json:"picSize"`
	PicWidth     int32     `gorm:"column:picWidth" json:"picWidth"`
	PicHeight    int32     `gorm:"column:picHeight" json:"picHeight"`
	PicScale     float64   `gorm:"column:picScale" json:"picScale"`
	PicFormat    string    `gorm:"column:picFormat;type:varchar(32)" json:"picFormat"`
	UserID       int64     `gorm:"column:userId;not null" json:"userId"`
	CreateTime   time.Time `gorm:"column:createTime;autoCreateTime" json:"createTime"`
	EditTime     time.Time `gorm:"column:editTime;autoCreateTime" json:"editTime"`
	UpdateTime   time.Time `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`
	IsDelete     int8      `gorm:"column:isDelete;default:0" json:"isDelete"`
}

// TableName 指定表名
func (Picture) TableName() string {
	return "picture"
}
