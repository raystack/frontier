package smoke_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/pkg/logger"
	"github.com/odpf/shield/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

const (
	fixturesDir = "testdata"
)

type PingSmokeTestSuite struct {
	suite.Suite

	close     func() error
	apiPort   int
	proxyPort int
}

func (s *PingSmokeTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Require().Nil(err)
	s.apiPort = apiPort
	grpcPort, err := testbench.GetFreePort()
	s.Require().Nil(err)
	proxyPort, err := testbench.GetFreePort()
	s.Require().Nil(err)
	s.proxyPort = proxyPort

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
					RulesPath: path.Join(testDataPath, "rule"),
				},
			},
		},
	}

	tb, err := testbench.Init(appConfig)
	s.Require().Nil(err)

	s.close = func() error {
		return tb.Close()
	}
}

func (s *PingSmokeTestSuite) TearDownSuite() {
	err := s.close()
	s.Assert().NoError(err)
}

func (s *PingSmokeTestSuite) TestPing() {
	s.Run("should be able to ping shield", func() {
		url := fmt.Sprintf("http://localhost:%d/admin/ping", s.apiPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		s.Require().NoError(err)

		res, err := http.DefaultClient.Do(req)
		s.Require().NoError(err)

		defer res.Body.Close()
		s.Assert().Equal(200, res.StatusCode)
		text, err := io.ReadAll(res.Body)
		s.Require().NoError(err)

		s.Assert().Equal("{\"status\":\"SERVING\"}\n", string(text))
	})
	s.Run("should be able to ping proxy", func() {
		url := fmt.Sprintf("http://localhost:%d/ping", s.proxyPort)
		res, err := http.Head(url)
		s.Require().NoError(err)

		s.Assert().Equal(200, res.StatusCode)
	})
}

func TestPingSmokeTest(t *testing.T) {
	suite.Run(t, new(PingSmokeTestSuite))
}
