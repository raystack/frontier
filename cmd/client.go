package cmd

import (
	"context"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func createConnection(ctx context.Context, host string, caCertFile string) (*grpc.ClientConn, error) {
	creds := insecure.NewCredentials()
	if caCertFile != "" {
		tlsCreds, err := credentials.NewClientTLSFromFile(caCertFile, "")
		if err != nil {
			return nil, err
		}
		creds = tlsCreds
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	}
	return grpc.DialContext(ctx, host, opts...)
}

func createClient(ctx context.Context, host string) (frontierv1beta1.FrontierServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(ctx, time.Second*2)
	conn, err := createConnection(dialTimeoutCtx, host, "")
	if err != nil {
		dialCancel()
		return nil, nil, err
	}
	cancel := func() {
		dialCancel()
		conn.Close()
	}

	client := frontierv1beta1.NewFrontierServiceClient(conn)
	return client, cancel, nil
}

func createAdminClient(ctx context.Context, host string) (frontierv1beta1.AdminServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(ctx, time.Second*2)
	conn, err := createConnection(dialTimeoutCtx, host, "")
	if err != nil {
		dialCancel()
		return nil, nil, err
	}
	cancel := func() {
		dialCancel()
		conn.Close()
	}

	client := frontierv1beta1.NewAdminServiceClient(conn)
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
	cmd.PersistentFlags().StringP("host", "h", "", "Frontier API service to connect to")
}
