package spicedb

import (
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"

	"github.com/raystack/salt/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type SpiceDB struct {
	client *authzed.ClientWithExperimental
}

func (s *SpiceDB) Check() error {
	_, err := s.client.ReadSchema(context.Background(), &authzedpb.ReadSchemaRequest{})
	grpCStatus := status.Convert(err)
	if grpCStatus.Code() == codes.Unavailable {
		return err
	}
	return nil
}

func New(config Config, logger log.Logger, clientMetrics *prometheus.ClientMetrics) (*SpiceDB, error) {
	endpoint := net.JoinHostPort(config.Host, config.Port)
	client, err := authzed.NewClientWithExperimentalAPIs(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(config.PreSharedKey),
		grpc.WithUnaryInterceptor(clientMetrics.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(clientMetrics.StreamClientInterceptor()),
	)
	if err != nil {
		return &SpiceDB{}, err
	}

	spiceDBClient := &SpiceDB{
		client: client,
	}
	if err := spiceDBClient.Check(); err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Connected to spiceDB: %s", endpoint))
	return spiceDBClient, nil
}
