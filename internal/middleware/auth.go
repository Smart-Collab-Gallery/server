package middleware

import (
	"context"
	"strings"

	v1 "smart-collab-gallery-server/api/user/v1"
	"smart-collab-gallery-server/internal/pkg"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	// AuthorizationKey HTTP Header 中的认证 key
	AuthorizationKey = "Authorization"
	// BearerPrefix Token 前缀
	BearerPrefix = "Bearer "
	// UserIDKey 上下文中存储用户 ID 的 key
	UserIDKey = "user_id"
	// UserAccountKey 上下文中存储用户账号的 key
	UserAccountKey = "user_account"
	// UserRoleKey 上下文中存储用户角色的 key
	UserRoleKey = "user_role"
)

// JWTAuth JWT 认证中间件
func JWTAuth(jwtManager *pkg.JWTManager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 从 HTTP Header 中获取 Token
			if tr, ok := transport.FromServerContext(ctx); ok {
				tokenString := tr.RequestHeader().Get(AuthorizationKey)

				if tokenString == "" {
					return nil, v1.ErrorNotLoginError("未登录")
				}

				// 去除 Bearer 前缀
				if !strings.HasPrefix(tokenString, BearerPrefix) {
					return nil, v1.ErrorInvalidToken("Token 格式错误")
				}
				tokenString = strings.TrimPrefix(tokenString, BearerPrefix)

				// 解析 Token
				claims, err := jwtManager.ParseToken(tokenString)
				if err != nil {
					return nil, v1.ErrorInvalidToken("Token 无效或已过期")
				}

				// 将用户信息存入上下文
				ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
				ctx = context.WithValue(ctx, UserAccountKey, claims.UserAccount)
				ctx = context.WithValue(ctx, UserRoleKey, claims.UserRole)
			}

			return handler(ctx, req)
		}
	}
}

// GetUserIDFromContext 从上下文中获取用户 ID
func GetUserIDFromContext(ctx context.Context) int64 {
	if userID, ok := ctx.Value(UserIDKey).(int64); ok {
		return userID
	}
	return 0
}

// GetUserAccountFromContext 从上下文中获取用户账号
func GetUserAccountFromContext(ctx context.Context) string {
	if account, ok := ctx.Value(UserAccountKey).(string); ok {
		return account
	}
	return ""
}

// GetUserRoleFromContext 从上下文中获取用户角色
func GetUserRoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}
