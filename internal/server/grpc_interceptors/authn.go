package grpc_interceptors

import (
	"context"

	"google.golang.org/grpc"
)

func AuthnUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		//TODO(kushsharma): should we deny all request by default?
		return handler(ctx, req)
	}
}
