package data

import (
	"context"

	"smart-collab-gallery-server/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo 创建用户仓储
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CreateUser 创建用户
func (r *userRepo) CreateUser(ctx context.Context, user *biz.User) (*biz.User, error) {
	userEntity := &User{
		UserAccount:  user.UserAccount,
		UserPassword: user.UserPassword,
		UserName:     user.UserName,
		UserRole:     user.UserRole,
	}

	if err := r.data.db.WithContext(ctx).Create(userEntity).Error; err != nil {
		r.log.Errorf("创建用户失败: %v", err)
		return nil, err
	}

	user.ID = userEntity.ID
	return user, nil
}

// GetUserByAccount 根据账号查询用户
func (r *userRepo) GetUserByAccount(ctx context.Context, account string) (*biz.User, error) {
	var userEntity User
	err := r.data.db.WithContext(ctx).
		Where("userAccount = ? AND isDelete = 0", account).
		First(&userEntity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("查询用户失败: %v", err)
		return nil, err
	}

	return &biz.User{
		ID:           userEntity.ID,
		UserAccount:  userEntity.UserAccount,
		UserPassword: userEntity.UserPassword,
		UserName:     userEntity.UserName,
		UserRole:     userEntity.UserRole,
	}, nil
}

// GetUserByAccountAndPassword 根据账号和密码查询用户
func (r *userRepo) GetUserByAccountAndPassword(ctx context.Context, account, password string) (*biz.User, error) {
	var userEntity User
	err := r.data.db.WithContext(ctx).
		Where("userAccount = ? AND userPassword = ? AND isDelete = 0", account, password).
		First(&userEntity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("根据账号密码查询用户失败: %v", err)
		return nil, err
	}

	return r.convertToUser(&userEntity), nil
}

// GetUserByID 根据 ID 查询用户
func (r *userRepo) GetUserByID(ctx context.Context, id int64) (*biz.User, error) {
	var userEntity User
	err := r.data.db.WithContext(ctx).
		Where("id = ? AND isDelete = 0", id).
		First(&userEntity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("根据 ID 查询用户失败: %v", err)
		return nil, err
	}

	return r.convertToUser(&userEntity), nil
}

// convertToUser 将数据库实体转换为业务对象
func (r *userRepo) convertToUser(userEntity *User) *biz.User {
	return &biz.User{
		ID:            userEntity.ID,
		UserAccount:   userEntity.UserAccount,
		UserPassword:  userEntity.UserPassword,
		UserName:      userEntity.UserName,
		UserAvatar:    userEntity.UserAvatar,
		UserProfile:   userEntity.UserProfile,
		UserRole:      userEntity.UserRole,
		VipNumber:     userEntity.VipNumber,
		VipExpireTime: userEntity.VipExpireTime,
		CreateTime:    userEntity.CreateTime,
		UpdateTime:    userEntity.UpdateTime,
	}
}
