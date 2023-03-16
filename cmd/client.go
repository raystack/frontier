package cmd

import (
	"context"
	"time"

	shieldv1beta1 "github.com/goto/shield/proto/v1beta1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func createConnection(ctx context.Context, host string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	return grpc.DialContext(ctx, host, opts...)
}

func createClient(ctx context.Context, host string) (shieldv1beta1.ShieldServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(ctx, time.Second*2)
	conn, err := createConnection(dialTimeoutCtx, host)
	if err != nil {
		dialCancel()
		return nil, nil, err
	}
	cancel := func() {
		dialCancel()
		conn.Close()
	}

	client := shieldv1beta1.NewShieldServiceClient(conn)
	return client, cancel, nil
}

func isClientCLI(cmd *cobra.Command) bool {
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["client"] == "true" {
			return true
		}
	}
	return false
}

func overrideClientConfigHost(cmd *cobra.Command, cliConfig *Config) error {
	if cliConfig == nil {
		return ErrClientConfigNotFound
	}

	host, err := cmd.Flags().GetString("host")
	if err == nil && host != "" {
		cliConfig.Host = host
		return nil
	}

	if cliConfig.Host == "" {
		return ErrClientConfigHostNotFound
	}

	return nil
}

func bindFlagsFromClientConfig(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("host", "h", "", "Shield API service to connect to")
}
