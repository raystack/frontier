package e2e_test

import (
	"context"
	"fmt"
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
			PAT: userpat.Config{Enabled: true, Prefix: "fpt", MaxPerUserPerOrg: 50, MaxLifetime: "8760h", DeniedPermissions: []string{"app_organization_administer"}},
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

func (s *PATRegressionTestSuite) createPAT(ctxAdmin context.Context, orgID, title string, scopes []*frontierv1beta1.PATScope) (string, string) {
	patResp, err := s.testBench.Client.CreateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
		Title:     title,
		OrgId:     orgID,
		Scopes:    scopes,
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
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

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-ov-pv", []*frontierv1beta1.PATScope{
		{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
		{RoleId: s.roleID(schema.RoleProjectViewer), ResourceType: schema.ProjectNamespace, ResourceIds: []string{proj1ID}},
	})
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

func (s *PATRegressionTestSuite) TestPATScope_OrgManager() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, _ := s.createOrgAndProjects(ctxAdmin, "org-pat-om", "pat-om-p1", "")

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-om", []*frontierv1beta1.PATScope{
		{RoleId: s.roleID(schema.RoleOrganizationManager), ResourceType: schema.OrganizationNamespace},
	})
	patCtx := getPATCtx(patToken)

	s.Run("org get allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission))
	})
	s.Run("org update allowed", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.UpdatePermission))
	})
	s.Run("project get inherited from org manager", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.GetPermission))
	})
	s.Run("project update inherited from org manager", func() {
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.UpdatePermission))
	})
}

func (s *PATRegressionTestSuite) TestPATScope_DeniedRole() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{Name: "org-pat-denied-role"},
	}))
	s.Require().NoError(err)
	orgID := createOrgResp.Msg.GetOrganization().GetId()

	s.Run("org_owner role is denied", func() {
		_, err := s.testBench.Client.CreateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
			Title: "denied-owner-pat",
			OrgId: orgID,
			Scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationOwner), ResourceType: schema.OrganizationNamespace},
			},
			ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		}))
		s.Assert().Error(err)
		s.Assert().Equal(connect.CodeInvalidArgument, connect.CodeOf(err))
	})
}

func (s *PATRegressionTestSuite) TestPATScope_OrgViewer_AllProjects() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, proj2ID := s.createOrgAndProjects(ctxAdmin, "org-pat-ov-ap", "pat-ov-ap-p1", "pat-ov-ap-p2")

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-ov-ap", []*frontierv1beta1.PATScope{
		{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
		{RoleId: s.roleID(schema.RoleProjectOwner), ResourceType: schema.ProjectNamespace}, // empty resource_ids = all projects
	})
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

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-bm", []*frontierv1beta1.PATScope{
		{RoleId: s.roleID("app_billing_manager"), ResourceType: schema.OrganizationNamespace},
	})
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

	_, patToken := s.createPAT(ctxAdmin, orgID, "pat-interceptor", []*frontierv1beta1.PATScope{
		{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
	})
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

	patID, _ := s.createPAT(ctxAdmin, orgID, "pat-federated", []*frontierv1beta1.PATScope{
		{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
	})

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

func (s *PATRegressionTestSuite) TestPATScope_RoleMatrix() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, proj2ID := s.createOrgAndProjects(ctxAdmin, "org-pat-matrix", "pat-matrix-p1", "pat-matrix-p2")

	type permCheck struct {
		namespace  string
		resourceID string
		permission string
		expected   bool
		label      string
	}

	tests := []struct {
		name   string
		scopes []*frontierv1beta1.PATScope
		checks []permCheck
	}{
		{
			name: "org_viewer only",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
			},
			checks: []permCheck{
				{schema.OrganizationNamespace, orgID, schema.GetPermission, true, "org:get"},
				{schema.OrganizationNamespace, orgID, schema.UpdatePermission, false, "org:update"},
				{schema.OrganizationNamespace, orgID, schema.DeletePermission, false, "org:delete"},
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, false, "proj1:get (no project scope)"},
			},
		},
		{
			name: "org_manager only",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationManager), ResourceType: schema.OrganizationNamespace},
			},
			checks: []permCheck{
				{schema.OrganizationNamespace, orgID, schema.GetPermission, true, "org:get"},
				{schema.OrganizationNamespace, orgID, schema.UpdatePermission, true, "org:update"},
				{schema.OrganizationNamespace, orgID, schema.DeletePermission, false, "org:delete"},
				// org manager inherits project get/update via synthetic permissions, but not delete
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, true, "proj1:get (inherited from org manager)"},
				{schema.ProjectNamespace, proj1ID, schema.UpdatePermission, true, "proj1:update (inherited from org manager)"},
				{schema.ProjectNamespace, proj1ID, schema.DeletePermission, false, "proj1:delete (manager cannot delete)"},
			},
		},
		{
			name: "org_viewer + project_viewer (specific project)",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
				{RoleId: s.roleID(schema.RoleProjectViewer), ResourceType: schema.ProjectNamespace, ResourceIds: []string{proj1ID}},
			},
			checks: []permCheck{
				{schema.OrganizationNamespace, orgID, schema.GetPermission, true, "org:get"},
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, true, "proj1:get (scoped)"},
				{schema.ProjectNamespace, proj1ID, schema.UpdatePermission, false, "proj1:update (viewer)"},
				{schema.ProjectNamespace, proj2ID, schema.GetPermission, false, "proj2:get (not in scope)"},
			},
		},
		{
			name: "org_viewer + project_manager (specific project)",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
				{RoleId: s.roleID(schema.RoleProjectManager), ResourceType: schema.ProjectNamespace, ResourceIds: []string{proj1ID}},
			},
			checks: []permCheck{
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, true, "proj1:get"},
				{schema.ProjectNamespace, proj1ID, schema.UpdatePermission, true, "proj1:update (manager)"},
				{schema.ProjectNamespace, proj1ID, schema.DeletePermission, false, "proj1:delete (manager, not owner)"},
				{schema.ProjectNamespace, proj2ID, schema.GetPermission, false, "proj2:get (not in scope)"},
			},
		},
		{
			name: "org_viewer + project_owner (specific project)",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
				{RoleId: s.roleID(schema.RoleProjectOwner), ResourceType: schema.ProjectNamespace, ResourceIds: []string{proj1ID}},
			},
			checks: []permCheck{
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, true, "proj1:get"},
				{schema.ProjectNamespace, proj1ID, schema.UpdatePermission, true, "proj1:update"},
				{schema.ProjectNamespace, proj1ID, schema.DeletePermission, true, "proj1:delete (owner)"},
				{schema.ProjectNamespace, proj2ID, schema.GetPermission, false, "proj2:get (not in scope)"},
			},
		},
		{
			name: "org_viewer + project_viewer (all projects)",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
				{RoleId: s.roleID(schema.RoleProjectViewer), ResourceType: schema.ProjectNamespace},
			},
			checks: []permCheck{
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, true, "proj1:get (all projects)"},
				{schema.ProjectNamespace, proj2ID, schema.GetPermission, true, "proj2:get (all projects)"},
				{schema.ProjectNamespace, proj1ID, schema.UpdatePermission, false, "proj1:update (viewer)"},
			},
		},
		{
			name: "org_viewer + project_owner (all projects)",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
				{RoleId: s.roleID(schema.RoleProjectOwner), ResourceType: schema.ProjectNamespace},
			},
			checks: []permCheck{
				{schema.ProjectNamespace, proj1ID, schema.DeletePermission, true, "proj1:delete (owner, all projects)"},
				{schema.ProjectNamespace, proj2ID, schema.DeletePermission, true, "proj2:delete (owner, all projects)"},
			},
		},
		{
			name: "billing_manager only",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID("app_billing_manager"), ResourceType: schema.OrganizationNamespace},
			},
			checks: []permCheck{
				{schema.OrganizationNamespace, orgID, schema.BillingViewPermission, true, "org:billingview"},
				{schema.OrganizationNamespace, orgID, schema.BillingManagePermission, true, "org:billingmanage"},
				{schema.OrganizationNamespace, orgID, schema.GetPermission, false, "org:get (no org access)"},
				{schema.OrganizationNamespace, orgID, schema.UpdatePermission, false, "org:update (no org access)"},
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, false, "proj1:get (no project access)"},
			},
		},
		{
			name: "org_manager + project_viewer (specific project)",
			scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationManager), ResourceType: schema.OrganizationNamespace},
				{RoleId: s.roleID(schema.RoleProjectViewer), ResourceType: schema.ProjectNamespace, ResourceIds: []string{proj1ID}},
			},
			checks: []permCheck{
				{schema.OrganizationNamespace, orgID, schema.UpdatePermission, true, "org:update (manager)"},
				{schema.ProjectNamespace, proj1ID, schema.GetPermission, true, "proj1:get"},
				// org manager inherits project get/update via synthetic permissions (not delete)
				{schema.ProjectNamespace, proj1ID, schema.UpdatePermission, true, "proj1:update (inherited from org manager)"},
				{schema.ProjectNamespace, proj2ID, schema.GetPermission, true, "proj2:get (inherited from org manager)"},
			},
		},
	}

	for i, tt := range tests {
		s.Run(tt.name, func() {
			title := fmt.Sprintf("matrix-pat-%d", i)
			_, patToken := s.createPAT(ctxAdmin, orgID, title, tt.scopes)
			patCtx := getPATCtx(patToken)

			for _, check := range tt.checks {
				result := s.checkPermission(patCtx, check.namespace, check.resourceID, check.permission)
				s.Assert().Equal(check.expected, result, check.label)
			}
		})
	}
}

func (s *PATRegressionTestSuite) TestPATCRUD_Lifecycle() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, proj1ID, _ := s.createOrgAndProjects(ctxAdmin, "org-pat-crud", "pat-crud-p1", "pat-crud-p2")

	var patID, patToken string

	s.Run("ListRolesForPAT returns roles excluding denied", func() {
		resp, err := s.testBench.Client.ListRolesForPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.ListRolesForPATRequest{
			Scopes: []string{schema.OrganizationNamespace, schema.ProjectNamespace},
		}))
		s.Require().NoError(err)
		s.Assert().NotEmpty(resp.Msg.GetRoles())
		// org_owner has app_organization_administer which is in DeniedPermissions
		for _, r := range resp.Msg.GetRoles() {
			s.Assert().NotEqual(schema.RoleOrganizationOwner, r.GetName(),
				"denied roles should be excluded from ListRolesForPAT")
		}
	})

	s.Run("CheckTitle available", func() {
		resp, err := s.testBench.Client.CheckCurrentUserPATTitle(ctxAdmin, connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
			OrgId: orgID,
			Title: "unique-crud-pat",
		}))
		s.Require().NoError(err)
		s.Assert().True(resp.Msg.GetAvailable())
	})

	s.Run("Create PAT", func() {
		resp, err := s.testBench.Client.CreateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
			Title: "unique-crud-pat",
			OrgId: orgID,
			Scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
				{RoleId: s.roleID(schema.RoleProjectViewer), ResourceType: schema.ProjectNamespace, ResourceIds: []string{proj1ID}},
			},
			ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		}))
		s.Require().NoError(err)
		pat := resp.Msg.GetPat()
		s.Assert().NotEmpty(pat.GetId())
		s.Assert().NotEmpty(pat.GetToken())
		s.Assert().Equal("unique-crud-pat", pat.GetTitle())
		s.Assert().Equal(orgID, pat.GetOrgId())
		s.Assert().Len(pat.GetScopes(), 2)
		s.Assert().Nil(pat.GetUsedAt())
		s.Assert().Nil(pat.GetRegeneratedAt())
		patID = pat.GetId()
		patToken = pat.GetToken()
	})

	s.Run("CheckTitle taken after create", func() {
		resp, err := s.testBench.Client.CheckCurrentUserPATTitle(ctxAdmin, connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
			OrgId: orgID,
			Title: "unique-crud-pat",
		}))
		s.Require().NoError(err)
		s.Assert().False(resp.Msg.GetAvailable())
	})

	s.Run("Get PAT", func() {
		resp, err := s.testBench.Client.GetCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
			Id: patID,
		}))
		s.Require().NoError(err)
		pat := resp.Msg.GetPat()
		s.Assert().Equal(patID, pat.GetId())
		s.Assert().Equal("unique-crud-pat", pat.GetTitle())
		s.Assert().Empty(pat.GetToken(), "token should not be returned on get")
		s.Assert().Len(pat.GetScopes(), 2)
	})

	s.Run("Search PATs returns created PAT", func() {
		resp, err := s.testBench.Client.SearchCurrentUserPATs(ctxAdmin, connect.NewRequest(&frontierv1beta1.SearchCurrentUserPATsRequest{
			OrgId: orgID,
		}))
		s.Require().NoError(err)
		s.Assert().GreaterOrEqual(len(resp.Msg.GetPats()), 1)
		found := false
		for _, pat := range resp.Msg.GetPats() {
			if pat.GetId() == patID {
				found = true
				s.Assert().Equal("unique-crud-pat", pat.GetTitle())
			}
		}
		s.Assert().True(found, "created PAT should appear in search results")
	})

	s.Run("PAT token authenticates and updates used_at", func() {
		patCtx := getPATCtx(patToken)
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission))

		// get the PAT again to verify used_at is now set
		resp, err := s.testBench.Client.GetCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
			Id: patID,
		}))
		s.Require().NoError(err)
		s.Assert().NotNil(resp.Msg.GetPat().GetUsedAt(), "used_at should be set after use")
	})

	s.Run("Verify permissions before update", func() {
		patCtx := getPATCtx(patToken)
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission), "org get should work")
		s.Assert().True(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.GetPermission), "project get should work before scope narrowing")
	})

	s.Run("Update PAT title and narrow scopes", func() {
		resp, err := s.testBench.Client.UpdateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateCurrentUserPATRequest{
			Id:    patID,
			Title: "updated-crud-pat",
			Scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
			},
		}))
		s.Require().NoError(err)
		pat := resp.Msg.GetPat()
		s.Assert().Equal("updated-crud-pat", pat.GetTitle())
		s.Assert().Len(pat.GetScopes(), 1, "scopes should be narrowed to org only")
	})

	s.Run("Verify permissions after narrowing scopes", func() {
		patCtx := getPATCtx(patToken)
		s.Assert().True(s.checkPermission(patCtx, schema.OrganizationNamespace, orgID, schema.GetPermission), "org get should still work")
		s.Assert().False(s.checkPermission(patCtx, schema.ProjectNamespace, proj1ID, schema.GetPermission), "project get should be denied after removing project scope")
	})

	s.Run("Regenerate PAT", func() {
		resp, err := s.testBench.Client.RegenerateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.RegenerateCurrentUserPATRequest{
			Id:        patID,
			ExpiresAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
		}))
		s.Require().NoError(err)
		pat := resp.Msg.GetPat()
		s.Assert().Equal(patID, pat.GetId())
		s.Assert().NotEmpty(pat.GetToken(), "new token should be returned")
		s.Assert().NotEqual(patToken, pat.GetToken(), "token should be different after regenerate")
		s.Assert().NotNil(pat.GetRegeneratedAt(), "regenerated_at should be set")

		// old token should no longer work
		oldPatCtx := getPATCtx(patToken)
		_, err = s.testBench.Client.CheckResourcePermission(oldPatCtx, connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, orgID),
			Permission: schema.GetPermission,
		}))
		s.Assert().Error(err, "old token should be rejected after regenerate")

		patToken = pat.GetToken()

		// verify new token works
		newPatCtx := getPATCtx(patToken)
		s.Assert().True(s.checkPermission(newPatCtx, schema.OrganizationNamespace, orgID, schema.GetPermission), "new token should work after regenerate")
	})

	s.Run("Delete PAT", func() {
		_, err := s.testBench.Client.DeleteCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.DeleteCurrentUserPATRequest{
			Id: patID,
		}))
		s.Require().NoError(err)

		// get should fail
		_, err = s.testBench.Client.GetCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
			Id: patID,
		}))
		s.Assert().Error(err)
		s.Assert().Equal(connect.CodeNotFound, connect.CodeOf(err))
	})

	s.Run("Deleted PAT token no longer authenticates", func() {
		patCtx := getPATCtx(patToken)
		_, err := s.testBench.Client.CheckResourcePermission(patCtx, connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
			Resource:   schema.JoinNamespaceAndResourceID(schema.OrganizationNamespace, orgID),
			Permission: schema.GetPermission,
		}))
		s.Assert().Error(err, "deleted PAT token should be rejected")
	})
}

func (s *PATRegressionTestSuite) TestPATCRUD_CreateErrors() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	orgID, _, _ := s.createOrgAndProjects(ctxAdmin, "org-pat-err", "pat-err-p1", "")

	s.Run("duplicate title", func() {
		s.createPAT(ctxAdmin, orgID, "dup-title", []*frontierv1beta1.PATScope{
			{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
		})
		_, err := s.testBench.Client.CreateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
			Title: "dup-title",
			OrgId: orgID,
			Scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
			},
			ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		}))
		s.Assert().Error(err)
		s.Assert().Equal(connect.CodeAlreadyExists, connect.CodeOf(err))
	})

	s.Run("denied role", func() {
		_, err := s.testBench.Client.CreateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
			Title: "denied-role-pat",
			OrgId: orgID,
			Scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationOwner), ResourceType: schema.OrganizationNamespace},
			},
			ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
		}))
		s.Assert().Error(err)
		s.Assert().Equal(connect.CodeInvalidArgument, connect.CodeOf(err))
	})

	s.Run("past expiry", func() {
		_, err := s.testBench.Client.CreateCurrentUserPAT(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
			Title: "past-expiry-pat",
			OrgId: orgID,
			Scopes: []*frontierv1beta1.PATScope{
				{RoleId: s.roleID(schema.RoleOrganizationViewer), ResourceType: schema.OrganizationNamespace},
			},
			ExpiresAt: timestamppb.New(time.Now().Add(-1 * time.Hour)),
		}))
		s.Assert().Error(err)
	})
}

func TestEndToEndPATRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(PATRegressionTestSuite))
}
