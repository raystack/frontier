package cmd

import (
	"context"
	"time"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
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

func createClient(cmd *cobra.Command) (shieldv1beta1.ShieldServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(cmd.Context(), time.Second*2)
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return nil, nil, err
	}
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
