package connectinterceptors

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
)

type AuditInterceptor struct {
	service *audit.Service
}

func NewAuditInterceptor(service *audit.Service) *AuditInterceptor {
	return &AuditInterceptor{
		service: service,
	}
}

func (a *AuditInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		ctx = audit.SetContextWithService(ctx, a.service)
		return next(ctx, req)
	})
}

func (a *AuditInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		return conn
	})
}

func (a *AuditInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx = audit.SetContextWithService(ctx, a.service)
		return next(ctx, conn)
	})
}
