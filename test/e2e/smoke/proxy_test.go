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

	"google.golang.org/grpc/metadata"

	"github.com/raystack/frontier/pkg/server"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	_ "embed"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/internal/proxy"
	"github.com/raystack/frontier/internal/store/spicedb"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

//go:embed testdata/rule_tmpl/rule.tmpl
var ruleTemplate string

type ProxySmokeTestSuite struct {
	suite.Suite

	sClient   frontierv1beta1.FrontierServiceClient
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

	ruleDir, err := os.MkdirTemp("", "frontier_rules_")
	s.Assert().NoError(err)
	ruleFile, err := os.CreateTemp(ruleDir, "frontier_rule_*.yml")
	s.Assert().NoError(err)
	ruleTmpl, err := template.New("rule").Parse(ruleTemplate)
	s.Assert().NoError(err)

	appConfig := &config.Frontier{
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
			URL:                 "postgres://frontier:12345@localhost:5432/frontier?sslmode=disable",
			MaxIdleConns:        10,
			MaxOpenConns:        10,
			ConnMaxLifeTime:     time.Millisecond * 10,
			MaxQueryTimeoutInMS: time.Millisecond * 100,
		},
		SpiceDB: spicedb.Config{
			Host:            "localhost",
			Port:            "50051",
			PreSharedKey:    "frontier",
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

	// setup pg for frontier
	_, connStringExternal, pgResource, err := testbench.StartPG(network, pool, "frontier")
	s.Assert().NoError(err)
	appConfig.DB.URL = connStringExternal

	// run migrations
	err = testbench.MigrateFrontier(logger, appConfig)
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

	testbench.StartFrontier(logger, appConfig)

	// let frontier start
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

	orgResp, err := sClient.ListOrganizations(ctx, &frontierv1beta1.ListOrganizationsRequest{})
	s.Assert().NoError(err)
	s.Assert().NotEqual(0, len(orgResp.Organizations))
	s.orgID = orgResp.Organizations[0].GetId()

	err = testbench.BootstrapProject(ctx, sClient, testbench.OrgAdminEmail)
	s.Assert().NoError(err)

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	projResp, err := sClient.ListOrganizationProjects(ctx, &frontierv1beta1.ListOrganizationProjectsRequest{
		Id: s.orgID,
	})
	s.Assert().NoError(err)
	s.Assert().NotEqual(0, len(projResp.Projects))
	s.projID = projResp.Projects[0].GetId()

	listUsers, err := sClient.ListUsers(ctx, &frontierv1beta1.ListUsersRequest{})
	s.Assert().NoError(err)
	s.userID = listUsers.Users[0].Id
}

func (s *ProxySmokeTestSuite) TearDownSuite() {
	proc, err := os.FindProcess(os.Getpid())
	s.Assert().NoError(err)
	proc.Signal(os.Interrupt)

	// let frontier finish
	time.Sleep(time.Second * 1)

	err = s.close()
	s.Assert().NoError(err)
}

func (s *ProxySmokeTestSuite) TestProxyToEchoServer() {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("should be able to proxy to an echo server", func() {
		url := fmt.Sprintf("http://localhost:%d/api/ping", s.proxyPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		req.Header.Set(testbench.IdentityHeader, "john.doe@raystack.org")

		res, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)

		defer res.Body.Close()
		s.Assert().Equal(200, res.StatusCode)
	})
	s.Run("resource created on echo server should persist in frontierDB", func() {
		url := fmt.Sprintf("http://localhost:%d/api/resource", s.proxyPort)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		s.Require().NoError(err)

		req.Header.Set(testbench.IdentityHeader, testbench.OrgAdminEmail)
		req.Header.Set("X-Frontier-Project", s.projID)
		req.Header.Set("X-Frontier-User", s.userID)
		req.Header.Set("X-Frontier-Name", "test-resource")
		req.Header.Set("X-Frontier-Resource-Type", "cart")

		res, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)

		defer res.Body.Close()

		resourceResp, err := s.sClient.ListProjectResources(ctx, &frontierv1beta1.ListProjectResourcesRequest{
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
