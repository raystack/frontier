package testbench

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

const (
	pgUname  = "test_user"
	pgPasswd = "test_pass"
)

func initPG(logger log.Logger, network *docker.Network, pool *dockertest.Pool, dbName string) (connStringInternal, connStringExternal string, res *dockertest.Resource, err error) {
	name := fmt.Sprintf("postgres-%s", uuid.New().String())
	res, err = pool.RunWithOptions(&dockertest.RunOptions{
		Name:       name,
		Repository: "postgres",
		Tag:        "12",
		Env: []string{
			"POSTGRES_PASSWORD=" + pgPasswd,
			"POSTGRES_USER=" + pgUname,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432/tcp"},
		NetworkID:    network.ID,
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return
	}

	pgPort := res.GetPort("5432/tcp")
	connStringInternal = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUname, pgPasswd, name, "5432", dbName)
	fmt.Printf("connStringInternal: %v\n", connStringInternal)
	connStringExternal = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUname, pgPasswd, "localhost", pgPort, dbName)
	fmt.Printf("connStringExternal: %v\n", connStringExternal)

	if err = res.Expire(120); err != nil {
		return
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	/*if err = pool.Retry(func() error {
		_, err := pgx.Connect(context.Background(), connStringExternal)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = fmt.Errorf("could not connect to docker: %w", err)
		return
	}*/

	return
}
