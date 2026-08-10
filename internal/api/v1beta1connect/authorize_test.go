package v1beta1connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/str"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testAuthzUserID    = "0d7a5b1c-3f2e-4b8a-9c6d-1e2f3a4b5c6d"
	testAuthzUserEmail = "jane@acme.io"
	testAuthzOrgID     = "11111111-1111-1111-1111-111111111111"
	testAuthzProjectID = "22222222-2222-2222-2222-222222222222"
	testAuthzGroupID   = "33333333-3333-3333-3333-333333333333"
	testAuthzInviteID  = "44444444-4444-4444-4444-444444444444"
	testAuthzSvUserID  = "55555555-5555-5555-5555-555555555555"
	testAuthzResID     = "66666666-6666-6666-6666-666666666666"
	testAuthzPATID     = "77777777-7777-7777-7777-777777777777"
)

type authorizeMocks struct {
	authn       *mocks.AuthnService
	resourceSvc *mocks.ResourceService
	orgSvc      *mocks.OrganizationService
	projectSvc  *mocks.ProjectService
	groupSvc    *mocks.GroupService
	inviteSvc   *mocks.InvitationService
	svUserSvc   *mocks.ServiceUserService
	patSvc      *mocks.UserPATService
}

func newAuthorizeHandler(t *testing.T) (*ConnectHandler, authorizeMocks) {
	t.Helper()
	m := authorizeMocks{
		authn:       mocks.NewAuthnService(t),
		resourceSvc: mocks.NewResourceService(t),
		orgSvc:      mocks.NewOrganizationService(t),
		projectSvc:  mocks.NewProjectService(t),
		groupSvc:    mocks.NewGroupService(t),
		inviteSvc:   mocks.NewInvitationService(t),
		svUserSvc:   mocks.NewServiceUserService(t),
		patSvc:      mocks.NewUserPATService(t),
	}
	h := &ConnectHandler{
		authnService:       m.authn,
		resourceService:    m.resourceSvc,
		orgService:         m.orgSvc,
		projectService:     m.projectSvc,
		groupService:       m.groupSvc,
		invitationService:  m.inviteSvc,
		serviceUserService: m.svUserSvc,
		userPATService:     m.patSvc,
	}
	return h, m
}

func expectUserPrincipal(m authorizeMocks) {
	m.authn.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
		ID:   testAuthzUserID,
		Type: schema.UserPrincipal,
		User: &user.User{ID: testAuthzUserID, Email: testAuthzUserEmail},
	}, nil)
}

func expectCheckAuthz(m authorizeMocks, object relation.Object, subjectID, permission string, allowed bool, err error) {
	m.resourceSvc.EXPECT().CheckAuthz(mock.Anything, resource.Check{
		Object:     object,
		Subject:    relation.Subject{Namespace: schema.UserPrincipal, ID: subjectID},
		Permission: permission,
	}).Return(allowed, err)
}

func TestIsAuthorized(t *testing.T) {
	request := connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{})

	tests := []struct {
		name       string
		object     relation.Object
		permission string
		ctx        context.Context
		setup      func(m authorizeMocks, object relation.Object)
		wantErr    error
	}{
		{
			name:   "denied caller gets PermissionDenied without any org lookup",
			object: relation.Object{Namespace: schema.ProjectNamespace, ID: testAuthzProjectID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, false, nil)
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrUnauthorized),
		},
		{
			name:   "org target in enabled state passes",
			object: relation.Object{Namespace: schema.OrganizationNamespace, ID: testAuthzOrgID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{ID: testAuthzOrgID}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "org target disabled fails precondition",
			object: relation.Object{Namespace: schema.OrganizationNamespace, ID: testAuthzOrgID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled),
		},
		{
			name:   "org target missing is not found",
			object: relation.Object{Namespace: schema.OrganizationNamespace, ID: testAuthzOrgID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name:   "project in disabled org fails precondition",
			object: relation.Object{Namespace: schema.ProjectNamespace, ID: testAuthzProjectID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.projectSvc.EXPECT().Get(mock.Anything, testAuthzProjectID).Return(project.Project{
					ID:           testAuthzProjectID,
					Organization: organization.Organization{ID: testAuthzOrgID},
				}, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled),
		},
		{
			name:   "hidden project row skips the gate",
			object: relation.Object{Namespace: schema.ProjectNamespace, ID: testAuthzProjectID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.projectSvc.EXPECT().Get(mock.Anything, testAuthzProjectID).Return(project.Project{}, project.ErrNotExist)
			},
			wantErr: nil,
		},
		{
			name:   "group in disabled org fails precondition",
			object: relation.Object{Namespace: schema.GroupNamespace, ID: testAuthzGroupID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.groupSvc.EXPECT().GetByIDs(mock.Anything, []string{testAuthzGroupID}, group.Filter{IncludeDisabled: true}).
					Return([]group.Group{{ID: testAuthzGroupID, OrganizationID: testAuthzOrgID}}, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled),
		},
		{
			name:       "disabled group in enabled org passes",
			object:     relation.Object{Namespace: schema.GroupNamespace, ID: testAuthzGroupID},
			permission: schema.DeletePermission,
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.DeletePermission, true, nil)
				m.groupSvc.EXPECT().GetByIDs(mock.Anything, []string{testAuthzGroupID}, group.Filter{IncludeDisabled: true}).
					Return([]group.Group{{ID: testAuthzGroupID, OrganizationID: testAuthzOrgID}}, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{ID: testAuthzOrgID}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "vanished group skips the gate",
			object: relation.Object{Namespace: schema.GroupNamespace, ID: testAuthzGroupID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.groupSvc.EXPECT().GetByIDs(mock.Anything, []string{testAuthzGroupID}, group.Filter{IncludeDisabled: true}).
					Return([]group.Group{}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "invitation allowed via email fallback still hits the gate",
			object: relation.Object{Namespace: schema.InvitationNamespace, ID: testAuthzInviteID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, false, nil)
				expectCheckAuthz(m, object, str.GenerateUserSlug(testAuthzUserEmail), schema.GetPermission, true, nil)
				m.inviteSvc.EXPECT().Get(mock.Anything, mock.Anything).Return(invitation.Invitation{OrgID: testAuthzOrgID}, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled),
		},
		{
			name:   "invitation with non uuid id skips the gate",
			object: relation.Object{Namespace: schema.InvitationNamespace, ID: "not-a-uuid"},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
			},
			wantErr: nil,
		},
		{
			name:       "service user in disabled org fails precondition",
			object:     relation.Object{Namespace: schema.ServiceUserPrincipal, ID: testAuthzSvUserID},
			permission: schema.ManagePermission,
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.ManagePermission, true, nil)
				m.svUserSvc.EXPECT().Get(mock.Anything, testAuthzSvUserID).Return(serviceuser.ServiceUser{OrgID: testAuthzOrgID}, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled),
		},
		{
			name:       "custom resource resolves through project to disabled org",
			object:     relation.Object{Namespace: "compute/instance", ID: testAuthzResID},
			permission: schema.UpdatePermission,
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.UpdatePermission, true, nil)
				m.resourceSvc.EXPECT().Get(mock.Anything, testAuthzResID).Return(resource.Resource{ID: testAuthzResID, ProjectID: testAuthzProjectID}, nil)
				m.projectSvc.EXPECT().Get(mock.Anything, testAuthzProjectID).Return(project.Project{
					ID:           testAuthzProjectID,
					Organization: organization.Organization{ID: testAuthzOrgID},
				}, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled),
		},
		{
			name:   "missing custom resource skips the gate",
			object: relation.Object{Namespace: "compute/instance", ID: testAuthzResID},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.resourceSvc.EXPECT().Get(mock.Anything, testAuthzResID).Return(resource.Resource{}, resource.ErrNotExist)
			},
			wantErr: nil,
		},
		{
			name:       "platform object skips the gate",
			object:     relation.Object{Namespace: schema.PlatformNamespace, ID: schema.PlatformID},
			permission: schema.PlatformCheckPermission,
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.PlatformCheckPermission, true, nil)
			},
			wantErr: nil,
		},
		{
			name:       "other predefined namespaces skip the gate",
			object:     relation.Object{Namespace: "app/role", ID: testAuthzResID},
			permission: schema.DeletePermission,
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.DeletePermission, true, nil)
			},
			wantErr: nil,
		},
		{
			name:   "superuser bypasses the gate on a disabled org",
			object: relation.Object{Namespace: schema.OrganizationNamespace, ID: testAuthzOrgID},
			ctx:    authenticate.SetSuperUserInContext(context.Background(), true),
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
			},
			wantErr: nil,
		},
		{
			name:   "member reaching a disabled org by name fails precondition after the permission check",
			object: relation.Object{Namespace: schema.OrganizationNamespace, ID: "acme-name"},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, true, nil)
				m.orgSvc.EXPECT().Get(mock.Anything, "acme-name").Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled),
		},
		{
			name:   "outsider probing a disabled org by name gets permission denied without any org lookup",
			object: relation.Object{Namespace: schema.OrganizationNamespace, ID: "acme-name"},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, false, nil)
			},
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrUnauthorized),
		},
		{
			name:   "missing org referenced by name stays not found",
			object: relation.Object{Namespace: schema.OrganizationNamespace, ID: "ghost-name"},
			setup: func(m authorizeMocks, object relation.Object) {
				expectUserPrincipal(m)
				expectCheckAuthz(m, object, testAuthzUserID, schema.GetPermission, false, organization.ErrNotExist)
			},
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name:    "empty object is rejected before authentication",
			object:  relation.Object{Namespace: schema.OrganizationNamespace, ID: ""},
			setup:   func(m authorizeMocks, object relation.Object) {},
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrInvalidNamesapceOrID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, m := newAuthorizeHandler(t)
			tt.setup(m, tt.object)

			ctx := tt.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			permission := tt.permission
			if permission == "" {
				permission = schema.GetPermission
			}

			err := h.IsAuthorized(ctx, tt.object, permission, request)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// TestResolveObjectOrgNamespaces pins which predefined namespaces the gate
// resolves and which it deliberately leaves alone. A new org-owned namespace
// must be added to the resolver switch or to the org-less list here.
func TestResolveObjectOrgNamespaces(t *testing.T) {
	resolved := map[string]func(m authorizeMocks){
		schema.OrganizationNamespace: func(m authorizeMocks) {},
		schema.ProjectNamespace: func(m authorizeMocks) {
			m.projectSvc.EXPECT().Get(mock.Anything, testAuthzResID).Return(project.Project{Organization: organization.Organization{ID: testAuthzOrgID}}, nil)
		},
		schema.GroupNamespace: func(m authorizeMocks) {
			m.groupSvc.EXPECT().GetByIDs(mock.Anything, []string{testAuthzResID}, group.Filter{IncludeDisabled: true}).
				Return([]group.Group{{OrganizationID: testAuthzOrgID}}, nil)
		},
		schema.InvitationNamespace: func(m authorizeMocks) {
			m.inviteSvc.EXPECT().Get(mock.Anything, mock.Anything).Return(invitation.Invitation{OrgID: testAuthzOrgID}, nil)
		},
		schema.ServiceUserPrincipal: func(m authorizeMocks) {
			m.svUserSvc.EXPECT().Get(mock.Anything, testAuthzResID).Return(serviceuser.ServiceUser{OrgID: testAuthzOrgID}, nil)
		},
	}
	// these predefined namespaces have no single owning org to enforce, or
	// are only reachable as policy targets; the gate must leave them alone
	orgless := []string{
		schema.PlatformNamespace,
		schema.UserPrincipal,
		schema.PATPrincipal,
		schema.RoleNamespace,
		schema.RoleBindingNamespace,
	}

	for ns, setup := range resolved {
		t.Run("resolves "+ns, func(t *testing.T) {
			h, m := newAuthorizeHandler(t)
			setup(m)
			// invitation ids must parse as uuids, so use a uuid for all
			orgID, found, err := h.resolveObjectOrg(context.Background(), relation.Object{Namespace: ns, ID: testAuthzResID})
			assert.NoError(t, err)
			assert.True(t, found)
			if ns == schema.OrganizationNamespace {
				assert.Equal(t, testAuthzResID, orgID)
			} else {
				assert.Equal(t, testAuthzOrgID, orgID)
			}
		})
	}

	for _, ns := range orgless {
		t.Run("skips "+ns, func(t *testing.T) {
			h, _ := newAuthorizeHandler(t)
			_, found, err := h.resolveObjectOrg(context.Background(), relation.Object{Namespace: ns, ID: testAuthzResID})
			assert.NoError(t, err)
			assert.False(t, found)
		})
	}
}

func TestEnsurePATOrgActive(t *testing.T) {
	t.Run("own token of a disabled org fails precondition", func(t *testing.T) {
		h, m := newAuthorizeHandler(t)
		expectUserPrincipal(m)
		m.patSvc.EXPECT().Get(mock.Anything, testAuthzUserID, testAuthzPATID).Return(patmodels.PAT{ID: testAuthzPATID, OrgID: testAuthzOrgID}, nil)
		m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{}, organization.ErrDisabled)

		err := h.EnsurePATOrgActive(context.Background(), testAuthzPATID)
		assert.Equal(t, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled), err)
	})

	t.Run("own token of an enabled org passes", func(t *testing.T) {
		h, m := newAuthorizeHandler(t)
		expectUserPrincipal(m)
		m.patSvc.EXPECT().Get(mock.Anything, testAuthzUserID, testAuthzPATID).Return(patmodels.PAT{ID: testAuthzPATID, OrgID: testAuthzOrgID}, nil)
		m.orgSvc.EXPECT().Get(mock.Anything, testAuthzOrgID).Return(organization.Organization{ID: testAuthzOrgID}, nil)

		assert.NoError(t, h.EnsurePATOrgActive(context.Background(), testAuthzPATID))
	})

	t.Run("unknown or someone else's token skips the check", func(t *testing.T) {
		h, m := newAuthorizeHandler(t)
		expectUserPrincipal(m)
		m.patSvc.EXPECT().Get(mock.Anything, testAuthzUserID, testAuthzPATID).Return(patmodels.PAT{}, paterrors.ErrNotFound)

		assert.NoError(t, h.EnsurePATOrgActive(context.Background(), testAuthzPATID))
	})

	t.Run("pat feature disabled skips the check", func(t *testing.T) {
		h, m := newAuthorizeHandler(t)
		expectUserPrincipal(m)
		m.patSvc.EXPECT().Get(mock.Anything, testAuthzUserID, testAuthzPATID).Return(patmodels.PAT{}, paterrors.ErrDisabled)

		assert.NoError(t, h.EnsurePATOrgActive(context.Background(), testAuthzPATID))
	})

	t.Run("service account principal skips the check", func(t *testing.T) {
		h, m := newAuthorizeHandler(t)
		m.authn.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
			ID:   testAuthzSvUserID,
			Type: schema.ServiceUserPrincipal,
		}, nil)

		assert.NoError(t, h.EnsurePATOrgActive(context.Background(), testAuthzPATID))
	})

	t.Run("malformed token id skips the check", func(t *testing.T) {
		h, _ := newAuthorizeHandler(t)

		assert.NoError(t, h.EnsurePATOrgActive(context.Background(), "not-a-uuid"))
	})

	t.Run("superuser bypasses the check", func(t *testing.T) {
		h, _ := newAuthorizeHandler(t)
		ctx := authenticate.SetSuperUserInContext(context.Background(), true)

		assert.NoError(t, h.EnsurePATOrgActive(ctx, testAuthzPATID))
	})
}
