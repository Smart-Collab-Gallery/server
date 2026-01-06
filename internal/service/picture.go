package service

import (
	"context"

	pb "smart-collab-gallery-server/api/picture/v1"
	"smart-collab-gallery-server/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PictureService struct {
	pb.UnimplementedPictureServer

	uc  *biz.PictureUsecase
	log *log.Helper
}

// NewPictureService 创建图片服务
func NewPictureService(uc *biz.PictureUsecase, logger log.Logger) *PictureService {
	return &PictureService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// UploadPicture 上传图片
func (s *PictureService) UploadPicture(ctx context.Context, req *pb.UploadPictureRequest) (*pb.UploadPictureReply, error) {
	// 从上下文获取用户信息
	loginUserID := s.getLoginUserID(ctx)
	if loginUserID == 0 {
		return nil, pb.ErrorUnauthorized("请先登录")
	}

	// 调用业务逻辑（直接传递 req）
	result, err := s.uc.UploadPicture(ctx, req, loginUserID)
	if err != nil {
		s.log.Errorf("上传图片失败: %v", err)
		return nil, err
	}

	// 转换返回结果
	return &pb.UploadPictureReply{
		Picture: s.convertToProtoPictureVO(result),
	}, nil
}

// GetPictureById 根据 ID 获取图片
func (s *PictureService) GetPictureById(ctx context.Context, req *pb.GetPictureByIdRequest) (*pb.GetPictureByIdReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorInvalidArgument("图片 ID 不能为空")
	}

	picture, err := s.uc.GetPictureByID(ctx, req.Id)
	if err != nil {
		s.log.Errorf("获取图片失败: %v", err)
		return nil, err
	}

	if picture == nil {
		return nil, pb.ErrorPictureNotFound("图片不存在")
	}

	return &pb.GetPictureByIdReply{
		Picture: s.convertToProtoPictureVO(picture),
	}, nil
}

// ListPictureByPage 分页查询图片
func (s *PictureService) ListPictureByPage(ctx context.Context, req *pb.ListPictureByPageRequest) (*pb.ListPictureByPageReply, error) {
	// 参数校验
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询参数
	params := &biz.PictureQueryParams{
		Current:      req.Current,
		PageSize:     req.PageSize,
		Name:         req.Name,
		Introduction: req.Introduction,
		Category:     req.Category,
		Tags:         req.Tags,
		SortField:    req.SortField,
		SortOrder:    req.SortOrder,
	}

	// 如果指定了用户ID，设置到查询参数
	if req.UserId > 0 {
		params.UserID = &req.UserId
	}

	// 调用业务逻辑
	page, err := s.uc.ListPictureByPage(ctx, params)
	if err != nil {
		s.log.Errorf("查询图片列表失败: %v", err)
		return nil, err
	}

	// 转换返回结果
	list := make([]*pb.PictureVO, 0, len(page.List))
	for _, pic := range page.List {
		list = append(list, s.convertToProtoPictureVO(pic))
	}

	return &pb.ListPictureByPageReply{
		Total: page.Total,
		List:  list,
	}, nil
}

// DeletePicture 删除图片
func (s *PictureService) DeletePicture(ctx context.Context, req *pb.DeletePictureRequest) (*pb.DeletePictureReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorInvalidArgument("图片 ID 不能为空")
	}

	// 从上下文获取用户信息
	loginUserID := s.getLoginUserID(ctx)
	if loginUserID == 0 {
		return nil, pb.ErrorUnauthorized("请先登录")
	}

	// 检查是否是管理员
	userRole := s.getUserRole(ctx)
	isAdmin := userRole == "admin"

	// 调用业务逻辑
	err := s.uc.DeletePicture(ctx, req.Id, loginUserID, isAdmin)
	if err != nil {
		s.log.Errorf("删除图片失败: %v", err)
		return nil, err
	}

	return &pb.DeletePictureReply{
		Success: true,
	}, nil
}

// UpdatePicture 更新图片
func (s *PictureService) UpdatePicture(ctx context.Context, req *pb.UpdatePictureRequest) (*pb.UpdatePictureReply, error) {
	if req.Id <= 0 {
		return nil, pb.ErrorInvalidArgument("图片 ID 不能为空")
	}

	// 从上下文获取用户信息
	loginUserID := s.getLoginUserID(ctx)
	if loginUserID == 0 {
		return nil, pb.ErrorUnauthorized("请先登录")
	}

	// 调用业务逻辑
	err := s.uc.UpdatePicture(ctx, req.Id, req.Name, req.Introduction, req.Category, req.Tags, loginUserID)
	if err != nil {
		s.log.Errorf("更新图片失败: %v", err)
		return nil, err
	}

	return &pb.UpdatePictureReply{
		Success: true,
	}, nil
}

// getLoginUserID 从上下文获取登录用户 ID
func (s *PictureService) getLoginUserID(ctx context.Context) int64 {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return 0
	}

	mapClaims, ok := claims.(jwtv5.MapClaims)
	if !ok {
		return 0
	}

	userID, ok := mapClaims["user_id"].(float64)
	if !ok {
		return 0
	}

	return int64(userID)
}

// getUserRole 从上下文获取用户角色
func (s *PictureService) getUserRole(ctx context.Context) string {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return ""
	}

	mapClaims, ok := claims.(jwtv5.MapClaims)
	if !ok {
		return ""
	}

	userRole, ok := mapClaims["user_role"].(string)
	if !ok {
		return ""
	}

	return userRole
}

// convertToProtoPictureVO 转换业务对象为 proto 对象
func (s *PictureService) convertToProtoPictureVO(vo *biz.PictureVO) *pb.PictureVO {
	if vo == nil {
		return nil
	}

	return &pb.PictureVO{
		Id:           vo.ID,
		Url:          vo.URL,
		Name:         vo.Name,
		Introduction: vo.Introduction,
		Category:     vo.Category,
		Tags:         vo.Tags,
		PicSize:      vo.PicSize,
		PicWidth:     vo.PicWidth,
		PicHeight:    vo.PicHeight,
		PicScale:     vo.PicScale,
		PicFormat:    vo.PicFormat,
		UserId:       vo.UserID,
		CreateTime:   timestamppb.New(vo.CreateTime),
		EditTime:     timestamppb.New(vo.EditTime),
		UpdateTime:   timestamppb.New(vo.UpdateTime),
		User:         s.convertToProtoUserVO(vo.User),
	}
}

// convertToProtoUserVO 转换用户对象为 proto 对象
func (s *PictureService) convertToProtoUserVO(userVO *biz.UserVO) *pb.UserVO {
	if userVO == nil {
		return nil
	}

	return &pb.UserVO{
		Id:          userVO.ID,
		UserAccount: userVO.UserAccount,
		UserName:    userVO.UserName,
		UserAvatar:  userVO.UserAvatar,
		UserProfile: userVO.UserProfile,
		UserRole:    userVO.UserRole,
	}
}
