package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/odpf/shield/internal/bootstrap/schema"

	"github.com/odpf/shield/core/organization"

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
	fixturesDir          = "testdata"
	potatoOrderNamespace = "potato/order"
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
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
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
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().NoError(err)

		userResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{Body: &shieldv1beta1.UserRequestBody{
			Name:  "acme 2 member",
			Email: "acme-member@odpf.io",
			Slug:  "acme_2_member",
		}})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         createOrgResp.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectId:        userResp.GetUser().GetId(),
			SubjectNamespace: schema.UserPrincipal,
			RelationName:     schema.MemberRole,
		}})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id: createOrgResp.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersResp.GetUsers()))
	})
	s.Run("3. deleting an org should delete all of its internal relations/projects/groups/resources", func() {
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &shieldv1beta1.CreateOrganizationRequest{
			Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "org acme 3",
				Slug: "org-acme-3",
			},
		})
		s.Assert().NoError(err)

		createUserResponse, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{Body: &shieldv1beta1.UserRequestBody{
			Name:  "acme 3 member 1",
			Email: "acme-member-1@odpf.io",
			Slug:  "acme_3_member_1",
		}})
		s.Assert().NoError(err)

		// attach user to org
		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         createOrgResp.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResponse.GetUser().GetId(),
			RelationName:     schema.MemberRole,
		}})
		s.Assert().NoError(err)

		createProjResp, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "proj-1",
				Slug:  "org-3-proj-1",
				OrgId: createOrgResp.GetOrganization().GetId(),
			},
		})
		s.Assert().NoError(err)

		createResourceResp, err := s.testBench.Client.CreateResource(ctxOrgAdminAuth, &shieldv1beta1.CreateResourceRequest{
			Body: &shieldv1beta1.ResourceRequestBody{
				Name:        "res-1",
				ProjectId:   createProjResp.GetProject().GetId(),
				NamespaceId: potatoOrderNamespace,
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createResourceResp)

		// check users
		listUsersBeforeDelete, err := s.testBench.Client.ListUsers(ctxOrgAdminAuth, &shieldv1beta1.ListUsersRequest{
			OrgId: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(listUsersBeforeDelete.Users))

		// delete org and all its items
		_, err = s.testBench.Client.DeleteOrganization(ctxOrgAdminAuth, &shieldv1beta1.DeleteOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)

		// check org
		_, err = s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		s.Assert().NotNil(err)

		// check project
		_, err = s.testBench.Client.GetProject(ctxOrgAdminAuth, &shieldv1beta1.GetProjectRequest{
			Id: createProjResp.Project.Id,
		})
		s.Assert().NotNil(err)

		// check resource
		_, err = s.testBench.Client.GetResource(ctxOrgAdminAuth, &shieldv1beta1.GetResourceRequest{
			Id: createResourceResp.Resource.Id,
		})
		s.Assert().NotNil(err)

		// check user relations
		listUsersAfterDelete, err := s.testBench.Client.ListUsers(ctxOrgAdminAuth, &shieldv1beta1.ListUsersRequest{
			OrgId: createOrgResp.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(listUsersAfterDelete.Users))
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
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().NoError(err)
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

	s.Run("9. list all projects attached/filtered to an org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-project-1",
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "org-project-1-p1",
				OrgId: existingOrg.Organization.GetId(),
			},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "org-project-1-p2",
				OrgId: existingOrg.Organization.GetId(),
			},
		})
		s.Assert().NoError(err)

		listResp, err := s.testBench.Client.ListOrganizationProjects(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationProjectsRequest{
			Id: existingOrg.Organization.GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(listResp.Projects))
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
				Name:  "group 1",
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
				Name:  "group 1",
				Slug:  "new-group",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new team with same name and org-id should conflict", func() {
		res, err := s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "group 1",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newGroup = res.GetGroup()
		s.Assert().NotNil(newGroup)

		_, err = s.testBench.Client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "group 1",
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
				Slug:  "new_user_123456",
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
		_, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "new-user-a@odpf.io",
				Slug:  "new_user_123456",
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
		_, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "",
				Slug:  "new_user_123456",
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
		res, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "new-user-a@odpf.io",
				Slug:  "new-user-123456",
				Metadata: &structpb.Struct{
					// TODO(kushsharma) add back foo fields once metadata jsonschema
					// is implemented
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
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
				Slug:  "new_user_123456",
				Metadata: &structpb.Struct{
					// TODO(kushsharma) add back foo fields once metadata jsonschema
					// is implemented
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. org admin update user with conflicted detail should not update the email and return nil error", func() {
		ExpectedEmail := "new-user-a@odpf.io"
		res, err := s.testBench.Client.UpdateUser(ctxOrgAdminAuth, &shieldv1beta1.UpdateUserRequest{
			Id: newUser.GetId(),
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "admin1-group2-org1@odpf.io",
				Slug:  "new_user_123456",
				Metadata: &structpb.Struct{
					// TODO(kushsharma) add back foo fields once metadata jsonschema
					// is implemented
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
		_, err := s.testBench.Client.UpdateCurrentUser(ctxCurrentUser, &shieldv1beta1.UpdateCurrentUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "",
				Slug:  "new_user_123456",
				Metadata: &structpb.Struct{
					// TODO(kushsharma) add back foo fields once metadata jsonschema
					// is implemented
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
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
				Slug:  "new_user_123456",
				Metadata: &structpb.Struct{
					// TODO(kushsharma) add back foo fields once metadata jsonschema
					// is implemented
					Fields: map[string]*structpb.Value{
						"description": structpb.NewStringValue("Description"),
					},
				},
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})
	s.Run("8. deleting a user should detach it from its respective relations", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-user-1",
		})
		s.Assert().NoError(err)
		existingGroup, err := s.testBench.Client.GetGroup(ctxOrgAdminAuth, &shieldv1beta1.GetGroupRequest{
			Id: "org1-group1",
		})
		s.Assert().NoError(err)
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user for org 1",
				Email: "user-1-for-org-1@odpf.io",
				Slug:  "user_1_for_org_1_odpf_io",
			},
		})
		s.Assert().NoError(err)
		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         existingOrg.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResp.GetUser().GetId(),
			RelationName:     schema.OwnerRole,
		}})
		s.Assert().NoError(err)
		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         existingGroup.GetGroup().GetId(),
			ObjectNamespace:  schema.GroupNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResp.GetUser().GetId(),
			RelationName:     schema.MemberRole,
		}})
		s.Assert().NoError(err)
		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))
		groupUsersResp, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &shieldv1beta1.ListGroupUsersRequest{
			Id: existingGroup.Group.Id,
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

		// delete user
		_, err = s.testBench.Client.DeleteUser(ctxOrgAdminAuth, &shieldv1beta1.DeleteUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)

		// check its existence
		getUserResp, err := s.testBench.Client.GetUser(ctxOrgAdminAuth, &shieldv1beta1.GetUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NotNil(err)
		s.Assert().Nil(getUserResp)

		// check its relations with org
		orgUsersRespAfterDeletion, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersRespAfterDeletion.GetUsers()))

		// check its relations with group
		groupUsersRespAfterDeletion, err := s.testBench.Client.ListGroupUsers(ctxOrgAdminAuth, &shieldv1beta1.ListGroupUsersRequest{
			Id: existingGroup.Group.Id,
		})
		s.Assert().NoError(err)
		for _, rel := range groupUsersRespAfterDeletion.GetUsers() {
			s.Assert().NotEqual(createUserResp.GetUser().GetId(), rel.GetId())
		}
	})
	s.Run("9. disabling a user should return not found in list/get api", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-user-1",
		})
		s.Assert().NoError(err)
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user for org 1",
				Email: "user-2-for-org-1@odpf.io",
				Slug:  "user_2_for_org_1_odpf_io",
			},
		})
		s.Assert().NoError(err)
		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         existingOrg.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResp.GetUser().GetId(),
			RelationName:     schema.OwnerRole,
		}})
		s.Assert().NoError(err)
		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))

		// disable user
		_, err = s.testBench.Client.DisableUser(ctxOrgAdminAuth, &shieldv1beta1.DisableUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)

		// check its existence
		getUserResp, err := s.testBench.Client.GetUser(ctxOrgAdminAuth, &shieldv1beta1.GetUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NotNil(err)
		s.Assert().Nil(getUserResp)

		// check its relations with org
		orgUsersRespAfterDisable, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersRespAfterDisable.GetUsers()))

		// enable user
		_, err = s.testBench.Client.EnableUser(ctxOrgAdminAuth, &shieldv1beta1.EnableUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)

		// check its existence
		getUserAfterEnableResp, err := s.testBench.Client.GetUser(ctxOrgAdminAuth, &shieldv1beta1.GetUserRequest{
			Id: createUserResp.GetUser().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(getUserAfterEnableResp)

		// check its relations with org
		orgUsersRespAfterEnable, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterEnable.GetUsers()))
	})
	s.Run("10. correctly filter users using list api in an org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-user-2",
		})
		s.Assert().NoError(err)
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user for org 2",
				Email: "user-1-for-org-2@odpf.io",
				Slug:  "user_1_for_org_2_odpf_io",
			},
		})
		s.Assert().NoError(err)

		listExistingUsers, err := s.testBench.Client.ListUsers(ctxCurrentUser, &shieldv1beta1.ListUsersRequest{
			OrgId: existingOrg.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(listExistingUsers.GetUsers()))

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         existingOrg.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResp.GetUser().GetId(),
			RelationName:     schema.OwnerRole,
		}})
		s.Assert().NoError(err)

		listNewUsers, err := s.testBench.Client.ListUsers(ctxCurrentUser, &shieldv1beta1.ListUsersRequest{
			OrgId: existingOrg.Organization.Id,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(listNewUsers.GetUsers()))
	})
	s.Run("11. correctly filter users using list api with user keyword", func() {
		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user",
				Email: "user-1-random-1@odpf.io",
				Slug:  "user_1_random_1_odpf_io",
			},
		})
		s.Assert().NoError(err)

		listExistingUsers, err := s.testBench.Client.ListUsers(ctxCurrentUser, &shieldv1beta1.ListUsersRequest{
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
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-relation-1",
		})
		s.Assert().NoError(err)

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user 1",
				Email: "new-user-for-rel-1@odpf.io",
				Slug:  "new_user_for_rel_1_odpf_io",
			},
		})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersResp.GetUsers()))

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         existingOrg.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResp.GetUser().GetId(),
			RelationName:     organization.AdminRole,
		}})
		s.Assert().NoError(err)

		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))
	})
	s.Run("2. creating a relation between org and user with editor role should provide view & edit permission in that org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-relation-2",
		})
		s.Assert().NoError(err)

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user 2",
				Email: "new-user-for-rel-2@odpf.io",
				Slug:  "new_user_for_rel_2_odpf_io",
			},
		})
		s.Assert().NoError(err)

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         existingOrg.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResp.GetUser().GetId(),
			RelationName:     organization.AdminRole,
		}})
		s.Assert().NoError(err)

		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: organization.AdminPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))

		checkViewPermResp, err := s.testBench.Client.CheckResourcePermission(ctxOrgAdminAuth, &shieldv1beta1.CheckResourcePermissionRequest{
			ObjectId:        existingOrg.GetOrganization().GetId(),
			ObjectNamespace: schema.OrganizationNamespace,
			Permission:      schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(true, checkViewPermResp.Status)

		checkEditPermResp, err := s.testBench.Client.CheckResourcePermission(ctxOrgAdminAuth, &shieldv1beta1.CheckResourcePermissionRequest{
			ObjectId:        existingOrg.GetOrganization().GetId(),
			ObjectNamespace: schema.OrganizationNamespace,
			Permission:      schema.UpdatePermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(true, checkEditPermResp.Status)
	})
	s.Run("3. deleting a relation between user and org should remove user from that org", func() {
		existingOrg, err := s.testBench.Client.GetOrganization(ctxOrgAdminAuth, &shieldv1beta1.GetOrganizationRequest{
			Id: "org-relation-3",
		})
		s.Assert().NoError(err)

		createUserResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user 3",
				Email: "new-user-for-rel-3@odpf.io",
				Slug:  "new_user_for_rel_3_odpf_io",
			},
		})
		s.Assert().NoError(err)

		orgUsersResp, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id: existingOrg.GetOrganization().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal(1, len(orgUsersResp.GetUsers()))

		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         existingOrg.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        createUserResp.GetUser().GetId(),
			RelationName:     schema.OwnerRole,
		}})
		s.Assert().NoError(err)

		orgUsersRespAfterRelation, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
			Id:               existingOrg.GetOrganization().GetId(),
			PermissionFilter: schema.GetPermission,
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(orgUsersRespAfterRelation.GetUsers()))

		_, err = s.testBench.Client.DeleteRelation(ctxOrgAdminAuth, &shieldv1beta1.DeleteRelationRequest{
			ObjectNamespace:  schema.OrganizationNamespace,
			ObjectId:         existingOrg.GetOrganization().GetId(),
			SubjectId:        createUserResp.GetUser().GetId(),
			SubjectNamespace: schema.UserPrincipal,
			Relation:         schema.OwnerRole,
		})
		s.Assert().NoError(err)

		orgUsersRespAfterRelationDelete, err := s.testBench.Client.ListOrganizationUsers(ctxOrgAdminAuth, &shieldv1beta1.ListOrganizationUsersRequest{
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
		createOrgResp, err := s.testBench.Client.CreateOrganization(ctxOrgAdminAuth, &shieldv1beta1.CreateOrganizationRequest{
			Body: &shieldv1beta1.OrganizationRequestBody{
				Name: "org 1",
				Slug: "org-resource-1",
			},
		})
		s.Assert().NoError(err)

		userResp, err := s.testBench.Client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{Body: &shieldv1beta1.UserRequestBody{
			Name:  "member 1",
			Email: "user-org-resource-1@odpf.io",
			Slug:  "user_org_resource_1",
		}})
		s.Assert().NoError(err)

		// attach user to org
		_, err = s.testBench.Client.CreateRelation(ctxOrgAdminAuth, &shieldv1beta1.CreateRelationRequest{Body: &shieldv1beta1.RelationRequestBody{
			ObjectId:         createOrgResp.GetOrganization().GetId(),
			ObjectNamespace:  schema.OrganizationNamespace,
			SubjectNamespace: schema.UserPrincipal,
			SubjectId:        userResp.GetUser().GetId(),
			RelationName:     schema.MemberRole,
		}})
		s.Assert().NoError(err)

		createProjResp, err := s.testBench.Client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "proj-1",
				Slug:  "org-1-proj-1",
				OrgId: createOrgResp.GetOrganization().GetId(),
			},
		})
		s.Assert().NoError(err)

		compassNamespacesResp, err := s.testBench.Client.GetNamespace(ctxOrgAdminAuth, &shieldv1beta1.GetNamespaceRequest{
			Id: potatoOrderNamespace,
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(compassNamespacesResp)

		createResourceResp, err := s.testBench.Client.CreateResource(ctxOrgAdminAuth, &shieldv1beta1.CreateResourceRequest{
			Body: &shieldv1beta1.ResourceRequestBody{
				Name:        "res-1",
				ProjectId:   createProjResp.GetProject().GetId(),
				NamespaceId: compassNamespacesResp.Namespace.Name,
				UserId:      userResp.GetUser().GetId(),
			},
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(createResourceResp)

		listResourcesResp, err := s.testBench.Client.ListProjectResources(ctxOrgAdminAuth, &shieldv1beta1.ListProjectResourcesRequest{
			ProjectId: createProjResp.GetProject().GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Equal("res-1", listResourcesResp.GetResources()[0].Name)
	})
}
func TestEndToEndAPIRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(APIRegressionTestSuite))
}
