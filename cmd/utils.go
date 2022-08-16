package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
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

func isClientCLI(cmd *cobra.Command) bool {
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["client"] == "true" {
			return true
		}
	}
	return false
}

func clientConfigHostExist(cmd *cobra.Command) bool {
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return false
	}
	if host != "" {
		return true
	}
	return false
}
