package interceptors

import (
	"context"
	"fmt"

	"github.com/raystack/shield/core/authenticate"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func EnrichCtxWithPassthroughEmail(identityHeader string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if len(identityHeader) == 0 {
			// if not configured, skip
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", fmt.Errorf("metadata for identity doesn't exist")
		}

		var email string
		if metadataValues := md.Get(identityHeader); len(metadataValues) > 0 {
			email = metadataValues[0]
		}

		ctx = authenticate.SetContextWithEmail(ctx, email)
		return handler(ctx, req)
	}
}
