package testbench

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/raystack/frontier/pkg/logger"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/internal/store/spicedb"
	"github.com/raystack/frontier/pkg/db"
)

const (
	preSharedKey         = "frontier"
	waitContainerTimeout = 60 * time.Second
)

var (
	RuleCacheRefreshDelay = time.Minute * 2
)

type TestBench struct {
	Pool        *dockertest.Pool
	Network     *docker.Network
	Resources   []*dockertest.Resource
	Client      frontierv1beta1.FrontierServiceClient
	AdminClient frontierv1beta1.AdminServiceClient
	close       func() error
}

func Init(appConfig *config.Frontier) (*TestBench, error) {
	var (
		err    error
		logger = logger.InitLogger(appConfig.Log)
		te     = &TestBench{}
	)
	te.Pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	// Upsert a bridge network for testing.
	// run `docker network prune` on local machine if failed to create a new network
	te.Network, err = te.Pool.Client.CreateNetwork(docker.CreateNetworkOptions{
		Name: fmt.Sprintf("bridge-%s", uuid.New().String()),
	})
	if err != nil {
		return nil, err
	}

	_, connMainPGExternal, pgResource, err := StartPG(te.Network, te.Pool, "frontier")
	if err != nil {
		return nil, err
	}

	spiceDBPort, spiceDBClose, err := StartSpiceDB(logger, te.Network, te.Pool, preSharedKey)
	if err != nil {
		return nil, err
	}

	appConfig.DB = db.Config{
		Driver:          "postgres",
		URL:             connMainPGExternal,
		MaxIdleConns:    10,
		MaxOpenConns:    10,
		ConnMaxLifeTime: time.Millisecond * 100,
		MaxQueryTimeout: time.Millisecond * 100,
	}
	appConfig.SpiceDB = spicedb.Config{
		Host:            "localhost",
		Port:            spiceDBPort,
		PreSharedKey:    preSharedKey,
		FullyConsistent: true,
	}

	if err = MigrateFrontier(logger, appConfig); err != nil {
		return nil, err
	}

	te.close = func() error {
		err1 := pgResource.Close()
		err2 := spiceDBClose()
		return errors.Join(err1, err2)
	}

	StartFrontier(logger, appConfig)

	// create fixtures
	sClient, sClose, err := CreateClient(context.Background(), net.JoinHostPort(appConfig.App.Host, strconv.Itoa(appConfig.App.GRPC.Port)))
	if err != nil {
		return nil, err
	}
	te.Client = sClient

	adClient, adClose, err := CreateAdminClient(context.Background(), net.JoinHostPort(appConfig.App.Host, strconv.Itoa(appConfig.App.GRPC.Port)))
	if err != nil {
		return nil, err
	}
	te.AdminClient = adClient

	te.close = func() error {
		err1 := pgResource.Close()
		err2 := spiceDBClose()
		err3 := sClose()
		err4 := adClose()
		return errors.Join(err1, err2, err3, err4)
	}

	// let frontier start
	time.Sleep(time.Second * 2)
	return te, nil
}

func (te *TestBench) Close() error {
	proc, err := os.FindProcess(os.Getpid())
	if err == nil {
		proc.Signal(os.Interrupt)
	}
	return errors.Join(err, te.close())
}
