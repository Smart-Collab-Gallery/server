package biz

import (
	"encoding/json"
	"time"
)

// PictureVO 图片视图对象
type PictureVO struct {
	ID           int64     `json:"id"`
	URL          string    `json:"url"`
	Name         string    `json:"name"`
	Introduction string    `json:"introduction"`
	Tags         []string  `json:"tags"`
	Category     string    `json:"category"`
	PicSize      int64     `json:"picSize"`
	PicWidth     int32     `json:"picWidth"`
	PicHeight    int32     `json:"picHeight"`
	PicScale     float64   `json:"picScale"`
	PicFormat    string    `json:"picFormat"`
	UserID       int64     `json:"userId"`
	CreateTime   time.Time `json:"createTime"`
	EditTime     time.Time `json:"editTime"`
	UpdateTime   time.Time `json:"updateTime"`
	User         *UserVO   `json:"user,omitempty"` // 创建用户信息
}

// Picture 业务对象
type Picture struct {
	ID           int64
	URL          string
	Name         string
	Introduction string
	Tags         string // JSON 字符串
	Category     string
	PicSize      int64
	PicWidth     int32
	PicHeight    int32
	PicScale     float64
	PicFormat    string
	UserID       int64
	CreateTime   time.Time
	EditTime     time.Time
	UpdateTime   time.Time
	IsDelete     int8
}

// PictureQueryParams 图片查询参数
type PictureQueryParams struct {
	Current      int64
	PageSize     int64
	Name         string
	Introduction string
	Category     string
	Tags         []string
	UserID       *int64
	SortField    string
	SortOrder    string // ascend 或 descend
}

// PicturePage 图片分页结果
type PicturePage struct {
	Total    int64
	List     []*PictureVO
	Current  int64
	PageSize int64
}

// ObjToVO 对象转 VO
func (p *Picture) ObjToVO() *PictureVO {
	if p == nil {
		return nil
	}

	vo := &PictureVO{
		ID:           p.ID,
		URL:          p.URL,
		Name:         p.Name,
		Introduction: p.Introduction,
		Category:     p.Category,
		PicSize:      p.PicSize,
		PicWidth:     p.PicWidth,
		PicHeight:    p.PicHeight,
		PicScale:     p.PicScale,
		PicFormat:    p.PicFormat,
		UserID:       p.UserID,
		CreateTime:   p.CreateTime,
		EditTime:     p.EditTime,
		UpdateTime:   p.UpdateTime,
	}

	// 解析 JSON 标签
	if p.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(p.Tags), &tags); err == nil {
			vo.Tags = tags
		}
	}

	return vo
}

// VOToObj VO 转对象
func VOToObj(vo *PictureVO) *Picture {
	if vo == nil {
		return nil
	}

	obj := &Picture{
		ID:           vo.ID,
		URL:          vo.URL,
		Name:         vo.Name,
		Introduction: vo.Introduction,
		Category:     vo.Category,
		PicSize:      vo.PicSize,
		PicWidth:     vo.PicWidth,
		PicHeight:    vo.PicHeight,
		PicScale:     vo.PicScale,
		PicFormat:    vo.PicFormat,
		UserID:       vo.UserID,
		CreateTime:   vo.CreateTime,
		EditTime:     vo.EditTime,
		UpdateTime:   vo.UpdateTime,
	}

	// 转换标签为 JSON
	if len(vo.Tags) > 0 {
		if tagsBytes, err := json.Marshal(vo.Tags); err == nil {
			obj.Tags = string(tagsBytes)
		}
	}

	return obj
}
