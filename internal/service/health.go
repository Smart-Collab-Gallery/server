package service

import (
	"context"

	v1 "smart-collab-gallery-server/api/health/v1"
)

// HealthService is a health check service.
type HealthService struct {
	v1.UnimplementedHealthServer
}

// NewHealthService new a health service.
func NewHealthService() *HealthService {
	return &HealthService{}
}

// Ping implements health.HealthServer.
func (s *HealthService) Ping(ctx context.Context, in *v1.PingRequest) (*v1.PingReply, error) {
	return &v1.PingReply{Status: "ok"}, nil
}
