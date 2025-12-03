package service

import (
	"smart-collab-gallery-server/internal/conf"
	"smart-collab-gallery-server/internal/pkg"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewGreeterService, NewUserService, NewJWTManager)

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(c *conf.Auth) *pkg.JWTManager {
	return pkg.NewJWTManager(c.JwtSecret, c.JwtExpire.AsDuration())
}
