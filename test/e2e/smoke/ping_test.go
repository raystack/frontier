package smoke_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/test/e2e/testbench"
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
	s.Assert().NoError(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Assert().NoError(err)
	s.apiPort = apiPort
	grpcPort, err := testbench.GetFreePort()
	s.Assert().NoError(err)
	proxyPort, err := testbench.GetFreePort()
	s.Assert().NoError(err)
	s.proxyPort = proxyPort

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
			ResourcesConfigPath: path.Join(testDataPath, "resource"),
		},
	}

	tb, err := testbench.Init(appConfig)
	s.Assert().NoError(err)

	s.close = func() error {
		return tb.Close()
	}
}

func (s *PingSmokeTestSuite) TearDownSuite() {
	err := s.close()
	s.Assert().NoError(err)
}

func (s *PingSmokeTestSuite) TestPing() {
	s.Run("should be able to ping frontier", func() {
		url := fmt.Sprintf("http://localhost:%d/ping", s.apiPort)
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
}

func TestPingSmokeTest(t *testing.T) {
	suite.Run(t, new(PingSmokeTestSuite))
}
