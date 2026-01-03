package server

import (
	"context"

	filev1 "smart-collab-gallery-server/api/file/v1"
	healthv1 "smart-collab-gallery-server/api/health/v1"
	v1 "smart-collab-gallery-server/api/helloworld/v1"
	userv1 "smart-collab-gallery-server/api/user/v1"
	"smart-collab-gallery-server/internal/conf"
	"smart-collab-gallery-server/internal/middleware"
	"smart-collab-gallery-server/internal/pkg"
	"smart-collab-gallery-server/internal/pkg/response"
	"smart-collab-gallery-server/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(bc *conf.Bootstrap, greeter *service.GreeterService, user *service.UserService, file *service.FileService, health *service.HealthService, jwtManager *pkg.JWTManager, logger log.Logger) *http.Server {
	c := bc.Server
	var opts = []http.ServerOption{
		// 应用统一响应格式编码器
		http.ResponseEncoder(response.ResponseEncoder),
		http.ErrorEncoder(response.ErrorEncoder),
		http.Middleware(
			recovery.Recovery(),
			middleware.MetricsServer(),
			// 选择性应用 JWT 认证中间件
			selector.Server(
				middleware.JWTAuth(jwtManager),
			).Match(NewWhiteListMatcher()).Build(),
			// 管理员权限中间件
			selector.Server(
				middleware.RequireAdmin(),
			).Match(NewAdminOnlyMatcher()).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// Prometheus metrics 端点
	srv.Handle("/metrics", promhttp.Handler())

	healthv1.RegisterHealthHTTPServer(srv, health)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	userv1.RegisterUserHTTPServer(srv, user)
	filev1.RegisterFileHTTPServer(srv, file)
	return srv
}

// NewWhiteListMatcher 创建白名单匹配器，不需要认证的接口
func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	// 不需要认证的接口路径
	whiteList["/api.user.v1.User/Register"] = struct{}{}
	whiteList["/api.user.v1.User/Login"] = struct{}{}
	whiteList["/api.helloworld.v1.Greeter/SayHello"] = struct{}{}
	whiteList["/api.health.v1.Health/Ping"] = struct{}{}

	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			// 在白名单中，不需要认证
			return false
		}
		// 不在白名单中，需要认证
		return true
	}
}

// NewAdminOnlyMatcher 创建管理员接口匹配器，仅管理员可访问的接口
func NewAdminOnlyMatcher() selector.MatchFunc {
	adminList := make(map[string]struct{})
	// 用户管理接口需要管理员权限
	adminList["/api.user.v1.User/AddUser"] = struct{}{}
	adminList["/api.user.v1.User/GetUserById"] = struct{}{}
	adminList["/api.user.v1.User/DeleteUser"] = struct{}{}
	adminList["/api.user.v1.User/UpdateUser"] = struct{}{}
	adminList["/api.user.v1.User/ListUserByPage"] = struct{}{}

	return func(ctx context.Context, operation string) bool {
		// 在管理员列表中，需要管理员权限
		_, ok := adminList[operation]
		return ok
	}
}
