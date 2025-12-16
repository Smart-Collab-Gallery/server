// 权限校验中间件示例 - 如何添加管理员权限接口
// 本文件仅作为示例，展示如何使用 RequireAdmin 中间件
// 注意：这是示例代码，不会被实际编译使用

//go:build example
// +build example

package middleware

import (
	"context"

	"smart-collab-gallery-server/internal/pkg"
	"smart-collab-gallery-server/internal/service"

	v1 "smart-collab-gallery-server/api/user/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// 示例 1: 在 HTTP Server 配置中使用管理员权限中间件
func NewHTTPServerWithAdminAuth(greeter *service.GreeterService,
	user *service.UserService, jwtManager *pkg.JWTManager, logger log.Logger) *http.Server {

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			MetricsServer(),

			// 第一层：JWT 认证 - 验证用户登录状态
			selector.Server(
				JWTAuth(jwtManager),
			).Match(NewAuthRequiredMatcher()).Build(),

			// 第二层：管理员权限校验 - 对特定接口要求管理员权限
			selector.Server(
				RequireAdmin(),
			).Match(NewAdminOnlyMatcher()).Build(),
		),
	}

	srv := http.NewServer(opts...)
	return srv
}

// 需要认证的接口匹配器（排除公开接口）
func NewAuthRequiredMatcher() selector.MatchFunc {
	publicAPIs := map[string]struct{}{
		"/api.user.v1.User/Register":          {},
		"/api.user.v1.User/Login":             {},
		"/api.helloworld.v1.Greeter/SayHello": {},
		"/api.health.v1.Health/Ping":          {},
	}

	return func(ctx context.Context, operation string) bool {
		_, isPublic := publicAPIs[operation]
		return !isPublic // 不是公开 API，需要认证
	}
}

// 仅管理员可访问的接口匹配器
func NewAdminOnlyMatcher() selector.MatchFunc {
	adminAPIs := map[string]struct{}{
		// 用户管理相关（管理员功能）
		"/api.user.v1.User/DeleteUser":     {},
		"/api.user.v1.User/UpdateUserRole": {},
		"/api.user.v1.User/BanUser":        {},
		"/api.user.v1.User/ListAllUsers":   {},

		// 系统管理相关（管理员功能）
		"/api.system.v1.System/GetSystemConfig":    {},
		"/api.system.v1.System/UpdateSystemConfig": {},

		// 内容审核相关（管理员功能）
		"/api.content.v1.Content/ReviewContent": {},
		"/api.content.v1.Content/DeleteContent": {},
	}

	return func(ctx context.Context, operation string) bool {
		_, requiresAdmin := adminAPIs[operation]
		return requiresAdmin
	}
}

// 示例 2: 在 Service 层手动校验权限
type ExampleService struct {
	log *log.Helper
}

func (s *ExampleService) DeleteUser(ctx context.Context, userID int64) error {
	// 方式 1: 手动检查管理员权限
	userRole := GetUserRoleFromContext(ctx)
	if UserRole(userRole) != RoleAdmin {
		return v1.ErrorNoAuthError("仅管理员可删除用户")
	}

	// 执行删除逻辑
	s.log.Infof("管理员正在删除用户: %d", userID)
	// ... 删除用户的具体实现

	return nil
}

func (s *ExampleService) UpdateContent(ctx context.Context, contentID int64, content string) error {
	// 获取当前用户信息
	userID := GetUserIDFromContext(ctx)
	userRole := GetUserRoleFromContext(ctx)

	// 方式 2: 复杂的权限判断逻辑
	// 管理员可以修改任何内容，普通用户只能修改自己的内容

	// 首先查询内容的所有者
	owner, err := s.getContentOwner(ctx, contentID)
	if err != nil {
		return err
	}

	// 权限校验
	isAdmin := UserRole(userRole) == RoleAdmin
	isOwner := owner == userID

	if !isAdmin && !isOwner {
		return v1.ErrorNoAuthError("只能修改自己的内容，或需要管理员权限")
	}

	// 执行更新逻辑
	s.log.Infof("用户 %d (角色: %s) 正在更新内容 %d", userID, userRole, contentID)
	// ... 更新内容的具体实现

	return nil
}

func (s *ExampleService) getContentOwner(ctx context.Context, contentID int64) (int64, error) {
	// 模拟查询内容所有者
	return 0, nil
}

// 示例 3: 使用 RequireLogin 仅要求登录
func NewHTTPServerWithLoginRequired(user *service.UserService,
	jwtManager *pkg.JWTManager, logger log.Logger) *http.Server {

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),

			// JWT 认证
			selector.Server(
				JWTAuth(jwtManager),
			).Match(NewAuthRequiredMatcher()).Build(),

			// 仅要求登录（不校验角色）
			selector.Server(
				RequireLogin(),
			).Match(NewLoginRequiredMatcher()).Build(),
		),
	}

	srv := http.NewServer(opts...)
	return srv
}

// 需要登录但不要求特定角色的接口
func NewLoginRequiredMatcher() selector.MatchFunc {
	loginRequiredAPIs := map[string]struct{}{
		"/api.user.v1.User/GetLoginUser":     {},
		"/api.user.v1.User/UpdateProfile":    {},
		"/api.user.v1.User/Logout":           {},
		"/api.content.v1.Content/CreatePost": {},
	}

	return func(ctx context.Context, operation string) bool {
		_, requiresLogin := loginRequiredAPIs[operation]
		return requiresLogin
	}
}

// 示例 4: 多层级权限控制
func NewHTTPServerWithMultiLevelAuth(greeter *service.GreeterService,
	user *service.UserService, jwtManager *pkg.JWTManager, logger log.Logger) *http.Server {

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),

			// 层级 1: JWT 认证（解析 Token）
			selector.Server(
				JWTAuth(jwtManager),
			).Match(func(ctx context.Context, operation string) bool {
				// 除了以下接口，其他都需要 JWT
				publicAPIs := []string{
					"/api.user.v1.User/Register",
					"/api.user.v1.User/Login",
					"/api.health.v1.Health/Ping",
				}
				for _, api := range publicAPIs {
					if operation == api {
						return false
					}
				}
				return true
			}).Build(),

			// 层级 2: 登录校验（要求用户 ID 存在）
			selector.Server(
				RequireLogin(),
			).Match(func(ctx context.Context, operation string) bool {
				// 需要登录的接口
				loginAPIs := []string{
					"/api.user.v1.User/GetLoginUser",
					"/api.user.v1.User/UpdateProfile",
					"/api.content.v1.Content/CreatePost",
				}
				for _, api := range loginAPIs {
					if operation == api {
						return true
					}
				}
				return false
			}).Build(),

			// 层级 3: 管理员权限校验
			selector.Server(
				RequireAdmin(),
			).Match(func(ctx context.Context, operation string) bool {
				// 需要管理员权限的接口
				adminAPIs := []string{
					"/api.user.v1.User/DeleteUser",
					"/api.user.v1.User/BanUser",
					"/api.system.v1.System/UpdateConfig",
				}
				for _, api := range adminAPIs {
					if operation == api {
						return true
					}
				}
				return false
			}).Build(),
		),
	}

	srv := http.NewServer(opts...)
	return srv
}

// 示例 5: 在 Biz 层使用权限校验辅助函数
type ExampleBiz struct {
	log *log.Helper
}

func (b *ExampleBiz) PerformAdminAction(ctx context.Context) error {
	// 获取用户信息
	userID := GetUserIDFromContext(ctx)
	userAccount := GetUserAccountFromContext(ctx)
	userRole := GetUserRoleFromContext(ctx)

	b.log.Infof("用户 [%d:%s] (角色: %s) 正在执行操作", userID, userAccount, userRole)

	// 检查是否为管理员
	if UserRole(userRole) != RoleAdmin {
		return v1.ErrorNoAuthError("需要管理员权限")
	}

	// 执行管理员操作
	// ...

	return nil
}

// 使用说明：
// 1. 在 internal/server/http.go 中配置中间件
// 2. 在 Service 或 Biz 层根据需要进行额外的权限校验
// 3. 使用 GetUserIDFromContext() 等函数获取用户信息
// 4. 使用 RoleAdmin 等常量进行角色比较
