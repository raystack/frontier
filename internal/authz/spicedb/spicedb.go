package spicedb

import (
	"fmt"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/odpf/shield/config"
	"google.golang.org/grpc"
)

type SpiceDB struct {
	client *authzed.Client
}

func (s *SpiceDB) Check() bool {
	return false
}

func New(config config.SpiceDBConfig) (*SpiceDB, error) {
	endpoint := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := authzed.NewClient(endpoint, grpc.WithInsecure(), grpcutil.WithInsecureBearerToken(config.PreSharedKey))
	if err != nil {
		return &SpiceDB{}, err
	}
	return &SpiceDB{
		client,
	}, nil
}
