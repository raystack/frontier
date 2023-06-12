package testbench

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/raystack/salt/log"
	"github.com/raystack/shield/internal/store/spicedb"
)

func migrateSpiceDB(logger log.Logger, network *docker.Network, pool *dockertest.Pool, pgConnString string) error {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "quay.io/authzed/spicedb",
		Tag:        "v1.0.0",
		Cmd:        []string{"spicedb", "migrate", "head", "--datastore-engine", "postgres", "--datastore-conn-uri", pgConnString},
		NetworkID:  network.ID,
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return err
	}

	if err := resource.Expire(120); err != nil {
		return err
	}

	waitCtx, cancel := context.WithTimeout(context.Background(), waitContainerTimeout)
	defer cancel()

	// Ensure the command completed successfully.
	status, err := pool.Client.WaitContainerWithContext(resource.Container.ID, waitCtx)
	if err != nil {
		return err
	}

	if status != 0 {
		stream := new(bytes.Buffer)

		if err = pool.Client.Logs(docker.LogsOptions{
			Context:      waitCtx,
			OutputStream: stream,
			ErrorStream:  stream,
			Stdout:       true,
			Stderr:       true,
			Container:    resource.Container.ID,
		}); err != nil {
			return err
		}

		return fmt.Errorf("got non-zero exit code %s", stream.String())
	}

	//purge
	if err := pool.Purge(resource); err != nil {
		return err
	}

	return nil
}

func startSpiceDB(logger log.Logger, network *docker.Network, pool *dockertest.Pool, pgConnString string, preSharedKey string) (extPort string, res *dockertest.Resource, err error) {
	res, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "quay.io/authzed/spicedb",
		Tag:          "v1.0.0",
		Cmd:          []string{"spicedb", "serve", "--log-level", "debug", "--grpc-preshared-key", preSharedKey, "--grpc-no-tls", "--datastore-engine", "postgres", "--datastore-conn-uri", pgConnString},
		ExposedPorts: []string{"50051/tcp"},
		NetworkID:    network.ID,
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return
	}

	extPort = res.GetPort("50051/tcp")

	if err := res.Expire(120); err != nil {
		return "", nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

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
	return
}
