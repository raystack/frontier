package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/odpf/shield/internal/schema"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/pkg/logger"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/odpf/shield/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	fixturesDir = "testdata"
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

	appConfig := &config.Shield{
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

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)

	ctx := context.Background()

	s.Require().NoError(testbench.BootstrapMetadataKey(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapUsers(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapOrganizations(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapProject(ctx, s.testBench.Client, testbench.OrgAdminEmail))
	s.Require().NoError(testbench.BootstrapGroup(ctx, s.testBench.Client, testbench.OrgAdminEmail))

	// validate
	uRes, err := s.testBench.Client.ListUsers(ctx, &shieldv1beta1.ListUsersRequest{})
	s.Require().NoError(err)
	s.Require().Equal(9, len(uRes.GetUsers()))

	oRes, err := s.testBench.Client.ListOrganizations(ctx, &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(1, len(oRes.GetOrganizations()))

	pRes, err := s.testBench.Client.ListProjects(ctx, &shieldv1beta1.ListProjectsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(1, len(pRes.GetProjects()))

	gRes, err := s.testBench.Client.ListGroups(ctx, &shieldv1beta1.ListGroupsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(3, len(gRes.GetGroups()))
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
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &shieldv1beta1.CreateOrganizationRequest{
			Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "org acme 1",
				Slug: "org-acme-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			},
		})
		s.Assert().Nil(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id: createOrgResp.GetOrganization().GetId(),
			//PermissionFilter: schema.EditPermission,
		})
		s.Assert().Nil(err)
		s.Assert().Equal(1, len(orgUsersResp.GetUsers()))
		s.Assert().Equal(testbench.OrgAdminEmail, orgUsersResp.GetUsers()[0].Email)
	})
	s.Run("2. user attached to an org as member should have no basic permission other than membership", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &shieldv1beta1.CreateOrganizationRequest{
			Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "org acme 2",
				Slug: "org-acme-2",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			},
		})
		s.Assert().Nil(err)

		userResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{Body: &shieldv1beta1.UserRequestBody{
			Name:  "acme 2 member",
			Email: "acme-member@odpf.io",
			Slug:  "acme-2-member",
		}})
		s.Assert().Nil(err)

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:        createOrgResp.GetOrganization().GetId(),
			ObjectNamespace: schema.OrganizationNamespace,
			Subject:         schema.UserPrincipal + ":" + userResp.GetUser().GetId(),
			RoleName:        schema.MemberRole,
		}})
		s.Assert().Nil(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().Nil(err)
		s.Assert().Equal(2, len(orgUsersResp.GetUsers()))
	})
}

func (s *APIRegressionTestSuite) TestProjectAPI() {
	var newProject *shieldv1beta1.Project

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	// get my org
	res, err := s.testBench.Client.ListOrganizations(context.Background(), &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(res.GetOrganizations()), 0)
	myOrg := res.GetOrganizations()[0]

	s.Run("1. org admin create a new project with empty auth email should not return unauthenticated error", func() {
		_, err := s.testBench.Client.CreateProject(context.Background(), &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project",
				OrgId: myOrg.GetId(),
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Nil(err)
	})

	s.Run("2. org admin create a new project with empty name should return invalid argument", func() {
		_, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Slug:  "new-project",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new project with wrong org id should return invalid argument", func() {
		_, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new project with same name and org-id should conflict", func() {
		res, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project-duplicate",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newProject = res.GetProject()
		s.Assert().NotNil(newProject)

		_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project-duplicate",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. org admin update a new project with empty body should return invalid argument", func() {
		_, err := s.testBench.Client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
			Id:   newProject.GetId(),
			Body: nil,
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("6. org admin update a new project with empty group id should return not found", func() {
		_, err := s.testBench.Client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
			Id: "",
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.NotFound, status.Convert(err).Code())
	})

	s.Run("7. org admin update a new project with unknown project id and not uuid should return not found", func() {
		_, err := s.testBench.Client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
			Id: "random",
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.NotFound, status.Convert(err).Code())
	})

	s.Run("8. org admin update a new project with same name and org-id but different id should return conflict", func() {
		_, err := s.testBench.Client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
			Id: newProject.GetId(),
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "project 1",
				Slug:  "project-1",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})
}

func (s *APIRegressionTestSuite) TestGroupAPI() {
	var newGroup *shieldv1beta1.Group

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	// get my org
	res, err := s.testBench.Client.ListOrganizations(context.Background(), &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(res.GetOrganizations()), 0)
	myOrg := res.GetOrganizations()[0]

	s.Run("1. org admin create a new team with empty auth email should return unauthenticated error", func() {
		_, err := s.testBench.Client.CreateGroup(context.Background(), &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	s.Run("2. org admin create a new team with empty name should return invalid argument", func() {
		_, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new team with wrong org id should return invalid argument", func() {
		_, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new team with same name and org-id should conflict", func() {
		res, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newGroup = res.GetGroup()
		s.Assert().NotNil(newGroup)

		_, err = s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. group admin update a new team with empty body should return invalid argument", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
			Id:   newGroup.GetId(),
			Body: nil,
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("6. group admin update a new team with empty group id should return not found", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
			Id: "",
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.NotFound, status.Convert(err).Code())
	})

	s.Run("7. group admin update a new team with unknown group id and not uuid should return not found", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
			Id: "random",
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.NotFound, status.Convert(err).Code())
	})

	s.Run("8. group admin update a new team with same name and org-id but different id should return conflict", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
			Id: newGroup.GetId(),
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "org1 group1",
				Slug:  "org1-group1",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})
}

func (s *APIRegressionTestSuite) TestUserAPI() {
	var newUser *shieldv1beta1.User

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	s.Run("1. org admin create a new user with empty auth email should return unauthenticated error", func() {
		_, err := s.testBench.Client.CreateUser(context.Background(), &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "new-user-a@odpf.io",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	s.Run("2. org admin create a new user with unparsable metadata should return invalid argument error", func() {
		_, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "new-user-a@odpf.io",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewNullValue(),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new user with empty email should use email from header", func() {
		_, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new user with same email should return conflict error", func() {
		res, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "new-user-a@odpf.io",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().NoError(err)
		newUser = res.GetUser()

		_, err = s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "new-user-a@odpf.io",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. org admin update user with conflicted detail should return conflict error", func() {
		_, err := s.testBench.Client.UpdateUser(ctxOrgAdminAuth, &shieldv1beta1.UpdateUserRequest{
			Id: newUser.GetId(),
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "admin1-group1-org1@odpf.io",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	ctxCurrentUser := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: newUser.GetEmail(),
	}))

	s.Run("6. update current user with empty email should return invalid argument error", func() {
		_, err := s.testBench.Client.UpdateCurrentUser(ctxCurrentUser, &shieldv1beta1.UpdateCurrentUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("7. update current user with different email in header and body should return invalid argument error", func() {
		_, err := s.testBench.Client.UpdateCurrentUser(ctxCurrentUser, &shieldv1beta1.UpdateCurrentUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "admin1-group1-org1@odpf.io",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})
}

func TestEndToEndAPIRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(APIRegressionTestSuite))
}
