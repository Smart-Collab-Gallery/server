package biz

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"

	v1 "smart-collab-gallery-server/api/user/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// User 用户业务对象
type User struct {
	ID           int64
	UserAccount  string
	UserPassword string
	UserName     string
	UserRole     string
}

// UserRepo 用户仓储接口
type UserRepo interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *User) (*User, error)
	// GetUserByAccount 根据账号查询用户
	GetUserByAccount(ctx context.Context, account string) (*User, error)
}

// UserUsecase 用户用例
type UserUsecase struct {
	repo UserRepo
	log  *log.Helper
}

// NewUserUsecase 创建用户用例
func NewUserUsecase(repo UserRepo, logger log.Logger) *UserUsecase {
	return &UserUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

const (
	// SALT 密码盐值
	SALT = "yupi"
)

// Register 用户注册
func (uc *UserUsecase) Register(ctx context.Context, userAccount, userPassword, checkPassword string) (int64, error) {
	// 1. 校验参数
	if err := uc.validateRegisterParams(userAccount, userPassword, checkPassword); err != nil {
		return 0, err
	}

	// 2. 检查账号是否已存在
	existUser, err := uc.repo.GetUserByAccount(ctx, userAccount)
	if err == nil && existUser != nil {
		return 0, v1.ErrorAccountDuplicate("账号已存在")
	}

	// 3. 加密密码
	encryptPassword := uc.encryptPassword(userPassword)

	// 4. 创建用户
	user := &User{
		UserAccount:  userAccount,
		UserPassword: encryptPassword,
		UserName:     "无名",
		UserRole:     "user",
	}

	newUser, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		uc.log.Errorf("创建用户失败: %v", err)
		return 0, v1.ErrorSystemError("注册失败，数据库错误")
	}

	return newUser.ID, nil
}

// validateRegisterParams 校验注册参数
func (uc *UserUsecase) validateRegisterParams(userAccount, userPassword, checkPassword string) error {
	// 参数为空检查
	if strings.TrimSpace(userAccount) == "" || strings.TrimSpace(userPassword) == "" || strings.TrimSpace(checkPassword) == "" {
		return v1.ErrorParamsError("参数为空")
	}

	// 账号长度检查
	if len(userAccount) < 4 {
		return v1.ErrorAccountTooShort("用户账号过短，至少4个字符")
	}

	// 密码长度检查
	if len(userPassword) < 8 || len(checkPassword) < 8 {
		return v1.ErrorPasswordTooShort("用户密码过短，至少8个字符")
	}

	// 两次密码一致性检查
	if userPassword != checkPassword {
		return v1.ErrorPasswordNotMatch("两次输入的密码不一致")
	}

	return nil
}

// encryptPassword 加密密码
func (uc *UserUsecase) encryptPassword(password string) string {
	hash := md5.New()
	hash.Write([]byte(SALT + password))
	return hex.EncodeToString(hash.Sum(nil))
}
