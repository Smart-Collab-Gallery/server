package service

import (
	"context"
	"time"

	v1 "smart-collab-gallery-server/api/user/v1"
	"smart-collab-gallery-server/internal/biz"
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

// convertToLoginUserVO 将 User 转换为 LoginUserVO
func (s *UserService) convertToLoginUserVO(user *biz.User) *v1.LoginUserVO {
	vo := &v1.LoginUserVO{
		Id:          user.ID,
		UserAccount: user.UserAccount,
		UserName:    user.UserName,
		UserAvatar:  user.UserAvatar,
		UserProfile: user.UserProfile,
		UserRole:    user.UserRole,
		VipNumber:   user.VipNumber,
		CreateTime:  user.CreateTime.Format(time.RFC3339),
	}

	if user.VipExpireTime != nil {
		vo.VipExpireTime = user.VipExpireTime.Format(time.RFC3339)
	}

	return vo
}
