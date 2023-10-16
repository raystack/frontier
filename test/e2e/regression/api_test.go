package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/raystack/frontier/pkg/utils"

	"golang.org/x/exp/slices"

	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/preference"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	fixturesDir           = "testdata"
	computeOrderNamespace = "compute/order"
	computeViewerRoleName = "compute_order_viewer"
)

type APIRegressionTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
}

func (s *APIRegressionTestSuite) SetupSuite() {
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

func (s *APIRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *APIRegressionTestSuite) TestOrganizationAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	s.Run("1. a user should successfully create a new org and become its admin", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org acme 1",
				Name:  "org-acme-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersResp.GetUsers()))
		s.Assert().Equal(testbench.OrgAdminEmail, orgUsersResp.GetUsers()[0].Email)
	})
	s.Run("2. user attached to an org as member should have no basic permission other than membership", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org acme 2",
				Name:  "org-acme-2",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().NoError(err)

		userResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
			Title: "acme 2 member",
			Email: "acme-member@raystack.org",
			Name:  "acme_2_member",
		}})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      createOrgResp.GetOrganization().GetId(),
			UserIds: []string{userResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Contains(utils.Map(orgUsersResp.GetUsers(), func(u *frontierv1beta1.User) string {
			return u.GetId()
		}), userResp.GetUser().GetId())
	})
	s.Run("3. deleting an org should delete all of its internal relations/projects/groups/resources", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org acme 3",
				Name:  "org-acme-3",
			},
		})
		s.Assert().NoError(err)

		createUserResponse, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
			Title: "acme 3 member 1",
			Email: "acme-member-1@raystack.org",
			Name:  "acme_3_member_1",
		}})
		s.Assert().NoError(err)

		// attach user to org
		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      createOrgResp.GetOrganization().GetId(),
			UserIds: []string{createUserResponse.GetUser().GetId()},
		})
		s.Assert().NoError(err)

		createProjResp, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "org-3-proj-1",
				OrgId: createOrgResp.GetOrganization().GetId(),
			},
		})
		s.Assert().NoError(err)

		createResourceResp, err := s.testBench.Client.CreateProjectResource(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectResourceRequest{
			ProjectId: createProjResp.GetProject().GetId(),
			Body: &frontierv1beta1.ResourceRequestBody{
				Name:      "res-1",
				Namespace: computeOrderNamespace,
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createResourceResp)

		// check users
		listUsersBeforeDelete, err := s.testBench.Client.ListUsers(ctxOrgAdminAuth, &frontierv1beta1.ListUsersRequest{
			OrgId: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Contains(utils.Map(listUsersBeforeDelete.GetUsers(), func(u *frontierv1beta1.User) string {
			return u.GetId()
		}), createUserResponse.GetUser().GetId())

		// delete org and all its items
		_, err = s.testBench.Client.DeleteOrganization(ctxOrgAdminAuth, &frontierv1beta1.DeleteOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)

		// check org
		_, err = s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		s.Assert().NotNil(err)

		// check project
		_, err = s.testBench.Client.GetProject(ctxOrgAdminAuth, &frontierv1beta1.GetProjectRequest{
			Id: createProjResp.Project.Id,
		})
		s.Assert().NotNil(err)

		// check resource
		_, err = s.testBench.Client.GetProjectResource(ctxOrgAdminAuth, &frontierv1beta1.GetProjectResourceRequest{
			Id: createResourceResp.Resource.Id,
		})
		s.Assert().NotNil(err)

		// check user relations
		listUsersAfterDelete, err := s.testBench.Client.ListUsers(ctxOrgAdminAuth, &frontierv1beta1.ListUsersRequest{
			OrgId: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listUsersAfterDelete.Users))
	})
	s.Run("4. removing a user from org should remove its access to all org resources", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org acme 4",
				Name:  "org-acme-4",
			},
		})
		s.Assert().NoError(err)

		createUserResponse, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
			Title: "acme 4 member 1",
			Email: "acme-4-member-1@raystack.org",
			Name:  "acme_4_member_1",
		}})
		s.Assert().NoError(err)

		// attach user to org
		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      createOrgResp.GetOrganization().GetId(),
			UserIds: []string{createUserResponse.GetUser().GetId()},
		})
		s.Assert().NoError(err)

		// check users
		listUsersBeforeDelete, err := s.testBench.Client.ListUsers(ctxOrgAdminAuth, &frontierv1beta1.ListUsersRequest{
			OrgId: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Contains(utils.Map(listUsersBeforeDelete.GetUsers(), func(u *frontierv1beta1.User) string {
			return u.GetId()
		}), createUserResponse.GetUser().GetId())

		// remove user from org
		_, err = s.testBench.Client.RemoveOrganizationUser(ctxOrgAdminAuth, &frontierv1beta1.RemoveOrganizationUserRequest{
			Id:     createOrgResp.GetOrganization().GetId(),
			UserId: createUserResponse.GetUser().GetId(),
		})
		s.Assert().NoError(err)

		// check users
		listUsersAfterDelete, err := s.testBench.Client.ListUsers(ctxOrgAdminAuth, &frontierv1beta1.ListUsersRequest{
			OrgId: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().NotContains(utils.Map(listUsersAfterDelete.GetUsers(), func(u *frontierv1beta1.User) string {
			return u.GetId()
		}), createUserResponse.GetUser().GetId())
	})
	s.Run("5. a user should successfully create a new org and list it even if it's disabled", func() {
		// enable disable_org_on_create preference
		disabledOrgs, err := s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
			Preferences: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.PlatformDisableOrgsOnCreate,
					Value: "true",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(disabledOrgs)

		ctxOrgUserAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: "normaluser@acme.org",
		}))
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgUserAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org acme 5",
				Name:  "org-acme-5",
			},
		})
		s.Assert().NoError(err)
		s.Assert().Equal(organization.Disabled.String(), createOrgResp.GetOrganization().GetState())

		// should not list org if it's disabled by default
		userEnabledOrgs, err := s.testBench.Client.ListOrganizationsByCurrentUser(ctxOrgUserAuth, &frontierv1beta1.ListOrganizationsByCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().False(slices.Contains(utils.Map(userEnabledOrgs.GetOrganizations(), func(o *frontierv1beta1.Organization) string {
			return o.GetName()
		}), createOrgResp.GetOrganization().GetName()))

		// should list org even if it's disabled
		userDisabledOrgs, err := s.testBench.Client.ListOrganizationsByCurrentUser(ctxOrgUserAuth, &frontierv1beta1.ListOrganizationsByCurrentUserRequest{
			State: organization.Disabled.String(),
		})
		s.Assert().NoError(err)
		s.Assert().True(slices.Contains(utils.Map(userDisabledOrgs.GetOrganizations(), func(o *frontierv1beta1.Organization) string {
			return o.GetName()
		}), createOrgResp.GetOrganization().GetName()))

		// reset disable_org_on_create preference
		_, err = s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
			Preferences: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.PlatformDisableOrgsOnCreate,
					Value: "false",
				},
			},
		})
		s.Assert().NoError(err)
	})
}

func (s *APIRegressionTestSuite) TestProjectAPI() {
	var newProject *frontierv1beta1.Project

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	// get my org
	res, err := s.testBench.Client.ListOrganizations(context.Background(), &frontierv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(res.GetOrganizations()), 0)
	myOrg := res.GetOrganizations()[0]

	s.Run("1. org admin create a new project successfully", func() {
		_, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "new-project",
				OrgId: myOrg.GetId(),
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().NoError(err)
	})

	s.Run("2. org admin create a new project with empty name should return invalid argument", func() {
		_, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new project with wrong org id should return not found", func() {
		_, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "new-project",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.NotFound, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new project with same name and org-id should conflict", func() {
		res, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "new-project-duplicate",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newProject = res.GetProject()
		s.Assert().NotNil(newProject)

		_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "new-project-duplicate",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. org admin update a new project with empty body should return invalid argument", func() {
		_, err := s.testBench.Client.UpdateProject(ctxOrgAdminAuth, &frontierv1beta1.UpdateProjectRequest{
			Id:   newProject.GetId(),
			Body: nil,
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("6. org admin update a new project with using project name instead of id should work", func() {
		_, err := s.testBench.Client.UpdateProject(ctxOrgAdminAuth, &frontierv1beta1.UpdateProjectRequest{
			Id: "new-project",
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "new-project",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
	})
	s.Run("7. list all projects attached/filtered to an org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-project-1",
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "org-project-1-p1",
				OrgId: existingOrg.Organization.GetId(),
			},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "org-project-1-p2",
				OrgId: existingOrg.Organization.GetId(),
			},
		})
		s.Assert().NoError(err)

		listResp, err := s.testBench.Client.ListOrganizationProjects(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationProjectsRequest{
			Id: existingOrg.Organization.GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(listResp.Projects))
	})
	s.Run("8. list all users who have access to a project", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-project-1",
		})
		s.Assert().NoError(err)

		createProjectP1Response, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "org-project-2-p1",
				OrgId: existingOrg.Organization.GetId(),
			},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "org-project-2-p2",
				OrgId: existingOrg.Organization.GetId(),
			},
		})
		s.Assert().NoError(err)

		// default
		listProjUsersRespBeforeAccess, err := s.testBench.Client.ListProjectUsers(ctxOrgAdminAuth, &frontierv1beta1.ListProjectUsersRequest{
			Id: "org-project-2-p1",
		})
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listProjUsersRespBeforeAccess.Users))

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Email: "user-for-org-project-2-p1@raystack.org",
				Name:  "user-for-org-project-2-p1",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createUserResp)

		// add user to project
		_, err = s.testBench.Client.CreatePolicy(ctxOrgAdminAuth, &frontierv1beta1.CreatePolicyRequest{
			Body: &frontierv1beta1.PolicyRequestBody{
				RoleId:    schema.RoleProjectViewer,
				Resource:  schema.JoinNamespaceAndResourceID(schema.ProjectNamespace, createProjectP1Response.GetProject().GetId()),
				Principal: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, createUserResp.GetUser().GetId()),
			},
		})
		s.Assert().NoError(err)

		listProjUsersResp, err := s.testBench.Client.ListProjectUsers(ctxOrgAdminAuth, &frontierv1beta1.ListProjectUsersRequest{
			Id: "org-project-2-p1",
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(listProjUsersResp.Users))

		listProjCurrentUsersResp, err := s.testBench.Client.ListProjectsByCurrentUser(ctxOrgAdminAuth, &frontierv1beta1.ListProjectsByCurrentUserRequest{})
		s.Assert().NoError(err)
		s.Assert().True(slices.ContainsFunc[[]*frontierv1beta1.Project](listProjCurrentUsersResp.GetProjects(), func(p *frontierv1beta1.Project) bool {
			return p.Name == "org-project-2-p1"
		}))
		s.Assert().True(slices.ContainsFunc[[]*frontierv1beta1.Project](listProjCurrentUsersResp.GetProjects(), func(p *frontierv1beta1.Project) bool {
			return p.Name == "org-project-2-p2"
		}))
	})
}

func (s *APIRegressionTestSuite) TestGroupAPI() {
	var newGroup *frontierv1beta1.Group
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	// get my org
	res, err := s.testBench.Client.ListOrganizations(context.Background(), &frontierv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(res.GetOrganizations()), 0)
	myOrg := res.GetOrganizations()[0]

	s.Run("1. org admin create a new team with empty auth email should return unauthenticated error", func() {
		_, err := s.testBench.Client.CreateGroup(context.Background(), &frontierv1beta1.CreateGroupRequest{
			OrgId: myOrg.GetId(),
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "group-basic-1",
			},
		})
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	s.Run("2. org admin create a new team with empty name should return invalid argument", func() {
		_, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &frontierv1beta1.CreateGroupRequest{
			OrgId: myOrg.GetId(),
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new team with wrong org id should return not found", func() {
		_, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &frontierv1beta1.CreateGroupRequest{
			OrgId: "not-uuid",
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "new-group",
			},
		})
		s.Assert().Equal(codes.NotFound, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new team with same name and org-id should conflict", func() {
		res, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &frontierv1beta1.CreateGroupRequest{
			OrgId: myOrg.GetId(),
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "new-group",
			},
		})
		s.Assert().NoError(err)
		newGroup = res.GetGroup()
		s.Assert().NotNil(newGroup)

		_, err = s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &frontierv1beta1.CreateGroupRequest{
			OrgId: myOrg.GetId(),
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "new-group",
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. group admin update a new team with empty body should return invalid argument", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxOrgAdminAuth, &frontierv1beta1.UpdateGroupRequest{
			Id:   newGroup.GetId(),
			Body: nil,
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("6. group admin update a new team with empty group id should return invalid arg", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxOrgAdminAuth, &frontierv1beta1.UpdateGroupRequest{
			Id:    "",
			OrgId: myOrg.GetId(),
			Body:  &frontierv1beta1.GroupRequestBody{},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("7. group admin update a new team without group id should fail", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxOrgAdminAuth, &frontierv1beta1.UpdateGroupRequest{
			OrgId: myOrg.GetId(),
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "org1-group1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Error(err)
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("8. create a group and add new member to it successfully", func() {
		createGroupResp, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &frontierv1beta1.CreateGroupRequest{
			OrgId: myOrg.GetId(),
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "group-8",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createGroupResp.GetGroup())

		listGroupUsers, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.ListGroupUsersRequest{
			Id:    createGroupResp.GetGroup().GetId(),
			OrgId: createGroupResp.GetGroup().GetOrgId(),
		})
		s.Assert().NoError(err)
		// only admin as member
		s.Assert().Len(listGroupUsers.GetUsers(), 1)

		// add a user
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Email: "user-for-group@raystack.org",
				Name:  "user-for-group",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createUserResp)
		addMemberResp, err := s.testBench.Client.AddGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.AddGroupUsersRequest{
			Id:      createGroupResp.GetGroup().GetId(),
			OrgId:   createGroupResp.GetGroup().GetOrgId(),
			UserIds: []string{createUserResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(addMemberResp)

		listGroupUsersAfterUser, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.ListGroupUsersRequest{
			Id:    createGroupResp.GetGroup().GetId(),
			OrgId: createGroupResp.GetGroup().GetOrgId(),
		})
		s.Assert().NoError(err)
		s.Assert().Len(listGroupUsersAfterUser.GetUsers(), 2)
	})
}

func (s *APIRegressionTestSuite) TestUserAPI() {
	var newUser *frontierv1beta1.User

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	s.Run("1. org admin create a new user with empty auth email should return unauthenticated error", func() {
		_, err := s.testBench.Client.CreateUser(context.Background(), &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "new-user-a@raystack.org",
				Name:  "new_user_123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	s.Run("2. org admin create a new user with unparsable metadata should return invalid argument error", func() {
		_, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "new-user-a@raystack.org",
				Name:  "new_user_123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewNullValue(),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new user with empty email should return invalid argument error", func() {
		_, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "",
				Name:  "new_user_123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new user with same email should return conflict error", func() {
		res, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "new-user-a@raystack.org",
				Name:  "new-user-123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().NoError(err)
		newUser = res.GetUser()

		_, err = s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "new-user-a@raystack.org",
				Name:  "new_user_123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. org admin update user with conflicted detail should not update the email and return nil error", func() {
		ExpectedEmail := "new-user-a@raystack.org"
		res, err := s.testBench.Client.UpdateUser(ctxOrgAdminAuth, &frontierv1beta1.UpdateUserRequest{
			Id: newUser.GetId(),
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "admin1-group2-org1@raystack.org",
				Name:  "new_user_123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(ExpectedEmail, res.User.Email)
		s.Assert().NoError(err)
	})

	ctxCurrentUser := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: newUser.GetEmail(),
	}))

	s.Run("6. update current user with empty email should return invalid argument error", func() {
		_, err := s.testBench.Client.UpdateCurrentUser(ctxCurrentUser, &frontierv1beta1.UpdateCurrentUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "",
				Name:  "new_user_123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})
	s.Run("7. update current user with different email in header and body should return invalid argument error", func() {
		_, err := s.testBench.Client.UpdateCurrentUser(ctxCurrentUser, &frontierv1beta1.UpdateCurrentUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user a",
				Email: "admin1-group1-org1@raystack.org",
				Name:  "new_user_123456",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})
	s.Run("8. deleting a user should detach it from its respective relations", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-1",
		})
		s.Assert().NoError(err)
		existingGroups, err := s.testBench.Client.ListOrganizationGroups(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationGroupsRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		existingGroup := existingGroups.GetGroups()[0]

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user for org 1",
				Email: "user-1-for-org-1@raystack.org",
				Name:  "user_1_for_org_1_raystack_io",
			},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      existingOrg.GetOrganization().GetId(),
			UserIds: []string{createUserResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)
		_, err = s.testBench.Client.AddGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.AddGroupUsersRequest{
			Id:      existingGroup.GetId(),
			OrgId:   existingGroup.GetOrgId(),
			UserIds: []string{createUserResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)

		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(3, len(orgUsersRespAfterRelation.GetUsers()))
		groupUsersResp, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.ListGroupUsersRequest{
			Id:    existingGroup.Id,
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		var userPartOfGroup bool
		for _, rel := range groupUsersResp.GetUsers() {
			if createUserResp.GetUser().GetId() == rel.GetId() {
				userPartOfGroup = true
				break
			}
		}
		s.Assert().True(userPartOfGroup)

		listUserGroups, err := s.testBench.Client.ListUserGroups(ctxOrgAdminAuth, &frontierv1beta1.ListUserGroupsRequest{
			Id:    createUserResp.GetUser().GetId(),
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(listUserGroups.GetGroups()))

		// delete user
		_, err = s.testBench.Client.DeleteUser(ctxOrgAdminAuth, &frontierv1beta1.DeleteUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)

		// check its existence
		getUserResp, err := s.testBench.Client.GetUser(ctxOrgAdminAuth, &frontierv1beta1.GetUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NotNil(err)
		s.Assert().Nil(getUserResp)

		// check its relations with org
		orgUsersRespAfterDeletion, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterDeletion.GetUsers()))

		// check its relations with group
		groupUsersRespAfterDeletion, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.ListGroupUsersRequest{
			Id:    existingGroup.Id,
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		for _, rel := range groupUsersRespAfterDeletion.GetUsers() {
			s.Assert().NotEqual(createUserResp.GetUser().GetId(), rel.GetId())
		}
	})
	s.Run("9. disabling a user should return not found in list/get api", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-user-1",
		})
		s.Assert().NoError(err)
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user for org 1",
				Email: "user-2-for-org-1@raystack.org",
				Name:  "user_2_for_org_1_raystack_io",
			},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      existingOrg.GetOrganization().GetId(),
			UserIds: []string{createUserResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)
		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))

		// disable user
		_, err = s.testBench.Client.DisableUser(ctxOrgAdminAuth, &frontierv1beta1.DisableUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)

		// check its existence
		getUserResp, err := s.testBench.Client.GetUser(ctxOrgAdminAuth, &frontierv1beta1.GetUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NotNil(err)
		s.Assert().Nil(getUserResp)

		// check its relations with org
		orgUsersRespAfterDisable, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersRespAfterDisable.GetUsers()))

		// enable user
		_, err = s.testBench.Client.EnableUser(ctxOrgAdminAuth, &frontierv1beta1.EnableUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)

		// check its existence
		getUserAfterEnableResp, err := s.testBench.Client.GetUser(ctxOrgAdminAuth, &frontierv1beta1.GetUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getUserAfterEnableResp)

		// check its relations with org
		orgUsersRespAfterEnable, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterEnable.GetUsers()))
	})
	s.Run("10. correctly filter users using list api in an org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-user-2",
		})
		s.Assert().NoError(err)
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user for org 2",
				Email: "user-1-for-org-2@raystack.org",
				Name:  "user_1_for_org_2_raystack_io",
			},
		})
		s.Assert().NoError(err)

		listExistingUsers, err := s.testBench.Client.ListUsers(ctxCurrentUser, &frontierv1beta1.ListUsersRequest{
			OrgId: existingOrg.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(listExistingUsers.GetUsers()))

		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      existingOrg.GetOrganization().GetId(),
			UserIds: []string{createUserResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)

		listNewUsers, err := s.testBench.Client.ListUsers(ctxCurrentUser, &frontierv1beta1.ListUsersRequest{
			OrgId: existingOrg.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(listNewUsers.GetUsers()))
	})
	s.Run("11. correctly filter users using list api with user keyword", func() {
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user",
				Email: "user-1-random-1@raystack.org",
				Name:  "user_1_random_1_raystack_io",
			},
		})
		s.Assert().NoError(err)

		listExistingUsers, err := s.testBench.Client.ListUsers(ctxCurrentUser, &frontierv1beta1.ListUsersRequest{
			Keyword: createUserResp.User.Email,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(listExistingUsers.GetUsers()))
	})
}

func (s *APIRegressionTestSuite) TestRelationAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	s.Run("1. creating a new relation between org and user should attach user to the org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-relation-1",
		})
		s.Assert().NoError(err)

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user 1",
				Email: "new-user-for-rel-1@raystack.org",
				Name:  "new_user_for_rel_1_raystack_io",
			},
		})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersResp.GetUsers()))

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &frontierv1beta1.CreateRelationRequest{Body: &frontierv1beta1.RelationRequestBody{
			Object:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
			Subject:  schema.JoinNamespaceAndResourceID(schema.UserPrincipal, createUserResp.GetUser().GetId()),
			Relation: organization.AdminRole,
		}})
		s.Assert().NoError(err)

		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))
	})
	s.Run("2. creating a relation between org and user with editor role should provide view & edit permission in that org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-relation-2",
		})
		s.Assert().NoError(err)

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user 2",
				Email: "new-user-for-rel-2@raystack.org",
				Name:  "new_user_for_rel_2_raystack_io",
			},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &frontierv1beta1.CreateRelationRequest{Body: &frontierv1beta1.RelationRequestBody{
			Object:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
			Subject:  schema.JoinNamespaceAndResourceID(schema.UserPrincipal, createUserResp.GetUser().GetId()),
			Relation: organization.AdminRole,
		}})
		s.Assert().NoError(err)

		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: organization.AdminPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))

		checkViewPermResp, err := s.testBench.Client.CheckResourcePermission(ctxOrgAdminAuth, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        existingOrg.GetOrganization().GetId(),
			ObjectNamespace: schema.OrganizationNamespace,
			Permission:      schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(true, checkViewPermResp.Status)

		checkEditPermResp, err := s.testBench.Client.CheckResourcePermission(ctxOrgAdminAuth, &frontierv1beta1.CheckResourcePermissionRequest{
			ObjectId:        existingOrg.GetOrganization().GetId(),
			ObjectNamespace: schema.OrganizationNamespace,
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(true, checkEditPermResp.Status)
	})
	s.Run("3. deleting a relation between user and org should remove user from that org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-relation-3",
		})
		s.Assert().NoError(err)

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user 3",
				Email: "new-user-for-rel-3@raystack.org",
				Name:  "new_user_for_rel_3_raystack_io",
			},
		})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersResp.GetUsers()))

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &frontierv1beta1.CreateRelationRequest{Body: &frontierv1beta1.RelationRequestBody{
			Object:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
			Subject:  schema.JoinNamespaceAndResourceID(schema.UserPrincipal, createUserResp.GetUser().GetId()),
			Relation: schema.OwnerRelationName,
		}})
		s.Assert().NoError(err)

		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))

		_, err = s.testBench.Client.DeleteRelation(ctxOrgAdminAuth, &frontierv1beta1.DeleteRelationRequest{
			Object:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
			Subject:  schema.JoinNamespaceAndResourceID(schema.UserPrincipal, createUserResp.GetUser().GetId()),
			Relation: schema.OwnerRelationName,
		})
		s.Assert().NoError(err)

		orgUsersRespAfterRelationDelete, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersRespAfterRelationDelete.GetUsers()))
		s.Assert().Equal(testbench.OrgAdminEmail, orgUsersRespAfterRelationDelete.GetUsers()[0].Email)
	})
}

func (s *APIRegressionTestSuite) TestResourceAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	s.Run("1. creating a resource under a project/org successfully", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org 1",
				Name:  "org-resource-1",
			},
		})
		s.Assert().NoError(err)

		userResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
			Title: "member 1",
			Email: "user-org-resource-1@raystack.org",
			Name:  "user_org_resource_1",
		}})
		s.Assert().NoError(err)

		// attach user to org
		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      createOrgResp.GetOrganization().GetId(),
			UserIds: []string{userResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)

		createProjResp, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  "org-1-proj-1",
				OrgId: createOrgResp.GetOrganization().GetId(),
			},
		})
		s.Assert().NoError(err)

		createResourceResp, err := s.testBench.Client.CreateProjectResource(ctxOrgAdminAuth, &frontierv1beta1.CreateProjectResourceRequest{
			ProjectId: createProjResp.GetProject().GetId(),
			Body: &frontierv1beta1.ResourceRequestBody{
				Name:      "res-1",
				Namespace: computeOrderNamespace,
				Principal: userResp.GetUser().GetId(),
				Metadata:  &structpb.Struct{},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createResourceResp)

		listResourcesResp, err := s.testBench.Client.ListProjectResources(ctxOrgAdminAuth, &frontierv1beta1.ListProjectResourcesRequest{
			ProjectId: createProjResp.GetProject().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal("res-1", listResourcesResp.GetResources()[0].Name)
	})
}

func (s *APIRegressionTestSuite) TestInvitationAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	// enable invite user with roles
	_, err := s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
		Preferences: []*frontierv1beta1.PreferenceRequestBody{
			{
				Name:  preference.PlatformInviteWithRoles,
				Value: "true",
			},
		},
	})
	s.Assert().NoError(err)

	s.Run("1. a user should successfully create a new invitation in org and accept it", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-invitation-1",
		})
		s.Assert().NoError(err)
		createGroupResp, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &frontierv1beta1.CreateGroupRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
			Body: &frontierv1beta1.GroupRequestBody{
				Name: "new-group",
			},
		})
		s.Assert().NoError(err)

		createRoleResp, err := s.testBench.Client.CreateOrganizationRole(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRoleRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
			Body: &frontierv1beta1.RoleRequestBody{
				Title: "invitation role 1",
				Name:  "invitation_role_1",
				Permissions: []string{
					"app.organization.groupcreate",
					"app.organization.grouplist",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createRoleResp)
		s.Assert().Equal("invitation role 1", createRoleResp.GetRole().GetTitle())

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user 1",
				Email: "new-user-for-invite-1@raystack.org",
				Name:  "new_user_for_invite_1_raystack_io",
			},
		})
		s.Assert().NoError(err)

		// check if the user already has permission in group
		ctxCurrentUserAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: "new-user-for-invite-1@raystack.org",
		}))
		checkResp, err := s.testBench.Client.CheckResourcePermission(ctxCurrentUserAuth, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
			Permission: schema.GroupCreatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().False(checkResp.GetStatus())

		createInviteResp, err := s.testBench.Client.CreateOrganizationInvitation(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationInvitationRequest{
			OrgId:    existingOrg.GetOrganization().GetId(),
			UserIds:  []string{createUserResp.GetUser().GetEmail()},
			GroupIds: []string{createGroupResp.GetGroup().GetId()},
			RoleIds:  []string{createRoleResp.GetRole().GetId()},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createInviteResp)

		createdInvite := createInviteResp.GetInvitations()[0]
		getInviteResp, err := s.testBench.Client.GetOrganizationInvitation(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationInvitationRequest{
			Id:    createdInvite.GetId(),
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getInviteResp)
		s.Assert().False(createdInvite.ExpiresAt.AsTime().IsZero())
		s.Assert().Equal(createdInvite.GetId(), getInviteResp.GetInvitation().GetId())

		listInviteByOrgResp, err := s.testBench.Client.ListOrganizationInvitations(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationInvitationsRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getInviteResp)
		s.Assert().Equal(createdInvite.GetId(), listInviteByOrgResp.GetInvitations()[0].GetId())

		listInviteByUserResp, err := s.testBench.Client.ListOrganizationInvitations(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationInvitationsRequest{
			UserId: createUserResp.GetUser().GetEmail(),
			OrgId:  existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getInviteResp)
		s.Assert().Equal(createdInvite.GetId(), listInviteByUserResp.GetInvitations()[0].GetId())

		// user should not be part of the org before accept
		userOrgsBeforeAcceptResp, err := s.testBench.Client.ListOrganizationsByUser(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationsByUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(userOrgsBeforeAcceptResp.GetOrganizations()))
		listGroupUsersBeforeAccept, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.ListGroupUsersRequest{
			Id:    createGroupResp.GetGroup().GetId(),
			OrgId: createGroupResp.GetGroup().GetOrgId(),
		})
		s.Assert().NoError(err)
		s.Assert().Len(listGroupUsersBeforeAccept.GetUsers(), 1)

		// accept invite should add user to org and delete it
		ctxOrgUserAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			testbench.IdentityHeader: createUserResp.GetUser().GetEmail(),
		}))
		_, err = s.testBench.Client.AcceptOrganizationInvitation(ctxOrgUserAuth, &frontierv1beta1.AcceptOrganizationInvitationRequest{
			Id:    createdInvite.GetId(),
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)

		// user should be part of the org
		userOrgsAfterAcceptResp, err := s.testBench.Client.ListOrganizationsByUser(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationsByUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(userOrgsAfterAcceptResp.GetOrganizations()))

		// invitation should be deleted
		_, err = s.testBench.Client.GetOrganizationInvitation(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationInvitationRequest{
			Id:    createdInvite.GetId(),
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().Error(err)

		// should be part of group
		listGroupUsersAfterAccept, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &frontierv1beta1.ListGroupUsersRequest{
			Id:    createGroupResp.GetGroup().GetId(),
			OrgId: createGroupResp.GetGroup().GetOrgId(),
		})
		s.Assert().NoError(err)
		s.Assert().Len(listGroupUsersAfterAccept.GetUsers(), 2)

		// user should have role permissions
		checkAfterAcceptResp, err := s.testBench.Client.CheckResourcePermission(ctxCurrentUserAuth, &frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, existingOrg.GetOrganization().GetId()),
			Permission: schema.GroupCreatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().True(checkAfterAcceptResp.GetStatus())
	})
	s.Run("2. users already part of an org shouldn't be invited again", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-invitation-2",
		})
		s.Assert().NoError(err)

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &frontierv1beta1.CreateUserRequest{
			Body: &frontierv1beta1.UserRequestBody{
				Title: "new user 1",
				Email: "new-user-for-invite-2@raystack.org",
				Name:  "new_user_for_invite_2_raystack_io",
			},
		})
		s.Assert().NoError(err)

		// attach user to org
		_, err = s.testBench.Client.AddOrganizationUsers(ctxOrgAdminAuth, &frontierv1beta1.AddOrganizationUsersRequest{
			Id:      existingOrg.GetOrganization().GetId(),
			UserIds: []string{createUserResp.GetUser().GetId()},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateOrganizationInvitation(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationInvitationRequest{
			OrgId:   existingOrg.GetOrganization().GetId(),
			UserIds: []string{createUserResp.GetUser().GetEmail()},
		})
		s.Assert().Error(err)
		s.Assert().ErrorContains(err, "already a member of organization")
	})

	// disable invite user with roles back
	_, err = s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
		Preferences: []*frontierv1beta1.PreferenceRequestBody{
			{
				Name:  preference.PlatformInviteWithRoles,
				Value: "false",
			},
		},
	})
	s.Assert().NoError(err)
}

func (s *APIRegressionTestSuite) TestOrganizationAuditLogsAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	dummyAuditLogs := []*frontierv1beta1.AuditLog{
		{
			Source: "frontier",
			Action: "user.login",
			Actor: &frontierv1beta1.AuditLogActor{
				Type: schema.UserPrincipal,
				Name: "john",
			},
			Target: &frontierv1beta1.AuditLogTarget{
				Name: "org-1",
				Type: schema.OrganizationNamespace,
			},
			Context: map[string]string{
				"usage": "test",
			},
		},
		{
			Source: "frontier",
			Action: "user.logout",
			Actor: &frontierv1beta1.AuditLogActor{
				Type: schema.UserPrincipal,
				Name: "john",
			},
			Target: &frontierv1beta1.AuditLogTarget{
				Name: "org-1",
				Type: schema.OrganizationNamespace,
			},
		},
	}
	s.Run("1. create a new log successfully under an org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-auditlogs-1",
		})
		s.Assert().NoError(err)

		createLogResp, err := s.testBench.Client.CreateOrganizationAuditLogs(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationAuditLogsRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
			Logs:  dummyAuditLogs,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createLogResp)
	})
	s.Run("2. list logs successfully under an org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-auditlogs-1",
		})
		s.Assert().NoError(err)

		listLogResp, err := s.testBench.Client.ListOrganizationAuditLogs(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationAuditLogsRequest{
			OrgId: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(listLogResp)
		unMatchedLogs := 2
		for _, log := range listLogResp.GetLogs() {
			if slices.ContainsFunc[[]*frontierv1beta1.AuditLog](dummyAuditLogs, func(l *frontierv1beta1.AuditLog) bool {
				return l.Action == log.Action && l.Source == log.Source
			}) {
				unMatchedLogs--
			}
		}
		s.Assert().Equal(0, unMatchedLogs)
	})
}

func (s *APIRegressionTestSuite) TestRolesAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. list all platform roles successfully", func() {
		platformRoles, err := s.testBench.Client.ListRoles(ctxOrgAdminAuth, &frontierv1beta1.ListRolesRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(platformRoles)
		s.Assert().True(len(platformRoles.GetRoles()) > 0)
	})
	s.Run("1. creating/updating platform role successfully", func() {
		createRole, err := s.testBench.AdminClient.CreateRole(ctxOrgAdminAuth, &frontierv1beta1.CreateRoleRequest{
			Body: &frontierv1beta1.RoleRequestBody{
				Title: "new role 1",
				Name:  "new_role_1",
				Permissions: []string{
					"app.organization.groupcreate",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createRole)
		s.Assert().Equal("new role 1", createRole.GetRole().GetTitle())

		// try updating it with different title
		updateRole, err := s.testBench.AdminClient.UpdateRole(ctxOrgAdminAuth, &frontierv1beta1.UpdateRoleRequest{
			Id: createRole.GetRole().GetId(),
			Body: &frontierv1beta1.RoleRequestBody{
				Title: "new role 1 updated",
				Name:  "new_role_1",
				Scopes: []string{
					schema.OrganizationNamespace,
				},
				Permissions: []string{
					"app.organization.groupcreate",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(updateRole)
		s.Assert().Equal("new role 1 updated", updateRole.GetRole().GetTitle())
	})
}

func (s *APIRegressionTestSuite) TestPreferencesAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. list all preference traits successfully", func() {
		prefTraitResp, err := s.testBench.Client.DescribePreferences(ctxOrgAdminAuth, &frontierv1beta1.DescribePreferencesRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(prefTraitResp)
		s.Assert().True(len(prefTraitResp.GetTraits()) > 0)
	})
	s.Run("2. create and fetch organization preference successfully", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationRequest{
			Id: "org-preferences-1",
		})
		s.Assert().NoError(err)
		createPrefResp, err := s.testBench.Client.CreateOrganizationPreferences(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationPreferencesRequest{
			Id: existingOrg.GetOrganization().GetId(),
			Bodies: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.OrganizationSocialLogin,
					Value: "true",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createPrefResp)
		s.Assert().True(len(createPrefResp.GetPreferences()) > 0)

		getPrefResp, err := s.testBench.Client.ListOrganizationPreferences(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationPreferencesRequest{
			Id: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getPrefResp)
		s.Assert().Equal("true", getPrefResp.GetPreferences()[0].GetValue())

		// try updating it with different value
		createPref2Resp, err := s.testBench.Client.CreateOrganizationPreferences(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationPreferencesRequest{
			Id: existingOrg.GetOrganization().GetId(),
			Bodies: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.OrganizationSocialLogin,
					Value: "false",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().True(len(createPref2Resp.GetPreferences()) > 0)

		getPref2Resp, err := s.testBench.Client.ListOrganizationPreferences(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationPreferencesRequest{
			Id: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal("false", getPref2Resp.GetPreferences()[0].GetValue())
	})
	s.Run("3. create and fetch platform preference successfully", func() {
		createPrefResp, err := s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
			Preferences: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.PlatformDisableOrgsOnCreate,
					Value: "false",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createPrefResp)
		s.Assert().True(len(createPrefResp.GetPreference()) > 0)

		// try updating it with different value
		createPref2Resp, err := s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
			Preferences: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.PlatformDisableOrgsOnCreate,
					Value: "true",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().True(len(createPref2Resp.GetPreference()) > 0)

		getPref2Resp, err := s.testBench.AdminClient.ListPreferences(ctxOrgAdminAuth, &frontierv1beta1.ListPreferencesRequest{})
		s.Assert().NoError(err)
		s.Assert().Equal("true", utils.Filter(getPref2Resp.GetPreferences(), func(p *frontierv1beta1.Preference) bool {
			return p.GetName() == preference.PlatformDisableOrgsOnCreate
		})[0].GetValue())
	})
	s.Run("4. PlatformDisableOrgsOnCreate if set to true should disable all orgs when created", func() {
		createPrefResp, err := s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
			Preferences: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.PlatformDisableOrgsOnCreate,
					Value: "true",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createPrefResp)
		s.Assert().True(len(createPrefResp.GetPreference()) > 0)

		// create a new org
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org 2",
				Name:  "org-preferences-2",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createOrgResp)
		s.Assert().Equal(organization.Disabled.String(), createOrgResp.GetOrganization().GetState())

		// reset it back to false
		updatePrefResp, err := s.testBench.AdminClient.CreatePreferences(ctxOrgAdminAuth, &frontierv1beta1.CreatePreferencesRequest{
			Preferences: []*frontierv1beta1.PreferenceRequestBody{
				{
					Name:  preference.PlatformDisableOrgsOnCreate,
					Value: "false",
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(updatePrefResp)
	})
}

func (s *APIRegressionTestSuite) TestOrganizationDomainsAPI() {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))
	s.Run("1. create and fetch organization domains successfully", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org 1",
				Name:  "org-domains-1",
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createOrgResp)

		createDomainResp, err := s.testBench.Client.CreateOrganizationDomain(ctxOrgAdminAuth, &frontierv1beta1.CreateOrganizationDomainRequest{
			OrgId:  createOrgResp.GetOrganization().GetId(),
			Domain: "org-domains-1.raystack.io",
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createDomainResp)

		listDomainResp, err := s.testBench.Client.ListOrganizationDomains(ctxOrgAdminAuth, &frontierv1beta1.ListOrganizationDomainsRequest{
			OrgId: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(listDomainResp)
		s.Assert().Equal("org-domains-1.raystack.io", listDomainResp.GetDomains()[0].Name)

		getDomainResp, err := s.testBench.Client.GetOrganizationDomain(ctxOrgAdminAuth, &frontierv1beta1.GetOrganizationDomainRequest{
			Id:    createDomainResp.GetDomain().GetId(),
			OrgId: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getDomainResp)
	})
}

func TestEndToEndAPIRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(APIRegressionTestSuite))
}
