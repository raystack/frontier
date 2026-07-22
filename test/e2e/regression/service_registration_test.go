package e2e_test

import (
	"context"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"
	testusers "github.com/raystack/frontier/core/authenticate/test_users"
	"github.com/raystack/frontier/pkg/server"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/frontier/test/e2e/testbench"
	"github.com/stretchr/testify/suite"
)

type ServiceRegistrationRegressionTestSuite struct {
	suite.Suite
	testBench   *testbench.TestBench
	adminCookie string
}

func (s *ServiceRegistrationRegressionTestSuite) SetupSuite() {
	connectPort, err := testbench.GetFreePort()
	s.Require().NoError(err)

	appConfig := &config.Frontier{
		Log: logger.Config{
			Level: "error",
		},
		App: server.Config{
			Host:    "localhost",
			Connect: server.ConnectConfig{Port: connectPort},
			Authentication: authenticate.Config{
				Session: authenticate.SessionConfig{
					HashSecretKey:  "hash-secret-should-be-32-chars--",
					BlockSecretKey: "hash-secret-should-be-32-chars--",
					Validity:       time.Hour,
				},
				MailOTP: authenticate.MailOTPConfig{
					Subject:  "{{.Otp}}",
					Body:     "{{.Otp}}",
					Validity: 10 * time.Minute,
				},
				TestUsers: testusers.Config{Enabled: true, Domain: "raystack.org", OTP: testbench.TestOTP},
			},
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
}

func (s *ServiceRegistrationRegressionTestSuite) TearDownSuite() {
	err := s.testBench.Close()
	s.Require().NoError(err)
}

func (s *ServiceRegistrationRegressionTestSuite) TestServiceRegistration() {
	ctx := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	s.Run("1. register a new service with custom permissions", func() {
		createPermResp, err := s.testBench.AdminClient.CreatePermission(ctx, connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
			Bodies: []*frontierv1beta1.PermissionRequestBody{
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
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(3, len(createPermResp.Msg.GetPermissions()))

		listPermResp, err := s.testBench.Client.ListPermissions(ctx, connect.NewRequest(&frontierv1beta1.ListPermissionsRequest{}))
		s.Assert().NoError(err)
		s.Assert().NotNil(listPermResp.Msg.GetPermissions())
		// check if list contains newly created permissions
		for _, perm := range createPermResp.Msg.GetPermissions() {
			s.Assert().Contains(listPermResp.Msg.GetPermissions(), perm)
		}
		// length of list should be greater than number of permissions created
		s.Assert().GreaterOrEqual(len(listPermResp.Msg.GetPermissions()), len(createPermResp.Msg.GetPermissions()))
	})
	s.Run("2. registering a new service should not remove existing permissions", func() {
		createPermResp, err := s.testBench.AdminClient.CreatePermission(ctx, connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
			Bodies: []*frontierv1beta1.PermissionRequestBody{
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
		}))
		s.Assert().NoError(err)
		s.Assert().Equal(2, len(createPermResp.Msg.GetPermissions()))

		listPermResp, err := s.testBench.Client.ListPermissions(ctx, connect.NewRequest(&frontierv1beta1.ListPermissionsRequest{}))
		s.Assert().NoError(err)
		s.Assert().NotNil(listPermResp.Msg.GetPermissions())
		// check if list contains newly created permissions
		for _, perm := range createPermResp.Msg.GetPermissions() {
			s.Assert().Contains(listPermResp.Msg.GetPermissions(), perm)
		}
		// list should contain permissions created in previous step
		var lastPermCount int
		for _, perm := range []string{"get", "update", "delete"} {
			for _, listPerm := range listPermResp.Msg.GetPermissions() {
				//nolint:staticcheck
				if listPerm.GetName() == perm && listPerm.GetNamespace() == "database/instance" {
					lastPermCount++
				}
			}
		}
		s.Assert().Equal(3, lastPermCount)
	})
}

// TestPermissionDeleteCascade asserts that deleting a (stray) permission sweeps
// the role->permission tuples that reference it, leaving no orphan relation on
// any role; and that a built-in permission (from the base schema) cannot be
// deleted (gap #1661.6).
func (s *ServiceRegistrationRegressionTestSuite) TestPermissionDeleteCascade() {
	ctx := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	const permSlug = "permcascade_res_act"

	createPermResp, err := s.testBench.AdminClient.CreatePermission(ctx, connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
		Bodies: []*frontierv1beta1.PermissionRequestBody{
			{Name: "act", Namespace: "permcascade/res"},
		},
	}))
	s.Require().NoError(err)
	s.Require().Len(createPermResp.Msg.GetPermissions(), 1)
	permissionID := createPermResp.Msg.GetPermissions()[0].GetId()

	createOrgResp, err := s.testBench.Client.CreateOrganization(ctx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{Name: "org-perm-delete-cascade"},
	}))
	s.Require().NoError(err)

	createRoleResp, err := s.testBench.Client.CreateOrganizationRole(ctx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRoleRequest{
		OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		Body: &frontierv1beta1.RoleRequestBody{
			Title:       "perm cascade role",
			Name:        "perm_cascade_role",
			Scopes:      []string{"permcascade/res"},
			Permissions: []string{permSlug},
		},
	}))
	s.Require().NoError(err)
	roleID := createRoleResp.Msg.GetRole().GetId()

	hasPermRelation := func() bool {
		resp, err := s.testBench.AdminClient.ListRelations(ctx, connect.NewRequest(&frontierv1beta1.ListRelationsRequest{
			Object: schema.JoinNamespaceAndResourceID(schema.RoleNamespace, roleID),
		}))
		s.Require().NoError(err)
		for _, rel := range resp.Msg.GetRelations() {
			if rel.GetRelation() == permSlug {
				return true
			}
		}
		return false
	}

	// the role->permission tuples exist before delete
	s.Assert().True(hasPermRelation())

	// delete the permission
	_, err = s.testBench.AdminClient.DeletePermission(ctx, connect.NewRequest(&frontierv1beta1.DeletePermissionRequest{
		Id: permissionID,
	}))
	s.Require().NoError(err)

	// no role->permission tuple should linger for the deleted permission
	s.Assert().False(hasPermRelation())

	// the role should no longer list the deleted permission in its definition
	getRoleResp, err := s.testBench.Client.GetOrganizationRole(ctx, connect.NewRequest(&frontierv1beta1.GetOrganizationRoleRequest{
		OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		Id:    roleID,
	}))
	s.Require().NoError(err)
	s.Assert().NotContains(getRoleResp.Msg.GetRole().GetPermissions(), permSlug)

	// a built-in permission (defined by the base schema) cannot be deleted —
	// bootstrap would just recreate it on the next boot.
	listResp, err := s.testBench.Client.ListPermissions(ctx, connect.NewRequest(&frontierv1beta1.ListPermissionsRequest{}))
	s.Require().NoError(err)
	var builtinID string
	for _, p := range listResp.Msg.GetPermissions() {
		//nolint:staticcheck
		if p.GetNamespace() == "app/organization" && p.GetName() == "get" {
			builtinID = p.GetId()
			break
		}
	}
	s.Require().NotEmpty(builtinID, "expected base permission app/organization:get to exist")

	_, err = s.testBench.AdminClient.DeletePermission(ctx, connect.NewRequest(&frontierv1beta1.DeletePermissionRequest{
		Id: builtinID,
	}))
	s.Require().Error(err)
	s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))
}

// TestPermissionDeleteBlockedByResource asserts that deleting the last
// permission of a namespace is rejected while resources of that type still
// exist — otherwise the namespace's SpiceDB definition would be dropped on the
// next boot and SpiceDB would refuse (relationships still exist), breaking
// startup. Once the resource is removed, the permission deletes cleanly.
func (s *ServiceRegistrationRegressionTestSuite) TestPermissionDeleteBlockedByResource() {
	ctx := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	// name it "delete" so the resource remains deletable (DeleteProjectResource
	// authorizes the "delete" permission on the resource's namespace)
	createPermResp, err := s.testBench.AdminClient.CreatePermission(ctx, connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
		Bodies: []*frontierv1beta1.PermissionRequestBody{
			{Name: "delete", Namespace: "orphanguard/widget"},
		},
	}))
	s.Require().NoError(err)
	permissionID := createPermResp.Msg.GetPermissions()[0].GetId()

	createOrgResp, err := s.testBench.Client.CreateOrganization(ctx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{Name: "org-orphan-guard"},
	}))
	s.Require().NoError(err)
	createProjResp, err := s.testBench.Client.CreateProject(ctx, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{Name: "orphan-guard-proj", OrgId: createOrgResp.Msg.GetOrganization().GetId()},
	}))
	s.Require().NoError(err)

	// a resource of that namespace writes #project/#owner tuples under the type
	createResourceResp, err := s.testBench.Client.CreateProjectResource(ctx, connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
		ProjectId: createProjResp.Msg.GetProject().GetId(),
		Body: &frontierv1beta1.ResourceRequestBody{
			Name:      "orphan-widget-1",
			Namespace: "orphanguard/widget",
		},
	}))
	s.Require().NoError(err)

	// deleting the namespace's only permission must be rejected while the
	// resource exists (it would orphan the namespace definition on next boot)
	_, err = s.testBench.AdminClient.DeletePermission(ctx, connect.NewRequest(&frontierv1beta1.DeletePermissionRequest{
		Id: permissionID,
	}))
	s.Require().Error(err)
	s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

	// remove the resource, then the permission deletes cleanly
	_, err = s.testBench.Client.DeleteProjectResource(ctx, connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
		ProjectId: createProjResp.Msg.GetProject().GetId(),
		Id:        createResourceResp.Msg.GetResource().GetId(),
	}))
	s.Require().NoError(err)

	_, err = s.testBench.AdminClient.DeletePermission(ctx, connect.NewRequest(&frontierv1beta1.DeletePermissionRequest{
		Id: permissionID,
	}))
	s.Require().NoError(err)
}

func TestEndToEndServiceRegistrationRegressionTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceRegistrationRegressionTestSuite))
}
