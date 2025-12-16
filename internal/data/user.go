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

// UpdateUser 更新用户
func (r *userRepo) UpdateUser(ctx context.Context, user *biz.User) error {
	updates := make(map[string]interface{})

	// 只更新非零值字段
	if user.UserAccount != "" {
		updates["userAccount"] = user.UserAccount
	}
	if user.UserName != "" {
		updates["userName"] = user.UserName
	}
	if user.UserAvatar != "" {
		updates["userAvatar"] = user.UserAvatar
	}
	if user.UserProfile != "" {
		updates["userProfile"] = user.UserProfile
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
