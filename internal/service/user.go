package service

import (
	"context"
	"time"

	v1 "smart-collab-gallery-server/api/user/v1"
	"smart-collab-gallery-server/internal/biz"
	"smart-collab-gallery-server/internal/middleware"
	"smart-collab-gallery-server/internal/pkg"

	"github.com/go-kratos/kratos/v2/log"
)

type UserService struct {
	v1.UnimplementedUserServer

	uc         *biz.UserUsecase
	jwtManager *pkg.JWTManager
	log        *log.Helper
}

func NewUserService(uc *biz.UserUsecase, jwtManager *pkg.JWTManager, logger log.Logger) *UserService {
	return &UserService{
		uc:         uc,
		jwtManager: jwtManager,
		log:        log.NewHelper(logger),
	}
}

func (s *UserService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterReply, error) {
	s.log.WithContext(ctx).Infof("用户注册请求: account=%s", req.UserAccount)

	userId, err := s.uc.Register(ctx, req.UserAccount, req.UserPassword, req.CheckPassword)
	if err != nil {
		s.log.WithContext(ctx).Errorf("用户注册失败: %v", err)
		return nil, err
	}

	return &v1.RegisterReply{
		UserId: userId,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	s.log.WithContext(ctx).Infof("用户登录请求: account=%s", req.UserAccount)

	// 1. 执行登录逻辑
	user, err := s.uc.Login(ctx, req.UserAccount, req.UserPassword)
	if err != nil {
		s.log.WithContext(ctx).Errorf("用户登录失败: %v", err)
		return nil, err
	}

	// 2. 生成 JWT Token
	token, err := s.jwtManager.GenerateToken(user.ID, user.UserAccount, user.UserRole)
	if err != nil {
		s.log.WithContext(ctx).Errorf("生成 Token 失败: %v", err)
		return nil, v1.ErrorSystemError("登录失败，生成令牌错误")
	}

	// 3. 构建返回的用户信息
	loginUserVO := s.convertToLoginUserVO(user)

	return &v1.LoginReply{
		Token: token,
		User:  loginUserVO,
	}, nil
}

func (s *UserService) GetLoginUser(ctx context.Context, req *v1.GetLoginUserRequest) (*v1.GetLoginUserReply, error) {
	// 从上下文中获取用户 ID（由 JWT 中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == 0 {
		return nil, v1.ErrorNotLoginError("未登录")
	}

	s.log.WithContext(ctx).Infof("获取登录用户信息: userID=%d", userID)

	// 从数据库查询用户信息
	user, err := s.uc.GetLoginUser(ctx, userID)
	if err != nil {
		s.log.WithContext(ctx).Errorf("获取登录用户失败: %v", err)
		return nil, err
	}

	// 构建返回的用户信息
	loginUserVO := s.convertToLoginUserVO(user)

	return &v1.GetLoginUserReply{
		User: loginUserVO,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutReply, error) {
	// 从上下文中获取用户 ID（由 JWT 中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == 0 {
		return nil, v1.ErrorNotLoginError("未登录")
	}

	s.log.WithContext(ctx).Infof("用户注销请求: userID=%d", userID)

	// 执行注销逻辑
	err := s.uc.Logout(ctx, userID)
	if err != nil {
		s.log.WithContext(ctx).Errorf("用户注销失败: %v", err)
		return nil, err
	}

	return &v1.LogoutReply{
		Success: true,
	}, nil
}

// convertToLoginUserVO 将 User 转换为 LoginUserVO
func (s *UserService) convertToLoginUserVO(user *biz.User) *v1.LoginUserVO {
	vo := &v1.LoginUserVO{
		Id:                  user.ID,
		UserAccount:         user.UserAccount,
		UserName:            user.UserName,
		UserAvatar:          user.UserAvatar,
		UserBackgroundImage: user.UserBackgroundImage,
		UserProfile:         user.UserProfile,
		UserEmail:           user.UserEmail,
		UserJob:             user.UserJob,
		UserAddress:         user.UserAddress,
		UserTags:            user.UserTags,
		UserRole:            user.UserRole,
		VipNumber:           user.VipNumber,
		CreateTime:          user.CreateTime.Format(time.RFC3339),
		UpdateTime:          user.UpdateTime.Format(time.RFC3339),
	}

	if user.VipExpireTime != nil {
		vo.VipExpireTime = user.VipExpireTime.Format(time.RFC3339)
	}

	return vo
}

// convertToUserVO 将 User 转换为 UserVO
func (s *UserService) convertToUserVO(user *biz.User) *v1.UserVO {
	if user == nil {
		return nil
	}

	vo := &v1.UserVO{
		Id:                  user.ID,
		UserAccount:         user.UserAccount,
		UserName:            user.UserName,
		UserAvatar:          user.UserAvatar,
		UserBackgroundImage: user.UserBackgroundImage,
		UserProfile:         user.UserProfile,
		UserEmail:           user.UserEmail,
		UserJob:             user.UserJob,
		UserAddress:         user.UserAddress,
		UserTags:            user.UserTags,
		UserRole:            user.UserRole,
		VipNumber:           user.VipNumber,
		CreateTime:          user.CreateTime.Format(time.RFC3339),
		UpdateTime:          user.UpdateTime.Format(time.RFC3339),
	}

	if user.VipExpireTime != nil {
		vo.VipExpireTime = user.VipExpireTime.Format(time.RFC3339)
	}

	return vo
}

// convertToUserVOList 将 User 列表转换为 UserVO 列表
func (s *UserService) convertToUserVOList(users []*biz.User) []*v1.UserVO {
	if users == nil || len(users) == 0 {
		return []*v1.UserVO{}
	}

	voList := make([]*v1.UserVO, 0, len(users))
	for _, user := range users {
		voList = append(voList, s.convertToUserVO(user))
	}
	return voList
}

// AddUser 创建用户（仅管理员）
func (s *UserService) AddUser(ctx context.Context, req *v1.AddUserRequest) (*v1.AddUserReply, error) {
	s.log.WithContext(ctx).Infof("创建用户请求: account=%s", req.UserAccount)

	user := &biz.User{
		UserAccount:         req.UserAccount,
		UserName:            req.UserName,
		UserAvatar:          req.UserAvatar,
		UserBackgroundImage: req.UserBackgroundImage,
		UserProfile:         req.UserProfile,
		UserRole:            req.UserRole,
	}

	userID, err := s.uc.AddUser(ctx, user)
	if err != nil {
		s.log.WithContext(ctx).Errorf("创建用户失败: %v", err)
		return nil, err
	}

	return &v1.AddUserReply{
		UserId: userID,
	}, nil
}

// GetUserById 根据 ID 获取用户（仅管理员）
func (s *UserService) GetUserById(ctx context.Context, req *v1.GetUserByIdRequest) (*v1.GetUserByIdReply, error) {
	if req.Id <= 0 {
		return nil, v1.ErrorParamsError("用户 ID 无效")
	}

	s.log.WithContext(ctx).Infof("获取用户: id=%d", req.Id)

	user, err := s.uc.GetUserByID(ctx, req.Id)
	if err != nil {
		s.log.WithContext(ctx).Errorf("获取用户失败: %v", err)
		return nil, err
	}

	reply := &v1.GetUserByIdReply{
		Id:                  user.ID,
		UserAccount:         user.UserAccount,
		UserPassword:        user.UserPassword,
		UserName:            user.UserName,
		UserAvatar:          user.UserAvatar,
		UserBackgroundImage: user.UserBackgroundImage,
		UserProfile:         user.UserProfile,
		UserEmail:           user.UserEmail,
		UserJob:             user.UserJob,
		UserAddress:         user.UserAddress,
		UserTags:            user.UserTags,
		UserRole:            user.UserRole,
		VipNumber:           user.VipNumber,
		CreateTime:          user.CreateTime.Format(time.RFC3339),
		UpdateTime:          user.UpdateTime.Format(time.RFC3339),
	}

	if user.VipExpireTime != nil {
		reply.VipExpireTime = user.VipExpireTime.Format(time.RFC3339)
	}

	return reply, nil
}

// GetUserVOById 根据 ID 获取用户 VO
func (s *UserService) GetUserVOById(ctx context.Context, req *v1.GetUserVOByIdRequest) (*v1.GetUserVOByIdReply, error) {
	if req.Id <= 0 {
		return nil, v1.ErrorParamsError("用户 ID 无效")
	}

	s.log.WithContext(ctx).Infof("获取用户 VO: id=%d", req.Id)

	user, err := s.uc.GetUserByID(ctx, req.Id)
	if err != nil {
		s.log.WithContext(ctx).Errorf("获取用户失败: %v", err)
		return nil, err
	}

	return &v1.GetUserVOByIdReply{
		User: s.convertToUserVO(user),
	}, nil
}

// DeleteUser 删除用户（仅管理员）
func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.DeleteUserReply, error) {
	if req.Id <= 0 {
		return nil, v1.ErrorParamsError("用户 ID 无效")
	}

	s.log.WithContext(ctx).Infof("删除用户: id=%d", req.Id)

	err := s.uc.DeleteUser(ctx, req.Id)
	if err != nil {
		s.log.WithContext(ctx).Errorf("删除用户失败: %v", err)
		return nil, err
	}

	return &v1.DeleteUserReply{
		Success: true,
	}, nil
}

// UpdateUser 更新用户（仅管理员）
func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserReply, error) {
	if req.Id <= 0 {
		return nil, v1.ErrorParamsError("用户 ID 无效")
	}

	s.log.WithContext(ctx).Infof("更新用户: id=%d", req.Id)

	user := &biz.User{
		ID:                  req.Id,
		UserAccount:         req.UserAccount,
		UserName:            req.UserName,
		UserAvatar:          req.UserAvatar,
		UserBackgroundImage: req.UserBackgroundImage,
		UserProfile:         req.UserProfile,
		UserRole:            req.UserRole,
	}

	err := s.uc.UpdateUser(ctx, user)
	if err != nil {
		s.log.WithContext(ctx).Errorf("更新用户失败: %v", err)
		return nil, err
	}

	return &v1.UpdateUserReply{
		Success: true,
	}, nil
}

// ListUserByPage 分页获取用户列表（仅管理员）
func (s *UserService) ListUserByPage(ctx context.Context, req *v1.ListUserByPageRequest) (*v1.ListUserByPageReply, error) {
	s.log.WithContext(ctx).Infof("分页查询用户: current=%d, pageSize=%d", req.Current, req.PageSize)

	// 构建查询参数
	params := &biz.UserQueryParams{
		UserAccount: req.UserAccount,
		UserName:    req.UserName,
		UserProfile: req.UserProfile,
		UserRole:    req.UserRole,
		SortField:   req.SortField,
		SortOrder:   req.SortOrder,
		Current:     req.Current,
		PageSize:    req.PageSize,
	}

	if req.Id > 0 {
		params.ID = &req.Id
	}

	// 查询用户列表
	userPage, err := s.uc.ListUserByPage(ctx, params)
	if err != nil {
		s.log.WithContext(ctx).Errorf("分页查询用户失败: %v", err)
		return nil, err
	}

	// 转换为 VO
	voList := s.convertToUserVOList(userPage.List)

	return &v1.ListUserByPageReply{
		Total:    userPage.Total,
		List:     voList,
		Current:  userPage.Current,
		PageSize: userPage.PageSize,
	}, nil
}

// UpdateMyInfo 更新个人信息（用户自己）
func (s *UserService) UpdateMyInfo(ctx context.Context, req *v1.UpdateMyInfoRequest) (*v1.UpdateMyInfoReply, error) {
	// 从上下文中获取用户 ID（由 JWT 中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == 0 {
		return nil, v1.ErrorNotLoginError("未登录")
	}

	s.log.WithContext(ctx).Infof("更新个人信息: userID=%d", userID)

	// 执行更新
	err := s.uc.UpdateMyInfo(ctx, userID, req.UserPassword, req.UserName, req.UserAvatar, req.UserBackgroundImage, req.UserProfile, req.UserJob, req.UserAddress, req.UserTags)
	if err != nil {
		s.log.WithContext(ctx).Errorf("更新个人信息失败: %v", err)
		return nil, err
	}

	return &v1.UpdateMyInfoReply{
		Success: true,
	}, nil
}

// SendEmailVerificationCode 发送邮箱验证码
func (s *UserService) SendEmailVerificationCode(ctx context.Context, req *v1.SendEmailVerificationCodeRequest) (*v1.SendEmailVerificationCodeReply, error) {
	// 从上下文中获取用户 ID（由 JWT 中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == 0 {
		return nil, v1.ErrorNotLoginError("未登录")
	}

	s.log.WithContext(ctx).Infof("发送邮箱验证码: userID=%d", userID)

	// 执行发送验证码逻辑
	message, err := s.uc.SendEmailVerificationCode(ctx, userID)
	if err != nil {
		s.log.WithContext(ctx).Errorf("发送邮箱验证码失败: %v", err)
		return nil, err
	}

	return &v1.SendEmailVerificationCodeReply{
		Success: true,
		Message: message,
	}, nil
}

// VerifyAndUpdateEmail 验证码校验并更新邮箱
func (s *UserService) VerifyAndUpdateEmail(ctx context.Context, req *v1.VerifyAndUpdateEmailRequest) (*v1.VerifyAndUpdateEmailReply, error) {
	// 从上下文中获取用户 ID（由 JWT 中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == 0 {
		return nil, v1.ErrorNotLoginError("未登录")
	}

	s.log.WithContext(ctx).Infof("验证码校验并更新邮箱: userID=%d, newEmail=%s", userID, req.Email)

	// 执行验证码校验和邮箱更新逻辑
	message, err := s.uc.VerifyAndUpdateEmail(ctx, userID, req.Code, req.Email)
	if err != nil {
		s.log.WithContext(ctx).Errorf("验证码校验失败: %v", err)
		return nil, err
	}

	return &v1.VerifyAndUpdateEmailReply{
		Success: true,
		Message: message,
	}, nil
}

// UpdatePassword 修改用户登录密码
func (s *UserService) UpdatePassword(ctx context.Context, req *v1.UpdatePasswordRequest) (*v1.UpdatePasswordReply, error) {
	// 从上下文中获取用户 ID（由 JWT 中间件设置）
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == 0 {
		return nil, v1.ErrorNotLoginError("未登录")
	}

	s.log.WithContext(ctx).Infof("修改用户密码: userID=%d", userID)

	// 执行修改密码逻辑
	message, err := s.uc.UpdatePassword(ctx, userID, req.OldPassword, req.NewPassword, req.CheckPassword)
	if err != nil {
		s.log.WithContext(ctx).Errorf("修改密码失败: %v", err)
		return nil, err
	}

	return &v1.UpdatePasswordReply{
		Success: true,
		Message: message,
	}, nil
}
