package membership_test

import (
	"context"
	"errors"
	"testing"

	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/membership/mocks"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	pat "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_ListPoliciesByPrincipal(t *testing.T) {
	ctx := context.Background()
	principalID := uuid.New().String()

	t.Run("returns every policy held by the principal", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: principalID, PrincipalType: schema.PATPrincipal}).
			Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		got, err := svc.ListPoliciesByPrincipal(ctx, principalID, schema.PATPrincipal)
		assert.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("surfaces policy list errors", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: principalID, PrincipalType: schema.PATPrincipal}).
			Return(nil, errors.New("db down"))

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		_, err := svc.ListPoliciesByPrincipal(ctx, principalID, schema.PATPrincipal)
		assert.Error(t, err)
	})
}

func TestService_ListPrincipalsByResource(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	projectID := uuid.New().String()
	groupID := uuid.New().String()
	user1, user2 := uuid.New().String(), uuid.New().String()
	suID := uuid.New().String()
	roleViewerID, roleOwnerID := uuid.New().String(), uuid.New().String()

	viewerRole := role.Role{ID: roleViewerID, Name: "viewer"}
	ownerRole := role.Role{ID: roleOwnerID, Name: schema.RoleOrganizationOwner}

	tests := []struct {
		name         string
		resourceID   string
		resourceType string
		filter       membership.MemberFilter
		setup        func(*mocks.PolicyService, *mocks.RoleService)
		want         []membership.Member
		wantErrIs    error
		wantErrMsg   string
	}{
		{
			name:         "rejects unsupported resource type",
			resourceID:   orgID,
			resourceType: "app/unknown",
			wantErrIs:    membership.ErrInvalidResourceType,
		},
		{
			name:         "returns empty when no policies exist",
			resourceID:   orgID,
			resourceType: schema.OrganizationNamespace,
			filter:       membership.MemberFilter{PrincipalType: schema.UserPrincipal},
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				ps.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return([]policy.Policy{}, nil)
			},
			want: []membership.Member{},
		},
		{
			name:         "lists users of an org, deduplicated across multiple policies",
			resourceID:   orgID,
			resourceType: schema.OrganizationNamespace,
			filter:       membership.MemberFilter{PrincipalType: schema.UserPrincipal},
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				orgPolicies := []policy.Policy{
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleViewerID},
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleOwnerID},
					{PrincipalID: user2, PrincipalType: schema.UserPrincipal, RoleID: roleViewerID},
				}
				ps.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return(orgPolicies, nil).Times(2)
				rs.EXPECT().List(ctx, mock.MatchedBy(func(f role.Filter) bool {
					return len(f.IDs) == 2
				})).Return([]role.Role{viewerRole, ownerRole}, nil)
			},
			want: []membership.Member{
				{PrincipalID: user1, PrincipalType: schema.UserPrincipal, Roles: []role.Role{viewerRole, ownerRole}},
				{PrincipalID: user2, PrincipalType: schema.UserPrincipal, Roles: []role.Role{viewerRole}},
			},
		},
		{
			name:         "filters by roles when RoleIDs provided",
			resourceID:   orgID,
			resourceType: schema.OrganizationNamespace,
			filter:       membership.MemberFilter{PrincipalType: schema.UserPrincipal, RoleIDs: []string{roleOwnerID}},
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				ps.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalType: schema.UserPrincipal,
					RoleIDs:       []string{roleOwnerID},
					ResourceType:  schema.OrganizationNamespace,
				}).Return([]policy.Policy{
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleOwnerID},
				}, nil)
				ps.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return([]policy.Policy{
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleViewerID},
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleOwnerID},
				}, nil)
				rs.EXPECT().List(ctx, mock.MatchedBy(func(f role.Filter) bool {
					return len(f.IDs) == 2
				})).Return([]role.Role{viewerRole, ownerRole}, nil)
			},
			want: []membership.Member{
				{PrincipalID: user1, PrincipalType: schema.UserPrincipal, Roles: []role.Role{viewerRole, ownerRole}},
			},
		},
		{
			name:         "enriches members with roles",
			resourceID:   projectID,
			resourceType: schema.ProjectNamespace,
			filter:       membership.MemberFilter{PrincipalType: schema.UserPrincipal},
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				projectPolicies := []policy.Policy{
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleViewerID},
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleOwnerID},
				}
				ps.EXPECT().List(ctx, policy.Filter{
					ProjectID:     projectID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.ProjectNamespace,
				}).Return(projectPolicies, nil).Times(2)
				rs.EXPECT().List(ctx, mock.MatchedBy(func(f role.Filter) bool {
					return len(f.IDs) == 2
				})).Return([]role.Role{viewerRole, ownerRole}, nil)
			},
			want: []membership.Member{
				{PrincipalID: user1, PrincipalType: schema.UserPrincipal, Roles: []role.Role{viewerRole, ownerRole}},
			},
		},
		{
			name:         "lists service users of a project",
			resourceID:   projectID,
			resourceType: schema.ProjectNamespace,
			filter:       membership.MemberFilter{PrincipalType: schema.ServiceUserPrincipal},
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				suPolicies := []policy.Policy{
					{PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal, RoleID: roleViewerID},
				}
				ps.EXPECT().List(ctx, policy.Filter{
					ProjectID:     projectID,
					PrincipalType: schema.ServiceUserPrincipal,
					ResourceType:  schema.ProjectNamespace,
				}).Return(suPolicies, nil).Times(2)
				rs.EXPECT().List(ctx, role.Filter{IDs: []string{roleViewerID}}).Return([]role.Role{viewerRole}, nil)
			},
			want: []membership.Member{
				{PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal, Roles: []role.Role{viewerRole}},
			},
		},
		{
			name:         "lists group members of a group",
			resourceID:   groupID,
			resourceType: schema.GroupNamespace,
			filter:       membership.MemberFilter{PrincipalType: schema.UserPrincipal},
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				groupPolicies := []policy.Policy{
					{PrincipalID: user1, PrincipalType: schema.UserPrincipal, RoleID: roleViewerID},
				}
				ps.EXPECT().List(ctx, policy.Filter{
					GroupID:       groupID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return(groupPolicies, nil).Times(2)
				rs.EXPECT().List(ctx, role.Filter{IDs: []string{roleViewerID}}).Return([]role.Role{viewerRole}, nil)
			},
			want: []membership.Member{
				{PrincipalID: user1, PrincipalType: schema.UserPrincipal, Roles: []role.Role{viewerRole}},
			},
		},
		{
			name:         "wraps policy list errors",
			resourceID:   orgID,
			resourceType: schema.OrganizationNamespace,
			filter:       membership.MemberFilter{PrincipalType: schema.UserPrincipal},
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				ps.EXPECT().List(ctx, mock.Anything).Return(nil, errors.New("db down"))
			},
			wantErrMsg: "db down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockRoleSvc := mocks.NewRoleService(t)
			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockRoleSvc)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

			got, err := svc.ListPrincipalsByResource(ctx, tt.resourceID, tt.resourceType, tt.filter)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
				return
			}
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_ListPrincipalIDsByResource(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	su1, su2 := uuid.New().String(), uuid.New().String()
	roleID := uuid.New().String()

	tests := []struct {
		name       string
		setup      func(*mocks.PolicyService, *mocks.RoleService)
		want       []string
		wantErrMsg string
	}{
		{
			name: "returns deduplicated principal IDs from a single policy query",
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				suPolicies := []policy.Policy{
					{PrincipalID: su1, PrincipalType: schema.ServiceUserPrincipal, RoleID: roleID},
					{PrincipalID: su1, PrincipalType: schema.ServiceUserPrincipal, RoleID: roleID},
					{PrincipalID: su2, PrincipalType: schema.ServiceUserPrincipal, RoleID: roleID},
				}
				ps.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalType: schema.ServiceUserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return(suPolicies, nil).Once()
				// no role-service calls: the ID-only path skips role enrichment
			},
			want: []string{su1, su2},
		},
		{
			name: "returns empty when no principals",
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				ps.EXPECT().List(ctx, mock.Anything).Return([]policy.Policy{}, nil)
			},
			want: []string{},
		},
		{
			name: "propagates errors",
			setup: func(ps *mocks.PolicyService, rs *mocks.RoleService) {
				ps.EXPECT().List(ctx, mock.Anything).Return(nil, errors.New("db down"))
			},
			wantErrMsg: "db down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockRoleSvc := mocks.NewRoleService(t)
			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockRoleSvc)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

			got, err := svc.ListPrincipalIDsByResource(ctx, orgID, schema.OrganizationNamespace, schema.ServiceUserPrincipal)
			if tt.wantErrMsg != "" {
				assert.ErrorContains(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestService_ListResourcesByPrincipal covers each resource type, role-based
// visibility filtering, group expansion, OrgID narrowing, and PAT intersection.
func TestService_ListResourcesByPrincipal(t *testing.T) {
	ctx := context.Background()

	// fixture IDs
	userID := uuid.New().String()
	suID := uuid.New().String()
	patID := uuid.New().String()
	orgA := uuid.New().String()
	orgB := uuid.New().String()
	project1, project2, project3 := uuid.New().String(), uuid.New().String(), uuid.New().String()
	groupA := uuid.New().String()

	roleOrgViewerID := uuid.New().String()
	roleOrgManagerID := uuid.New().String()
	roleOrgOwnerID := uuid.New().String()
	roleOrgCustomID := uuid.New().String()
	roleProjectViewerID := uuid.New().String()
	roleProjectOwnerID := uuid.New().String()

	type mockSet struct {
		policy  *mocks.PolicyService
		role    *mocks.RoleService
		project *mocks.ProjectService
		group   *mocks.GroupService
	}

	tests := []struct {
		name         string
		principal    authenticate.Principal
		resourceType string
		filter       membership.ResourceFilter
		setup        func(m *mockSet)
		want         []string
		wantErrIs    error
	}{
		{
			name:         "rejects unsupported resource type",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: "app/unknown",
			setup:        func(m *mockSet) {},
			wantErrIs:    membership.ErrInvalidResourceType,
		},
		{
			name:         "lists orgs from direct policies without role-permission filter",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.OrganizationNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return([]policy.Policy{
					{ResourceID: orgA, RoleID: roleOrgViewerID},
					{ResourceID: orgB, RoleID: roleOrgManagerID},
				}, nil)
			},
			want: []string{orgA, orgB},
		},
		{
			name:         "deduplicates org IDs across multiple policies on the same org",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.OrganizationNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return([]policy.Policy{
					{ResourceID: orgA, RoleID: roleOrgViewerID},
					{ResourceID: orgA, RoleID: roleOrgOwnerID},
				}, nil)
			},
			want: []string{orgA},
		},
		{
			name:         "stale-relation regression: returns empty when no policies, ignoring any SpiceDB state",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.OrganizationNamespace,
			setup: func(m *mockSet) {
				// Even if SpiceDB still had an org#owner@U tuple from a
				// pre-demotion state, this method only consults policies.
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return([]policy.Policy{}, nil)
			},
			want: []string{},
		},
		{
			name:         "lists groups from direct policies, no inheritance",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.GroupNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{
					{ResourceID: groupA, RoleID: uuid.New().String()},
				}, nil)
			},
			want: []string{groupA},
		},
		{
			name:         "project listing: direct policy with role granting project visibility is returned",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.ProjectNamespace,
			filter:       membership.ResourceFilter{NonInherited: true},
			setup: func(m *mockSet) {
				// direct project policies — gated by RolePermissions at policy.Filter
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{
					{ResourceID: project1, RoleID: roleProjectViewerID},
				}, nil)
				// group expansion: principal has no groups
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				// NonInherited=true → org-inheritance branch skipped
			},
			want: []string{project1},
		},
		{
			name:         "project listing: owner role on org expands to all org projects via inheritance",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.ProjectNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{
					{ResourceID: orgA, RoleID: roleOrgOwnerID},
				}, nil)
				m.project.EXPECT().List(ctx, project.Filter{OrgIDs: []string{orgA}}).Return([]project.Project{
					{ID: project1}, {ID: project2}, {ID: project3},
				}, nil)
			},
			want: []string{project1, project2, project3},
		},
		{
			name:         "project listing: manager role on org expands via app_project_get inheritance",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.ProjectNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{
					{ResourceID: orgA, RoleID: roleOrgManagerID},
				}, nil)
				m.project.EXPECT().List(ctx, project.Filter{OrgIDs: []string{orgA}}).Return([]project.Project{
					{ID: project1}, {ID: project2},
				}, nil)
			},
			want: []string{project1, project2},
		},
		{
			name:         "project listing: viewer role on org does NOT expand (no inheritance)",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.ProjectNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				// SQL filter excludes the viewer's policy (role doesn't grant
				// any OrganizationProjectInheritPerms) — empty result, no
				// follow-up projectService.List call.
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{}, nil)
			},
			want: []string{},
		},
		{
			name:         "project listing: custom org role with app_project_administer expands",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.ProjectNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{
					{ResourceID: orgA, RoleID: roleOrgCustomID},
				}, nil)
				m.project.EXPECT().List(ctx, project.Filter{OrgIDs: []string{orgA}}).Return([]project.Project{
					{ID: project1},
				}, nil)
			},
			want: []string{project1},
		},
		{
			name:         "project listing: group expansion adds group-policied projects (even with NonInherited=true)",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.ProjectNamespace,
			filter:       membership.ResourceFilter{NonInherited: true},
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{}, nil)
				// recursion to list groups for the user (no RolePermissions —
				// group listing isn't role-permission-gated)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{
					{ResourceID: groupA, RoleID: uuid.New().String()},
				}, nil)
				// then project policies on those groups, also gated
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalType:   schema.GroupPrincipal,
					PrincipalIDs:    []string{groupA},
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{
					{ResourceID: project2, RoleID: roleProjectViewerID},
				}, nil)
			},
			want: []string{project2},
		},
		{
			name:         "project listing: OrgID narrows the result set via projectService.List",
			principal:    authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			resourceType: schema.ProjectNamespace,
			filter:       membership.ResourceFilter{OrgID: orgA, NonInherited: true},
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{
					{ResourceID: project1, RoleID: roleProjectViewerID},
					{ResourceID: project2, RoleID: roleProjectViewerID},
				}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				// narrowing: re-issue projectService.List with the OrgID filter,
				// returning only project1 (project2 was filtered out by org_id).
				m.project.EXPECT().List(ctx, mock.MatchedBy(func(f project.Filter) bool {
					return f.OrgID == orgA && len(f.ProjectIDs) == 2
				})).Return([]project.Project{{ID: project1}}, nil)
			},
			want: []string{project1},
		},
		{
			name:         "serviceuser principal: org listing uses ServiceUserPrincipal type",
			principal:    authenticate.Principal{ID: suID, Type: schema.ServiceUserPrincipal},
			resourceType: schema.OrganizationNamespace,
			setup: func(m *mockSet) {
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   suID,
					PrincipalType: schema.ServiceUserPrincipal,
					ResourceType:  schema.OrganizationNamespace,
				}).Return([]policy.Policy{
					{ResourceID: orgA, RoleID: roleOrgViewerID},
				}, nil)
			},
			want: []string{orgA},
		},
		{
			name: "no-PAT path: Principal{Type: UserPrincipal, PAT: nil} skips the recursive PAT pass",
			principal: authenticate.Principal{
				ID:   userID,
				Type: schema.UserPrincipal,
				PAT:  nil,
			},
			resourceType: schema.ProjectNamespace,
			filter:       membership.ResourceFilter{NonInherited: true},
			setup: func(m *mockSet) {
				// only the user-pass queries fire; no second list under the PAT principal type
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{
					{ResourceID: project1, RoleID: roleProjectViewerID},
				}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
			},
			want: []string{project1},
		},
		{
			name: "PAT all-projects scope with ProjectOwner role resolves via org inheritance",
			principal: authenticate.Principal{
				ID:   userID,
				Type: schema.UserPrincipal,
				PAT:  &pat.PAT{ID: patID, UserID: userID, OrgID: orgA},
			},
			resourceType: schema.ProjectNamespace,
			setup: func(m *mockSet) {
				// user pass — user is org owner, expands via inheritance
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{
					{ResourceID: orgA, RoleID: roleOrgOwnerID},
				}, nil)
				m.project.EXPECT().List(ctx, project.Filter{OrgIDs: []string{orgA}}).Return([]project.Project{
					{ID: project1}, {ID: project2}, {ID: project3},
				}, nil)
				// PAT pass — all-projects scope is one pat_granted policy on the org
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     patID,
					PrincipalType:   schema.PATPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   patID,
					PrincipalType: schema.PATPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     patID,
					PrincipalType:   schema.PATPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{
					// grant_relation here would be pat_granted in production;
					// listing doesn't filter on it, so the value doesn't matter
					// for behavior — only the role's permissions do.
					{ResourceID: orgA, RoleID: roleProjectOwnerID},
				}, nil)
				m.project.EXPECT().List(ctx, project.Filter{OrgIDs: []string{orgA}}).Return([]project.Project{
					{ID: project1}, {ID: project2}, {ID: project3},
				}, nil)
			},
			// PAT can see all of OrgA. User can also see all. Intersection = all.
			want: []string{project1, project2, project3},
		},
		{
			name: "PAT narrows: user is org viewer with direct P1, PAT scoped to P2 only → empty intersection",
			principal: authenticate.Principal{
				ID:   userID,
				Type: schema.UserPrincipal,
				PAT:  &pat.PAT{ID: patID, UserID: userID, OrgID: orgA},
			},
			resourceType: schema.ProjectNamespace,
			setup: func(m *mockSet) {
				// user pass — viewer role on org doesn't pass the inheritance
				// gate, so the org-inheritance query returns []
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{
					{ResourceID: project1, RoleID: roleProjectViewerID},
				}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     userID,
					PrincipalType:   schema.UserPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{}, nil)
				// PAT pass
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     patID,
					PrincipalType:   schema.PATPrincipal,
					ResourceType:    schema.ProjectNamespace,
					RolePermissions: schema.ProjectDirectVisibilityPerms,
				}).Return([]policy.Policy{
					{ResourceID: project2, RoleID: roleProjectViewerID},
				}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   patID,
					PrincipalType: schema.PATPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{}, nil)
				m.policy.EXPECT().List(ctx, policy.Filter{
					PrincipalID:     patID,
					PrincipalType:   schema.PATPrincipal,
					ResourceType:    schema.OrganizationNamespace,
					RolePermissions: schema.OrganizationProjectInheritPerms,
				}).Return([]policy.Policy{}, nil)
			},
			// user sees [P1], PAT sees [P2], intersection = []
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := mocks.NewPolicyService(t)
			mr := mocks.NewRoleService(t)
			mpr := mocks.NewProjectService(t)
			mg := mocks.NewGroupService(t)

			tt.setup(&mockSet{policy: mp, role: mr, project: mpr, group: mg})

			svc := membership.NewService(
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				mp,
				mocks.NewRelationService(t),
				mr,
				mocks.NewOrgService(t),
				mocks.NewUserService(t),
				mpr,
				mg,
				mocks.NewServiceuserService(t),
				mocks.NewAuditRecordRepository(t),
			)

			got, err := svc.ListResourcesByPrincipal(ctx, tt.principal, tt.resourceType, tt.filter)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
				return
			}
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestService_ListGroupsByPrincipal(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New().String()
	patID := uuid.New().String()
	orgA := uuid.New().String()
	groupA := uuid.New().String()
	groupB := uuid.New().String()
	roleGroupMemberID := uuid.New().String()

	tests := []struct {
		name      string
		principal authenticate.Principal
		orgID     string
		setup     func(p *mocks.PolicyService, g *mocks.GroupService)
		want      []string
	}{
		{
			name:      "user principal — reads user's group policies",
			principal: authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			setup: func(p *mocks.PolicyService, _ *mocks.GroupService) {
				p.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{
					{ResourceID: groupA, RoleID: roleGroupMemberID},
					{ResourceID: groupB, RoleID: roleGroupMemberID},
				}, nil)
			},
			want: []string{groupA, groupB},
		},
		{
			name: "PAT principal — resolves to underlying user, no PAT-side query",
			principal: authenticate.Principal{
				ID:   userID,
				Type: schema.UserPrincipal,
				PAT:  &pat.PAT{ID: patID, UserID: userID, OrgID: orgA},
			},
			setup: func(p *mocks.PolicyService, _ *mocks.GroupService) {
				// only the user-side lookup; no policy.List for PrincipalType=PAT
				p.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{
					{ResourceID: groupA, RoleID: roleGroupMemberID},
				}, nil)
			},
			want: []string{groupA},
		},
		{
			name: "PAT principal + orgID — narrows result via groupService",
			principal: authenticate.Principal{
				ID:   userID,
				Type: schema.UserPrincipal,
				PAT:  &pat.PAT{ID: patID, UserID: userID, OrgID: orgA},
			},
			orgID: orgA,
			setup: func(p *mocks.PolicyService, g *mocks.GroupService) {
				p.EXPECT().List(ctx, policy.Filter{
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
					ResourceType:  schema.GroupNamespace,
				}).Return([]policy.Policy{
					{ResourceID: groupA, RoleID: roleGroupMemberID},
					{ResourceID: groupB, RoleID: roleGroupMemberID},
				}, nil)
				g.EXPECT().List(ctx, group.Filter{
					OrganizationID: orgA,
					GroupIDs:       []string{groupA, groupB},
				}).Return([]group.Group{{ID: groupA}}, nil)
			},
			want: []string{groupA},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := mocks.NewPolicyService(t)
			mg := mocks.NewGroupService(t)

			tt.setup(mp, mg)

			svc := membership.NewService(
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				mp,
				mocks.NewRelationService(t),
				mocks.NewRoleService(t),
				mocks.NewOrgService(t),
				mocks.NewUserService(t),
				mocks.NewProjectService(t),
				mg,
				mocks.NewServiceuserService(t),
				mocks.NewAuditRecordRepository(t),
			)

			got, err := svc.ListGroupsByPrincipal(ctx, tt.principal, tt.orgID)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestService_ListProjectsByPrincipal(t *testing.T) {
	ctx := context.Background()

	userID := uuid.New().String()
	patID := uuid.New().String()
	orgA := uuid.New().String()
	projDirect := uuid.New().String()
	projPATScope := uuid.New().String()

	t.Run("user principal with NonInherited=true skips org-inheritance branch", func(t *testing.T) {
		mp := mocks.NewPolicyService(t)
		// Direct project policies fetch.
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:     userID,
			PrincipalType:   schema.UserPrincipal,
			ResourceType:    schema.ProjectNamespace,
			RolePermissions: schema.ProjectDirectVisibilityPerms,
		}).Return([]policy.Policy{{ResourceID: projDirect}}, nil)
		// Group expansion: principal has no groups (NonInherited=true on inner call).
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:   userID,
			PrincipalType: schema.UserPrincipal,
			ResourceType:  schema.GroupNamespace,
		}).Return([]policy.Policy{}, nil)
		// NO org-inheritance fetch must happen — that's the NonInherited contract.

		svc := membership.NewService(
			slog.New(slog.NewTextHandler(io.Discard, nil)),
			mp,
			mocks.NewRelationService(t),
			mocks.NewRoleService(t),
			mocks.NewOrgService(t),
			mocks.NewUserService(t),
			mocks.NewProjectService(t),
			mocks.NewGroupService(t),
			mocks.NewServiceuserService(t),
			mocks.NewAuditRecordRepository(t),
		)

		got, err := svc.ListProjectsByPrincipal(
			ctx,
			authenticate.Principal{ID: userID, Type: schema.UserPrincipal},
			"",
			true,
		)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{projDirect}, got)
	})

	t.Run("PAT principal — runs both user-side and PAT-side queries and intersects (unlike groups)", func(t *testing.T) {
		mp := mocks.NewPolicyService(t)
		// User-side: direct project policies + (no groups) + org-inheritance branch.
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:     userID,
			PrincipalType:   schema.UserPrincipal,
			ResourceType:    schema.ProjectNamespace,
			RolePermissions: schema.ProjectDirectVisibilityPerms,
		}).Return([]policy.Policy{
			{ResourceID: projDirect},
			{ResourceID: projPATScope},
		}, nil)
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:   userID,
			PrincipalType: schema.UserPrincipal,
			ResourceType:  schema.GroupNamespace,
		}).Return([]policy.Policy{}, nil)
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:     userID,
			PrincipalType:   schema.UserPrincipal,
			ResourceType:    schema.OrganizationNamespace,
			RolePermissions: schema.OrganizationProjectInheritPerms,
		}).Return([]policy.Policy{}, nil)

		// PAT-side: same fanout under PAT principal type — PAT only scopes projPATScope.
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:     patID,
			PrincipalType:   schema.PATPrincipal,
			ResourceType:    schema.ProjectNamespace,
			RolePermissions: schema.ProjectDirectVisibilityPerms,
		}).Return([]policy.Policy{{ResourceID: projPATScope}}, nil)
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:   patID,
			PrincipalType: schema.PATPrincipal,
			ResourceType:  schema.GroupNamespace,
		}).Return([]policy.Policy{}, nil)
		mp.EXPECT().List(ctx, policy.Filter{
			PrincipalID:     patID,
			PrincipalType:   schema.PATPrincipal,
			ResourceType:    schema.OrganizationNamespace,
			RolePermissions: schema.OrganizationProjectInheritPerms,
		}).Return([]policy.Policy{}, nil)

		svc := membership.NewService(
			slog.New(slog.NewTextHandler(io.Discard, nil)),
			mp,
			mocks.NewRelationService(t),
			mocks.NewRoleService(t),
			mocks.NewOrgService(t),
			mocks.NewUserService(t),
			mocks.NewProjectService(t),
			mocks.NewGroupService(t),
			mocks.NewServiceuserService(t),
			mocks.NewAuditRecordRepository(t),
		)

		got, err := svc.ListProjectsByPrincipal(
			ctx,
			authenticate.Principal{
				ID:   userID,
				Type: schema.UserPrincipal,
				PAT:  &pat.PAT{ID: patID, UserID: userID, OrgID: orgA},
			},
			"",
			false,
		)
		assert.NoError(t, err)
		// PAT narrows: user sees [direct, patScope]; PAT sees [patScope]; intersect → [patScope].
		assert.ElementsMatch(t, []string{projPATScope}, got)
	})
}
