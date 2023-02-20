package e2e_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/pkg/logger"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/odpf/shield/test/e2e_test/testbench"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type EndToEndAPIRegressionTestSuite struct {
	suite.Suite
	client       shieldv1beta1.ShieldServiceClient
	cancelClient func()
	testBench    *testbench.TestBench
}

func (s *EndToEndAPIRegressionTestSuite) SetupTest() {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	parent := filepath.Dir(wd)
	testDataPath := parent + "/testbench/testdata"

	proxyPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	apiPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	appConfig := &config.Shield{
		Log: logger.Config{
			Level: "fatal",
		},
		App: server.Config{
			HTTPPort:            apiPort,
			IdentityProxyHeader: testbench.IdentityHeader,
			ResourcesConfigPath: fmt.Sprintf("file://%s/%s", wd, "testdata/configs/resources"),
			RulesPath:           fmt.Sprintf("file://%s/%s", wd, "testdata/configs/rules"),
		},
		Proxy: proxy.ServicesConfig{
			Services: []proxy.Config{
				{
					Name:      "base",
					Port:      proxyPort,
					RulesPath: fmt.Sprintf("file://%s/%s", wd, "testdata/configs/rules"),
				},
			},
		},
	}

	s.testBench, appConfig, err = testbench.Init(appConfig)
	s.Require().NoError(err)

	ctx := context.Background()
	s.client, s.cancelClient, err = testbench.CreateClient(ctx, fmt.Sprintf("localhost:%d", apiPort))
	s.Require().NoError(err)

	s.Require().NoError(testbench.BootstrapMetadataKey(ctx, s.client, testbench.OrgAdminEmail, testDataPath))
	s.Require().NoError(testbench.BootstrapUser(ctx, s.client, testbench.OrgAdminEmail, testDataPath))
	s.Require().NoError(testbench.BootstrapOrganization(ctx, s.client, testbench.OrgAdminEmail, testDataPath))
	s.Require().NoError(testbench.BootstrapProject(ctx, s.client, testbench.OrgAdminEmail, testDataPath))
	s.Require().NoError(testbench.BootstrapGroup(ctx, s.client, testbench.OrgAdminEmail, testDataPath))

	// validate
	uRes, err := s.client.ListUsers(ctx, &shieldv1beta1.ListUsersRequest{})
	s.Require().NoError(err)
	s.Require().Equal(9, len(uRes.GetUsers()))

	oRes, err := s.client.ListOrganizations(ctx, &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(1, len(oRes.GetOrganizations()))

	pRes, err := s.client.ListProjects(ctx, &shieldv1beta1.ListProjectsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(1, len(pRes.GetProjects()))

	gRes, err := s.client.ListGroups(ctx, &shieldv1beta1.ListGroupsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(3, len(gRes.GetGroups()))
}

func (s *EndToEndAPIRegressionTestSuite) TearDownTest() {
	s.cancelClient()
	// Clean tests
	err := s.testBench.CleanUp()
	s.Require().NoError(err)
}

func (s *EndToEndAPIRegressionTestSuite) TestProjectAPI() {
	var newProject *shieldv1beta1.Project

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	// get my org
	res, err := s.client.ListOrganizations(context.Background(), &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(res.GetOrganizations()), 0)
	myOrg := res.GetOrganizations()[0]

	s.Run("1. org admin create a new project with empty auth email should return unauthenticated error", func() {
		_, err := s.client.CreateProject(context.Background(), &shieldv1beta1.CreateProjectRequest{
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
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	s.Run("2. org admin create a new project with empty name should return invalid argument", func() {
		_, err := s.client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Slug:  "new-project",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new project with wrong org id should return invalid argument", func() {
		_, err := s.client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new project with same name and org-id should conflict", func() {
		res, err := s.client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newProject = res.GetProject()
		s.Assert().NotNil(newProject)

		_, err = s.client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
			Body: &shieldv1beta1.ProjectRequestBody{
				Name:  "new project",
				Slug:  "new-project",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. org admin update a new project with empty body should return invalid argument", func() {
		_, err := s.client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
			Id:   newProject.GetId(),
			Body: nil,
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("6. org admin update a new project with empty group id should return not found", func() {
		_, err := s.client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
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
		_, err := s.client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
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
		_, err := s.client.UpdateProject(ctxOrgAdminAuth, &shieldv1beta1.UpdateProjectRequest{
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

func (s *EndToEndAPIRegressionTestSuite) TestGroupAPI() {
	var newGroup *shieldv1beta1.Group

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	// get my org
	res, err := s.client.ListOrganizations(context.Background(), &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(res.GetOrganizations()), 0)
	myOrg := res.GetOrganizations()[0]

	s.Run("1. org admin create a new team with empty auth email should return unauthenticated error", func() {
		_, err := s.client.CreateGroup(context.Background(), &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	s.Run("2. org admin create a new team with empty name should return invalid argument", func() {
		_, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. org admin create a new team with wrong org id should return invalid argument", func() {
		_, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. org admin create a new team with same name and org-id should conflict", func() {
		res, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newGroup = res.GetGroup()
		s.Assert().NotNil(newGroup)

		_, err = s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.AlreadyExists, status.Convert(err).Code())
	})

	s.Run("5. group admin update a new team with empty body should return invalid argument", func() {
		_, err := s.client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
			Id:   newGroup.GetId(),
			Body: nil,
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("6. group admin update a new team with empty group id should return not found", func() {
		_, err := s.client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
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
		_, err := s.client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
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
		_, err := s.client.UpdateGroup(ctxOrgAdminAuth, &shieldv1beta1.UpdateGroupRequest{
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

func (s *EndToEndAPIRegressionTestSuite) TestUserAPI() {
	var newUser *shieldv1beta1.User

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	s.Run("1. org admin create a new user with empty auth email should return unauthenticated error", func() {
		_, err := s.client.CreateUser(context.Background(), &shieldv1beta1.CreateUserRequest{
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
		_, err := s.client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
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

	s.Run("3. org admin create a new user with empty email should return invalid argument error", func() {
		_, err := s.client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
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

	s.Run("4. org admin create a new user with same email should return conflict error", func() {
		res, err := s.client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
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

		_, err = s.client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
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
		_, err := s.client.UpdateUser(ctxOrgAdminAuth, &shieldv1beta1.UpdateUserRequest{
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
		_, err := s.client.UpdateCurrentUser(ctxCurrentUser, &shieldv1beta1.UpdateCurrentUserRequest{
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
		_, err := s.client.UpdateCurrentUser(ctxCurrentUser, &shieldv1beta1.UpdateCurrentUserRequest{
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
	suite.Run(t, new(EndToEndAPIRegressionTestSuite))
}
