package data

import (
	"context"
	"fmt"

	v1 "smart-collab-gallery-server/api/user/v1"
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
		ID:                  userEntity.ID,
		UserAccount:         userEntity.UserAccount,
		UserPassword:        userEntity.UserPassword,
		UserName:            userEntity.UserName,
		UserAvatar:          userEntity.UserAvatar,
		UserBackgroundImage: userEntity.UserBackgroundImage,
		UserProfile:         userEntity.UserProfile,
		UserEmail:           userEntity.UserEmail,
		UserJob:             userEntity.UserJob,
		UserAddress:         userEntity.UserAddress,
		UserTags:            userEntity.UserTags,
		UserRole:            userEntity.UserRole,
		VipNumber:           userEntity.VipNumber,
		VipExpireTime:       userEntity.VipExpireTime,
		CreateTime:          userEntity.CreateTime,
		UpdateTime:          userEntity.UpdateTime,
	}
}

// UpdateUser 更新用户
func (r *userRepo) UpdateUser(ctx context.Context, user *biz.User) error {
	updates := make(map[string]interface{})

	// 只更新非零值字段
	if user.UserAccount != "" {
		updates["userAccount"] = user.UserAccount
	}
	if user.UserPassword != "" {
		updates["userPassword"] = user.UserPassword
	}
	if user.UserName != "" {
		updates["userName"] = user.UserName
	}
	if user.UserAvatar != "" {
		updates["userAvatar"] = user.UserAvatar
	}
	if user.UserBackgroundImage != "" {
		updates["userBackgroundImage"] = user.UserBackgroundImage
	}
	if user.UserProfile != "" {
		updates["userProfile"] = user.UserProfile
	}
	if user.UserEmail != "" {
		updates["userEmail"] = user.UserEmail
	}
	if user.UserJob != "" {
		updates["userJob"] = user.UserJob
	}
	if user.UserAddress != "" {
		updates["userAddress"] = user.UserAddress
	}
	if user.UserTags != "" {
		updates["userTags"] = user.UserTags
	}
	if user.UserRole != "" {
		updates["userRole"] = user.UserRole
	}

	err := r.data.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ? AND isDelete = 0", user.ID).
		Updates(updates).Error

	if err != nil {
		r.log.Errorf("更新用户失败: %v", err)
		return err
	}

	return nil
}

// DeleteUser 删除用户（逻辑删除）
func (r *userRepo) DeleteUser(ctx context.Context, id int64) error {
	err := r.data.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", id).
		Update("isDelete", 1).Error

	if err != nil {
		r.log.Errorf("删除用户失败: %v", err)
		return err
	}

	return nil
}

// ListUserByPage 分页查询用户
func (r *userRepo) ListUserByPage(ctx context.Context, params *biz.UserQueryParams) (*biz.UserPage, error) {
	var userEntities []User
	var total int64

	// 构建查询条件
	query := r.data.db.WithContext(ctx).Model(&User{}).Where("isDelete = 0")

	// 添加查询条件
	if params.ID != nil {
		query = query.Where("id = ?", *params.ID)
	}
	if params.UserAccount != "" {
		query = query.Where("userAccount LIKE ?", "%"+params.UserAccount+"%")
	}
	if params.UserName != "" {
		query = query.Where("userName LIKE ?", "%"+params.UserName+"%")
	}
	if params.UserProfile != "" {
		query = query.Where("userProfile LIKE ?", "%"+params.UserProfile+"%")
	}
	if params.UserRole != "" {
		query = query.Where("userRole = ?", params.UserRole)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("查询用户总数失败: %v", err)
		return nil, err
	}

	// 排序
	if params.SortField != "" {
		order := "ASC"
		if params.SortOrder == "descend" {
			order = "DESC"
		}
		query = query.Order(params.SortField + " " + order)
	} else {
		query = query.Order("createTime DESC")
	}

	// 分页查询
	offset := (params.Current - 1) * params.PageSize
	err := query.Offset(int(offset)).Limit(int(params.PageSize)).Find(&userEntities).Error

	if err != nil {
		r.log.Errorf("分页查询用户失败: %v", err)
		return nil, err
	}

	// 转换为业务对象
	users := make([]*biz.User, 0, len(userEntities))
	for i := range userEntities {
		users = append(users, r.convertToUser(&userEntities[i]))
	}

	return &biz.UserPage{
		Total:    total,
		List:     users,
		Current:  params.Current,
		PageSize: params.PageSize,
	}, nil
}

// SaveEmailVerificationCode 保存邮箱验证码到 Redis
func (r *userRepo) SaveEmailVerificationCode(ctx context.Context, userID int64, code, newEmail string) error {
	key := r.getEmailVerifyKey(userID)
	value := code + ":" + newEmail
	// 设置过期时间为5分钟
	err := r.data.rdb.Set(ctx, key, value, 5*60*1000000000).Err()
	if err != nil {
		r.log.Errorf("保存验证码到Redis失败: userID=%d, err=%v", userID, err)
		return err
	}
	return nil
}

// GetAndVerifyEmailCode 获取并验证邮箱验证码
func (r *userRepo) GetAndVerifyEmailCode(ctx context.Context, userID int64, code string) (string, error) {
	key := r.getEmailVerifyKey(userID)
	value, err := r.data.rdb.Get(ctx, key).Result()
	if err != nil {
		r.log.Errorf("从Redis获取验证码失败: userID=%d, err=%v", userID, err)
		return "", v1.ErrorVerificationCodeExpired("验证码已过期或不存在")
	}

	// 解析 value，格式为 {code}:{newEmail}
	parts := splitTwo(value, ":")
	if len(parts) != 2 {
		r.log.Errorf("验证码格式错误: userID=%d, value=%s", userID, value)
		return "", v1.ErrorSystemError("验证码格式错误")
	}

	storedCode := parts[0]
	newEmail := parts[1]

	// 验证验证码是否匹配
	if storedCode != code {
		r.log.Errorf("验证码不匹配: userID=%d, expected=%s, actual=%s", userID, storedCode, code)
		return "", v1.ErrorVerificationCodeError("验证码错误")
	}

	return newEmail, nil
}

// DeleteEmailVerificationCode 删除邮箱验证码
func (r *userRepo) DeleteEmailVerificationCode(ctx context.Context, userID int64) error {
	key := r.getEmailVerifyKey(userID)
	return r.data.rdb.Del(ctx, key).Err()
}

// SendEmailVerificationCode 发送邮箱验证码
func (r *userRepo) SendEmailVerificationCode(ctx context.Context, email, code string) error {
	// TODO: 实现实际的邮件发送逻辑
	// 这里暂时只打印日志，实际项目中需要集成邮件服务
	r.log.Infof("发送验证码到邮箱: email=%s, code=%s", email, code)
	r.log.Infof("【模拟邮件】您的验证码是: %s，有效期5分钟", code)

	// 实际使用时需要集成 SMTP 或第三方邮件服务
	// 例如：使用 gomail 库发送邮件
	// 参考实现：
	// m := gomail.NewMessage()
	// m.SetHeader("From", "noreply@example.com")
	// m.SetHeader("To", email)
	// m.SetHeader("Subject", "邮箱验证码")
	// m.SetBody("text/html", fmt.Sprintf("您的验证码是: <b>%s</b>，有效期5分钟", code))
	// d := gomail.NewDialer("smtp.example.com", 587, "username", "password")
	// return d.DialAndSend(m)

	return nil
}

// getEmailVerifyKey 获取邮箱验证码的 Redis key
func (r *userRepo) getEmailVerifyKey(userID int64) string {
	return fmt.Sprintf("email_verify:%d", userID)
}

// splitTwo 将字符串按分隔符分割为两部分
func splitTwo(s, sep string) []string {
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}
