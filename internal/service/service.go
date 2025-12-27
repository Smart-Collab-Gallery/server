package service

import (
	"smart-collab-gallery-server/internal/conf"
	"smart-collab-gallery-server/internal/pkg"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewGreeterService, NewUserService, NewFileService, NewHealthService, NewJWTManager, NewCOSManager)

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(c *conf.Auth) *pkg.JWTManager {
	return pkg.NewJWTManager(c.JwtSecret, c.JwtExpire.AsDuration())
}

// NewCOSManager 创建 COS 管理器
func NewCOSManager(c *conf.Cos, logger log.Logger) (*pkg.COSManager, error) {
	return pkg.NewCOSManager(c, logger)
}
