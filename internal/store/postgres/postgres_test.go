package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/internal/store/postgres/migrations"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

const (
	logLevelDebug = "debug"
	pg_uname      = "test_user"
	pg_passwd     = "test_pass"
	pg_dbname     = "test_db"
)

func newTestClient(logger log.Logger) (*db.Client, *dockertest.Pool, *dockertest.Resource, error) {
	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12",
		Env: []string{
			"POSTGRES_PASSWORD=" + pg_passwd,
			"POSTGRES_USER=" + pg_uname,
			"POSTGRES_DB=" + pg_dbname,
		},
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create dockertest pool: %w", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	pg_port := resource.GetPort("5432/tcp")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse external port of container to int: %w", err)
	}

	// attach terminal logger to container if exists
	// for debugging purpose
	if logger.Level() == logLevelDebug {
		logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
			Container:    resource.Container.ID,
			OutputStream: logger.Writer(),
			ErrorStream:  logger.Writer(),
			Stderr:       true,
			Stdout:       true,
			Stream:       true,
		})
		if err != nil {
			logger.Fatal("could not connect to postgres container log output", "error", err)
		}
		defer func() {
			err = logWaiter.Close()
			if err != nil {
				logger.Fatal("could not close container log", "error", err)
			}

			err = logWaiter.Wait()
			if err != nil {
				logger.Fatal("could not wait for container log to close", "error", err)
			}
		}()
	}

	// Tell docker to hard kill the container in 120 seconds
	if err := resource.Expire(120); err != nil {
		return nil, nil, nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	pgConfig := db.Config{
		Driver:              "pgx",
		URL:                 fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", pg_uname, pg_passwd, pg_port, pg_dbname),
		MaxIdleConns:        10,
		MaxOpenConns:        10,
		ConnMaxLifeTime:     time.Millisecond * 100,
		MaxQueryTimeoutInMS: time.Millisecond * 100,
	}
	var pgClient *db.Client
	if err = pool.Retry(func() error {
		pgClient, err = db.New(pgConfig)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	err = setup(context.Background(), logger, pgClient, pgConfig)
	if err != nil {
		logger.Fatal("failed to setup and migrate DB", "error", err)
	}
	return pgClient, pool, resource, nil
}

func purgeDocker(pool *dockertest.Pool, resource *dockertest.Resource) error {
	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}

func setup(ctx context.Context, logger log.Logger, client *db.Client, cfg db.Config) (err error) {
	var queries = []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
	}

	err = execQueries(ctx, client, queries)
	if err != nil {
		return
	}

	err = db.RunMigrations(cfg, migrations.MigrationFs, migrations.ResourcePath)
	return
}

// ExecQueries is used for executing list of db query
func execQueries(ctx context.Context, client *db.Client, queries []string) error {
	for _, query := range queries {
		if _, err := client.DB.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func bootstrapAction(client *db.Client) ([]action.Action, error) {
	actionRepository := postgres.NewActionRepository(client)
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mock-action.json")
	if err != nil {
		return nil, err
	}

	var data []action.Action
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []action.Action
	for _, d := range data {
		act, err := actionRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, act)
	}

	return insertedData, nil
}

func bootstrapNamespace(client *db.Client) ([]namespace.Namespace, error) {
	namespaceRepository := postgres.NewNamespaceRepository(client)
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mock-namespace.json")
	if err != nil {
		return nil, err
	}

	var data []namespace.Namespace
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []namespace.Namespace
	for _, d := range data {
		domain, err := namespaceRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapUser(client *db.Client) ([]user.User, error) {
	userRepository := postgres.NewUserRepository(client)
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mock-user.json")
	if err != nil {
		return nil, err
	}

	var data []user.User
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []user.User
	for _, d := range data {
		domain, err := userRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapRole(client *db.Client) ([]string, error) {
	roleRepository := postgres.NewRoleRepository(client)
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mock-role.json")
	if err != nil {
		return nil, err
	}

	var data []role.Role
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []string
	for _, d := range data {
		domain, err := roleRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapPolicy(client *db.Client) ([]string, error) {
	policyRepository := postgres.NewPolicyRepository(client)
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mock-policy.json")
	if err != nil {
		return nil, err
	}

	var data []policy.Policy
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []string
	for _, d := range data {
		domain, err := policyRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapRelation(client *db.Client) ([]relation.Relation, error) {
	relationRepository := postgres.NewRelationRepository(client)
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mock-relation.json")
	if err != nil {
		return nil, err
	}

	var data []relation.Relation
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []relation.Relation
	for _, d := range data {
		domain, err := relationRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}
