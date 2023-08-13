package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/golang/protobuf/jsonpb"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/pkg/server"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type AuthenticationRegressionTestSuite struct {
	suite.Suite
	testBench      *testbench.TestBench
	apiPort        int
	mockOIDCServer *mockoidc.MockOIDC
	callbackPort   int
}

func (s *AuthenticationRegressionTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	grpcPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	s.apiPort = apiPort
	callbackPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	s.callbackPort = callbackPort

	s.mockOIDCServer, err = mockoidc.Run()
	s.Require().NoError(err)

	// mock callback host

	appConfig := &config.Frontier{
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
					Issuer:  "frontier",
				},
				OIDCConfig: map[string]authenticate.OIDCConfig{
					"mock": {
						ClientID:     s.mockOIDCServer.Config().ClientID,
						ClientSecret: s.mockOIDCServer.Config().ClientSecret,
						IssuerUrl:    s.mockOIDCServer.Issuer(),
					},
				},
				CallbackURLs: []string{fmt.Sprintf("http://localhost:%d/callback", s.callbackPort)},
			},
		},
	}

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)
}

func (s *AuthenticationRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
	err = s.mockOIDCServer.Shutdown()
	s.Require().NoError(err)
}

func (s *AuthenticationRegressionTestSuite) TestUserSession() {
	ctx := context.Background()
	s.Run("1. return authenticate strategies of oidc", func() {
		authStrategyResp, err := s.testBench.Client.ListAuthStrategies(ctx, &frontierv1beta1.ListAuthStrategiesRequest{})
		s.Assert().NoError(err)
		s.Assert().Equal("mock", authStrategyResp.GetStrategies()[0].GetName())
	})
	s.Run("2. authenticate a user successfully using oidc and create a session via cookies", func() {
		// start registration flow
		authResp, err := s.testBench.Client.Authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
			StrategyName:    "mock",
			RedirectOnstart: false,
			ReturnTo:        "",
			Email:           mockoidc.DefaultUser().Email,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(authResp.Endpoint)

		// mock oidc code
		parsedEndpoint, err := url.Parse(authResp.Endpoint)
		s.Assert().NoError(err)
		mockAuth0Code := "012345"
		s.mockOIDCServer.QueueCode(mockAuth0Code)

		// prepare mock callback server
		callbackMux := http.NewServeMux()
		callbackMux.Handle("/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			state := r.URL.Query().Get("state")
			s.Assert().Equal(mockAuth0Code, code)
			s.Assert().Equal(parsedEndpoint.Query().Get("state"), state)
			w.WriteHeader(http.StatusOK)
		}))
		srv := &http.Server{
			Addr:    fmt.Sprintf("localhost:%d", s.callbackPort),
			Handler: callbackMux,
		}
		// clean up callback server
		defer srv.Shutdown(ctx)
		go func() {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				s.Assert().NoError(err)
			}
		}()

		// start session in oidc server
		endpointRes, err := http.Get(authResp.Endpoint)
		s.Assert().NoError(err)
		s.Assert().Equal(http.StatusOK, endpointRes.StatusCode)

		// callback to frontier and get valid cookies
		authCallbackFinalResp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1beta1/auth/callback?code=%s&state=%s",
			s.apiPort, mockAuth0Code, parsedEndpoint.Query().Get("state")))
		s.Assert().NoError(err)
		s.Assert().Equal(http.StatusOK, authCallbackFinalResp.StatusCode)
		s.Assert().Equal("sid", authCallbackFinalResp.Cookies()[0].Name)

		// verify if session is created
		getUserReq, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/v1beta1/users/self", s.apiPort), nil)
		getUserReq.AddCookie(authCallbackFinalResp.Cookies()[0])

		userResp, err := http.DefaultClient.Do(getUserReq)
		s.Assert().NoError(err)
		s.Assert().Equal(http.StatusOK, userResp.StatusCode)

		user := &frontierv1beta1.GetCurrentUserResponse{}
		s.Assert().NoError(jsonpb.Unmarshal(userResp.Body, user))
		s.Assert().Equal(mockoidc.DefaultUser().Email, user.GetUser().Email)
	})
}

func TestEndToEndAuthenticationRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationRegressionTestSuite))
}
