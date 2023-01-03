package e2e_test

import (
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
}

func (s *EndToEndProxySmokeTestSuite) SetupTest() {
	wd, err := os.Getwd()
	s.Require().Nil(err)

	parent := filepath.Dir(wd)
	testDataPath := parent + "/../../"
	fmt.Printf("file://%s%s", testDataPath, "resources_config")

	proxyPort, err := testbench.GetFreePort()
	s.Require().Nil(err)
	fmt.Printf("proxyPort: %v\n", proxyPort)

	s.proxyport = proxyPort

	apiPort, err := testbench.GetFreePort()
	s.Require().Nil(err)
	fmt.Printf("apiPort: %v\n", apiPort)

	appConfig := &config.Shield{
		Log: logger.Config{
			Level: "fatal",
		},
		App: server.Config{
			Port:                      apiPort,
			IdentityProxyHeader:       testbench.IdentityHeader,
			ResourcesConfigPath:       "file:///Users/ishanarya/Desktop/Work/shield/resources_config",
			ResourcesConfigPathSecret: "",
			RulesPath:                 "file:///Users/ishanarya/Desktop/Work/shield/rules",
		},
		Proxy: proxy.ServicesConfig{
			Services: []proxy.Config{
				{
					Name:      "base",
					Host:      "localhost",
					Port:      proxyPort,
					RulesPath: "file:///Users/ishanarya/Desktop/Work/shield/rules",
				},
			},
		},
		DB: db.Config{
			Driver:              "postgres",
			URL:                 "postgres://shield:12345@localhost:5432/shield-docs?sslmode=disable",
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

	fmt.Printf("appConfig.App.ResourcesConfigPath: %v\n", appConfig.App.ResourcesConfigPath)
	fmt.Printf("appConfig.App.RulesPath: %v\n", appConfig.App.RulesPath)

	spiceDBClient, err := spicedb.New(appConfig.SpiceDB)
	if err != nil {
		logger.Fatal("failed to create spiceDB client", err)
		return
	}

	dbClient, err := testbench.SetupDB(appConfig.DB)
	if err != nil {
		logger.Fatal("failed to setup database")
		return
	}
	/*defer func() {
		logger.Info("cleaning up db")
		dbClient.Close()
	}()*/

	query := "INSERT INTO users (name,email) VALUES ('Ishan', 'ishan.arya@gojek.com') ON CONFLICT DO NOTHING"

	_, err = dbClient.DB.Query(query)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	resourceBlobFS, err := blob.NewStore(ctx, appConfig.App.ResourcesConfigPath, appConfig.App.ResourcesConfigPathSecret)
	if err != nil {
		fmt.Printf("f to create new blob store: %v\n", err)
		logger.Fatal("failed to create new blob store", err)
		return
	}

	resourceBlobRepository := blob.NewResourcesRepository(logger, resourceBlobFS)
	if err := resourceBlobRepository.InitCache(ctx, testbench.RuleCacheRefreshDelay); err != nil {
		fmt.Printf("failed to initialize cache: %v\n", err)
		logger.Fatal("failed to Initialise cache", err)
		return
	}
	/*defer func() {
		logger.Info("cleaning up resource blob")
		defer resourceBlobRepository.Close()
	}()*/

	deps, err := testbench.BuildAPIDependencies(ctx, logger, resourceBlobRepository, dbClient, spiceDBClient)
	if err != nil {
		fmt.Printf("f to build API dep: %v\n", err)
		logger.Fatal("failed to build API dependencies", err)
		return
	}

	// serving proxies
	_, _, err = testbench.ServeProxies(ctx, logger, appConfig.App.IdentityProxyHeader, appConfig.App.UserIDHeader, appConfig.Proxy, deps.ResourceService, deps.RelationService, deps.UserService, deps.ProjectService)
	if err != nil {
		fmt.Printf("failed to serve proxies: %v\n", err)
		logger.Fatal("failed to serve proxies", err)
		return
	}
	fmt.Printf("\"proxy served\": %v\n", "proxy served")
	/*defer func() {
		// clean up stage
		logger.Info("cleaning up rules proxy blob")
		for _, f := range cbs {
			if err := f(); err != nil {
				logger.Warn("error occurred during shutdown rules proxy blob storages", "err", err)
			}
		}

		logger.Info("cleaning up proxies")
		for _, f := range cps {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*20)
			if err := f(shutdownCtx); err != nil {
				shutdownCancel()
				logger.Warn("error occurred during shutdown proxies", "err", err)
				continue
			}
			shutdownCancel()
		}
	}()*/
}

func (s *EndToEndProxySmokeTestSuite) TearDownTest() {
	err := s.testBench.CleanUp()
	s.Require().NoError(err)
}

func (s *EndToEndProxySmokeTestSuite) TestSmokeTestAdmin() {
	// sleep needed to compensate transaction done in spice db

	s.Run("1. org admin could create a new team", func() {
		// check permission

		url := fmt.Sprintf("http://localhost:%d/api/ping", s.proxyport)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		req.Header.Set(testbench.IdentityHeader, "ishan.arya@gojek.com")
		//req.Header.Set("Accept", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}

		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
		if err != nil {
			fmt.Print(err)
		}

		fmt.Printf("res.StatusCode: %v\n", res.StatusCode)
		fmt.Println("body", string(b))

		s.Assert().Equal("", "")
	})
}

// TODO proxy test
// - member who does not belong to any team cannot access resources
// - could not access resource in different org
// - Calling create upstream resource should create a resource in shield DB
// - Team member can access the created resource created by the same team
// - Team admin can access the created resource created by the same team
// - Org admin who is not team member or admin or creator can access the created resource in the same org
// - Project admin who is not team member or admin or creator can access the created resource in the same project

func TestEndToEndProxySmokeTestSuite(t *testing.T) {
	suite.Run(t, new(EndToEndProxySmokeTestSuite))
}
