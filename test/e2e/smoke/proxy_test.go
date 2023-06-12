package smoke_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"testing"
	"text/template"
	"time"

	"github.com/raystack/shield/pkg/server"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"

	_ "embed"

	"github.com/raystack/shield/config"
	"github.com/raystack/shield/internal/proxy"
	"github.com/raystack/shield/internal/store/spicedb"
	"github.com/raystack/shield/pkg/db"
	"github.com/raystack/shield/pkg/logger"
	"github.com/raystack/shield/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

//go:embed testdata/rule_tmpl/rule.tmpl
var ruleTemplate string

type ProxySmokeTestSuite struct {
	suite.Suite

	sClient   shieldv1beta1.ShieldServiceClient
	proxyPort int
	orgID     string
	projID    string
	userID    string
	close     func() error
}

func (s *ProxySmokeTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Assert().NoError(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	proxyPort, err := testbench.GetFreePort()
	s.Assert().NoError(err)
	s.proxyPort = proxyPort
	apiPort, err := testbench.GetFreePort()
	s.Assert().NoError(err)
	grpcPort, err := testbench.GetFreePort()
	s.Assert().NoError(err)

	ruleDir, err := os.MkdirTemp("", "shield_rules_")
	s.Assert().NoError(err)
	ruleFile, err := os.CreateTemp(ruleDir, "shield_rule_*.yml")
	s.Assert().NoError(err)
	ruleTmpl, err := template.New("rule").Parse(ruleTemplate)
	s.Assert().NoError(err)

	appConfig := &config.Shield{
		Log: logger.Config{
			Level: "fatal",
		},
		App: server.Config{
			Host: "localhost",
			Port: apiPort,
			GRPC: server.GRPCConfig{
				Port:           grpcPort,
				MaxRecvMsgSize: 2 << 10,
				MaxSendMsgSize: 2 << 10,
			},
			IdentityProxyHeader: testbench.IdentityHeader,
			UserIDHeader:        "user-id-header-value",
			ResourcesConfigPath: path.Join(testDataPath, "resource"),
		},
		Proxy: proxy.ServicesConfig{
			Services: []proxy.Config{
				{
					Name:      "base",
					Host:      "localhost",
					Port:      proxyPort,
					RulesPath: "file://" + ruleDir,
				},
			},
		},
		DB: db.Config{
			Driver:              "postgres",
			URL:                 "postgres://shield:12345@localhost:5432/shield?sslmode=disable",
			MaxIdleConns:        10,
			MaxOpenConns:        10,
			ConnMaxLifeTime:     time.Millisecond * 10,
			MaxQueryTimeoutInMS: time.Millisecond * 100,
		},
		SpiceDB: spicedb.Config{
			Host:            "localhost",
			Port:            "50051",
			PreSharedKey:    "shield",
			FullyConsistent: true,
		},
	}

	ctx := context.Background()
	logger := logger.InitLogger(appConfig.Log)
	ctx, networkCancel := context.WithCancel(ctx)

	// create a pool of containers
	pool, err := dockertest.NewPool("")
	s.Assert().NoError(err)

	// Upsert a bridge network for docker containers to communicate
	network, err := pool.Client.CreateNetwork(docker.CreateNetworkOptions{
		Name:    fmt.Sprintf("bridge-%s", uuid.New().String()),
		Context: ctx,
	})
	s.Assert().NoError(err)

	// setup spiceDB
	spiceDBPort, spiceDBClose, err := testbench.StartSpiceDB(logger, network, pool, appConfig.SpiceDB.PreSharedKey)
	s.Assert().NoError(err)
	appConfig.SpiceDB.Port = spiceDBPort

	// setup pg for shield
	_, connStringExternal, pgResource, err := testbench.StartPG(network, pool, "shield")
	s.Assert().NoError(err)
	appConfig.DB.URL = connStringExternal

	// run migrations
	err = testbench.MigrateShield(logger, appConfig)
	s.Assert().NoError(err)

	// echo server for proxy
	echoPort, echoResource, err := testbench.EchoServer(network, pool)
	s.Assert().NoError(err)
	// update ruleset with echo server port
	err = ruleTmpl.Execute(ruleFile, map[string]string{
		"host": "localhost",
		"port": echoPort,
	})
	s.Assert().NoError(err)

	testbench.StartShield(logger, appConfig)

	// let shield start
	time.Sleep(time.Second * 3)

	// create fixtures
	sClient, sClose, err := testbench.CreateClient(ctx, net.JoinHostPort(appConfig.App.Host, strconv.Itoa(appConfig.App.GRPC.Port)))
	s.sClient = sClient
	s.Assert().NoError(err)

	s.close = func() error {
		err1 := pgResource.Close()
		err2 := spiceDBClose()
		err3 := sClose()
		err4 := echoResource.Close()
		networkCancel()
		os.Remove(ruleFile.Name())
		return errors.Join(err1, err2, err3, err4)
	}

	err = testbench.BootstrapUsers(ctx, sClient, testbench.OrgAdminEmail)
	s.Assert().NoError(err)

	err = testbench.BootstrapOrganizations(ctx, sClient, testbench.OrgAdminEmail)
	s.Assert().NoError(err)

	orgResp, err := sClient.ListOrganizations(ctx, &shieldv1beta1.ListOrganizationsRequest{})
	s.Assert().NoError(err)
	s.Assert().NotEqual(0, len(orgResp.Organizations))
	s.orgID = orgResp.Organizations[0].GetId()

	err = testbench.BootstrapProject(ctx, sClient, testbench.OrgAdminEmail)
	s.Assert().NoError(err)

	projResp, err := sClient.ListOrganizationProjects(ctx, &shieldv1beta1.ListOrganizationProjectsRequest{
		Id: s.orgID,
	})
	s.Assert().NoError(err)
	s.Assert().NotEqual(0, len(projResp.Projects))
	s.projID = projResp.Projects[0].GetId()

	listUsers, err := sClient.ListUsers(ctx, &shieldv1beta1.ListUsersRequest{})
	s.Assert().NoError(err)
	s.userID = listUsers.Users[0].Id
}

func (s *ProxySmokeTestSuite) TearDownSuite() {
	proc, err := os.FindProcess(os.Getpid())
	s.Assert().NoError(err)
	proc.Signal(os.Interrupt)

	// let shield finish
	time.Sleep(time.Second * 1)

	err = s.close()
	s.Assert().NoError(err)
}

func (s *ProxySmokeTestSuite) TestProxyToEchoServer() {
	s.Run("should be able to proxy to an echo server", func() {
		url := fmt.Sprintf("http://localhost:%d/api/ping", s.proxyPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		req.Header.Set(testbench.IdentityHeader, "john.doe@raystack.com")

		res, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)

		defer res.Body.Close()
		s.Assert().Equal(200, res.StatusCode)
	})
	s.Run("resource created on echo server should persist in shieldDB", func() {
		url := fmt.Sprintf("http://localhost:%d/api/resource", s.proxyPort)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		s.Require().NoError(err)

		req.Header.Set(testbench.IdentityHeader, testbench.OrgAdminEmail)
		req.Header.Set("X-Shield-Project", s.projID)
		req.Header.Set("X-Shield-User", s.userID)
		req.Header.Set("X-Shield-Name", "test-resource")
		req.Header.Set("X-Shield-Resource-Type", "cart")

		res, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)

		defer res.Body.Close()

		resourceResp, err := s.sClient.ListProjectResources(context.Background(), &shieldv1beta1.ListProjectResourcesRequest{
			ProjectId: s.projID,
		})
		s.Assert().NoError(err)

		s.Assert().Equal(1, len(resourceResp.GetResources()))
		s.Assert().Equal(200, res.StatusCode)
		s.Assert().Equal("test-resource", resourceResp.GetResources()[0].Name)
	})
}

func TestProxySmokeTestSuite(t *testing.T) {
	suite.Run(t, new(ProxySmokeTestSuite))
}
