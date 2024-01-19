package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate/strategy"
	"github.com/raystack/frontier/pkg/mailer"
	"github.com/raystack/frontier/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/golang/protobuf/jsonpb"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/pkg/logger"
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

	smtpServer *smtpmock.Server
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

	s.smtpServer = smtpmock.New(smtpmock.ConfigurationAttr{
		LogToStdout:       false,
		LogServerActivity: false,
	})

	// To start smtpServer use Start() method
	if err := s.smtpServer.Start(); err != nil {
		s.Assert().NoError(err)
	}
	// Server's port will be assigned dynamically after smtpServer.Start()
	// for case when portNumber wasn't specified
	smtpHostAddress, smtpPortNumber := "127.0.0.1", s.smtpServer.PortNumber()

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
					Domain:         "",
					SameSite:       "lax",
					Validity:       time.Hour,
					Secure:         false,
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
				MailOTP: authenticate.MailOTPConfig{
					Subject: "{{.Otp}}",
					Body:    "{{.Otp}}",
				},
			},
			Mailer: mailer.Config{
				SMTPHost:      smtpHostAddress,
				SMTPPort:      smtpPortNumber,
				SMTPInsecure:  true,
				SMTPTLSPolicy: "none",
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
	err = s.smtpServer.Stop()
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
		s.Assert().NotNil(authResp.GetEndpoint())

		// mock oidc code
		parsedEndpoint, err := url.Parse(authResp.GetEndpoint())
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
		endpointRes, err := http.Get(authResp.GetEndpoint())
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
		s.Assert().Equal(mockoidc.DefaultUser().Email, user.GetUser().GetEmail())
	})
	s.Run("3. authenticate a user successfully using mailotp", func() {
		// start registration flow
		authResp, err := s.testBench.Client.Authenticate(ctx, &frontierv1beta1.AuthenticateRequest{
			StrategyName:    strategy.MailOTPAuthMethod,
			RedirectOnstart: false,
			ReturnTo:        "",
			Email:           mockoidc.DefaultUser().Email,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(authResp.GetState())

		// check if mail is sent
		messages := s.smtpServer.Messages()
		s.Assert().NotEmpty(messages)
		mailMsg := messages[0].MsgRequest()
		// extract mail headers
		mailParts := strings.Split(mailMsg, "\r\n")
		emailOTP := ""
		for _, part := range mailParts {
			if strings.HasPrefix(part, "Subject: ") {
				emailOTP = strings.TrimPrefix(part, "Subject: ")
			}
		}
		s.Assert().NotEmpty(emailOTP)

		// verify incorrect otp
		// extract grpc headers
		md := metadata.MD{}
		_, err = s.testBench.Client.AuthCallback(ctx, &frontierv1beta1.AuthCallbackRequest{
			StrategyName: strategy.MailOTPAuthMethod,
			Code:         "123456",
			State:        authResp.GetState(),
		}, grpc.Header(&md))
		s.Assert().Error(err)
		s.Assert().Empty(md["gateway-session-id"])

		// verify correct otp
		// extract grpc headers
		md = metadata.MD{}
		_, err = s.testBench.Client.AuthCallback(ctx, &frontierv1beta1.AuthCallbackRequest{
			StrategyName: strategy.MailOTPAuthMethod,
			Code:         emailOTP,
			State:        authResp.GetState(),
		}, grpc.Header(&md))
		s.Assert().NoError(err)
		s.Assert().NotEmpty(md["gateway-session-id"])
	})
}

func TestEndToEndAuthenticationRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationRegressionTestSuite))
}
