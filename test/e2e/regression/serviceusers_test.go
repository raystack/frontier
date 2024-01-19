//go:build !race

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
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/server/consts"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/metadata"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/test/e2e/testbench"
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
			IdentityProxyHeader: testbench.IdentityHeader,
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
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
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
	var svUserKey *frontierv1beta1.KeyCredential
	var svKeyToken []byte
	s.Run("1. create a service user in an org and generate a key", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserKeyResp, err := s.testBench.Client.CreateServiceUserKey(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserKeyRequest{
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

		getCurrentUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithKey, &frontierv1beta1.GetCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)
		s.Assert().Equal(svUserKey.GetPrincipalId(), getCurrentUserResp.GetServiceuser().GetId())
	})
	s.Run("3. ensure request is authenticated using service user key with user-token header", func() {
		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			consts.UserTokenRequestKey: string(svKeyToken),
		}))

		getCurrentUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithKey, &frontierv1beta1.GetCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)
		s.Assert().Equal(svUserKey.GetPrincipalId(), getCurrentUserResp.GetServiceuser().GetId())
	})
	s.Run("4. passing invalid type of jwt should fail", func() {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			consts.UserTokenRequestKey: sampleHMACJwt,
		}))
		_, err := s.testBench.Client.GetCurrentUser(ctx, &frontierv1beta1.GetCurrentUserRequest{})
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
		_, err := s.testBench.Client.CreateOrganization(context.Background(), &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-su-test-1",
			},
		})
		s.Assert().Error(err)

		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Bearer " + string(svKeyToken),
		}))
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxWithKey, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-su-test-1",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createOrgResp)

		checkPermResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        createOrgResp.GetOrganization().GetId(),
			ObjectNamespace: "organization",
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermResp.GetStatus())
	})
	s.Run("7. service user should be allowed to assign role", func() {
		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Bearer " + string(svKeyToken),
		}))
		existingOrg, err := s.testBench.Client.GetOrganization(ctxWithKey, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		// by default, it should not have any permission
		checkPermResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID("organization", existingOrg.GetOrganization().GetId()),
			Permission: schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermResp.GetStatus())

		// assign role
		_, err = s.testBench.Client.CreatePolicy(ctxOrgAdminAuth, &frontierv1beta1.CreatePolicyRequest{
			Body: &frontierv1beta1.PolicyRequestBody{
				RoleId:    schema.RoleOrganizationManager,
				Resource:  schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
				Principal: schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, svUserKey.GetPrincipalId()),
			},
		})
		s.Assert().NoError(err)

		checkPermAfterResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        existingOrg.GetOrganization().GetId(),
			ObjectNamespace: "organization",
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermAfterResp.GetStatus())
	})
	s.Run("8. a service account should not have access to modify another service account", func() {
		ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: testbench.OrgAdminEmail,
		}))
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		// create another service user
		createServiceUser2Resp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUser2Resp)
		createServiceUser2KeyResp, err := s.testBench.Client.CreateServiceUserKey(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserKeyRequest{
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
		checkPermAfterResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey2, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
			Permission: schema.ServiceUserManagePermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermAfterResp.GetStatus())
	})
}

func (s *ServiceUsersRegressionTestSuite) TestServiceUserWithSecret() {
	var svUserSecret *frontierv1beta1.SecretCredential
	var svKeySecret string
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. create a service user in an org and generate a secret", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-sv-user-1",
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserSecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp)
		svUserSecret = createServiceUserSecretResp.GetSecret()
		svKeySecret = fmt.Sprintf("%s:%s", svUserSecret.GetId(),
			svUserSecret.GetSecret())
		svKeySecret = base64.StdEncoding.EncodeToString([]byte(svKeySecret))

		// list service user secrets
		listServiceUserSecretResp, err := s.testBench.Client.ListServiceUserSecrets(ctxOrgAdminAuth, &frontierv1beta1.ListServiceUserSecretsRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(listServiceUserSecretResp)
	})
	s.Run("2. fetch current profile and ensure request is authenticated using service user key", func() {
		ctxWithSecret := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Basic " + svKeySecret,
		}))

		getCurrentUserResp, err := s.testBench.Client.GetCurrentUser(ctxWithSecret, &frontierv1beta1.GetCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(getCurrentUserResp)
	})
	s.Run("3. passing invalid type of jwt should fail", func() {
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Basic randomsecret",
		}))
		_, err := s.testBench.Client.GetCurrentUser(ctx, &frontierv1beta1.GetCurrentUserRequest{})
		s.Assert().Error(err)
	})
	s.Run("4. service user should support organization roles", func() {
		testNamespace := "compute/machine"
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-sv-user-2",
			},
		})
		s.Assert().NoError(err)
		projectResp, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "project-sv-user-1",
				OrgId: createOrgResp.GetOrganization().GetId(),
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(projectResp)

		// create a service account
		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserSecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp)
		ctxWithSecret := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", createServiceUserSecretResp.GetSecret().GetId(),
				createServiceUserSecretResp.GetSecret().GetSecret()))),
		}))

		// create dummy permissions
		_, err = s.testBench.AdminClient.CreatePermission(ctxOrgAdminAuth, &frontierv1beta1.CreatePermissionRequest{
			Bodies: []*frontierv1beta1.PermissionRequestBody{
				{
					Key: "compute.machine.get",
				},
				{
					Key: "compute.machine.create",
				},
				{
					Key: "compute.machine.delete",
				},
			},
		})
		s.Assert().NoError(err)

		// create role without delete permission
		createdRoleResponse, err := s.testBench.AdminClient.CreateRole(ctxOrgAdminAuth, &frontierv1beta1.CreateRoleRequest{
			Body: &frontierv1beta1.RoleRequestBody{
				Name: "compute_machine_manager",
				Permissions: []string{
					"compute.machine.get",
					"compute.machine.create",
				},
			},
		})
		s.Assert().NoError(err)

		// create compute machine resource
		createResourceResp, err := s.testBench.Client.CreateProjectResource(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectResourceRequest{
			ProjectId: projectResp.GetProject().GetId(),
			Body: &frontierv1beta1.ResourceRequestBody{
				Name:      "resource1",
				Namespace: testNamespace,
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createResourceResp)

		// by default, it should not have any permission
		checkPermResp, err := s.testBench.Client.CheckResourcePermission(ctxWithSecret, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(testNamespace, createResourceResp.GetResource().GetId()),
			Permission: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermResp.GetStatus())

		// create policy binding
		_, err = s.testBench.Client.CreatePolicy(ctxOrgAdminAuth, &frontierv1beta1.CreatePolicyRequest{
			Body: &frontierv1beta1.PolicyRequestBody{
				RoleId:    createdRoleResponse.GetRole().GetId(),
				Resource:  schema.JoinNamespaceAndResourceID(schema.ProjectNamespace, projectResp.GetProject().GetId()),
				Principal: schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, createServiceUserResp.GetServiceuser().GetId()),
			},
		})
		s.Assert().NoError(err)

		// it will have get permission but not delete
		checkPermAfterResp, err := s.testBench.Client.CheckResourcePermission(ctxWithSecret, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(testNamespace, createResourceResp.GetResource().GetId()),
			Permission: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermAfterResp.GetStatus())
		checkPermAfterRespWithDelete, err := s.testBench.Client.CheckResourcePermission(ctxWithSecret, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(testNamespace, createResourceResp.GetResource().GetId()),
			Permission: schema.DeletePermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermAfterRespWithDelete.GetStatus())

		// update role in place to add delete permission
		_, err = s.testBench.AdminClient.UpdateRole(ctxOrgAdminAuth, &frontierv1beta1.UpdateRoleRequest{
			Id: createdRoleResponse.GetRole().GetId(),
			Body: &frontierv1beta1.RoleRequestBody{
				Name: "compute_machine_manager",
				Permissions: []string{
					"compute.machine.get",
					"compute.machine.create",
					"compute.machine.delete",
				},
			},
		})
		s.Assert().NoError(err)

		// should have permission now
		checkPermAfterDelete, err := s.testBench.Client.CheckResourcePermission(ctxWithSecret, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(testNamespace, createResourceResp.GetResource().GetId()),
			Permission: schema.DeletePermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermAfterDelete.GetStatus())

		// update role in place to remove delete permission again
		_, err = s.testBench.AdminClient.UpdateRole(ctxOrgAdminAuth, &frontierv1beta1.UpdateRoleRequest{
			Id: createdRoleResponse.GetRole().GetId(),
			Body: &frontierv1beta1.RoleRequestBody{
				Name: "compute_machine_manager",
				Permissions: []string{
					"compute.machine.get",
					"compute.machine.create",
				},
			},
		})
		s.Assert().NoError(err)

		// removing of permission should also reflect
		checkPermAfterDeleteRemoved, err := s.testBench.Client.BatchCheckPermission(ctxWithSecret, &frontierv1beta1.BatchCheckPermissionRequest{
			Bodies: []*frontierv1beta1.BatchCheckPermissionBody{
				{
					Resource:   schema.JoinNamespaceAndResourceID(testNamespace, createResourceResp.GetResource().GetId()),
					Permission: schema.DeletePermission,
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermAfterDeleteRemoved.GetPairs()[0].GetStatus())
	})
	s.Run("5. service user should be allowed to create resources for projects", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-svuser-1",
		})
		s.Assert().NoError(err)

		createSVUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
			Body: &frontierv1beta1.ServiceUserRequestBody{
				Title: "org-svuser-1-sv-user-1",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createSVUserResp)

		createServiceUserSecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id:    createSVUserResp.GetServiceuser().GetId(),
			Title: "org-svuser-1-sv-user-1-key-1",
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp)

		createdSVKey := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", createServiceUserSecretResp.GetSecret().GetId(),
			createServiceUserSecretResp.GetSecret().GetSecret())))
		ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Basic " + createdSVKey,
		}))

		// by default, it should not have any permission
		checkPermResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID("organization", existingOrg.GetOrganization().GetId()),
			Permission: schema.ProjectCreatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkPermResp.GetStatus())

		// assign role
		_, err = s.testBench.Client.CreatePolicy(ctxOrgAdminAuth, &frontierv1beta1.CreatePolicyRequest{
			Body: &frontierv1beta1.PolicyRequestBody{
				RoleId:    "app_project_manager",
				Resource:  schema.JoinNamespaceAndResourceID("organization", existingOrg.GetOrganization().GetId()),
				Principal: schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, createSVUserResp.GetServiceuser().GetId()),
			},
		})
		s.Assert().NoError(err)

		checkPermAfterResp, err := s.testBench.Client.CheckResourcePermission(ctxWithKey, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        existingOrg.GetOrganization().GetId(),
			ObjectNamespace: "organization",
			Permission:      schema.ProjectCreatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermAfterResp.GetStatus())

		// create a project
		createProjectResp, err := s.testBench.Client.CreateProject(ctxWithKey, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Title: "org-svuser-1-sv-user-1-project-1",
				OrgId: existingOrg.GetOrganization().GetId(),
				Name:  "proj1",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createProjectResp)

		// register a new service using custom permission
		createServiceResp, err := s.testBench.AdminClient.CreatePermission(ctxOrgAdminAuth, &frontierv1beta1.CreatePermissionRequest{
			Bodies: []*frontierv1beta1.PermissionRequestBody{
				{
					Key: "resource.workflow.run",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceResp)

		// check if service user has permission to create workflow
		checkPermAfterResp, err = s.testBench.Client.CheckResourcePermission(ctxWithKey, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   "project:" + createProjectResp.GetProject().GetId(),
			Permission: "resource_workflow_run",
		})
		s.Assert().NoError(err)
		s.Assert().True(checkPermAfterResp.GetStatus())
	})
	s.Run("6. listing serviceuser secrets only list it for that service user", func() {
		// first org
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-sv-user-6-1",
			},
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserSecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp)
		createServiceUserSecretResp2, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp2)

		// list service user secrets
		listServiceUserSecretResp, err := s.testBench.Client.ListServiceUserSecrets(ctxOrgAdminAuth, &frontierv1beta1.ListServiceUserSecretsRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Len(listServiceUserSecretResp.GetSecrets(), 2)

		// first org su key
		createdOrg1SVKey := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", createServiceUserSecretResp.GetSecret().GetId(),
			createServiceUserSecretResp.GetSecret().GetSecret())))
		ctxOrg1SVUWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"Authorization": "Basic " + createdOrg1SVKey,
		}))

		// second org
		createOrg2Resp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-sv-user-6-2",
			},
		})
		s.Assert().NoError(err)

		createServiceUserRespOrg2, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: createOrg2Resp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserRespOrg2)

		createServiceUser2SecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserRespOrg2.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUser2SecretResp)

		// list service user secrets should only get 1
		listServiceUser2SecretResp, err := s.testBench.Client.ListServiceUserSecrets(ctxOrgAdminAuth, &frontierv1beta1.ListServiceUserSecretsRequest{
			Id: createServiceUserRespOrg2.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Len(listServiceUser2SecretResp.GetSecrets(), 1)

		// org 1 should not list secrets for org 2
		_, err = s.testBench.Client.ListServiceUserSecrets(ctxOrg1SVUWithKey, &frontierv1beta1.ListServiceUserSecretsRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().Error(err)
	})
}

func (s *ServiceUsersRegressionTestSuite) TestServiceUserAsPlatformMember() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. create a service user in an org and make it platform superuser", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-sv-user-pl-1",
			},
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserSecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp)

		ctxWithKey := getSVUCtx(createServiceUserSecretResp.GetSecret())

		// before giving access, it should return error
		_, err = s.testBench.AdminClient.ListRelations(ctxWithKey, &frontierv1beta1.ListRelationsRequest{})
		s.Assert().Error(err)

		// make service user platform admin
		_, err = s.testBench.AdminClient.AddPlatformUser(ctxOrgAdminAuth, &frontierv1beta1.AddPlatformUserRequest{
			ServiceuserId: createServiceUserResp.GetServiceuser().GetId(),
			Relation:      schema.AdminRelationName,
		})
		s.Assert().NoError(err)

		// check if we have su permissions by listing relations
		listRelationsResp, err := s.testBench.AdminClient.ListRelations(ctxWithKey, &frontierv1beta1.ListRelationsRequest{
			Subject: schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, createServiceUserResp.GetServiceuser().GetId()),
			Object:  schema.JoinNamespaceAndResourceID(schema.PlatformNamespace, schema.PlatformID),
		})
		s.Assert().NoError(err)
		s.Assert().Len(listRelationsResp.GetRelations(), 1)
	})
	s.Run("2. create a service user in an org and make it platform member", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-sv-user-pl-2",
			},
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		createServiceUserSecretResp, err := s.testBench.Client.CreateServiceUserSecret(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserSecretRequest{
			Id: createServiceUserResp.GetServiceuser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserSecretResp)

		ctxWithKey := getSVUCtx(createServiceUserSecretResp.GetSecret())

		// make service user platform member
		_, err = s.testBench.AdminClient.AddPlatformUser(ctxOrgAdminAuth, &frontierv1beta1.AddPlatformUserRequest{
			ServiceuserId: createServiceUserResp.GetServiceuser().GetId(),
			Relation:      schema.MemberRelationName,
		})
		s.Assert().NoError(err)

		// check if we have su permissions by checking federated resource permission

		// this should be false as we are a member
		checkResp, err := s.testBench.AdminClient.CheckFederatedResourcePermission(ctxOrgAdminAuth,
			&frontierv1beta1.CheckFederatedResourcePermissionRequest{
				Subject:    schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, createServiceUserResp.GetServiceuser().GetId()),
				Resource:   schema.JoinNamespaceAndResourceID(schema.PlatformNamespace, schema.PlatformID),
				Permission: schema.PlatformSudoPermission,
			})
		s.Assert().NoError(err)
		s.Assert().False(checkResp.GetStatus())

		// should return true as we are a member
		checkResp, err = s.testBench.AdminClient.CheckFederatedResourcePermission(ctxWithKey,
			&frontierv1beta1.CheckFederatedResourcePermissionRequest{
				Subject:    schema.JoinNamespaceAndResourceID(schema.ServiceUserPrincipal, createServiceUserResp.GetServiceuser().GetId()),
				Resource:   schema.JoinNamespaceAndResourceID(schema.PlatformNamespace, schema.PlatformID),
				Permission: schema.PlatformCheckPermission,
			})
		s.Assert().NoError(err)
		s.Assert().True(checkResp.GetStatus())
	})
	s.Run("3. list users & service users of platform", func() {
		// create a service user
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "org-sv-user-pl-3",
			},
		})
		s.Assert().NoError(err)

		createServiceUserResp, err := s.testBench.Client.CreateServiceUser(ctxOrgAdminAuth, &frontierv1beta1.CreateServiceUserRequest{
			OrgId: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createServiceUserResp)

		// check if we have su permissions by listing users
		listUsersBeforeResp, err := s.testBench.AdminClient.ListPlatformUsers(ctxOrgAdminAuth, &frontierv1beta1.ListPlatformUsersRequest{})
		s.Assert().NoError(err)
		s.Assert().False(utils.ContainsFunc(listUsersBeforeResp.GetServiceusers(), func(user *frontierv1beta1.ServiceUser) bool {
			return user.GetId() == createServiceUserResp.GetServiceuser().GetId()
		}))

		// make service user platform member
		_, err = s.testBench.AdminClient.AddPlatformUser(ctxOrgAdminAuth, &frontierv1beta1.AddPlatformUserRequest{
			ServiceuserId: createServiceUserResp.GetServiceuser().GetId(),
			Relation:      schema.MemberRelationName,
		})
		s.Assert().NoError(err)

		// check if we have su permissions by listing users
		listUsersResp, err := s.testBench.AdminClient.ListPlatformUsers(ctxOrgAdminAuth, &frontierv1beta1.ListPlatformUsersRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(listUsersResp)
		s.Assert().Len(listUsersResp.GetUsers(), 1)
		s.Assert().True(utils.ContainsFunc(listUsersResp.GetServiceusers(), func(user *frontierv1beta1.ServiceUser) bool {
			return user.GetId() == createServiceUserResp.GetServiceuser().GetId()
		}))
	})
}

func TestEndToEndServiceUsersRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceUsersRegressionTestSuite))
}

func getSVUCtx(cred *frontierv1beta1.SecretCredential) context.Context {
	ctxWithKey := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cred.GetId(),
			cred.GetSecret()))),
	}))
	return ctxWithKey
}
