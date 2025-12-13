package server

import (
	"context"

	healthv1 "smart-collab-gallery-server/api/health/v1"
	v1 "smart-collab-gallery-server/api/helloworld/v1"
	userv1 "smart-collab-gallery-server/api/user/v1"
	"smart-collab-gallery-server/internal/conf"
	"smart-collab-gallery-server/internal/middleware"
	"smart-collab-gallery-server/internal/pkg"
	"smart-collab-gallery-server/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, user *service.UserService, health *service.HealthService, jwtManager *pkg.JWTManager, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			metrics.Server(),
			// 选择性应用 JWT 认证中间件
			selector.Server(
				middleware.JWTAuth(jwtManager),
			).Match(NewWhiteListMatcher()).Build(),
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
