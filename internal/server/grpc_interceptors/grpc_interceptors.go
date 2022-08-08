package grpc_interceptors

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func EnrichCtxWithIdentity(identityHeader string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", fmt.Errorf("metadata for identity doesn't exist")
		}

		var email string
		metadataValues := md.Get(identityHeader)
		if len(metadataValues) > 0 {
			email = metadataValues[0]
		}

		ctx = user.SetContextWithEmail(ctx, email)
		return handler(ctx, req)
	}
}
