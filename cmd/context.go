package cmd

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

func setCtxHeader(ctx context.Context, header string) context.Context {
	s := strings.Split(header, ":")
	key := s[0]
	val := s[1]

	md := metadata.New(map[string]string{key: val})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}
