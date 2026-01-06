package biz

import (
	"context"
	"encoding/json"
	"math"
	"time"

	v1 "smart-collab-gallery-server/api/picture/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// PictureRepo 图片仓储接口
type PictureRepo interface {
	// CreatePicture 创建图片
	CreatePicture(ctx context.Context, picture *Picture) (*Picture, error)
	// GetPictureByID 根据 ID 查询图片
	GetPictureByID(ctx context.Context, id int64) (*Picture, error)
	// UpdatePicture 更新图片
	UpdatePicture(ctx context.Context, picture *Picture) error
	// DeletePicture 删除图片（逻辑删除）
	DeletePicture(ctx context.Context, id int64) error
	// ListPictureByPage 分页查询图片
	ListPictureByPage(ctx context.Context, params *PictureQueryParams) (*PicturePage, error)
}

// PictureUsecase 图片用例
type PictureUsecase struct {
	pictureRepo PictureRepo
	userRepo    UserRepo // 用于获取用户信息
	log         *log.Helper
}

// NewPictureUsecase 创建图片用例
func NewPictureUsecase(pictureRepo PictureRepo, userRepo UserRepo, logger log.Logger) *PictureUsecase {
	return &PictureUsecase{
		pictureRepo: pictureRepo,
		userRepo:    userRepo,
		log:         log.NewHelper(logger),
	}
}

// UploadPicture 上传图片
func (uc *PictureUsecase) UploadPicture(ctx context.Context, req *v1.UploadPictureRequest, userID int64) (*PictureVO, error) {
	uc.log.WithContext(ctx).Infof("上传图片: userID=%d, name=%s", userID, req.Name)

	// 如果 ID 不为空，表示更新
	if req.Id > 0 {
		// 检查图片是否存在
		existPicture, err := uc.pictureRepo.GetPictureByID(ctx, req.Id)
		if err != nil {
			return nil, v1.ErrorPictureNotFound("图片不存在")
		}

		// 检查权限：只能更新自己的图片
		if existPicture.UserID != userID {
			return nil, v1.ErrorPictureNoAuth("无权限操作该图片")
		}
	}

	// 构造图片对象
	picture := &Picture{
		ID:           req.Id,
		URL:          req.Url,
		Name:         req.Name,
		Introduction: req.Introduction,
		Category:     req.Category,
		PicSize:      req.PicSize,
		PicWidth:     req.PicWidth,
		PicHeight:    req.PicHeight,
		PicFormat:    req.PicFormat,
		UserID:       userID,
	}

	// 计算图片宽高比
	if req.PicHeight > 0 {
		picture.PicScale = math.Round(float64(req.PicWidth)/float64(req.PicHeight)*100) / 100
	}

	// 转换标签为 JSON
	if len(req.Tags) > 0 {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, v1.ErrorParamsError("标签格式错误")
		}
		picture.Tags = string(tagsBytes)
	}

	var result *Picture
	var err error

	if req.Id > 0 {
		// 更新图片
		picture.EditTime = time.Now()
		err = uc.pictureRepo.UpdatePicture(ctx, picture)
		if err != nil {
			return nil, v1.ErrorPictureUpdateFailed("图片更新失败")
		}
		result = picture
	} else {
		// 创建新图片
		result, err = uc.pictureRepo.CreatePicture(ctx, picture)
		if err != nil {
			return nil, v1.ErrorPictureUploadFailed("图片上传失败")
		}
	}

	// 转换为 VO
	pictureVO := result.ObjToVO()

	// 填充用户信息
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err == nil && user != nil {
		pictureVO.User = &UserVO{
			ID:          user.ID,
			UserAccount: user.UserAccount,
			UserName:    user.UserName,
			UserAvatar:  user.UserAvatar,
			UserProfile: user.UserProfile,
			UserRole:    user.UserRole,
		}
	}

	return pictureVO, nil
}

// GetPictureByID 根据 ID 获取图片
func (uc *PictureUsecase) GetPictureByID(ctx context.Context, id int64) (*PictureVO, error) {
	uc.log.WithContext(ctx).Infof("获取图片: id=%d", id)

	picture, err := uc.pictureRepo.GetPictureByID(ctx, id)
	if err != nil {
		return nil, v1.ErrorPictureNotFound("图片不存在")
	}

	if picture == nil {
		return nil, v1.ErrorPictureNotFound("图片不存在")
	}

	// 转换为 VO
	pictureVO := picture.ObjToVO()

	// 填充用户信息
	user, err := uc.userRepo.GetUserByID(ctx, picture.UserID)
	if err == nil && user != nil {
		pictureVO.User = &UserVO{
			ID:          user.ID,
			UserAccount: user.UserAccount,
			UserName:    user.UserName,
			UserAvatar:  user.UserAvatar,
			UserProfile: user.UserProfile,
			UserRole:    user.UserRole,
		}
	}

	return pictureVO, nil
}

// ListPictureByPage 分页查询图片
func (uc *PictureUsecase) ListPictureByPage(ctx context.Context, params *PictureQueryParams) (*PicturePage, error) {
	uc.log.WithContext(ctx).Infof("分页查询图片: current=%d, pageSize=%d", params.Current, params.PageSize)

	page, err := uc.pictureRepo.ListPictureByPage(ctx, params)
	if err != nil {
		return nil, err
	}

	// 填充用户信息
	userIDs := make(map[int64]bool)
	for _, pic := range page.List {
		userIDs[pic.UserID] = true
	}

	// 批量查询用户信息（简化实现，实际可以优化为批量查询）
	userMap := make(map[int64]*UserVO)
	for userID := range userIDs {
		user, err := uc.userRepo.GetUserByID(ctx, userID)
		if err == nil && user != nil {
			userMap[userID] = &UserVO{
				ID:          user.ID,
				UserAccount: user.UserAccount,
				UserName:    user.UserName,
				UserAvatar:  user.UserAvatar,
				UserProfile: user.UserProfile,
				UserRole:    user.UserRole,
			}
		}
	}

	// 填充用户信息到图片 VO
	for _, pic := range page.List {
		if user, ok := userMap[pic.UserID]; ok {
			pic.User = user
		}
	}

	return page, nil
}

// DeletePicture 删除图片
func (uc *PictureUsecase) DeletePicture(ctx context.Context, id int64, userID int64, isAdmin bool) error {
	uc.log.WithContext(ctx).Infof("删除图片: id=%d, userID=%d", id, userID)

	// 检查图片是否存在
	picture, err := uc.pictureRepo.GetPictureByID(ctx, id)
	if err != nil {
		return v1.ErrorPictureNotFound("图片不存在")
	}

	if picture == nil {
		return v1.ErrorPictureNotFound("图片不存在")
	}

	// 检查权限：只能删除自己的图片或者管理员可以删除任何图片
	if picture.UserID != userID && !isAdmin {
		return v1.ErrorPictureNoAuth("无权限操作该图片")
	}

	// 逻辑删除
	err = uc.pictureRepo.DeletePicture(ctx, id)
	if err != nil {
		return v1.ErrorPictureDeleteFailed("图片删除失败")
	}

	return nil
}

// UpdatePicture 更新图片信息
func (uc *PictureUsecase) UpdatePicture(ctx context.Context, id int64, name, introduction, category string, tags []string, userID int64) error {
	uc.log.WithContext(ctx).Infof("更新图片: id=%d, userID=%d", id, userID)

	// 检查图片是否存在
	picture, err := uc.pictureRepo.GetPictureByID(ctx, id)
	if err != nil {
		return v1.ErrorPictureNotFound("图片不存在")
	}

	if picture == nil {
		return v1.ErrorPictureNotFound("图片不存在")
	}

	// 检查权限：只能更新自己的图片
	if picture.UserID != userID {
		return v1.ErrorPictureNoAuth("无权限操作该图片")
	}

	// 更新字段
	picture.Name = name
	picture.Introduction = introduction
	picture.Category = category
	picture.EditTime = time.Now()

	// 转换标签为 JSON
	if len(tags) > 0 {
		tagsBytes, err := json.Marshal(tags)
		if err != nil {
			return v1.ErrorParamsError("标签格式错误")
		}
		picture.Tags = string(tagsBytes)
	}

	err = uc.pictureRepo.UpdatePicture(ctx, picture)
	if err != nil {
		return v1.ErrorPictureUpdateFailed("图片更新失败")
	}

	return nil
}

// EditPicture 编辑图片（用户使用）
func (uc *PictureUsecase) EditPicture(ctx context.Context, id int64, name, introduction, category string, tags []string, userID int64) error {
	uc.log.WithContext(ctx).Infof("编辑图片: id=%d, userID=%d", id, userID)

	// 检查图片是否存在
	picture, err := uc.pictureRepo.GetPictureByID(ctx, id)
	if err != nil {
		return v1.ErrorPictureNotFound("图片不存在")
	}

	if picture == nil {
		return v1.ErrorPictureNotFound("图片不存在")
	}

	// 检查权限：只能编辑自己的图片
	if picture.UserID != userID {
		return v1.ErrorPictureNoAuth("无权限操作该图片")
	}

	// 更新字段
	picture.Name = name
	picture.Introduction = introduction
	picture.Category = category
	picture.EditTime = time.Now()

	// 转换标签为 JSON
	if len(tags) > 0 {
		tagsBytes, err := json.Marshal(tags)
		if err != nil {
			return v1.ErrorParamsError("标签格式错误")
		}
		picture.Tags = string(tagsBytes)
	}

	err = uc.pictureRepo.UpdatePicture(ctx, picture)
	if err != nil {
		return v1.ErrorPictureUpdateFailed("图片编辑失败")
	}

	return nil
}
