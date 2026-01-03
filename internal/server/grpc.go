package server

import (
	healthv1 "smart-collab-gallery-server/api/health/v1"
	v1 "smart-collab-gallery-server/api/helloworld/v1"
	userv1 "smart-collab-gallery-server/api/user/v1"
	"smart-collab-gallery-server/internal/conf"
	"smart-collab-gallery-server/internal/middleware"
	"smart-collab-gallery-server/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(bc *conf.Bootstrap, greeter *service.GreeterService, user *service.UserService, health *service.HealthService, logger log.Logger) *grpc.Server {
	c := bc.Server
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			middleware.MetricsServer(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	healthv1.RegisterHealthServer(srv, health)
	v1.RegisterGreeterServer(srv, greeter)
	userv1.RegisterUserServer(srv, user)
	return srv
}
