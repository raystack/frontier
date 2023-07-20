package interceptors

import (
	"context"

	"github.com/raystack/shield/core/audit"
	"google.golang.org/grpc"
)

func UnaryCtxWithAudit(service *audit.Service) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = audit.SetContextWithService(ctx, service)
		return handler(ctx, req)
	}
}
