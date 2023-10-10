package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type BillingRegressionTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
}

func (s *BillingRegressionTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	grpcPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	appConfig := &config.Frontier{
		Log: logger.Config{
			Level:       "error",
			AuditEvents: "db",
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

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)

	ctx := context.Background()

	s.Require().NoError(testbench.BootstrapUsers(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapOrganizations(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapProject(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapGroup(ctx, s.testBench.Client, testbench.OrgAdminEmail))
}

func (s *BillingRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *BillingRegressionTestSuite) TestBillingCustomerAPI() {
}

func TestEndToEndBillingRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(BillingRegressionTestSuite))
}
