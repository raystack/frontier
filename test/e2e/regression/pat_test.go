//go:build !race

package e2e_test

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	testusers "github.com/raystack/frontier/core/authenticate/test_users"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PATRegressionTestSuite struct {
	suite.Suite
	testBench   *testbench.TestBench
	adminCookie string
	roleIDs     map[string]string // role name -> UUID
}

func (s *PATRegressionTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().Nil(err)
	testDataPath := path.Join("file://", wd, fixturesDir)

	connectPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	appConfig := &config.Frontier{
		Log: logger.Config{
			Level: "error",
		},
		App: server.Config{
			Host:                "localhost",
			Connect:             server.ConnectConfig{Port: connectPort},
			ResourcesConfigPath: path.Join(testDataPath, "resource"),
			Authentication: authenticate.Config{
				Session: authenticate.SessionConfig{
					HashSecretKey:  "hash-secret-should-be-32-chars--",
					BlockSecretKey: "hash-secret-should-be-32-chars--",
					Validity:       time.Hour,
				},
				Token: authenticate.TokenConfig{
					RSAPath: "testdata/jwks.json",
					Issuer:  "frontier",
				},
				MailOTP: authenticate.MailOTPConfig{
					Subject:  "{{.Otp}}",
					Body:     "{{.Otp}}",
					Validity: 10 * time.Minute,
				},
				TestUsers: testusers.Config{Enabled: true, Domain: "raystack.org", OTP: testbench.TestOTP},
			},
			PAT: userpat.Config{Enabled: true, Prefix: "fpt", MaxPerUserPerOrg: 50, MaxLifetime: "8760h"},
		},
	}

	s.testBench, err = testbench.Init(appConfig)
	s.Require().NoError(err)

	ctx := context.Background()

	adminCookie, err := testbench.AuthenticateUser(ctx, s.testBench.Client, testbench.OrgAdminEmail)
	s.Require().NoError(err)
	s.adminCookie = adminCookie

	s.Require().NoError(testbench.BootstrapUsers(ctx, s.testBench.Client, adminCookie))
	s.Require().NoError(testbench.BootstrapOrganizations(ctx, s.testBench.Client, adminCookie))
	s.Require().NoError(testbench.BootstrapProject(ctx, s.testBench.Client, adminCookie))
	s.Require().NoError(testbench.BootstrapGroup(ctx, s.testBench.Client, adminCookie))

	// build role name → UUID map for PAT creation (requires UUIDs)
	ctxAdmin := testbench.ContextWithAuth(ctx, adminCookie)
	rolesResp, err := s.testBench.Client.ListRoles(ctxAdmin, connect.NewRequest(&frontierv1beta1.ListRolesRequest{}))
	s.Require().NoError(err)
	s.roleIDs = make(map[string]string, len(rolesResp.Msg.GetRoles()))
	for _, r := range rolesResp.Msg.GetRoles() {
		s.roleIDs[r.GetName()] = r.GetId()
	}
}

func (s *PATRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *PATRegressionTestSuite) roleID(name string) string {
	id, ok := s.roleIDs[name]
	s.Require().True(ok, "role %q not found in platform roles", name)
	return id
}

func getPATCtx(token string) context.Context {
	return testbench.ContextWithHeaders(context.Background(), map[string]string{
		"Authorization": "Bearer " + token,
	})
}

func (s *PATRegressionTestSuite) createOrgAndProjects(ctxAdmin context.Context, orgName, proj1Name, proj2Name string) (string, string, string) {
	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: orgName,
		},
	}))
	s.Require().NoError(err)
	orgID := createOrgResp.Msg.GetOrganization().GetId()

	proj1Resp, err := s.testBench.Client.CreateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Name:  proj1Name,
			OrgId: orgID,
		},
	}))
	s.Require().NoError(err)
	proj1ID := proj1Resp.Msg.GetProject().GetId()

	var proj2ID string
	if proj2Name != "" {
		proj2Resp, err := s.testBench.Client.CreateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Name:  proj2Name,
				OrgId: orgID,
			},
		}))
		s.Require().NoError(err)
		proj2ID = proj2Resp.Msg.GetProject().GetId()
	}

	return orgID, proj1ID, proj2ID
}

func (s *PATRegressionTestSuite) createPAT(ctxAdmin context.Context, orgID, title string, roleIDs, projectIDs []string) (string, string) {
	patResp, err := s.testBench.Client.CreateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
		Title:      title,
		OrgId:      orgID,
		RoleIds:    roleIDs,
		ProjectIds: projectIDs,
		ExpiresAt:  timestamppb.New(time.Now().Add(24 * time.Hour)),
	}))
	s.Require().NoError(err)
	s.Require().NotEmpty(patResp.Msg.GetPat().GetToken())
	return patResp.Msg.GetPat().GetId(), patResp.Msg.GetPat().GetToken()
}

func (s *PATRegressionTestSuite) checkPermission(ctx context.Context, namespace, id, permission string) bool {
	resp, err := s.testBench.Client.CheckResourcePermission(ctx, connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
		Resource:   schema.JoinNamespaceAndResourceID(namespace, id),
		Permission: permission,
	}))
	s.Require().NoError(err)
	return resp.Msg.GetStatus()
}

func (s *PATRegressionTestSuite) TestPATScope_OrgViewer_ProjectViewer() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, proj2ID := s.createOrgAndProjects(ctxAdmin, "org-pat-ov-pv", "pat-ov-pv-p1", "pat-ov-pv-p2")

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-ov-pv",
		[]string{s.roleID(schema.RoleOrganizationViewer), s.roleID(schema.RoleProjectViewer)},
		[]string{proj1ID},
	)
	patCtx := getPATCtx(patToken)

	s.Run("org get allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission))
	})
	s.Run("org update denied", func() {
		s.Assert().False(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.UpdatePermission))
	})
	s.Run("scoped project get allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.GetPermission))
	})
	s.Run("scoped project update denied", func() {
		s.Assert().False(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.UpdatePermission))
	})
	s.Run("unscoped project get denied", func() {
		s.Assert().False(s.checkPermission(patCtx, schema.ProjectNamespace, proj2ID, schema.GetPermission))
	})
	s.Run("batch check mixed results", func() {
		batchResp, err := s.testBench.Client.BatchCheckPermission(patCtx, connect.NewRequest(&frontierv1beta1.BatchCheckPermissionRequest{
			Bodies: []*frontierv1beta1.BatchCheckPermissionBody{
				{
					Resource:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, orgID),
					Permission: schema.GetPermission,
				},
				{
					Resource:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, orgID),
					Permission: schema.UpdatePermission,
				},
				{
					Resource:   schema.JoinNamespaceAndResourceID(schema.ProjectNamespace, proj1ID),
					Permission: schema.GetPermission,
				},
			},
		}))
		s.Require().NoError(err)
		pairs := batchResp.Msg.GetPairs()
		s.Require().Len(pairs, 3)
		s.Assert().True(pairs[0].GetStatus(), "org:get should be true")
		s.Assert().False(pairs[1].GetStatus(), "org:update should be false")
		s.Assert().True(pairs[2].GetStatus(), "proj1:get should be true")
	})
}

func (s *PATRegressionTestSuite) TestPATScope_OrgOwner() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, _ := s.createOrgAndProjects(ctxAdmin, "org-pat-oo", "pat-oo-p1", "")

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-oo",
		[]string{s.roleID(schema.RoleOrganizationOwner)},
		nil,
	)
	patCtx := getPATCtx(patToken)

	s.Run("org get allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission))
	})
	s.Run("org update allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.UpdatePermission))
	})
	s.Run("project get inherited from org owner", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.GetPermission))
	})
	s.Run("project update inherited from org owner", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.UpdatePermission))
	})
}

func (s *PATRegressionTestSuite) TestPATScope_OrgViewer_AllProjects() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, proj2ID := s.createOrgAndProjects(ctxAdmin, "org-pat-ov-ap", "pat-ov-ap-p1", "pat-ov-ap-p2")

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-ov-ap",
		[]string{s.roleID(schema.RoleOrganizationViewer), s.roleID(schema.RoleProjectOwner)},
		nil, // empty = all projects
	)
	patCtx := getPATCtx(patToken)

	s.Run("org get allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission))
	})
	s.Run("org update denied", func() {
		s.Assert().False(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.UpdatePermission))
	})
	s.Run("proj1 update allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.UpdatePermission))
	})
	s.Run("proj2 update allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj2ID, schema.UpdatePermission))
	})
}

func (s *PATRegressionTestSuite) TestPATScope_BillingManager() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, _ := s.createOrgAndProjects(ctxAdmin, "org-pat-bm", "pat-bm-p1", "")

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-bm",
		[]string{s.roleID("app_billing_manager")},
		nil,
	)
	patCtx := getPATCtx(patToken)

	s.Run("org billingview allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.BillingViewPermission))
	})
	s.Run("org billingmanage allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.BillingManagePermission))
	})
	s.Run("org get denied", func() {
		s.Assert().False(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission))
	})
	s.Run("org update denied", func() {
		s.Assert().False(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.UpdatePermission))
	})
	s.Run("project get denied", func() {
		s.Assert().False(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.GetPermission))
	})
}

func (s *PATRegressionTestSuite) TestPATScope_Interceptor() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: "org-pat-interceptor",
		},
	}))
	s.Require().NoError(err)
	orgID := createOrgResp.Msg.GetOrganization().GetId()

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-interceptor",
		[]string{s.roleID(schema.RoleOrganizationViewer)},
		nil,
	)
	patCtx := getPATCtx(patToken)

	// UpdateOrganization requires update permission — PAT only has viewer scope
	_, err = s.testBench.Client.UpdateOrganization(patCtx, connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
		Id: orgID,
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name:  "org-pat-interceptor",
			Title: "updated title",
		},
	}))
	s.Assert().Error(err)
	s.Assert().Equal(connect.CodePermissionDenied, connect.CodeOf(err))
}

func (s *PATRegressionTestSuite) TestPATScope_FederatedCheck() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Name: "org-pat-federated",
		},
	}))
	s.Require().NoError(err)
	orgID := createOrgResp.Msg.GetOrganization().GetId()

	patID, _ := s.createPAT(ctxAdmin, orgID, "pat-federated",
		[]string{s.roleID(schema.RoleOrganizationViewer)},
		nil,
	)

	patSubject := schema.JoinNamespaceAndResourceID(schema.PATPrincipal, patID)
	orgResource := schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, orgID)

	s.Run("federated check get allowed", func() {
		resp, err := s.testBench.AdminClient.CheckFederatedResourcePermission(ctxAdmin,
			connect.NewRequest(&frontierv1beta1.CheckFederatedResourcePermissionRequest{
				Subject:    patSubject,
				Resource:   orgResource,
				Permission: schema.GetPermission,
			}))
		s.Require().NoError(err)
		s.Assert().True(resp.Msg.GetStatus())
	})
	s.Run("federated check update denied", func() {
		resp, err := s.testBench.AdminClient.CheckFederatedResourcePermission(ctxAdmin,
			connect.NewRequest(&frontierv1beta1.CheckFederatedResourcePermissionRequest{
				Subject:    patSubject,
				Resource:   orgResource,
				Permission: schema.UpdatePermission,
			}))
		s.Require().NoError(err)
		s.Assert().False(resp.Msg.GetStatus())
	})
}

func TestEndToEndPATRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(PATRegressionTestSuite))
}
