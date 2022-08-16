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
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/odpf/shield/test/e2e_test/testbench"
	"github.com/stretchr/testify/suite"
)

type EndToEndAPITestSuite struct {
	suite.Suite
	client       shieldv1beta1.ShieldServiceClient
	cancelClient func()
	testBench    *testbench.TestBench
}

func (s *EndToEndAPITestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)

	proxyPort, err := GetFreePort()
	s.Require().Nil(err)

	apiPort, err := GetFreePort()
	s.Require().Nil(err)

	appConfig := &config.Shield{
		App: server.Config{
			Port:                apiPort,
			IdentityProxyHeader: "X-Shield-Email",
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

	s.Require().Nil(bootstrapUser(ctx, s.client))
	s.Require().Nil(bootstrapOrganization(ctx, s.client))
	s.Require().Nil(bootstrapProject(ctx, s.client))
	s.Require().Nil(bootstrapGroup(ctx, s.client))

	//validate
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

func (s *EndToEndAPITestSuite) TearDownSuite() {
	s.cancelClient()
	// Clean tests
	if err := s.testBench.CleanUp(); err != nil {
		log.Fatal(err)
	}
}

func (s *EndToEndAPITestSuite) TestEntityCreation() {
	// Test goes here
}

func TestEndToEndAPITestSuite(t *testing.T) {
	suite.Run(t, new(EndToEndAPITestSuite))
}
