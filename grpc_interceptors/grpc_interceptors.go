package grpc_interceptors

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	identityCtx = "identityCtx"
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

		ctx = context.WithValue(ctx, identityCtx, email)
		return handler(ctx, req)
	}
}

func GetIdentityHeader(ctx context.Context) (string, bool) {
	identity, ok := ctx.Value(identityCtx).(string)
	return identity, ok
}
