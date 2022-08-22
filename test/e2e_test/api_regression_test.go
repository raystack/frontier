package e2e_test

import (
	"context"
	"fmt"
	"log"
	"os"
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
	s.Require().Nil(err)

	proxyPort, err := GetFreePort()
	s.Require().Nil(err)

	apiPort, err := GetFreePort()
	s.Require().Nil(err)

	appConfig := &config.Shield{
		Log: logger.Config{
			Level: "error",
		},
		App: server.Config{
			Port:                apiPort,
			IdentityProxyHeader: identityHeader,
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

	s.testBench, err = testbench.Init(appConfig)
	s.Require().Nil(err)

	ctx := context.Background()
	s.client, s.cancelClient, err = createClient(ctx, fmt.Sprintf("localhost:%d", apiPort))
	s.Require().Nil(err)

	s.Require().Nil(bootstrapUser(ctx, s.client, orgAdminEmail))
	s.Require().Nil(bootstrapOrganization(ctx, s.client, orgAdminEmail))
	s.Require().Nil(bootstrapProject(ctx, s.client, orgAdminEmail))
	s.Require().Nil(bootstrapGroup(ctx, s.client, orgAdminEmail))

	// validate
	uRes, err := s.client.ListUsers(ctx, &shieldv1beta1.ListUsersRequest{})
	s.Require().Nil(err)
	s.Require().Equal(9, len(uRes.GetUsers()))

	oRes, err := s.client.ListOrganizations(ctx, &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().Nil(err)
	s.Require().Equal(1, len(oRes.GetOrganizations()))

	pRes, err := s.client.ListProjects(ctx, &shieldv1beta1.ListProjectsRequest{})
	s.Require().Nil(err)
	s.Require().Equal(1, len(pRes.GetProjects()))

	gRes, err := s.client.ListGroups(ctx, &shieldv1beta1.ListGroupsRequest{})
	s.Require().Nil(err)
	s.Require().Equal(3, len(gRes.GetGroups()))
}

func (s *EndToEndAPIRegressionTestSuite) TearDownTest() {
	s.cancelClient()
	// Clean tests
	if err := s.testBench.CleanUp(); err != nil {
		log.Fatal(err)
	}
}

func (s *EndToEndAPIRegressionTestSuite) TestGroupAPI() {
	var newGroup *shieldv1beta1.Group

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		identityHeader: orgAdminEmail,
	}))

	// get my org
	loRes, err := s.client.ListOrganizations(context.Background(), &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(loRes.GetOrganizations()), 0)
	myOrg := loRes.GetOrganizations()[0]

	s.Run("1. org admin create a new team with empty auth email should return unauthenticated error", func() {
		_, err := s.client.CreateGroup(context.Background(), &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	// org admin is currently the group admin
	s.Run("2. group admin create a new team with empty name should return invalid argument", func() {
		_, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. group admin create a new team with wrong org id should return invalid argument", func() {
		_, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. group admin create a new team with same name and org-id should conflict", func() {
		cgRes, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newGroup = cgRes.GetGroup()
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
		identityHeader: orgAdminEmail,
	}))

	// get my org
	loRes, err := s.client.ListOrganizations(context.Background(), &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(loRes.GetOrganizations()), 0)
	myOrg := loRes.GetOrganizations()[0]

	s.Run("1. org admin create a new user with empty auth email should return unauthenticated error", func() {
		_, err := s.client.CreateUser(context.Background(), &shieldv1beta1.CreateUserRequest{
			Body: &shieldv1beta1.UserRequestBody{
				Name:  "new user a",
				Email: "new-user-a@odpf.ipo",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewBoolValue(true),
					},
				},
			},
		})
		s.Assert().Equal(codes.Unauthenticated, status.Convert(err).Code())
	})

	// org admin is currently the group admin
	s.Run("2. group admin create a new team with empty name should return invalid argument", func() {
		_, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("3. group admin create a new team with wrong org id should return invalid argument", func() {
		_, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: "not-uuid",
			},
		})
		s.Assert().Equal(codes.InvalidArgument, status.Convert(err).Code())
	})

	s.Run("4. group admin create a new team with same name and org-id should conflict", func() {
		cgRes, err := s.client.CreateGroup(ctxOrgAdminAuth, &shieldv1beta1.CreateGroupRequest{
			Body: &shieldv1beta1.GroupRequestBody{
				Name:  "new group",
				Slug:  "new-group",
				OrgId: myOrg.GetId(),
			},
		})
		s.Assert().NoError(err)
		newGroup = cgRes.GetGroup()
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

func TestEndToEndAPIRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(EndToEndAPIRegressionTestSuite))
}
