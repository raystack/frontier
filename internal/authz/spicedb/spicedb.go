package spicedb

import (
	"context"
	"fmt"

	"github.com/odpf/salt/log"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/odpf/shield/config"
	"google.golang.org/grpc"
)

type SpiceDB struct {
	Policy *Policy
}

type Policy struct {
	client *authzed.Client
}

func (s *SpiceDB) Check() bool {
	return false
}

func (p *Policy) AddPolicy(ctx context.Context, schema string) error {
	request := &pb.WriteSchemaRequest{Schema: schema}
	_, err := p.client.WriteSchema(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func New(config config.SpiceDBConfig, logger log.Logger) (*SpiceDB, error) {
	endpoint := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := authzed.NewClient(endpoint, grpc.WithInsecure(), grpcutil.WithInsecureBearerToken(config.PreSharedKey))
	if err != nil {
		return &SpiceDB{}, err
	}

	logger.Info(fmt.Sprintf("Connected to spiceDB: %s", endpoint))

	policy := &Policy{
		client,
	}
	return &SpiceDB{
		policy,
	}, nil
}
