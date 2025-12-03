package service

import (
	"context"

	v1 "smart-collab-gallery-server/api/user/v1"
	"smart-collab-gallery-server/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type UserService struct {
	v1.UnimplementedUserServer

	uc  *biz.UserUsecase
	log *log.Helper
}

func NewUserService(uc *biz.UserUsecase, logger log.Logger) *UserService {
	return &UserService{
		uc:  uc,
		log: log.NewHelper(logger),
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
