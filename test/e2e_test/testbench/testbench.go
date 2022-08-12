package testbench

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/store/spicedb"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

const (
	preSharedKey         = "shield"
	waitContainerTimeout = 60 * time.Second
)

type TestBench struct {
	PGConfig          db.Config
	SpiceDBConfig     spicedb.Config
	bridgeNetworkName string
	pool              *dockertest.Pool
	network           *docker.Network
	resources         []*dockertest.Resource
}

func Init(appConfig *config.Shield) (*TestBench, error) {
	var (
		err    error
		logger = log.NewZap()
	)

	te := &TestBench{
		bridgeNetworkName: fmt.Sprintf("bridge-%s", uuid.New().String()),
		resources:         []*dockertest.Resource{},
	}

	te.pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	// Create a bridge network for testing.
	te.network, err = te.pool.Client.CreateNetwork(docker.CreateNetworkOptions{
		Name: te.bridgeNetworkName,
	})
	if err != nil {
		return nil, err
	}

	// pg 1
	logger.Info("creating main postgres...")
	_, connMainPGExternal, res, err := initPG(logger, te.network, te.pool, "test_db")
	if err != nil {
		return nil, err
	}
	te.resources = append(te.resources, res)
	logger.Info("main postgres is created")

	// pg 2
	logger.Info("creating spicedb postgres...")
	connSpicePGInternal, _, res, err := initPG(logger, te.network, te.pool, "spicedb")
	if err != nil {
		return nil, err
	}
	te.resources = append(te.resources, res)
	logger.Info("spicedb postgres is created")

	logger.Info("migrating spicedb...")
	if err = migrateSpiceDB(logger, te.network, te.pool, connSpicePGInternal); err != nil {
		return nil, err
	}
	logger.Info("spicedb is migrated")

	logger.Info("starting up spicedb...")
	spiceDBPort, res, err := startSpiceDB(logger, te.network, te.pool, connSpicePGInternal, preSharedKey)
	if err != nil {
		return nil, err
	}
	te.resources = append(te.resources, res)
	logger.Info("spicedb is up")

	te.PGConfig = db.Config{
		Driver:              "postgres",
		URL:                 connMainPGExternal,
		MaxIdleConns:        10,
		MaxOpenConns:        10,
		ConnMaxLifeTime:     time.Millisecond * 100,
		MaxQueryTimeoutInMS: time.Millisecond * 100,
	}

	te.SpiceDBConfig = spicedb.Config{
		Host:         "localhost",
		Port:         spiceDBPort,
		PreSharedKey: preSharedKey,
	}

	appConfig.DB = te.PGConfig
	appConfig.SpiceDB = te.SpiceDBConfig

	logger.Info("migrating shield...")
	if err = migrateShield(appConfig); err != nil {
		return nil, err
	}
	logger.Info("shield is migrated")

	logger.Info("starting up shield...")
	startShield(appConfig)
	logger.Info("shield is up")

	return te, nil
}

func (te *TestBench) CleanUp() error {
	for _, r := range te.resources {
		if err := r.Close(); err != nil {
			return fmt.Errorf("could not purge resource: %w", err)
		}
	}
	if err := te.pool.Client.RemoveNetwork(te.network.ID); err != nil {
		return err
	}
	return nil
}
