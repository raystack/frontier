package spicedb

import (
	"context"
	"fmt"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"

	"github.com/goto/salt/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type SpiceDB struct {
	client *authzed.Client
}

func (s *SpiceDB) Check() error {
	_, err := s.client.ReadSchema(context.Background(), &authzedpb.ReadSchemaRequest{})
	grpCStatus := status.Convert(err)
	if grpCStatus.Code() == codes.Unavailable {
		return err
	}
	return nil
}

func New(config Config, logger log.Logger) (*SpiceDB, error) {
	endpoint := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := authzed.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(config.PreSharedKey),
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
