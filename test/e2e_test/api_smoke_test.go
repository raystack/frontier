package e2e_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/pkg/logger"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/odpf/shield/test/e2e_test/testbench"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	orgAdminEmail  = "admin1-group1-org1@odpf.io"
	identityHeader = "X-Shield-Email"
)

type EndToEndAPISmokeTestSuite struct {
	suite.Suite
	client       shieldv1beta1.ShieldServiceClient
	cancelClient func()
	testBench    *testbench.TestBench
}

func (s *EndToEndAPISmokeTestSuite) SetupTest() {
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

func (s *EndToEndAPISmokeTestSuite) TearDownTest() {
	s.cancelClient()
	// Clean tests
	if err := s.testBench.CleanUp(); err != nil {
		log.Fatal(err)
	}
}

func (s *EndToEndAPISmokeTestSuite) TestSmokeTestAdmin() {
	var (
		listOfUsers  []*shieldv1beta1.User
		listOfGroups []*shieldv1beta1.Group
		mySelf       *shieldv1beta1.User
		newGroup     *shieldv1beta1.Group
		group1       *shieldv1beta1.Group
	)

	// sleep needed to compensate transaction done in spice db
	time.Sleep(5 * time.Second)

	// Validation and Preparation
	// get list of users
	luRes, err := s.client.ListUsers(context.Background(), &shieldv1beta1.ListUsersRequest{})
	s.Require().NoError(err)
	listOfUsers = luRes.GetUsers()

	// get list of groups
	lgRes, err := s.client.ListGroups(context.Background(), &shieldv1beta1.ListGroupsRequest{})
	s.Require().NoError(err)
	listOfGroups = lgRes.GetGroups()

	// get my self
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		identityHeader: orgAdminEmail,
	}))
	gcuRes, err := s.client.GetCurrentUser(ctxOrgAdminAuth, &shieldv1beta1.GetCurrentUserRequest{})
	s.Require().NoError(err)
	mySelf = gcuRes.GetUser()

	// get my org
	loRes, err := s.client.ListOrganizations(context.Background(), &shieldv1beta1.ListOrganizationsRequest{})
	s.Require().NoError(err)
	s.Require().Greater(len(loRes.GetOrganizations()), 0)
	myOrg := loRes.GetOrganizations()[0]

	// Verify org admin
	loaRes, err := s.client.ListOrganizationAdmins(context.Background(), &shieldv1beta1.ListOrganizationAdminsRequest{
		Id: myOrg.Id,
	})
	s.Require().NoError(err)
	s.Require().Contains(loaRes.GetUsers(), mySelf)

	s.Run("1. org admin could create a new team", func() {
		// check permission
		_, err := s.client.CheckResourcePermission(ctxOrgAdminAuth, &shieldv1beta1.ResourceActionAuthzRequest{
			ResourceId:  myOrg.GetId(),
			ActionId:    action.DefinitionCreateTeam.ID,
			NamespaceId: namespace.DefinitionOrg.ID,
		})
		s.Assert().NoError(err)

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
	})

	var user1 *shieldv1beta1.User
	s.Run("2. group admin could add member to team", func() {
		// Get a random group that is not a new group
		for _, g := range listOfGroups {
			if g.GetId() != newGroup.GetId() {
				group1 = g
				break
			}
		}

		// check permission
		_, err := s.client.CheckResourcePermission(ctxOrgAdminAuth, &shieldv1beta1.ResourceActionAuthzRequest{
			ResourceId:  group1.GetId(),
			ActionId:    action.DefinitionManageTeam.ID,
			NamespaceId: namespace.DefinitionTeam.ID,
		})
		s.Assert().NoError(err)

		// verify number of users
		lguRes, err := s.client.ListGroupUsers(ctxOrgAdminAuth, &shieldv1beta1.ListGroupUsersRequest{
			Id: group1.GetId(),
		})
		s.Require().NoError(err)
		s.Require().Len(lguRes.GetUsers(), 1)

		// Get a random user that is not me
		for _, u := range listOfUsers {
			if u.GetId() != mySelf.GetId() {
				user1 = u
				break
			}
		}

		// Verify group admin
		lgaRes, err := s.client.ListGroupAdmins(context.Background(), &shieldv1beta1.ListGroupAdminsRequest{
			Id: group1.GetId(),
		})
		s.Require().NoError(err)
		s.Require().Contains(lgaRes.GetUsers(), mySelf)

		aguRes, err := s.client.AddGroupUser(ctxOrgAdminAuth, &shieldv1beta1.AddGroupUserRequest{
			Id: group1.GetId(),
			Body: &shieldv1beta1.AddGroupUserRequestBody{
				UserIds: []string{user1.GetId()},
			},
		})
		s.Assert().NoError(err)
		s.Assert().Len(aguRes.GetUsers(), 2)
	})

	s.Run("3. group admin could make a member an admin", func() {
		// verify number of admins
		lgaRes, err := s.client.ListGroupAdmins(ctxOrgAdminAuth, &shieldv1beta1.ListGroupAdminsRequest{
			Id: group1.GetId(),
		})
		s.Require().NoError(err)
		s.Require().Greater(len(lgaRes.GetUsers()), 0)
		numAdmin := len(lgaRes.GetUsers())

		agaRes, err := s.client.AddGroupAdmin(ctxOrgAdminAuth, &shieldv1beta1.AddGroupAdminRequest{
			Id: group1.GetId(),
			Body: &shieldv1beta1.AddGroupAdminRequestBody{
				UserIds: []string{user1.GetId()},
			},
		})
		s.Assert().NoError(err)
		s.Assert().Len(agaRes.GetUsers(), numAdmin+1)
	})

	s.Run("4. group admin could remove admin role of a member", func() {
		// verify number of admins
		lgaRes, err := s.client.ListGroupAdmins(ctxOrgAdminAuth, &shieldv1beta1.ListGroupAdminsRequest{
			Id: group1.GetId(),
		})
		s.Require().NoError(err)
		s.Require().Greater(len(lgaRes.GetUsers()), 0)
		numAdmin := len(lgaRes.GetUsers())

		_, err = s.client.RemoveGroupAdmin(ctxOrgAdminAuth, &shieldv1beta1.RemoveGroupAdminRequest{
			Id:     group1.GetId(),
			UserId: user1.GetId(),
		})
		s.Require().NoError(err)

		// verify number of admins
		lgaRes, err = s.client.ListGroupAdmins(ctxOrgAdminAuth, &shieldv1beta1.ListGroupAdminsRequest{
			Id: group1.GetId(),
		})
		s.Assert().NoError(err)
		s.Assert().Len(lgaRes.GetUsers(), numAdmin-1)
	})

	s.Run("5. group admin could remove member from a team", func() {
		// verify number of users
		lguRes, err := s.client.ListGroupUsers(ctxOrgAdminAuth, &shieldv1beta1.ListGroupUsersRequest{
			Id: group1.GetId(),
		})
		s.Require().NoError(err)
		s.Require().Greater(len(lguRes.GetUsers()), 0)
		numMembers := len(lguRes.GetUsers())

		_, err = s.client.RemoveGroupUser(ctxOrgAdminAuth, &shieldv1beta1.RemoveGroupUserRequest{
			Id:     group1.GetId(),
			UserId: user1.GetId(),
		})
		s.Assert().NoError(err)

		// verify number of users
		lguRes, err = s.client.ListGroupUsers(ctxOrgAdminAuth, &shieldv1beta1.ListGroupUsersRequest{
			Id: group1.GetId(),
		})
		s.Require().NoError(err)
		s.Require().Len(lguRes.GetUsers(), numMembers-1)
	})

	s.Run("6. group admin could add same user in multiple teams", func() {
		// verify number of users
		lguRes, err := s.client.ListGroupUsers(ctxOrgAdminAuth, &shieldv1beta1.ListGroupUsersRequest{
			Id: group1.GetId(),
		})
		s.Require().NoError(err)
		s.Require().Greater(len(lguRes.GetUsers()), 0)

		// Get a random user that is not me
		for _, g := range listOfGroups {
			_, err := s.client.AddGroupUser(ctxOrgAdminAuth, &shieldv1beta1.AddGroupUserRequest{
				Id: g.GetId(),
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: []string{user1.GetId()},
				},
			})
			s.Assert().NoError(err)
		}
	})
}

func (s *EndToEndAPISmokeTestSuite) TestSmokeTestMember() {
	var (
		members      []*shieldv1beta1.User
		listOfGroups []*shieldv1beta1.Group
	)

	// sleep needed to compensate transaction done in spice db
	time.Sleep(5 * time.Second)

	// get list of users
	luRes, err := s.client.ListUsers(context.Background(), &shieldv1beta1.ListUsersRequest{})
	s.Require().NoError(err)
	members = luRes.GetUsers()

	// get list of groups
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		identityHeader: orgAdminEmail,
	}))
	lgRes, err := s.client.ListGroups(ctxOrgAdminAuth, &shieldv1beta1.ListGroupsRequest{})
	s.Require().NoError(err)
	listOfGroups = lgRes.GetGroups()

	// Remove admin user from list
	adminIdx := 0
	for _, u := range members {
		adminIdx = 0
		if u.GetEmail() == orgAdminEmail {
			break
		}
		adminIdx = adminIdx + 1
	}
	members = append(members[:adminIdx], members[adminIdx+1:]...)
	s.Require().Greater(len(members), 0)

	ctxMemberAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		identityHeader: members[0].GetEmail(),
	}))

	s.Run("1. member unable to add member to team", func() {
		// check permission
		_, err := s.client.CheckResourcePermission(ctxMemberAuth, &shieldv1beta1.ResourceActionAuthzRequest{
			ResourceId:  listOfGroups[0].GetId(),
			ActionId:    action.DefinitionManageTeam.ID,
			NamespaceId: namespace.DefinitionTeam.ID,
		})
		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())

		_, err = s.client.AddGroupUser(ctxMemberAuth, &shieldv1beta1.AddGroupUserRequest{
			Id: listOfGroups[0].GetId(),
			Body: &shieldv1beta1.AddGroupUserRequestBody{
				UserIds: []string{members[0].GetId()},
			},
		})
		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())
	})

	s.Run("2. member unable to add admin to team", func() {
		// check permission
		_, err := s.client.CheckResourcePermission(ctxMemberAuth, &shieldv1beta1.ResourceActionAuthzRequest{
			ResourceId:  listOfGroups[0].GetId(),
			ActionId:    action.DefinitionManageTeam.ID,
			NamespaceId: namespace.DefinitionTeam.ID,
		})
		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())

		_, err = s.client.AddGroupAdmin(ctxMemberAuth, &shieldv1beta1.AddGroupAdminRequest{
			Id: listOfGroups[0].GetId(),
			Body: &shieldv1beta1.AddGroupAdminRequestBody{
				UserIds: []string{members[0].GetId()},
			},
		})
		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())
	})

	s.Run("3. member unable to remove admin from team", func() {
		// check permission
		_, err := s.client.CheckResourcePermission(ctxMemberAuth, &shieldv1beta1.ResourceActionAuthzRequest{
			ResourceId:  listOfGroups[0].GetId(),
			ActionId:    action.DefinitionManageTeam.ID,
			NamespaceId: namespace.DefinitionTeam.ID,
		})

		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())

		_, err = s.client.RemoveGroupAdmin(ctxMemberAuth, &shieldv1beta1.RemoveGroupAdminRequest{
			Id:     listOfGroups[0].GetId(),
			UserId: members[0].GetId(),
		})
		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())
	})

	s.Run("4. member unable to remove member from team", func() {
		// check permission
		_, err := s.client.CheckResourcePermission(ctxMemberAuth, &shieldv1beta1.ResourceActionAuthzRequest{
			ResourceId:  listOfGroups[0].GetId(),
			ActionId:    action.DefinitionManageTeam.ID,
			NamespaceId: namespace.DefinitionTeam.ID,
		})
		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())

		_, err = s.client.RemoveGroupUser(ctxMemberAuth, &shieldv1beta1.RemoveGroupUserRequest{
			Id:     listOfGroups[0].GetId(),
			UserId: members[0].GetId(),
		})
		s.Assert().Equal(codes.PermissionDenied, status.Convert(err).Code())
	})
}

// TODO proxy test
// - member who does not belong to any team cannot access resources
// - could not access resource in different org
// - Calling create upstream resource should create a resource in shield DB
// - Team member can access the created resource created by the same team
// - Team admin can access the created resource created by the same team
// - Org admin who is not team member or admin or creator can access the created resource in the same org
// - Project admin who is not team member or admin or creator can access the created resource in the same project

func TestEndToEndAPISmokeTestSuite(t *testing.T) {
	suite.Run(t, new(EndToEndAPISmokeTestSuite))
}
