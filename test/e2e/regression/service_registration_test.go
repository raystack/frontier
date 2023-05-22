package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/pkg/logger"
	"github.com/odpf/shield/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type ServiceRegistrationRegressionTestSuite struct {
	suite.Suite
	testBench *testbench.TestBench
}

func (s *ServiceRegistrationRegressionTestSuite) SetupSuite() {
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

func (s *ServiceRegistrationRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *ServiceRegistrationRegressionTestSuite) TestServiceRegistration() {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		testbench.IdentityHeader: testbench.OrgAdminEmail,
	}))

	s.Run("1. register a new service with custom permissions", func() {
		createPermResp, err := s.testBench.AdminClient.CreatePermission(ctx, &shieldv1beta1.CreatePermissionRequest{
			Bodies: []*shieldv1beta1.PermissionRequestBody{
				{
					Name:      "get",
					Namespace: "database/instance",
					Title:     "",
				},
				{
					Name:      "update",
					Namespace: "database/instance",
					Title:     "update db instance",
				},
				{
					Name:      "delete",
					Namespace: "database/instance",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"description": structpb.NewStringValue("bar"),
						},
					},
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().Equal(3, len(createPermResp.GetPermissions()))

		listPermResp, err := s.testBench.Client.ListPermissions(ctx, &shieldv1beta1.ListPermissionsRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(listPermResp.GetPermissions())
		// check if list contains newly created permissions
		for _, perm := range createPermResp.GetPermissions() {
			s.Assert().Contains(listPermResp.GetPermissions(), perm)
		}
		// length of list should be greater than number of permissions created
		s.Assert().GreaterOrEqual(len(listPermResp.GetPermissions()), len(createPermResp.GetPermissions()))
	})
	s.Run("2. registering a new service should not remove existing permissions", func() {
		createPermResp, err := s.testBench.AdminClient.CreatePermission(ctx, &shieldv1beta1.CreatePermissionRequest{
			Bodies: []*shieldv1beta1.PermissionRequestBody{
				{
					Name:      "update",
					Namespace: "database/alert",
					Title:     "update db alert",
				},
				{
					Name:      "delete",
					Namespace: "database/alert",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"description": structpb.NewStringValue("bar"),
						},
					},
				},
			},
		})
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(createPermResp.GetPermissions()))

		listPermResp, err := s.testBench.Client.ListPermissions(ctx, &shieldv1beta1.ListPermissionsRequest{})
		s.Assert().NoError(err)
		s.Assert().NotNil(listPermResp.GetPermissions())
		// check if list contains newly created permissions
		for _, perm := range createPermResp.GetPermissions() {
			s.Assert().Contains(listPermResp.GetPermissions(), perm)
		}
		// list should contain permissions created in previous step
		var lastPermCount int
		for _, perm := range []string{"get", "update", "delete"} {
			for _, listPerm := range listPermResp.GetPermissions() {
				if listPerm.Name == perm && listPerm.Namespace == "database/instance" {
					lastPermCount++
				}
			}
		}
		s.Assert().Equal(3, lastPermCount)
	})
}

func TestEndToEndServiceRegistrationRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceRegistrationRegressionTestSuite))
}
