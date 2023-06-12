package testbench

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/raystack/salt/log"
	"github.com/raystack/shield/internal/store/spicedb"
)

const (
	spiceDBImage   = "quay.io/authzed/spicedb"
	spiceDBVersion = "v1.19.1"
)

func StartSpiceDB(logger log.Logger, network *docker.Network, pool *dockertest.Pool, preSharedKey string) (extPort string, close func() error, err error) {
	pgConnString, _, pgResource, err := StartPG(network, pool, "spicedb")
	if err != nil {
		return "", nil, err
	}
	if err = migrateSpiceDB(network, pool, pgConnString); err != nil {
		return "", nil, err
	}

	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   spiceDBImage,
		Tag:          spiceDBVersion,
		Cmd:          []string{"serve", "--log-level", "debug", "--grpc-preshared-key", preSharedKey, "--datastore-engine", "postgres", "--datastore-conn-uri", pgConnString},
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
		err1 := pgResource.Close()
		err2 := res.Close()
		return errors.Join(err1, err2)
	}
	return
}

func migrateSpiceDB(network *docker.Network, pool *dockertest.Pool, pgConnString string) error {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: spiceDBImage,
		Tag:        spiceDBVersion,
		Cmd:        []string{"migrate", "head", "--datastore-engine", "postgres", "--datastore-conn-uri", pgConnString},
		NetworkID:  network.ID,
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
		config.AutoRemove = true
	})
	if err != nil {
		return err
	}
	if err = resource.Expire(60); err != nil {
		return err
	}

	waitCtx, cancel := context.WithTimeout(context.Background(), waitContainerTimeout)
	defer cancel()

	// Ensure the command completed successfully.
	status, err := pool.Client.WaitContainerWithContext(resource.Container.ID, context.Background())
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
	_ = resource.Close()
	return nil
}
