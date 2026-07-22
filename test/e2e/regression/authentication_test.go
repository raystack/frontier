package e2e_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/raystack/frontier/core/authenticate/strategy"
	"github.com/raystack/frontier/pkg/mailer"
	"github.com/raystack/frontier/pkg/server"

	"github.com/lestrrat-go/jwx/v2/jwt"
	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/token"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/logger"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthenticationRegressionTestSuite struct {
	suite.Suite
	testBench      *testbench.TestBench
	connectPort    int
	mockOIDCServer *mockoidc.MockOIDC
	callbackPort   int

	smtpServer *smtpmock.Server
}

func (s *AuthenticationRegressionTestSuite) SetupSuite() {
	connectPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	s.connectPort = connectPort
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
			Connect: server.ConnectConfig{
				Port: connectPort,
			},
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
					RSAPath:  "testdata/jwks.json",
					Issuer:   "frontier",
					Validity: time.Hour,
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
			PAT: userpat.Config{Enabled: true, Prefix: "fpt", MaxPerUserPerOrg: 50, MaxLifetime: "8760h"},
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
		authStrategyResp, err := s.testBench.Client.ListAuthStrategies(ctx, connect.NewRequest(&frontierv1beta1.ListAuthStrategiesRequest{}))
		s.Assert().NoError(err)
		s.Assert().Equal("mock", authStrategyResp.Msg.GetStrategies()[0].GetName())
	})
	s.Run("2. authenticate a user successfully using oidc and create a session via cookies", func() {
		// start registration flow
		authResp, err := s.testBench.Client.Authenticate(ctx, connect.NewRequest(&frontierv1beta1.AuthenticateRequest{
			StrategyName:    "mock",
			RedirectOnstart: false,
			ReturnTo:        "",
			Email:           mockoidc.DefaultUser().Email,
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(authResp.Msg.GetEndpoint())

		// mock oidc code
		parsedEndpoint, err := url.Parse(authResp.Msg.GetEndpoint())
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
		defer srv.Shutdown(ctx) //nolint:errcheck
		go func() {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				s.Assert().NoError(err)
			}
		}()

		// start session in oidc server
		endpointRes, err := http.Get(authResp.Msg.GetEndpoint())
		s.Assert().NoError(err)
		s.Assert().Equal(http.StatusOK, endpointRes.StatusCode)

		// callback to frontier via ConnectRPC and get valid session cookie
		authCallbackResp, err := s.testBench.Client.AuthCallback(ctx, connect.NewRequest(&frontierv1beta1.AuthCallbackRequest{
			Code:  mockAuth0Code,
			State: parsedEndpoint.Query().Get("state"),
		}))
		s.Assert().NoError(err)
		setCookie := authCallbackResp.Header().Get("Set-Cookie")
		s.Assert().NotEmpty(setCookie)
		cookie := strings.SplitN(setCookie, ";", 2)[0]

		// verify if session is created
		ctxWithSession := testbench.ContextWithAuth(ctx, cookie)
		getUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithSession, connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{}))
		s.Assert().NoError(err)
		s.Assert().Equal(mockoidc.DefaultUser().Email, getUserResp.Msg.GetUser().GetEmail())
	})
	var mailOTPCtx context.Context
	s.Run("3. authenticate a user successfully using mailotp", func() {
		// start registration flow
		authResp, err := s.testBench.Client.Authenticate(ctx, connect.NewRequest(&frontierv1beta1.AuthenticateRequest{
			StrategyName:    strategy.MailOTPAuthMethod,
			RedirectOnstart: false,
			ReturnTo:        "",
			Email:           mockoidc.DefaultUser().Email,
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(authResp.Msg.GetState())

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
		// For the error case - we don't get response headers on error with connect
		_, err = s.testBench.Client.AuthCallback(ctx, connect.NewRequest(&frontierv1beta1.AuthCallbackRequest{
			StrategyName: strategy.MailOTPAuthMethod,
			Code:         "123456",
			State:        authResp.Msg.GetState(),
		}))
		s.Assert().Error(err)

		// verify correct otp
		// For the success case - get headers from connect response
		authCallbackResp, err := s.testBench.Client.AuthCallback(ctx, connect.NewRequest(&frontierv1beta1.AuthCallbackRequest{
			StrategyName: strategy.MailOTPAuthMethod,
			Code:         emailOTP,
			State:        authResp.Msg.GetState(),
		}))
		s.Assert().NoError(err)
		setCookie := authCallbackResp.Header().Get("Set-Cookie")
		s.Assert().NotEmpty(setCookie)
		cookie := strings.SplitN(setCookie, ";", 2)[0]

		// Create context with session cookie for subsequent calls
		ctxWithSession := testbench.ContextWithAuth(ctx, cookie)
		getUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithSession, connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{}))
		s.Assert().NoError(err)
		s.Assert().Equal(mockoidc.DefaultUser().Email, getUserResp.Msg.GetUser().GetEmail())
		mailOTPCtx = ctxWithSession
	})
	s.Run("4. authenticate a service user successfully using jwt", func() {
		// create organization via session
		createOrgResp, err := s.testBench.Client.CreateOrganization(mailOTPCtx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-svuser-1",
			},
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createOrgResp)

		// create service user and it's opaque token to authenticate using it
		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(mailOTPCtx, connect.NewRequest(&frontierv1beta1.CreateServiceUserRequest{
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserTokenResp, err := s.testBench.Client.CreateServiceUserToken(mailOTPCtx, connect.NewRequest(&frontierv1beta1.CreateServiceUserTokenRequest{
			Id:    createServiceUserResp.Msg.GetServiceuser().GetId(),
			OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserTokenResp)
		svUserToken := createServiceUserTokenResp.Msg.GetToken()
		svKeyToken := fmt.Sprintf("%s:%s", svUserToken.GetId(),
			svUserToken.GetToken())
		svKeyToken = base64.StdEncoding.EncodeToString([]byte(svKeyToken))

		ctxWithSVSecret := testbench.ContextWithHeaders(context.Background(), map[string]string{
			"Authorization": "Basic " + svKeyToken,
		})

		// verify sv user token works
		getCurrentUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithSVSecret, connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{}))
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)

		// generate jwt token using sv user authenticator
		jwtTokenResp, err := s.testBench.Client.AuthToken(ctxWithSVSecret, connect.NewRequest(&frontierv1beta1.AuthTokenRequest{
			GrantType:    "client_credentials",
			ClientId:     svUserToken.GetId(),
			ClientSecret: svUserToken.GetToken(),
			Assertion:    "",
		}))
		s.Assert().NoError(err)
		s.Assert().NotNil(jwtTokenResp)

		// verify if the jwt token works
		ctxWithJWT := testbench.ContextWithHeaders(context.Background(), map[string]string{
			"Authorization": "Bearer " + jwtTokenResp.Msg.GetAccessToken(),
		})
		getCurrentUserResp, err = s.testBench.Client.GetCurrentUser(ctxWithJWT, connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{}))
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)
	})
	s.Run("5. exchange a PAT for a JWT at AuthToken, and reject re-exchange of the minted token", func() {
		// org owned by the mailotp user, so a PAT can be scoped within it
		orgResp, err := s.testBench.Client.CreateOrganization(mailOTPCtx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{Name: "org-pat-authtoken"},
		}))
		s.Require().NoError(err)
		orgID := orgResp.Msg.GetOrganization().GetId()

		rolesResp, err := s.testBench.Client.ListRoles(mailOTPCtx, connect.NewRequest(&frontierv1beta1.ListRolesRequest{}))
		s.Require().NoError(err)
		var viewerRoleID string
		for _, r := range rolesResp.Msg.GetRoles() {
			if r.GetName() == schema.RoleOrganizationViewer {
				viewerRoleID = r.GetId()
				break
			}
		}
		s.Require().NotEmpty(viewerRoleID)

		patResp, err := s.testBench.Client.CreateCurrentUserPAT(mailOTPCtx, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
			Title: "authtoken-pat",
			OrgId: orgID,
			Scopes: []*frontierv1beta1.PATScope{
				{RoleId: viewerRoleID, ResourceType: schema.OrganizationNamespace},
			},
			ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		}))
		s.Require().NoError(err)
		patToken := patResp.Msg.GetPat().GetToken()
		s.Require().NotEmpty(patToken)

		// the gateways present the PAT as a bearer token; AuthToken exchanges it for a jwt
		ctxWithPAT := testbench.ContextWithHeaders(context.Background(), map[string]string{
			"Authorization": "Bearer " + patToken,
		})
		patTokenResp, err := s.testBench.Client.AuthToken(ctxWithPAT, connect.NewRequest(&frontierv1beta1.AuthTokenRequest{}))
		s.Require().NoError(err)
		mintedToken := patTokenResp.Msg.GetAccessToken()
		s.Require().NotEmpty(mintedToken)

		// the minted jwt identifies a PAT principal
		parsed, err := jwt.ParseInsecure([]byte(mintedToken))
		s.Require().NoError(err)
		subType, ok := parsed.Get(token.SubTypeClaimsKey)
		s.Assert().True(ok)
		s.Assert().Equal(schema.PATPrincipal, subType)

		// and the minted jwt authenticates
		ctxWithMinted := testbench.ContextWithHeaders(context.Background(), map[string]string{
			"Authorization": "Bearer " + mintedToken,
		})
		_, err = s.testBench.Client.GetCurrentUser(ctxWithMinted, connect.NewRequest(&frontierv1beta1.GetCurrentUserRequest{}))
		s.Assert().NoError(err)

		// the minted access token itself cannot be re-exchanged at AuthToken
		_, err = s.testBench.Client.AuthToken(ctxWithMinted, connect.NewRequest(&frontierv1beta1.AuthTokenRequest{}))
		s.Assert().Equal(connect.CodeUnauthenticated, connect.CodeOf(err))
	})
}

func TestEndToEndAuthenticationRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationRegressionTestSuite))
}
