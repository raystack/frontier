package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/internal/store/blob"
	"github.com/odpf/shield/internal/store/postgres/migrations"
	"github.com/odpf/shield/internal/store/spicedb"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/logger"
	"github.com/odpf/shield/test/e2e_test/testbench"
	"github.com/stretchr/testify/suite"
)

type EndToEndProxySmokeTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
	proxyport int
	orgID     string
	projID    string
	dbClient  *db.Client
}

func (s *EndToEndProxySmokeTestSuite) SetupTest() {
	var (
		orgID  string
		projID string
	)

	wd, err := os.Getwd()
	s.Require().Nil(err)

	parent := filepath.Dir(wd)
	testDataPath := parent + "/../../test/e2e_test/smoke/testdata/"

	proxyPort, err := testbench.GetFreePort()
	s.Require().Nil(err)

	s.proxyport = proxyPort

	apiPort, err := testbench.GetFreePort()
	s.Require().Nil(err)

	appConfig := &config.Shield{
		Log: logger.Config{
			Level: "fatal",
		},
		App: server.Config{
			Port:                      apiPort,
			IdentityProxyHeader:       testbench.IdentityHeader,
			UserIDHeader:              "user-id-header-value",
			ResourcesConfigPath:       fmt.Sprintf("file://%s%s", testDataPath, "resource"),
			ResourcesConfigPathSecret: "",
			RulesPath:                 fmt.Sprintf("file://%s%s", testDataPath, "rule"),
		},
		Proxy: proxy.ServicesConfig{
			Services: []proxy.Config{
				{
					Name:      "base",
					Host:      "localhost",
					Port:      proxyPort,
					RulesPath: fmt.Sprintf("file://%s%s", testDataPath, "rule"),
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
			Host:         "localhost",
			Port:         "50051",
			PreSharedKey: "shield",
		},
	}

	ctx := context.Background()
	logger := logger.InitLogger(appConfig.Log)

	spiceDBClient, err := spicedb.New(appConfig.SpiceDB, logger)
	if err != nil {
		logger.Fatal("failed to create spiceDB client", err)
		return
	}

	dbClient, err := testbench.SetupDB(appConfig.DB)
	if err != nil {
		logger.Fatal("failed to setup database")
		return
	}
	s.dbClient = dbClient

	err = db.RunMigrations(db.Config{
		Driver: appConfig.DB.Driver,
		URL:    appConfig.DB.URL,
	}, migrations.MigrationFs, migrations.ResourcePath)
	fmt.Printf("migrations.ResourcePath: %v\n", migrations.ResourcePath)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to run migration: %s", err))
	}

	userCreationQuery := "INSERT INTO users (name,email) VALUES ('John', 'john.doe@odpf.com') ON CONFLICT DO NOTHING"
	_, err = dbClient.DB.Query(userCreationQuery)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to run query: %s", err))
	}

	orgCreationQuery := "INSERT INTO organizations (name, slug) VALUES ('ODPF', 'odpf org') ON CONFLICT DO NOTHING"
	_, err = dbClient.DB.Query(orgCreationQuery)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to run query: %s", err))
	}

	orgSelectQuery := "SELECT id FROM organizations"
	orgs, err := dbClient.DB.Query(orgSelectQuery)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to run query: %s", err))
	}
	defer orgs.Close()

	for orgs.Next() {
		if err := orgs.Scan(&orgID); err != nil {
			fmt.Println(err)
		}
	}
	s.orgID = orgID

	projCreationQuery := fmt.Sprintf("INSERT INTO projects (name, slug, org_id) VALUES ('Shield', 'shield proj', '%s') ON CONFLICT DO NOTHING", orgID)
	_, err = dbClient.DB.Query(projCreationQuery)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to run query: %s", err))
	}

	projSelectQuery := "SELECT id FROM projects"
	projs, err := dbClient.DB.Query(projSelectQuery)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to run query: %s", err))
	}
	defer projs.Close()

	for projs.Next() {
		if err := projs.Scan(&projID); err != nil {
			fmt.Println(err)
		}
	}
	s.projID = projID

	resourceBlobFS, err := blob.NewStore(ctx, appConfig.App.ResourcesConfigPath, appConfig.App.ResourcesConfigPathSecret)
	if err != nil {
		logger.Fatal("failed to create new blob store", err)
		return
	}

	resourceBlobRepository := blob.NewResourcesRepository(logger, resourceBlobFS)
	if err := resourceBlobRepository.InitCache(ctx, testbench.RuleCacheRefreshDelay); err != nil {
		logger.Fatal("failed to Initialise cache", err)
		return
	}

	deps, err := testbench.BuildAPIDependencies(ctx, logger, resourceBlobRepository, dbClient, spiceDBClient)
	if err != nil {
		logger.Fatal("failed to build API dependencies", err)
		return
	}

	// serving proxies
	_, _, err = testbench.ServeProxies(ctx, logger, appConfig.App.IdentityProxyHeader, appConfig.App.UserIDHeader, appConfig.Proxy, deps.ResourceService, deps.RelationService, deps.UserService, deps.ProjectService)
	if err != nil {
		logger.Fatal("failed to serve proxies", err)
		return
	}
}

func (s *EndToEndProxySmokeTestSuite) TearDownTest() {
	err := s.testBench.CleanUp()
	s.Require().NoError(err)
}

func (s *EndToEndProxySmokeTestSuite) TestProxyToEchoServer() {
	s.Run("1. should be able to proxy to an echo server", func() {
		url := fmt.Sprintf("http://localhost:%d/api/ping", s.proxyport)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		req.Header.Set(testbench.IdentityHeader, "john.doe@odpf.com")

		res, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)

		defer res.Body.Close()
		s.Assert().Equal(res.StatusCode, 200)
	})

	s.Run("2. resource created on echo server should persist in shieldDB", func() {
		buff := bytes.NewReader([]byte(`{"name": "test-proxy-resource", "type": "firehose"}`))
		url := fmt.Sprintf("http://localhost:%d/api/resource", s.proxyport)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		s.Require().NoError(err)
		req.Body = io.NopCloser(buff)

		req.Header.Set(testbench.IdentityHeader, "john.doe@odpf.com")
		req.Header.Set("X-Shield-Project", s.projID)
		req.Header.Set("X-Shield-Org", s.orgID)
		req.Header.Set("X-Shield-Name", "test-resource")
		req.Header.Set("X-Shield-Resource-Type", "firehose")

		resourceSelectQuery := "SELECT name FROM resources"
		resources, err := s.dbClient.DB.Query(resourceSelectQuery)
		s.Require().NoError(err)
		defer resources.Close()

		var resourceName = ""
		for resources.Next() {
			if err := resources.Scan(&resourceName); err != nil {
				s.Require().NoError(err)
			}
		}

		res, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)

		defer res.Body.Close()
		s.Assert().Equal(200, res.StatusCode)
		s.Assert().Equal("test-resource", resourceName)
	})
}

func TestEndToEndProxySmokeTestSuite(t *testing.T) {
	suite.Run(t, new(EndToEndProxySmokeTestSuite))
}
