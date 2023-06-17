package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/raystack/shield/core/authenticate"
	"github.com/raystack/shield/pkg/server"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"google.golang.org/grpc/metadata"

	"github.com/raystack/shield/config"
	"github.com/raystack/shield/pkg/logger"
	"github.com/raystack/shield/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type PassthroughEmailRegressionTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
	apiPort   int
}

func (s *PassthroughEmailRegressionTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	grpcPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	s.apiPort = apiPort

	appConfig := &config.Shield{
		Log: logger.Config{
			Level: "error",
		},
		App: server.Config{
			Host: "localhost",
			Port: apiPort,
			GRPC: server.GRPCConfig{
				Port:           grpcPort,
				MaxRecvMsgSize: 2 << 10,
				MaxSendMsgSize: 2 << 10,
			},
			ResourcesConfigPath: path.Join(testDataPath, "resource"),
			Authentication: authenticate.Config{
				Session: authenticate.SessionConfig{
					HashSecretKey:  "hash-secret-should-be-32-chars--",
					BlockSecretKey: "hash-secret-should-be-32-chars--",
				},
				Token: authenticate.TokenConfig{
					RSAPath: "testdata/jwks.json",
					Issuer:  "shield",
				},
			},
		},
	}

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)
}

func (s *PassthroughEmailRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *PassthroughEmailRegressionTestSuite) TestWithoutHeader() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. passing no context header should fail", func() {
		ctx := context.Background()
		_, err := s.testBench.Client.GetCurrentUser(ctx, &shieldv1beta1.GetCurrentUserRequest{})
		s.Assert().Error(err)
	})
	s.Run("2. passing context with header should fail if not configured", func() {
		_, err := s.testBench.Client.GetCurrentUser(ctxOrgAdminAuth, &shieldv1beta1.GetCurrentUserRequest{})
		s.Assert().Error(err)
	})
	s.Run("3. passing context with header should fail if not configured", func() {
		profileRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/v1beta1/users/self", s.apiPort), nil)
		s.Assert().NoError(err)
		profileRequest.Header.Set(testbench.IdentityHeader, testbench.OrgAdminEmail)

		currentUserResp, err := http.DefaultClient.Do(profileRequest)
		s.Assert().NoError(err)
		s.Assert().Equal(http.StatusUnauthorized, currentUserResp.StatusCode)
	})
}

func TestEndToEndPassthroughEmailRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(PassthroughEmailRegressionTestSuite))
}
