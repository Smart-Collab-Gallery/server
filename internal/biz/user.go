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
	ID                  int64
	UserAccount         string
	UserPassword        string
	UserName            string
	UserAvatar          string
	UserBackgroundImage string
	UserProfile         string
	UserEmail           string
	UserJob             string
	UserAddress         string
	UserTags            string
	UserRole            string
	VipNumber           int64
	VipExpireTime       *time.Time
	CreateTime          time.Time
	UpdateTime          time.Time
}

// UserQueryParams 用户查询参数
type UserQueryParams struct {
	ID          *int64
	UserAccount string
	UserName    string
	UserProfile string
	UserRole    string
	SortField   string
	SortOrder   string // ascend 或 descend
	Current     int64
	PageSize    int64
}

// UserPage 用户分页结果
type UserPage struct {
	Total    int64
	List     []*User
	Current  int64
	PageSize int64
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
	// UpdateUser 更新用户
	UpdateUser(ctx context.Context, user *User) error
	// DeleteUser 删除用户
	DeleteUser(ctx context.Context, id int64) error
	// ListUserByPage 分页查询用户
	ListUserByPage(ctx context.Context, params *UserQueryParams) (*UserPage, error)
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

// Logout 用户注销（逻辑删除用户）
func (uc *UserUsecase) Logout(ctx context.Context, userID int64) error {
	// 验证用户是否已登录
	if userID <= 0 {
		return v1.ErrorNotLoginError("未登录")
	}

	// 检查用户是否存在
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		uc.log.Errorf("查询用户失败: userID=%d, err=%v", userID, err)
		return v1.ErrorSystemError("查询用户失败")
	}
	if user == nil {
		return v1.ErrorUserNotFound("用户不存在")
	}

	// 记录注销日志
	uc.log.Infof("用户注销（逻辑删除）: userID=%d, account=%s", userID, user.UserAccount)

	// 执行逻辑删除（设置 isDelete = 1）
	err = uc.repo.DeleteUser(ctx, userID)
	if err != nil {
		uc.log.Errorf("用户注销失败: userID=%d, err=%v", userID, err)
		return v1.ErrorSystemError("用户注销失败")
	}

	return nil
}

// AddUser 创建用户（管理员功能）
func (uc *UserUsecase) AddUser(ctx context.Context, user *User) (int64, error) {
	// 1. 参数校验
	if strings.TrimSpace(user.UserAccount) == "" {
		return 0, v1.ErrorParamsError("用户账号不能为空")
	}

	// 校验用户简介长度
	if len([]rune(user.UserProfile)) > 50 {
		return 0, v1.ErrorParamsError("用户简介不能超过50个字")
	}

	// 校验标签长度
	if len([]rune(user.UserTags)) > 100 {
		return 0, v1.ErrorParamsError("用户标签不能超过100个字")
	}

	// 2. 检查账号是否已存在
	existUser, err := uc.repo.GetUserByAccount(ctx, user.UserAccount)
	if err == nil && existUser != nil {
		return 0, v1.ErrorAccountDuplicate("账号已存在")
	}

	// 3. 设置默认密码 12345678
	const DEFAULT_PASSWORD = "12345678"
	user.UserPassword = uc.encryptPassword(DEFAULT_PASSWORD)

	// 4. 设置默认值
	if user.UserName == "" {
		user.UserName = "无名"
	}
	if user.UserRole == "" {
		user.UserRole = "user"
	}

	// 5. 创建用户
	newUser, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		uc.log.Errorf("创建用户失败: %v", err)
		return 0, v1.ErrorSystemError("创建用户失败")
	}

	return newUser.ID, nil
}

// GetUserByID 根据 ID 获取用户（管理员功能）
func (uc *UserUsecase) GetUserByID(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, v1.ErrorParamsError("用户 ID 无效")
	}

	user, err := uc.repo.GetUserByID(ctx, id)
	if err != nil {
		uc.log.Errorf("查询用户失败: id=%d, err=%v", id, err)
		return nil, v1.ErrorSystemError("查询用户失败")
	}

	if user == nil {
		return nil, v1.ErrorUserNotFound("用户不存在")
	}

	return user, nil
}

// DeleteUser 删除用户（管理员功能）
func (uc *UserUsecase) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return v1.ErrorParamsError("用户 ID 无效")
	}

	// 检查用户是否存在
	user, err := uc.repo.GetUserByID(ctx, id)
	if err != nil {
		return v1.ErrorSystemError("查询用户失败")
	}
	if user == nil {
		return v1.ErrorUserNotFound("用户不存在")
	}

	// 删除用户
	err = uc.repo.DeleteUser(ctx, id)
	if err != nil {
		uc.log.Errorf("删除用户失败: id=%d, err=%v", id, err)
		return v1.ErrorSystemError("删除用户失败")
	}

	return nil
}

// UpdateUser 更新用户（管理员功能）
func (uc *UserUsecase) UpdateUser(ctx context.Context, user *User) error {
	if user.ID <= 0 {
		return v1.ErrorParamsError("用户 ID 无效")
	}

	// 校验用户简介长度
	if user.UserProfile != "" && len([]rune(user.UserProfile)) > 50 {
		return v1.ErrorParamsError("用户简介不能超过50个字")
	}

	// 校验标签长度
	if user.UserTags != "" && len([]rune(user.UserTags)) > 100 {
		return v1.ErrorParamsError("用户标签不能超过100个字")
	}

	// 检查用户是否存在
	existUser, err := uc.repo.GetUserByID(ctx, user.ID)
	if err != nil {
		return v1.ErrorSystemError("查询用户失败")
	}
	if existUser == nil {
		return v1.ErrorUserNotFound("用户不存在")
	}

	// 如果更新账号，检查新账号是否已被使用
	if user.UserAccount != "" && user.UserAccount != existUser.UserAccount {
		duplicateUser, _ := uc.repo.GetUserByAccount(ctx, user.UserAccount)
		if duplicateUser != nil && duplicateUser.ID != user.ID {
			return v1.ErrorAccountDuplicate("账号已存在")
		}
	}

	// 更新用户
	err = uc.repo.UpdateUser(ctx, user)
	if err != nil {
		uc.log.Errorf("更新用户失败: id=%d, err=%v", user.ID, err)
		return v1.ErrorSystemError("更新用户失败")
	}

	return nil
}

// ListUserByPage 分页查询用户（管理员功能）
func (uc *UserUsecase) ListUserByPage(ctx context.Context, params *UserQueryParams) (*UserPage, error) {
	// 参数校验
	if params.Current <= 0 {
		params.Current = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 10
	}

	// 查询用户列表
	userPage, err := uc.repo.ListUserByPage(ctx, params)
	if err != nil {
		uc.log.Errorf("分页查询用户失败: %v", err)
		return nil, v1.ErrorSystemError("查询用户列表失败")
	}

	return userPage, nil
}

// UpdateMyInfo 更新个人信息（用户自己）
func (uc *UserUsecase) UpdateMyInfo(ctx context.Context, userID int64, password, name, avatar, backgroundImage, profile, email, job, address, tags string) error {
	if userID <= 0 {
		return v1.ErrorNotLoginError("未登录")
	}

	// 检查用户是否存在
	existUser, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return v1.ErrorSystemError("查询用户失败")
	}
	if existUser == nil {
		return v1.ErrorUserNotFound("用户不存在")
	}

	// 构建更新对象，只更新提供的字段
	user := &User{
		ID: userID,
	}

	// 如果提供了密码，需要验证并加密
	if strings.TrimSpace(password) != "" {
		// 密码长度检查
		if len(password) < 8 {
			return v1.ErrorPasswordTooShort("密码过短，至少8个字符")
		}
		user.UserPassword = uc.encryptPassword(password)
	}

	// 更新昵称
	if strings.TrimSpace(name) != "" {
		user.UserName = name
	}

	// 更新头像
	if strings.TrimSpace(avatar) != "" {
		user.UserAvatar = avatar
	}

	// 更新背景图片
	if strings.TrimSpace(backgroundImage) != "" {
		user.UserBackgroundImage = backgroundImage
	}

	// 更新简介
	if strings.TrimSpace(profile) != "" {
		// 校验简介长度
		if len([]rune(profile)) > 50 {
			return v1.ErrorParamsError("用户简介不能超过50个字")
		}
		user.UserProfile = profile
	}

	// 更新邮箱
	if strings.TrimSpace(email) != "" {
		user.UserEmail = email
	}

	// 更新职业
	if strings.TrimSpace(job) != "" {
		user.UserJob = job
	}

	// 更新地址
	if strings.TrimSpace(address) != "" {
		user.UserAddress = address
	}

	// 更新标签
	if strings.TrimSpace(tags) != "" {
		// 校验标签长度
		if len([]rune(tags)) > 100 {
			return v1.ErrorParamsError("用户标签不能超过100个字")
		}
		user.UserTags = tags
	}

	// 执行更新
	err = uc.repo.UpdateUser(ctx, user)
	if err != nil {
		uc.log.Errorf("更新个人信息失败: userID=%d, err=%v", userID, err)
		return v1.ErrorSystemError("更新个人信息失败")
	}

	return nil
}
