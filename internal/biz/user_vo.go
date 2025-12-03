package biz

import "time"

// UserVO 用户视图对象（用于返回给前端，不包含敏感信息）
type UserVO struct {
	ID            int64      `json:"id"`
	UserAccount   string     `json:"userAccount"`
	UserName      string     `json:"userName"`
	UserAvatar    string     `json:"userAvatar"`
	UserProfile   string     `json:"userProfile"`
	UserRole      string     `json:"userRole"`
	VipNumber     int64      `json:"vipNumber"`
	VipExpireTime *time.Time `json:"vipExpireTime"`
	CreateTime    time.Time  `json:"createTime"`
}
