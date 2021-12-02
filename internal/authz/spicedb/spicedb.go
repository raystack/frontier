package spicedb

import (
	"context"
	"fmt"

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

func (p *Policy) AddPolicy(schema string) error {
	request := &pb.WriteSchemaRequest{Schema: schema}
	resp, err := p.client.WriteSchema(context.Background(), request)
	if err != nil {
		return err
	}
	fmt.Println("Output", resp)
	return nil
}

func New(config config.SpiceDBConfig) (*SpiceDB, error) {
	endpoint := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := authzed.NewClient(endpoint, grpc.WithInsecure(), grpcutil.WithInsecureBearerToken(config.PreSharedKey))
	if err != nil {
		return &SpiceDB{}, err
	}

	policy := &Policy{
		client,
	}
	return &SpiceDB{
		policy,
	}, nil
}
