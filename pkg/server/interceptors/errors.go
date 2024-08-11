package interceptors

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternalServer       = errors.New("internal server error")
	grpcInternalServerError = status.Errorf(codes.Internal, ErrInternalServer.Error())
)

// UnaryErrorHandler is a unary server interceptor that captures the request and response and overrides the error message
func UnaryErrorHandler() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			return nil, errWrapper(ctx, err)
		}
		return resp, nil
	}
}

func errWrapper(ctx context.Context, err error) error {
	grpczap.Extract(ctx).Error(err.Error())

	// check if grpc status error
	if _, ok := status.FromError(err); ok {
		return err
	}

	// check if err was context canceled
	if errors.Is(err, context.Canceled) {
		// override grpc status with context canceled
		return status.Error(codes.Canceled, err.Error())
	}

	// defaults to internal server error
	return grpcInternalServerError
}
