package data

import (
	"context"
	"encoding/json"

	"smart-collab-gallery-server/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type pictureRepo struct {
	data *Data
	log  *log.Helper
}

// NewPictureRepo 创建图片仓储
func NewPictureRepo(data *Data, logger log.Logger) biz.PictureRepo {
	return &pictureRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreatePicture 创建图片
func (r *pictureRepo) CreatePicture(ctx context.Context, picture *biz.Picture) (*biz.Picture, error) {
	pictureEntity := &Picture{
		URL:          picture.URL,
		Name:         picture.Name,
		Introduction: picture.Introduction,
		Category:     picture.Category,
		Tags:         picture.Tags,
		PicSize:      picture.PicSize,
		PicWidth:     picture.PicWidth,
		PicHeight:    picture.PicHeight,
		PicScale:     picture.PicScale,
		PicFormat:    picture.PicFormat,
		UserID:       picture.UserID,
	}

	if err := r.data.db.WithContext(ctx).Create(pictureEntity).Error; err != nil {
		r.log.Errorf("创建图片失败: %v", err)
		return nil, err
	}

	picture.ID = pictureEntity.ID
	picture.CreateTime = pictureEntity.CreateTime
	picture.UpdateTime = pictureEntity.UpdateTime
	picture.EditTime = pictureEntity.EditTime

	return picture, nil
}

// GetPictureByID 根据 ID 查询图片
func (r *pictureRepo) GetPictureByID(ctx context.Context, id int64) (*biz.Picture, error) {
	var pictureEntity Picture
	err := r.data.db.WithContext(ctx).
		Where("id = ? AND isDelete = 0", id).
		First(&pictureEntity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("查询图片失败: %v", err)
		return nil, err
	}

	return r.convertToPicture(&pictureEntity), nil
}

// UpdatePicture 更新图片
func (r *pictureRepo) UpdatePicture(ctx context.Context, picture *biz.Picture) error {
	updates := map[string]interface{}{
		"name":         picture.Name,
		"introduction": picture.Introduction,
		"category":     picture.Category,
		"tags":         picture.Tags,
		"editTime":     picture.EditTime,
	}

	// 如果是重新上传，更新图片信息
	if picture.URL != "" {
		updates["url"] = picture.URL
		updates["picSize"] = picture.PicSize
		updates["picWidth"] = picture.PicWidth
		updates["picHeight"] = picture.PicHeight
		updates["picScale"] = picture.PicScale
		updates["picFormat"] = picture.PicFormat
	}

	err := r.data.db.WithContext(ctx).
		Model(&Picture{}).
		Where("id = ? AND isDelete = 0", picture.ID).
		Updates(updates).Error

	if err != nil {
		r.log.Errorf("更新图片失败: %v", err)
		return err
	}

	return nil
}

// DeletePicture 删除图片（逻辑删除）
func (r *pictureRepo) DeletePicture(ctx context.Context, id int64) error {
	err := r.data.db.WithContext(ctx).
		Model(&Picture{}).
		Where("id = ? AND isDelete = 0", id).
		Update("isDelete", 1).Error

	if err != nil {
		r.log.Errorf("删除图片失败: %v", err)
		return err
	}

	return nil
}

// ListPictureByPage 分页查询图片
func (r *pictureRepo) ListPictureByPage(ctx context.Context, params *biz.PictureQueryParams) (*biz.PicturePage, error) {
	var total int64
	var pictures []Picture

	// 构建查询
	query := r.data.db.WithContext(ctx).Model(&Picture{}).Where("isDelete = 0")

	// 搜索词查询（同时搜索名称和简介）
	if params.SearchText != "" {
		query = query.Where("name LIKE ? OR introduction LIKE ?",
			"%"+params.SearchText+"%", "%"+params.SearchText+"%")
	}

	// 条件查询
	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Introduction != "" {
		query = query.Where("introduction LIKE ?", "%"+params.Introduction+"%")
	}
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}
	if params.UserID != nil {
		query = query.Where("userId = ?", *params.UserID)
	}

	// 标签查询（JSON 数组）
	if len(params.Tags) > 0 {
		for _, tag := range params.Tags {
			query = query.Where("JSON_CONTAINS(tags, ?)", `"`+tag+`"`)
		}
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("统计图片总数失败: %v", err)
		return nil, err
	}

	// 排序
	sortField := "createTime"
	if params.SortField != "" {
		sortField = params.SortField
	}
	sortOrder := "desc"
	if params.SortOrder == "ascend" {
		sortOrder = "asc"
	}
	query = query.Order(sortField + " " + sortOrder)

	// 分页
	if params.Current > 0 && params.PageSize > 0 {
		offset := (params.Current - 1) * params.PageSize
		query = query.Offset(int(offset)).Limit(int(params.PageSize))
	}

	// 查询数据
	if err := query.Find(&pictures).Error; err != nil {
		r.log.Errorf("查询图片列表失败: %v", err)
		return nil, err
	}

	// 转换为 VO
	list := make([]*biz.PictureVO, 0, len(pictures))
	for _, pic := range pictures {
		bizPic := r.convertToPicture(&pic)
		list = append(list, bizPic.ObjToVO())
	}

	return &biz.PicturePage{
		Total:    total,
		List:     list,
		Current:  params.Current,
		PageSize: params.PageSize,
	}, nil
}

// convertToPicture 转换实体为业务对象
func (r *pictureRepo) convertToPicture(entity *Picture) *biz.Picture {
	return &biz.Picture{
		ID:           entity.ID,
		URL:          entity.URL,
		Name:         entity.Name,
		Introduction: entity.Introduction,
		Category:     entity.Category,
		Tags:         entity.Tags,
		PicSize:      entity.PicSize,
		PicWidth:     entity.PicWidth,
		PicHeight:    entity.PicHeight,
		PicScale:     entity.PicScale,
		PicFormat:    entity.PicFormat,
		UserID:       entity.UserID,
		CreateTime:   entity.CreateTime,
		EditTime:     entity.EditTime,
		UpdateTime:   entity.UpdateTime,
		IsDelete:     entity.IsDelete,
	}
}

// convertToEntity 转换业务对象为实体
func (r *pictureRepo) convertToEntity(picture *biz.Picture) *Picture {
	return &Picture{
		ID:           picture.ID,
		URL:          picture.URL,
		Name:         picture.Name,
		Introduction: picture.Introduction,
		Category:     picture.Category,
		Tags:         picture.Tags,
		PicSize:      picture.PicSize,
		PicWidth:     picture.PicWidth,
		PicHeight:    picture.PicHeight,
		PicScale:     picture.PicScale,
		PicFormat:    picture.PicFormat,
		UserID:       picture.UserID,
		CreateTime:   picture.CreateTime,
		EditTime:     picture.EditTime,
		UpdateTime:   picture.UpdateTime,
		IsDelete:     picture.IsDelete,
	}
}

// tagsToJSON 标签数组转 JSON 字符串
func tagsToJSON(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	bytes, err := json.Marshal(tags)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}

// tagsFromJSON JSON 字符串转标签数组
func tagsFromJSON(tagsJSON string) []string {
	if tagsJSON == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
		return []string{}
	}
	return tags
}
