package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

type OnboardingRegressionTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
}

func (s *OnboardingRegressionTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	apiPort, err := testbench.GetFreePort()
	s.Require().NoError(err)
	grpcPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

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

func (s *OnboardingRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *OnboardingRegressionTestSuite) TestOnboardOrganizationWithUser() {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	var orgID = ""
	var projectID = ""
	var adminID = ""
	var newUserID = ""
	var newUserEmail = ""
	var resourceID = ""
	var roleToLookFor = "app_project_owner"
	var roleID = ""
	s.Run("1. a user should successfully create a new org and become its admin", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctx, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org acme 1",
				Name:  "org-acme-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("User description"),
					},
				},
			},
		})
		s.Assert().NoError(err)
		orgID = createOrgResp.Organization.Id

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctx, &frontierv1beta1.ListOrganizationUsersRequest{
			Id: orgID,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersResp.GetUsers()))
		s.Assert().Equal(testbench.OrgAdminEmail, orgUsersResp.GetUsers()[0].Email)
		adminID = orgUsersResp.Users[0].Id
	})
	s.Run("2. org admin should be able to create a new project", func() {
		projResponse, err := s.testBench.Client.CreateProject(ctx, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "new-project",
				OrgId: orgID,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Project description"),
					},
				},
			},
		})
		s.Assert().NoError(err)
		projectID = projResponse.Project.Id
	})
	s.Run("3. org admin should be able to create a new resource inside project", func() {
		createResourceResp, err := s.testBench.Client.CreateProjectResource(ctx, &frontierv1beta1.CreateProjectResourceRequest{
			ProjectId: projectID,
			Body: &frontierv1beta1.ResourceRequestBody{
				Name:      "res-1",
				Namespace: computeOrderNamespace,
				Principal: adminID,
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createResourceResp)
		resourceID = createResourceResp.Resource.Id
	})
	s.Run("4. org admin should have access to the resource created", func() {
		createResourceResp, err := s.testBench.Client.CheckResourcePermission(ctx, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        resourceID,
			ObjectNamespace: computeOrderNamespace,
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createResourceResp)
		s.Assert().True(createResourceResp.Status)
	})
	s.Run("5. list all predefined roles/permissions successfully", func() {
		listRolesResp, err := s.testBench.Client.ListRoles(ctx, &frontierv1beta1.ListRolesRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(listRolesResp)
		s.Assert().Len(listRolesResp.GetRoles(), 12)
		for _, r := range listRolesResp.GetRoles() {
			if r.Name == roleToLookFor {
				roleID = r.Id
			}
		}

		listPermissionsResp, err := s.testBench.Client.ListPermissions(ctx, &frontierv1beta1.ListPermissionsRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(listPermissionsResp)
		s.Assert().Len(listPermissionsResp.GetPermissions(), 28)
	})
	s.Run("6. creating role with bad body should fail", func() {
		_, err := s.testBench.Client.CreateOrganizationRole(ctx, &frontierv1beta1.CreateOrganizationRoleRequest{
			OrgId: orgID,
			Body: &frontierv1beta1.RoleRequestBody{
				Name:        "should-fail-without-permission",
				Permissions: nil,
			},
		})
		s.Assert().Error(err)

		_, err = s.testBench.Client.CreateOrganizationRole(ctx, &frontierv1beta1.CreateOrganizationRoleRequest{
			OrgId: orgID,
			Body: &frontierv1beta1.RoleRequestBody{
				Name:        "should-fail",
				Permissions: []string{"unknown-permission"},
			},
		})
		s.Assert().Error(err)
	})
	s.Run("7. list all custom roles successfully", func() {
		createRoleResp, err := s.testBench.Client.CreateOrganizationRole(ctx, &frontierv1beta1.CreateOrganizationRoleRequest{
			OrgId: orgID,
			Body: &frontierv1beta1.RoleRequestBody{
				Name:        "something_owner",
				Permissions: []string{"app_organization_get"},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createRoleResp)

		listRolesResp, err := s.testBench.Client.ListOrganizationRoles(ctx, &frontierv1beta1.ListOrganizationRolesRequest{
			OrgId: orgID,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(listRolesResp)
		s.Assert().Len(listRolesResp.GetRoles(), 1)
	})
	s.Run("8. create a new user and create a policy to make it a project manager", func() {
		createUserResp, err := s.testBench.Client.CreateUser(ctx, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user for org 1",
				Email: "user-1-for-org-1@raystack.org",
				Name:  "user_1_for_org_1_raystack_io",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createUserResp)
		newUserID = createUserResp.User.Id
		newUserEmail = createUserResp.User.Email

		// make user member of the org
		_, err = s.testBench.Client.AddOrganizationUsers(ctx, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      orgID,
			UserIds: []string{newUserID},
		})
		s.Assert().NoError(err)

		// assign new user as project admin
		createPolicyResp, err := s.testBench.Client.CreatePolicy(ctx, &frontierv1beta1.CreatePolicyRequest{Body: &frontierv1beta1.PolicyRequestBody{
			RoleId:    roleID,
			Resource:  schema.JoinNamespaceAndResourceID(schema.ProjectNamespace, projectID),
			Principal: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, newUserID),
		}})
		s.Assert().NoError(err)
		s.Assert().NotNil(createPolicyResp)
	})
	s.Run("9. new user should have access to that project it is managing and all of its resources", func() {
		userCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: newUserEmail,
		}))

		checkUpdateProjectResp, err := s.testBench.Client.CheckResourcePermission(userCtx, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        projectID,
			ObjectNamespace: schema.ProjectNamespace,
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(checkUpdateProjectResp)
		s.Assert().True(checkUpdateProjectResp.Status)

		// resources under the project
		checkUpdateResourceResp, err := s.testBench.Client.CheckResourcePermission(userCtx, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        resourceID,
			ObjectNamespace: computeOrderNamespace,
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(checkUpdateResourceResp)
		s.Assert().True(checkUpdateResourceResp.Status)
	})
	s.Run("10. new user should not have access to update the parent organization it is part of", func() {
		userCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: newUserEmail,
		}))
		checkUpdateOrgResp, err := s.testBench.Client.CheckResourcePermission(userCtx, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        orgID,
			ObjectNamespace: schema.OrganizationNamespace,
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(checkUpdateOrgResp)
		s.Assert().False(checkUpdateOrgResp.Status)
	})
	s.Run("11. a role assigned at org level for a resource should have access across projects", func() {
		createUserResp, err := s.testBench.Client.CreateUser(ctx, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user for org 1",
				Email: "user-2-for-org-1@raystack.org",
				Name:  "user_2_for_org_1_raystack_io",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createUserResp)

		// make user member of the org
		_, err = s.testBench.Client.AddOrganizationUsers(ctx, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      orgID,
			UserIds: []string{createUserResp.User.Id},
		})
		s.Assert().NoError(err)

		resourceViewerRole := ""
		listRolesResp, err := s.testBench.Client.ListRoles(ctx, &frontierv1beta1.ListRolesRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(listRolesResp)
		for _, r := range listRolesResp.GetRoles() {
			if r.Name == computeViewerRoleName {
				resourceViewerRole = r.Id
			}
		}

		// assign new user resource role across org
		createPolicyResp, err := s.testBench.Client.CreatePolicy(ctx, &frontierv1beta1.CreatePolicyRequest{Body: &frontierv1beta1.PolicyRequestBody{
			RoleId:    resourceViewerRole,
			Resource:  schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, orgID),
			Principal: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, createUserResp.User.Id),
		}})
		s.Assert().NoError(err)
		s.Assert().NotNil(createPolicyResp)

		// items under the org > project > resources
		userCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: createUserResp.User.Email,
		}))

		checkGetResourceResp, err := s.testBench.Client.CheckResourcePermission(userCtx, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        resourceID,
			ObjectNamespace: computeOrderNamespace,
			Permission:      schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(checkGetResourceResp)
		s.Assert().True(checkGetResourceResp.Status)

		checkUpdateResourceResp, err := s.testBench.Client.CheckResourcePermission(userCtx, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        resourceID,
			ObjectNamespace: computeOrderNamespace,
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(checkUpdateResourceResp)
		s.Assert().False(checkUpdateResourceResp.Status)
	})
}

func TestEndToEndOnboardingRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(OnboardingRegressionTestSuite))
}
