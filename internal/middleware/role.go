package middleware

import (
	"context"

	v1 "smart-collab-gallery-server/api/user/v1"

	"github.com/go-kratos/kratos/v2/middleware"
)

// UserRole 用户角色类型
type UserRole string

const (
	// RoleUser 普通用户角色
	RoleUser UserRole = "user"
	// RoleAdmin 管理员角色
	RoleAdmin UserRole = "admin"
)

// RequireRole 权限校验中间件 - 要求指定角色才能访问
// 类似于 Java Spring 的 @AuthCheck 注解功能
// mustRole: 必须的角色，传入 RoleAdmin 或 RoleUser，空字符串表示不需要权限校验
func RequireRole(mustRole UserRole) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 不需要权限，放行
			if mustRole == "" {
				return handler(ctx, req)
			}

			// 从上下文中获取当前登录用户的角色
			currentRole := GetUserRoleFromContext(ctx)

			// 没有获取到用户角色，说明用户未登录或 JWT 中间件未执行
			if currentRole == "" {
				return nil, v1.ErrorNotLoginError("未登录")
			}

			// 获取当前用户的角色枚举
			currentRoleEnum := UserRole(currentRole)

			// 要求必须有管理员权限，但用户没有管理员权限，拒绝
			if mustRole == RoleAdmin && currentRoleEnum != RoleAdmin {
				return nil, v1.ErrorNoAuthError("无权限")
			}

			// 通过权限校验，放行
			return handler(ctx, req)
		}
	}
}

// RequireAdmin 要求管理员权限的便捷中间件
// 等价于 RequireRole(RoleAdmin)
func RequireAdmin() middleware.Middleware {
	return RequireRole(RoleAdmin)
}

// RequireLogin 仅要求登录，不校验角色
// 等价于 RequireRole("")，但语义更清晰
func RequireLogin() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 从上下文中获取用户 ID，检查是否已登录
			userID := GetUserIDFromContext(ctx)
			if userID == 0 {
				return nil, v1.ErrorNotLoginError("未登录")
			}

			// 已登录，放行
			return handler(ctx, req)
		}
	}
}
