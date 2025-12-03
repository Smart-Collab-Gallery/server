package data

import (
	"time"
)

// User 用户实体
type User struct {
	ID            int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserAccount   string     `gorm:"column:userAccount;type:varchar(256);not null;uniqueIndex:uk_userAccount" json:"userAccount"`
	UserPassword  string     `gorm:"column:userPassword;type:varchar(512);not null" json:"-"`
	UserName      string     `gorm:"column:userName;type:varchar(256);index:idx_userName" json:"userName"`
	UserAvatar    string     `gorm:"column:userAvatar;type:varchar(1024)" json:"userAvatar"`
	UserProfile   string     `gorm:"column:userProfile;type:varchar(512)" json:"userProfile"`
	UserRole      string     `gorm:"column:userRole;type:varchar(256);not null;default:user" json:"userRole"`
	VipExpireTime *time.Time `gorm:"column:vipExpireTime" json:"vipExpireTime"`
	VipCode       string     `gorm:"column:vipCode;type:varchar(128)" json:"vipCode"`
	VipNumber     int64      `gorm:"column:vipNumber" json:"vipNumber"`
	ShareCode     string     `gorm:"column:shareCode;type:varchar(20)" json:"shareCode"`
	InviteUser    int64      `gorm:"column:inviteUser" json:"inviteUser"`
	CreateTime    time.Time  `gorm:"column:createTime;autoCreateTime" json:"createTime"`
	UpdateTime    time.Time  `gorm:"column:updateTime;autoUpdateTime" json:"updateTime"`
	EditTime      time.Time  `gorm:"column:editTime;autoCreateTime" json:"editTime"`
	IsDelete      int8       `gorm:"column:isDelete;not null;default:0" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}

// UserRoleEnum 用户角色枚举
type UserRoleEnum string

const (
	UserRoleUser  UserRoleEnum = "user"
	UserRoleAdmin UserRoleEnum = "admin"
)
