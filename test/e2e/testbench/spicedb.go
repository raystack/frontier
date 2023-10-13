package testbench

import (
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/raystack/frontier/internal/store/spicedb"
	"github.com/raystack/salt/log"
)

const (
	spiceDBImage   = "quay.io/authzed/spicedb"
	spiceDBVersion = "v1.25.0"
)

func StartSpiceDB(logger log.Logger, network *docker.Network, pool *dockertest.Pool, preSharedKey string) (extPort string, close func() error, err error) {
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   spiceDBImage,
		Tag:          spiceDBVersion,
		Cmd:          []string{"serve", "--log-level", "debug", "--grpc-preshared-key", preSharedKey, "--datastore-engine", "memory"},
		ExposedPorts: []string{"50051/tcp"},
		NetworkID:    network.ID,
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
		config.AutoRemove = true
	})
	if err != nil {
		return
	}

	if err = res.Expire(120); err != nil {
		return "", nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	extPort = res.GetPort("50051/tcp")
	var spiceClient *spicedb.SpiceDB
	if err = pool.Retry(func() error {
		if spiceClient == nil {
			spiceClient, err = spicedb.New(spicedb.Config{
				Host: "localhost",
				Port: extPort,
			}, logger)
			if err != nil {
				return err
			}
		}
		return spiceClient.Check()
	}); err != nil {
		err = fmt.Errorf("could not connect to docker: %w", err)
		return
	}

	close = func() error {
		return res.Close()
	}
	return
}
