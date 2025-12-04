package biz

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"
	"time"

	v1 "smart-collab-gallery-server/api/user/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// User 用户业务对象
type User struct {
	ID            int64
	UserAccount   string
	UserPassword  string
	UserName      string
	UserAvatar    string
	UserProfile   string
	UserRole      string
	VipNumber     int64
	VipExpireTime *time.Time
	CreateTime    time.Time
	UpdateTime    time.Time
}

// UserRepo 用户仓储接口
type UserRepo interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *User) (*User, error)
	// GetUserByAccount 根据账号查询用户
	GetUserByAccount(ctx context.Context, account string) (*User, error)
	// GetUserByAccountAndPassword 根据账号和密码查询用户
	GetUserByAccountAndPassword(ctx context.Context, account, password string) (*User, error)
	// GetUserByID 根据 ID 查询用户
	GetUserByID(ctx context.Context, id int64) (*User, error)
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

// Login 用户登录
func (uc *UserUsecase) Login(ctx context.Context, userAccount, userPassword string) (*User, error) {
	// 1. 校验参数
	if err := uc.validateLoginParams(userAccount, userPassword); err != nil {
		return nil, err
	}

	// 2. 加密密码
	encryptPassword := uc.encryptPassword(userPassword)

	// 3. 查询用户是否存在
	user, err := uc.repo.GetUserByAccountAndPassword(ctx, userAccount, encryptPassword)
	if err != nil {
		uc.log.Errorf("用户登录失败，账号或密码错误: account=%s, err=%v", userAccount, err)
		return nil, v1.ErrorUserNotExistOrPasswordError("用户不存在或密码错误")
	}

	if user == nil {
		uc.log.Infof("user login failed, userAccount cannot match userPassword: %s", userAccount)
		return nil, v1.ErrorUserNotExistOrPasswordError("用户不存在或密码错误")
	}

	return user, nil
}

// validateLoginParams 校验登录参数
func (uc *UserUsecase) validateLoginParams(userAccount, userPassword string) error {
	// 参数为空检查
	if strings.TrimSpace(userAccount) == "" || strings.TrimSpace(userPassword) == "" {
		return v1.ErrorParamsError("参数为空")
	}

	// 账号长度检查
	if len(userAccount) < 4 {
		return v1.ErrorAccountError("账号错误")
	}

	// 密码长度检查
	if len(userPassword) < 8 {
		return v1.ErrorPasswordError("密码错误")
	}

	return nil
}

// GetLoginUser 获取当前登录用户
func (uc *UserUsecase) GetLoginUser(ctx context.Context, userID int64) (*User, error) {
	if userID <= 0 {
		return nil, v1.ErrorNotLoginError("未登录")
	}

	// 从数据库查询用户
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		uc.log.Errorf("查询用户失败: userID=%d, err=%v", userID, err)
		return nil, v1.ErrorSystemError("查询用户失败")
	}

	if user == nil {
		return nil, v1.ErrorUserNotFound("用户不存在")
	}

	return user, nil
}

// Logout 用户注销
func (uc *UserUsecase) Logout(ctx context.Context, userID int64) error {
	// 验证用户是否已登录
	if userID <= 0 {
		return v1.ErrorNotLoginError("未登录")
	}

	// 记录注销日志
	uc.log.Infof("用户注销: userID=%d", userID)

	// JWT 是无状态的，实际的注销由客户端删除 Token 完成
	// 这里可以做一些额外的清理工作，比如：
	// 1. 记录注销日志到数据库
	// 2. 清理用户相关的缓存
	// 3. 发送注销通知等

	return nil
}
