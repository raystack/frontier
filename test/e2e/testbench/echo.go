package testbench

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func EchoServer(network *docker.Network, pool *dockertest.Pool) (string, *dockertest.Resource, error) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "ealen/echo-server",
		Tag:          "0.7.0",
		NetworkID:    network.ID,
		ExposedPorts: []string{"80/tcp"},
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
		config.AutoRemove = true
	})
	if err != nil {
		return "", resource, err
	}
	if err = resource.Expire(120); err != nil {
		return "", resource, err
	}
	accessPort := resource.GetPort("80/tcp")

	pool.MaxWait = 60 * time.Second
	if err = pool.Retry(func() error {
		resp, err := http.Head("http://localhost:" + accessPort)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("not ready")
		}
		return nil
	}); err != nil {
		return "", resource, fmt.Errorf("could not boot up echo server in docker: %w", err)
	}

	return accessPort, resource, nil
}
