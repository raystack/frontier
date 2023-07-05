package e2e_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/shield/internal/bootstrap/schema"
	"github.com/raystack/shield/pkg/server/consts"
	"github.com/raystack/shield/pkg/utils"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"google.golang.org/grpc/metadata"

	"github.com/raystack/shield/core/authenticate"
	"github.com/raystack/shield/pkg/server"

	"github.com/raystack/shield/config"
	"github.com/raystack/shield/pkg/logger"
	"github.com/raystack/shield/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type ServiceUsersRegressionTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
	apiPort   int
}

func (s *ServiceUsersRegressionTestSuite) SetupSuite() {
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
			IdentityProxyHeader: testbench.IdentityHeader,
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

	ctx := context.Background()
	s.Require().NoError(testbench.BootstrapUsers(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapOrganizations(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapProject(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapGroup(ctx, s.testBench.Client, testbench.OrgAdminEmail))
}

func (s *ServiceUsersRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *ServiceUsersRegressionTestSuite) TestServiceUserWithKey() {
	/*
		{
		  "alg": "HS256",
		  "typ": "JWT"
		}
		.
		{
		  "sub": "1234567890",
		  "name": "John Doe",
		  "iat": 1516239022
		}
		.
		HMACSHA256(password)
	*/
	sampleHMACJwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.kXSdJhhUKTJemgs8O0rfIJmUaxoSIDdClL_OPmaC7Eo"
	var svUserKey *shieldv1beta1.KeyCredential
	var svKeyToken []byte
	s.Run("1. create a service user in an org and generate a key", func() {
		ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: testbench.OrgAdminEmail,
		}))
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &shieldv1beta1.CreateServiceUserRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserKeyResp, err := s.testBench.Client.CreateServiceUserKey(ctxOrgAdminAuth, &shieldv1beta1.CreateServiceUserKeyRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserKeyResp)
		svUserKey = createServiceUserKeyResp.GetKey()

		// generate a token out of key
		rsaKey, err := jwk.ParseKey([]byte(svUserKey.GetPrivateKey()), jwk.WithPEM(true))
		s.Assert().NoError(err)
		s.Assert().NotNil(rsaKey)
		_ = rsaKey.Set(jwk.KeyIDKey, svUserKey.GetKid())

		svKeyToken, err = utils.BuildToken(rsaKey, "custom", svUserKey.GetPrincipalId(),
			time.Minute*5, nil)
		s.Assert().NoError(err)
		s.Assert().NotNil(svKeyToken)
	})
	s.Run("2. fetch current profile and ensure request is authenticated using service user key", func() {
		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Bearer " + string(svKeyToken),
		}))

		getCurrentUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithKey, &shieldv1beta1.GetCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)
		s.Assert().Equal(svUserKey.GetPrincipalId(), getCurrentUserResp.GetServiceuser().GetId())
	})
	s.Run("3. ensure request is authenticated using service user key with user-token header", func() {
		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			consts.UserTokenRequestKey: string(svKeyToken),
		}))

		getCurrentUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithKey, &shieldv1beta1.GetCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)
		s.Assert().Equal(svUserKey.GetPrincipalId(), getCurrentUserResp.GetServiceuser().GetId())
	})
	s.Run("4. passing invalid type of jwt should fail", func() {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			consts.UserTokenRequestKey: sampleHMACJwt,
		}))
		_, err := s.testBench.Client.GetCurrentUser(ctx, &shieldv1beta1.GetCurrentUserRequest{})
		s.Assert().Error(err)
	})
	s.Run("5. fetch current profile and pass additional headers via rest", func() {
		profileRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/v1beta1/users/self", s.apiPort), nil)
		s.Assert().NoError(err)
		profileRequest.Header.Set("Authorization", "Bearer 123")
		profileRequest.Header.Set(consts.UserTokenRequestKey, string(svKeyToken))

		currentUserResp, err := http.DefaultClient.Do(profileRequest)
		s.Assert().NoError(err)
		s.Assert().NotNil(currentUserResp.Body)
	})
	s.Run("6. service user should be able to create an organization with full permission", func() {
		_, err := s.testBench.Client.CreateOrganization(context.Background(), &shieldv1beta1.CreateOrganizationRequest{
			Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "org-su-test-1",
			},
		})
		s.Assert().Error(err)

		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Bearer " + string(svKeyToken),
		}))
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxWithKey, &shieldv1beta1.CreateOrganizationRequest{
			Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "org-su-test-1",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createOrgResp)

		checkPermResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &shieldv1beta1.CheckResourcePermissionRequest{
			ObjectId:        createOrgResp.Organization.Id,
			ObjectNamespace: "organization",
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermResp.Status)
	})
	s.Run("7. service user should be allowed to assign role", func() {
		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Bearer " + string(svKeyToken),
		}))
		existingOrg, err := s.testBench.Client.GetOrganization(ctxWithKey, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		// by default it should not have any permission
		checkPermResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &shieldv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID("organization", existingOrg.Organization.Id),
			Permission: schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermResp.Status)

		// assign role
		_, err = s.testBench.Client.CreatePolicy(ctxWithKey, &shieldv1beta1.CreatePolicyRequest{
			Body: &shieldv1beta1.PolicyRequestBody{
				RoleId:    "app_organization_manager",
				Resource:  schema.JoinNamespaceAndResourceID("organization", existingOrg.Organization.Id),
				Principal: schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, svUserKey.GetPrincipalId()),
			},
		})
		s.Assert().NoError(err)

		checkPermAfterResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &shieldv1beta1.CheckResourcePermissionRequest{
			ObjectId:        existingOrg.Organization.Id,
			ObjectNamespace: "organization",
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermAfterResp.Status)
	})
	s.Run("8. a service account should not have access to modify another service account", func() {
		ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: testbench.OrgAdminEmail,
		}))
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		// create another service user
		createServiceUser2Resp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &shieldv1beta1.CreateServiceUserRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUser2Resp)
		createServiceUser2KeyResp, err := s.testBench.Client.CreateServiceUserKey(ctxOrgAdminAuth, &shieldv1beta1.CreateServiceUserKeyRequest{
			Id: createServiceUser2Resp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUser2KeyResp)

		// generate a token out of key
		rsaKey, err := jwk.ParseKey([]byte(createServiceUser2KeyResp.GetKey().GetPrivateKey()), jwk.WithPEM(true))
		s.Assert().NoError(err)
		s.Assert().NotNil(rsaKey)
		_ = rsaKey.Set(jwk.KeyIDKey, createServiceUser2KeyResp.GetKey().GetKid())

		sv2KeyToken, err := utils.BuildToken(rsaKey, "custom", svUserKey.GetPrincipalId(),
			time.Minute*5, nil)
		s.Assert().NoError(err)
		s.Assert().NotNil(sv2KeyToken)
		ctxWithKey2 := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Bearer " + string(sv2KeyToken),
		}))

		// by default it should not have any permission
		checkPermAfterResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey2, &shieldv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, svUserKey.GetPrincipalId()),
			Permission: schema.ManagePermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermAfterResp.Status)
	})
}

func (s *ServiceUsersRegressionTestSuite) TestServiceUserWithSecret() {
	var svUserSecret *shieldv1beta1.SecretCredential
	var svKeySecret string
	s.Run("1. create a service user in an org and generate a secret", func() {
		ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: testbench.OrgAdminEmail,
		}))
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &shieldv1beta1.CreateServiceUserRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserSecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &shieldv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp)
		svUserSecret = createServiceUserSecretResp.GetSecret()
		svKeySecret = fmt.Sprintf("%s:%s", svUserSecret.GetId(),
			svUserSecret.GetSecret())
		svKeySecret = base64.StdEncoding.EncodeToString([]byte(svKeySecret))
	})
	s.Run("2. fetch current profile and ensure request is authenticated using service user key", func() {
		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Basic " + svKeySecret,
		}))

		getCurrentUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithKey, &shieldv1beta1.GetCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)
	})
	s.Run("3. passing invalid type of jwt should fail", func() {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Basic randomsecret",
		}))
		_, err := s.testBench.Client.GetCurrentUser(ctx, &shieldv1beta1.GetCurrentUserRequest{})
		s.Assert().Error(err)
	})
}

func TestEndToEndServiceUsersRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceUsersRegressionTestSuite))
}
