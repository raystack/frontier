package membership_test

import (
	"context"
	"errors"
	"testing"

	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/membership/mocks"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	pat "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_AddOrganizationMember(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()
	viewerRoleID := uuid.New().String()

	enabledUser := user.User{ID: userID, Title: "test-user", Email: "test@acme.dev", State: user.Enabled}
	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}

	tests := []struct {
		name           string
		setup          func(*mocks.PolicyService, *mocks.RelationService, *mocks.RoleService, *mocks.OrgService, *mocks.UserService, *mocks.AuditRecordRepository)
		orgID          string
		userID         string
		principalType  string
		roleID         string
		wantErr        error
		wantErrContain string
	}{
		{
			name: "should return error if principal type is unsupported",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, orgSvc *mocks.OrgService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
			},
			orgID:         orgID,
			userID:        userID,
			principalType: "app/unknown",
			roleID:        viewerRoleID,
			wantErr:       membership.ErrInvalidPrincipal,
		},
		{
			name: "should return error if org does not exist",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, orgSvc *mocks.OrgService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: organization.ErrNotExist,
		},
		{
			name: "should return error if user does not exist",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{}, user.ErrNotExist)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: user.ErrNotExist,
		},
		{
			name: "should return error if user is disabled",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID, State: user.Disabled}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: user.ErrDisabled,
		},
		{
			name: "should return error if role does not exist",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{}, role.ErrNotExist)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: role.ErrNotExist,
		},
		{
			name: "should return error if role is not valid for org scope",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.ProjectNamespace}}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: membership.ErrInvalidOrgRole,
		},
		{
			name: "should return error if org-specific role has project scope",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				// custom role created for this org but scoped to project, not org
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, OrgID: orgID, Scopes: []string{schema.ProjectNamespace}}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: membership.ErrInvalidOrgRole,
		},
		{
			name: "should return error if user is already a member",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "existing-policy"}}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: membership.ErrAlreadyMember,
		},
		{
			name: "should succeed adding a new member with viewer role",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: viewerRoleID, ResourceID: orgID, ResourceType: schema.OrganizationNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				relSvc.EXPECT().Create(ctx, relation.Relation{
					Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
					Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
					RelationName: schema.MemberRelationName,
				}).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: nil,
		},
		{
			name: "should succeed adding a new member with owner role and create owner relation",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, ownerRoleID).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: ownerRoleID, ResourceID: orgID, ResourceType: schema.OrganizationNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				relSvc.EXPECT().Create(ctx, relation.Relation{
					Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
					Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
					RelationName: schema.OwnerRelationName,
				}).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  ownerRoleID,
			wantErr: nil,
		},
		{
			name: "should succeed with org-specific custom role",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, OrgID: orgID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
				relSvc.EXPECT().Create(ctx, mock.Anything).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: nil,
		},
		{
			name: "should return error if listing existing policies fails",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return(nil, errors.New("db connection error"))
			},
			orgID:          orgID,
			userID:         userID,
			roleID:         viewerRoleID,
			wantErrContain: "db connection error",
		},
		{
			name: "should return error if policy creation fails",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, errors.New("policy create failed"))
			},
			orgID:          orgID,
			userID:         userID,
			roleID:         viewerRoleID,
			wantErrContain: "policy create failed",
		},
		{
			name: "should return error and cleanup policy if relation creation fails",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{ID: "created-policy-1"}, nil)
				relSvc.EXPECT().Create(ctx, mock.Anything).Return(relation.Relation{}, errors.New("spicedb unavailable"))
				// compensating delete should be called
				policySvc.EXPECT().Delete(ctx, "created-policy-1").Return(nil)
			},
			orgID:          orgID,
			userID:         userID,
			roleID:         viewerRoleID,
			wantErrContain: "spicedb unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockRelSvc := mocks.NewRelationService(t)
			mockRoleSvc := mocks.NewRoleService(t)
			mockOrgSvc := mocks.NewOrgService(t)
			mockUserSvc := mocks.NewUserService(t)
			mockAuditRepo := mocks.NewAuditRecordRepository(t)

			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mockUserSvc, mockAuditRepo)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mockUserSvc, mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)

			principalType := tt.principalType
			if principalType == "" {
				principalType = schema.UserPrincipal
			}
			err := svc.AddOrganizationMember(ctx, tt.orgID, tt.userID, principalType, tt.roleID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.wantErrContain != "" {
				assert.ErrorContains(t, err, tt.wantErrContain)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_AddOrganizationMember_ServiceUser(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	suID := uuid.New().String()
	viewerRoleID := uuid.New().String()

	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}

	t.Run("should succeed adding a service user", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockSuSvc := mocks.NewServiceuserService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockSuSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID, Title: "test-su", State: string(serviceuser.Enabled)}, nil)
		mockRoleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal}).Return([]policy.Policy{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
		mockRelSvc.EXPECT().Create(ctx, mock.Anything).Return(relation.Relation{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mockSuSvc, mockAuditRepo)
		err := svc.AddOrganizationMember(ctx, orgID, suID, schema.ServiceUserPrincipal, viewerRoleID)
		assert.NoError(t, err)
	})

	t.Run("should reject service user from different org", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockSuSvc := mocks.NewServiceuserService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockSuSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: "other-org", State: string(serviceuser.Enabled)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mockSuSvc, mocks.NewAuditRecordRepository(t))
		err := svc.AddOrganizationMember(ctx, orgID, suID, schema.ServiceUserPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, membership.ErrPrincipalNotInOrg)
	})

	t.Run("should reject disabled service user", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockSuSvc := mocks.NewServiceuserService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockSuSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID, State: string(serviceuser.Disabled)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mockSuSvc, mocks.NewAuditRecordRepository(t))
		err := svc.AddOrganizationMember(ctx, orgID, suID, schema.ServiceUserPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, serviceuser.ErrDisabled)
	})
}

func TestService_SetOrganizationMemberRole(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()
	viewerRoleID := uuid.New().String()
	managerRoleID := uuid.New().String()

	enabledUser := user.User{ID: userID, Title: "test-user", Email: "test@acme.dev", State: user.Enabled}
	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}

	orgRelation := func(name string) relation.Relation {
		return relation.Relation{
			Object:       relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
			Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
			RelationName: name,
		}
	}

	tests := []struct {
		name           string
		setup          func(*mocks.PolicyService, *mocks.RelationService, *mocks.RoleService, *mocks.OrgService, *mocks.UserService, *mocks.AuditRecordRepository)
		principalType  string
		roleID         string
		wantErr        error
		wantErrContain string
	}{
		{
			name: "should return error if principal type is unsupported",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, orgSvc *mocks.OrgService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
			},
			principalType: "app/unknown",
			roleID:        viewerRoleID,
			wantErr:       membership.ErrInvalidPrincipal,
		},
		{
			name: "should return error if user is not a member",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
			},
			roleID:  viewerRoleID,
			wantErr: membership.ErrNotMember,
		},
		{
			name: "should skip write when role is unchanged",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "existing-p1", RoleID: viewerRoleID}}, nil)
				// no Delete or Create should be called
			},
			roleID:  viewerRoleID,
			wantErr: nil,
		},
		{
			name: "should return error if demoting last owner",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, managerRoleID).Return(role.Role{ID: managerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				// user is the only owner
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
			},
			roleID:  managerRoleID,
			wantErr: membership.ErrLastOwnerRole,
		},
		{
			name: "should return ErrLastOwnerRole when DB guard rejects concurrent demotion",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				// app-level check passes (sees 2 owners)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
				// DB-level guard rejects (concurrent request already deleted the other owner)
				policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(policy.ErrLastRoleGuard)
			},
			roleID:  viewerRoleID,
			wantErr: membership.ErrLastOwnerRole,
		},
		{
			name: "should succeed demoting owner to viewer with multiple owners",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}, {ID: "p2", RoleID: ownerRoleID}}, nil)
				// replace policy with owner guard
				policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: viewerRoleID, ResourceID: orgID, ResourceType: schema.OrganizationNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{ID: "new-p"}, nil)
				// replace relation: delete both owner and member, then create member
				relSvc.EXPECT().Delete(ctx, orgRelation(schema.OwnerRelationName)).Return(nil)
				relSvc.EXPECT().Delete(ctx, orgRelation(schema.MemberRelationName)).Return(nil)
				relSvc.EXPECT().Create(ctx, orgRelation(schema.MemberRelationName)).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			roleID:  viewerRoleID,
			wantErr: nil,
		},
		{
			name: "should succeed promoting viewer to owner",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, ownerRoleID).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: viewerRoleID}}, nil)
				// promoting to owner — min-owner constraint doesn't apply
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				// existing policy is viewer (non-owner), uses plain Delete
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: ownerRoleID, ResourceID: orgID, ResourceType: schema.OrganizationNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{ID: "new-p"}, nil)
				// replace relation: delete both, create owner
				relSvc.EXPECT().Delete(ctx, orgRelation(schema.OwnerRelationName)).Return(nil)
				relSvc.EXPECT().Delete(ctx, orgRelation(schema.MemberRelationName)).Return(nil)
				relSvc.EXPECT().Create(ctx, orgRelation(schema.OwnerRelationName)).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			roleID:  ownerRoleID,
			wantErr: nil,
		},
		{
			name: "should return error and log if relation delete fails",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
				policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
				// relation delete fails with a real error — logged, no rollback
				relSvc.EXPECT().Delete(ctx, orgRelation(schema.OwnerRelationName)).Return(errors.New("spicedb connection error"))
			},
			roleID:         viewerRoleID,
			wantErrContain: "delete relation owner",
		},
		{
			name: "should ignore not-found on relation delete",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: managerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				// existing policy is manager (non-owner), uses plain Delete
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
				// both relation deletes return not-found — that's fine, should continue
				relSvc.EXPECT().Delete(ctx, orgRelation(schema.OwnerRelationName)).Return(relation.ErrNotExist)
				relSvc.EXPECT().Delete(ctx, orgRelation(schema.MemberRelationName)).Return(relation.ErrNotExist)
				relSvc.EXPECT().Create(ctx, orgRelation(schema.MemberRelationName)).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			roleID:  viewerRoleID,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockRelSvc := mocks.NewRelationService(t)
			mockRoleSvc := mocks.NewRoleService(t)
			mockOrgSvc := mocks.NewOrgService(t)
			mockUserSvc := mocks.NewUserService(t)
			mockAuditRepo := mocks.NewAuditRecordRepository(t)

			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mockUserSvc, mockAuditRepo)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mockUserSvc, mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)

			principalType := tt.principalType
			if principalType == "" {
				principalType = schema.UserPrincipal
			}
			err := svc.SetOrganizationMemberRole(ctx, orgID, userID, principalType, tt.roleID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.wantErrContain != "" {
				assert.ErrorContains(t, err, tt.wantErrContain)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_SetOrganizationMemberRole_ServiceUser(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	suID := uuid.New().String()
	viewerRoleID := uuid.New().String()
	ownerRoleID := uuid.New().String()

	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}

	t.Run("should succeed changing service user role", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockSuSvc := mocks.NewServiceuserService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockSuSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID, Title: "test-su", State: string(serviceuser.Enabled)}, nil)
		mockRoleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
		mockRoleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
		mockPolicySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
		mockRelSvc.EXPECT().Delete(ctx, mock.Anything).Return(relation.ErrNotExist).Times(2)
		mockRelSvc.EXPECT().Create(ctx, mock.Anything).Return(relation.Relation{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mockSuSvc, mockAuditRepo)
		err := svc.SetOrganizationMemberRole(ctx, orgID, suID, schema.ServiceUserPrincipal, viewerRoleID)
		assert.NoError(t, err)
	})

	t.Run("should reject service user from different org", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockSuSvc := mocks.NewServiceuserService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockSuSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: "other-org", State: string(serviceuser.Enabled)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mockSuSvc, mocks.NewAuditRecordRepository(t))
		err := svc.SetOrganizationMemberRole(ctx, orgID, suID, schema.ServiceUserPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, membership.ErrPrincipalNotInOrg)
	})

	t.Run("should reject disabled service user", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockSuSvc := mocks.NewServiceuserService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockSuSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID, State: string(serviceuser.Disabled)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mockSuSvc, mocks.NewAuditRecordRepository(t))
		err := svc.SetOrganizationMemberRole(ctx, orgID, suID, schema.ServiceUserPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, serviceuser.ErrDisabled)
	})
}

func TestService_RemoveOrganizationMember(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()
	viewerRoleID := uuid.New().String()
	projectID := uuid.New().String()
	groupID := uuid.New().String()

	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}

	orgObj := relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace}
	grpObj := relation.Object{ID: groupID, Namespace: schema.GroupNamespace}
	userSub := relation.Subject{ID: userID, Namespace: schema.UserPrincipal}

	type testDeps struct {
		policySvc *mocks.PolicyService
		relSvc    *mocks.RelationService
		roleSvc   *mocks.RoleService
		orgSvc    *mocks.OrgService
		projSvc   *mocks.ProjectService
		grpSvc    *mocks.GroupService
		auditRepo *mocks.AuditRecordRepository
	}

	tests := []struct {
		name           string
		principalType  string
		setup          func(d testDeps)
		wantErr        error
		wantErrContain string
	}{
		{
			name:          "should return error for invalid principal type",
			principalType: "app/invalid",
			setup:         func(d testDeps) {},
			wantErr:       membership.ErrInvalidPrincipalType,
		},
		{
			name: "should return error if org does not exist",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			wantErr: organization.ErrNotExist,
		},
		{
			name: "should return error if org is disabled",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			wantErr: organization.ErrDisabled,
		},
		{
			name: "should return ErrNotMember if principal has no org policies",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
			},
			wantErr: membership.ErrNotMember,
		},
		{
			name: "should return ErrLastOwnerRole when removing the last owner",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
			},
			wantErr: membership.ErrLastOwnerRole,
		},
		{
			name: "should return error if listing projects fails",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return(nil, errors.New("db error"))
			},
			wantErrContain: "list org projects",
		},
		{
			name: "should return error if listing groups fails",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return(nil, errors.New("db error"))
			},
			wantErrContain: "list org groups",
		},
		{
			name: "should remove viewer with no sub-resource policies",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID},
				}, nil)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: orgObj, Subject: userSub, RelationName: schema.OwnerRelationName}).Return(relation.ErrNotExist)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: orgObj, Subject: userSub, RelationName: schema.MemberRelationName}).Return(nil)
				d.auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "should cascade remove policies from projects and groups and clean up relations",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: ownerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "org-p1"}, {ID: "other-owner-p"}}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{{ID: projectID}}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{{ID: groupID}}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID},
					{ID: "proj-p1", ResourceType: schema.ProjectNamespace, ResourceID: projectID},
					{ID: "grp-p1", ResourceType: schema.GroupNamespace, ResourceID: groupID},
					{ID: "other-org-p", ResourceType: schema.OrganizationNamespace, ResourceID: "other-org-id"},
				}, nil)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.policySvc.EXPECT().Delete(ctx, "proj-p1").Return(nil)
				d.policySvc.EXPECT().Delete(ctx, "grp-p1").Return(nil)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: orgObj, Subject: userSub, RelationName: schema.OwnerRelationName}).Return(nil)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: orgObj, Subject: userSub, RelationName: schema.MemberRelationName}).Return(relation.ErrNotExist)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: grpObj, Subject: userSub, RelationName: schema.OwnerRelationName}).Return(nil)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: grpObj, Subject: userSub, RelationName: schema.MemberRelationName}).Return(relation.ErrNotExist)
				d.auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "should not delete policies belonging to other orgs",
			setup: func(d testDeps) {
				otherOrgID := uuid.New().String()
				otherProjectID := uuid.New().String()
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID},
					{ID: "other-org-p", ResourceType: schema.OrganizationNamespace, ResourceID: otherOrgID},
					{ID: "other-proj-p", ResourceType: schema.ProjectNamespace, ResourceID: otherProjectID},
				}, nil)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: orgObj, Subject: userSub, RelationName: schema.OwnerRelationName}).Return(relation.ErrNotExist)
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: orgObj, Subject: userSub, RelationName: schema.MemberRelationName}).Return(nil)
				d.auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "should return error if policy deletion fails during cascade",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{{ID: projectID}}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "proj-p1", ResourceType: schema.ProjectNamespace, ResourceID: projectID},
				}, nil)
				d.policySvc.EXPECT().Delete(ctx, "proj-p1").Return(errors.New("delete failed"))
			},
			wantErrContain: "delete sub-resource policy",
		},
		{
			name: "should return error if org relation removal fails after org policies deleted",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID, RoleID: viewerRoleID},
				}, nil)
				// org policy deleted first (viewer, plain Delete)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				// then relation removal fails
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: orgObj, Subject: userSub, RelationName: schema.OwnerRelationName}).Return(errors.New("spicedb down"))
			},
			wantErrContain: "remove org relations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := testDeps{
				policySvc: mocks.NewPolicyService(t),
				relSvc:    mocks.NewRelationService(t),
				roleSvc:   mocks.NewRoleService(t),
				orgSvc:    mocks.NewOrgService(t),
				projSvc:   mocks.NewProjectService(t),
				grpSvc:    mocks.NewGroupService(t),
				auditRepo: mocks.NewAuditRecordRepository(t),
			}
			if tt.setup != nil {
				tt.setup(d)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), d.policySvc, d.relSvc, d.roleSvc, d.orgSvc, mocks.NewUserService(t), d.projSvc, d.grpSvc, mocks.NewServiceuserService(t), d.auditRepo)

			principalType := tt.principalType
			if principalType == "" {
				principalType = schema.UserPrincipal
			}
			err := svc.RemoveOrganizationMember(ctx, orgID, userID, principalType)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.wantErrContain != "" {
				assert.ErrorContains(t, err, tt.wantErrContain)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_SetProjectMemberRole(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New().String()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	suID := uuid.New().String()
	groupID := uuid.New().String()
	roleID := uuid.New().String()

	prj := project.Project{
		ID:           projectID,
		Organization: organization.Organization{ID: orgID},
	}

	tests := []struct {
		name          string
		setup         func(*mocks.PolicyService, *mocks.RoleService, *mocks.ProjectService, *mocks.UserService, *mocks.ServiceuserService, *mocks.GroupService, *mocks.AuditRecordRepository)
		principalID   string
		principalType string
		roleID        string
		wantErr       error
	}{
		{
			name: "should return error if project does not exist",
			setup: func(_ *mocks.PolicyService, _ *mocks.RoleService, prjSvc *mocks.ProjectService, _ *mocks.UserService, _ *mocks.ServiceuserService, _ *mocks.GroupService, _ *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(project.Project{}, project.ErrNotExist)
			},
			principalID: userID, principalType: schema.UserPrincipal, roleID: roleID,
			wantErr: project.ErrNotExist,
		},
		{
			name: "should return error if role is not project-scoped",
			setup: func(_ *mocks.PolicyService, roleSvc *mocks.RoleService, prjSvc *mocks.ProjectService, _ *mocks.UserService, _ *mocks.ServiceuserService, _ *mocks.GroupService, _ *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
			},
			principalID: userID, principalType: schema.UserPrincipal, roleID: roleID,
			wantErr: membership.ErrInvalidProjectRole,
		},
		{
			name: "should return error if user is not org member",
			setup: func(policySvc *mocks.PolicyService, roleSvc *mocks.RoleService, prjSvc *mocks.ProjectService, userSvc *mocks.UserService, _ *mocks.ServiceuserService, _ *mocks.GroupService, _ *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID, State: user.Enabled}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
			},
			principalID: userID, principalType: schema.UserPrincipal, roleID: roleID,
			wantErr: membership.ErrNotOrgMember,
		},
		{
			name: "should return error if service user is not in org",
			setup: func(_ *mocks.PolicyService, roleSvc *mocks.RoleService, prjSvc *mocks.ProjectService, _ *mocks.UserService, suSvc *mocks.ServiceuserService, _ *mocks.GroupService, _ *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				suSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: "other-org"}, nil)
			},
			principalID: suID, principalType: schema.ServiceUserPrincipal, roleID: roleID,
			wantErr: membership.ErrNotOrgMember,
		},
		{
			name: "should succeed adding new user to project",
			setup: func(policySvc *mocks.PolicyService, roleSvc *mocks.RoleService, prjSvc *mocks.ProjectService, userSvc *mocks.UserService, _ *mocks.ServiceuserService, _ *mocks.GroupService, auditRepo *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID, State: user.Enabled}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1"}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: roleID, ResourceID: projectID, ResourceType: schema.ProjectNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			principalID: userID, principalType: schema.UserPrincipal, roleID: roleID,
		},
		{
			name: "should succeed adding service user to project",
			setup: func(policySvc *mocks.PolicyService, roleSvc *mocks.RoleService, prjSvc *mocks.ProjectService, _ *mocks.UserService, suSvc *mocks.ServiceuserService, _ *mocks.GroupService, auditRepo *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				suSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: roleID, ResourceID: projectID, ResourceType: schema.ProjectNamespace,
					PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal,
				}).Return(policy.Policy{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			principalID: suID, principalType: schema.ServiceUserPrincipal, roleID: roleID,
		},
		{
			name: "should succeed adding group to project",
			setup: func(policySvc *mocks.PolicyService, roleSvc *mocks.RoleService, prjSvc *mocks.ProjectService, _ *mocks.UserService, _ *mocks.ServiceuserService, grpSvc *mocks.GroupService, auditRepo *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				grpSvc.EXPECT().Get(ctx, groupID).Return(group.Group{ID: groupID, OrganizationID: orgID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: groupID, PrincipalType: schema.GroupPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: roleID, ResourceID: projectID, ResourceType: schema.ProjectNamespace,
					PrincipalID: groupID, PrincipalType: schema.GroupPrincipal,
				}).Return(policy.Policy{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			principalID: groupID, principalType: schema.GroupPrincipal, roleID: roleID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockRoleSvc := mocks.NewRoleService(t)
			mockPrjSvc := mocks.NewProjectService(t)
			mockUserSvc := mocks.NewUserService(t)
			mockSuSvc := mocks.NewServiceuserService(t)
			mockGrpSvc := mocks.NewGroupService(t)
			mockAuditRepo := mocks.NewAuditRecordRepository(t)

			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockRoleSvc, mockPrjSvc, mockUserSvc, mockSuSvc, mockGrpSvc, mockAuditRepo)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mocks.NewOrgService(t), mockUserSvc, mockPrjSvc, mockGrpSvc, mockSuSvc, mockAuditRepo)
			err := svc.SetProjectMemberRole(ctx, projectID, tt.principalID, tt.principalType, tt.roleID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_RemoveProjectMember(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New().String()
	userID := uuid.New().String()
	suID := uuid.New().String()

	prj := project.Project{
		ID:           projectID,
		Title:        "Test Project",
		Organization: organization.Organization{ID: uuid.New().String()},
	}

	tests := []struct {
		name          string
		setup         func(*mocks.PolicyService, *mocks.ProjectService, *mocks.AuditRecordRepository)
		principalID   string
		principalType string
		wantErr       error
	}{
		{
			name:          "should return error for invalid principal type",
			principalID:   userID,
			principalType: "app/invalid",
			wantErr:       membership.ErrInvalidPrincipalType,
		},
		{
			name: "should return error if not a member",
			setup: func(policySvc *mocks.PolicyService, prjSvc *mocks.ProjectService, _ *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
			},
			principalID:   userID,
			principalType: schema.UserPrincipal,
			wantErr:       membership.ErrNotMember,
		},
		{
			name: "should succeed removing a user",
			setup: func(policySvc *mocks.PolicyService, prjSvc *mocks.ProjectService, auditRepo *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1"}}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			principalID:   userID,
			principalType: schema.UserPrincipal,
		},
		{
			name: "should succeed removing a service user",
			setup: func(policySvc *mocks.PolicyService, prjSvc *mocks.ProjectService, auditRepo *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal}).Return([]policy.Policy{{ID: "p1"}}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			principalID:   suID,
			principalType: schema.ServiceUserPrincipal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockPrjSvc := mocks.NewProjectService(t)
			mockAuditRepo := mocks.NewAuditRecordRepository(t)

			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockPrjSvc, mockAuditRepo)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mockPrjSvc, mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
			err := svc.RemoveProjectMember(ctx, projectID, tt.principalID, tt.principalType)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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

func TestService_SetGroupMemberRole(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	groupID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()
	memberRoleID := uuid.New().String()

	enabledUser := user.User{ID: userID, Title: "test-user", Email: "test@acme.dev", State: user.Enabled}
	grp := group.Group{ID: groupID, OrganizationID: orgID, Title: "Test Group"}

	groupMemberRelation := func(name string) relation.Relation {
		return relation.Relation{
			Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
			Subject:      relation.Subject{ID: userID, Namespace: schema.UserPrincipal},
			RelationName: name,
		}
	}

	tests := []struct {
		name           string
		setup          func(*mocks.PolicyService, *mocks.RelationService, *mocks.RoleService, *mocks.GroupService, *mocks.UserService, *mocks.AuditRecordRepository)
		principalType  string
		roleID         string
		wantErr        error
		wantErrContain string
	}{
		{
			name: "should add member on upsert when no existing group policy and user is in org",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: schema.GroupMemberRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				// org-membership check: user must be in org
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1"}}, nil)
				// create policy + relation
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: memberRoleID, ResourceID: groupID, ResourceType: schema.GroupNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{ID: "new-p"}, nil)
				relSvc.EXPECT().Create(ctx, groupMemberRelation(schema.MemberRelationName)).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			roleID:  memberRoleID,
			wantErr: nil,
		},
		{
			name: "should reject upsert-add if principal is not a member of the org",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: schema.GroupMemberRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
			},
			roleID:  memberRoleID,
			wantErr: membership.ErrNotOrgMember,
		},
		{
			name: "should skip write when role is unchanged",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: schema.GroupMemberRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: memberRoleID}}, nil)
			},
			roleID:  memberRoleID,
			wantErr: nil,
		},
		{
			name: "should return error if demoting last owner",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: schema.GroupMemberRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
			},
			roleID:  memberRoleID,
			wantErr: membership.ErrLastGroupOwnerRole,
		},
		{
			name: "should succeed demoting owner to member with multiple owners (relation flips owner->member)",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: schema.GroupMemberRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
				// deleting an owner-role policy uses the atomic guard
				policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: memberRoleID, ResourceID: groupID, ResourceType: schema.GroupNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{ID: "new-p"}, nil)
				relSvc.EXPECT().Delete(ctx, groupMemberRelation(schema.OwnerRelationName)).Return(nil)
				relSvc.EXPECT().Delete(ctx, groupMemberRelation(schema.MemberRelationName)).Return(relation.ErrNotExist)
				relSvc.EXPECT().Create(ctx, groupMemberRelation(schema.MemberRelationName)).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			roleID:  memberRoleID,
			wantErr: nil,
		},
		{
			name: "should surface ErrLastGroupOwnerRole when DeleteWithMinRoleGuard races (TOCTOU)",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: schema.GroupMemberRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole}, nil)
				// pre-check sees two owners
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
				// but a concurrent delete makes this the last owner; the DB guard catches it
				policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(policy.ErrLastRoleGuard)
			},
			roleID:  memberRoleID,
			wantErr: membership.ErrLastGroupOwnerRole,
		},
		{
			name: "should succeed promoting member to owner (relation flips member->owner)",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, ownerRoleID).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: memberRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{ID: "new-p"}, nil)
				relSvc.EXPECT().Delete(ctx, groupMemberRelation(schema.OwnerRelationName)).Return(relation.ErrNotExist)
				relSvc.EXPECT().Delete(ctx, groupMemberRelation(schema.MemberRelationName)).Return(nil)
				relSvc.EXPECT().Create(ctx, groupMemberRelation(schema.OwnerRelationName)).Return(relation.Relation{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			roleID:  ownerRoleID,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockRelSvc := mocks.NewRelationService(t)
			mockRoleSvc := mocks.NewRoleService(t)
			mockGrpSvc := mocks.NewGroupService(t)
			mockUserSvc := mocks.NewUserService(t)
			mockAuditRepo := mocks.NewAuditRecordRepository(t)

			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockRelSvc, mockRoleSvc, mockGrpSvc, mockUserSvc, mockAuditRepo)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mocks.NewOrgService(t), mockUserSvc, mocks.NewProjectService(t), mockGrpSvc, mocks.NewServiceuserService(t), mockAuditRepo)

			principalType := tt.principalType
			if principalType == "" {
				principalType = schema.UserPrincipal
			}
			err := svc.SetGroupMemberRole(ctx, groupID, userID, principalType, tt.roleID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.wantErrContain != "" {
				assert.ErrorContains(t, err, tt.wantErrContain)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_OnGroupCreated(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	groupID := uuid.New().String()
	creatorID := uuid.New().String()
	ownerRoleID := uuid.New().String()

	enabledUser := user.User{ID: creatorID, Title: "creator", Email: "creator@acme.dev", State: user.Enabled}
	grp := group.Group{ID: groupID, OrganizationID: orgID, Title: "Test Group"}

	groupOrgRelation := relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
	}
	orgGroupMemberRelation := relation.Relation{
		Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
		Subject: relation.Subject{
			ID:              groupID,
			Namespace:       schema.GroupNamespace,
			SubRelationName: schema.MemberRelationName,
		},
		RelationName: schema.MemberRelationName,
	}
	creatorOwnerRelation := relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: creatorID, Namespace: schema.UserPrincipal},
		RelationName: schema.OwnerRelationName,
	}

	t.Run("should link group<->org and add creator as owner", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockGrpSvc := mocks.NewGroupService(t)
		mockUserSvc := mocks.NewUserService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockRelSvc.EXPECT().Create(ctx, groupOrgRelation).Return(relation.Relation{}, nil)
		mockRelSvc.EXPECT().Create(ctx, orgGroupMemberRelation).Return(relation.Relation{}, nil)

		mockGrpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
		mockUserSvc.EXPECT().GetByID(ctx, creatorID).Return(enabledUser, nil)
		mockRoleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole, Scopes: []string{schema.GroupNamespace}}, nil)
		mockUserSvc.EXPECT().GetByID(ctx, creatorID).Return(enabledUser, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: creatorID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1"}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: creatorID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{ID: "new-p"}, nil)
		mockRelSvc.EXPECT().Create(ctx, creatorOwnerRelation).Return(relation.Relation{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mocks.NewOrgService(t), mockUserSvc, mocks.NewProjectService(t), mockGrpSvc, mocks.NewServiceuserService(t), mockAuditRepo)

		err := svc.OnGroupCreated(ctx, groupID, orgID, creatorID, schema.UserPrincipal)
		assert.NoError(t, err)
	})

	t.Run("should return error if hierarchy relation creation fails", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)

		mockRelSvc.EXPECT().Create(ctx, groupOrgRelation).Return(relation.Relation{}, errors.New("spicedb unavailable"))

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

		err := svc.OnGroupCreated(ctx, groupID, orgID, creatorID, schema.UserPrincipal)
		assert.ErrorContains(t, err, "link group to org")
	})

	t.Run("should rollback first hierarchy relation if second fails", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)

		mockRelSvc.EXPECT().Create(ctx, groupOrgRelation).Return(relation.Relation{}, nil)
		mockRelSvc.EXPECT().Create(ctx, orgGroupMemberRelation).Return(relation.Relation{}, errors.New("spicedb unavailable"))
		// rollback: delete the first hierarchy relation
		mockRelSvc.EXPECT().Delete(ctx, groupOrgRelation).Return(nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

		err := svc.OnGroupCreated(ctx, groupID, orgID, creatorID, schema.UserPrincipal)
		assert.ErrorContains(t, err, "add group as org member")
	})

	t.Run("should rollback both hierarchy relations if owner add fails", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockGrpSvc := mocks.NewGroupService(t)
		mockUserSvc := mocks.NewUserService(t)

		// linkGroupToOrg succeeds
		mockRelSvc.EXPECT().Create(ctx, groupOrgRelation).Return(relation.Relation{}, nil)
		mockRelSvc.EXPECT().Create(ctx, orgGroupMemberRelation).Return(relation.Relation{}, nil)

		// AddGroupMember fails before policy creation (group fetch fails)
		mockGrpSvc.EXPECT().Get(ctx, groupID).Return(group.Group{}, errors.New("db down"))

		// unused mocks: only here for completeness, won't be called
		_ = mockRoleSvc
		_ = mockUserSvc

		// rollback: delete both hierarchy relations
		mockRelSvc.EXPECT().Delete(ctx, groupOrgRelation).Return(nil)
		mockRelSvc.EXPECT().Delete(ctx, orgGroupMemberRelation).Return(nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mocks.NewOrgService(t), mockUserSvc, mocks.NewProjectService(t), mockGrpSvc, mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

		err := svc.OnGroupCreated(ctx, groupID, orgID, creatorID, schema.UserPrincipal)
		assert.ErrorContains(t, err, "db down")
	})
}

func TestService_RemoveGroupMember(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	groupID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()
	memberRoleID := uuid.New().String()

	enabledUser := user.User{ID: userID, Title: "test-user", Email: "test@acme.dev", State: user.Enabled}
	grp := group.Group{ID: groupID, OrganizationID: orgID, Title: "Test Group"}

	expectOwnerRoleLookup := func(roleSvc *mocks.RoleService) {
		roleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole}, nil)
	}

	tests := []struct {
		name           string
		setup          func(*mocks.PolicyService, *mocks.RelationService, *mocks.RoleService, *mocks.GroupService, *mocks.UserService, *mocks.AuditRecordRepository)
		principalType  string
		wantErr        error
		wantErrContain string
	}{
		{
			name: "should return error if group does not exist",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, grpSvc *mocks.GroupService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(group.Group{}, group.ErrNotExist)
			},
			wantErr: group.ErrNotExist,
		},
		{
			name: "should return error if principal type is unsupported",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, grpSvc *mocks.GroupService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
			},
			principalType: schema.ServiceUserPrincipal,
			wantErr:       membership.ErrInvalidPrincipalType,
		},
		{
			name: "should return ErrNotMember if principal has no group policy",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
			},
			wantErr: membership.ErrNotMember,
		},
		{
			name: "should return ErrLastGroupOwnerRole when removing last owner (pre-check fires)",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				expectOwnerRoleLookup(roleSvc)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}}, nil)
			},
			wantErr: membership.ErrLastGroupOwnerRole,
		},
		{
			name: "should surface ErrLastGroupOwnerRole when DeleteWithMinRoleGuard races (TOCTOU)",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, _ *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				expectOwnerRoleLookup(roleSvc)
				// pre-check sees two owners, but the DB guard catches the race
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
				policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(policy.ErrLastRoleGuard)
			},
			wantErr: membership.ErrLastGroupOwnerRole,
		},
		{
			name: "should remove a member (non-owner) and delete both relations",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: memberRoleID}}, nil)
				expectOwnerRoleLookup(roleSvc)
				// member-role policy → plain Delete (no guard)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				obj := relation.Object{ID: groupID, Namespace: schema.GroupNamespace}
				sub := relation.Subject{ID: userID, Namespace: schema.UserPrincipal}
				relSvc.EXPECT().Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: schema.OwnerRelationName}).Return(relation.ErrNotExist)
				relSvc.EXPECT().Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: schema.MemberRelationName}).Return(nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "should remove an owner via atomic guard when more owners remain",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				expectOwnerRoleLookup(roleSvc)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
				policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "p1", ownerRoleID).Return(nil)
				obj := relation.Object{ID: groupID, Namespace: schema.GroupNamespace}
				sub := relation.Subject{ID: userID, Namespace: schema.UserPrincipal}
				relSvc.EXPECT().Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: schema.OwnerRelationName}).Return(nil)
				relSvc.EXPECT().Delete(ctx, relation.Relation{Object: obj, Subject: sub, RelationName: schema.MemberRelationName}).Return(relation.ErrNotExist)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicySvc := mocks.NewPolicyService(t)
			mockRelSvc := mocks.NewRelationService(t)
			mockRoleSvc := mocks.NewRoleService(t)
			mockGrpSvc := mocks.NewGroupService(t)
			mockUserSvc := mocks.NewUserService(t)
			mockAuditRepo := mocks.NewAuditRecordRepository(t)

			if tt.setup != nil {
				tt.setup(mockPolicySvc, mockRelSvc, mockRoleSvc, mockGrpSvc, mockUserSvc, mockAuditRepo)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mocks.NewOrgService(t), mockUserSvc, mocks.NewProjectService(t), mockGrpSvc, mocks.NewServiceuserService(t), mockAuditRepo)

			principalType := tt.principalType
			if principalType == "" {
				principalType = schema.UserPrincipal
			}
			err := svc.RemoveGroupMember(ctx, groupID, userID, principalType)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.wantErrContain != "" {
				assert.ErrorContains(t, err, tt.wantErrContain)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_RemoveAllGroupMembers(t *testing.T) {
	ctx := context.Background()
	groupID := uuid.New().String()
	userA := uuid.New().String()
	userB := uuid.New().String()

	relFor := func(name, principalID string) relation.Relation {
		return relation.Relation{
			Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
			Subject:      relation.Subject{ID: principalID, Namespace: schema.UserPrincipal},
			RelationName: name,
		}
	}

	t.Run("removes policies and per-principal relations, dedupes principals across policies", func(t *testing.T) {
		policySvc := mocks.NewPolicyService(t)
		relSvc := mocks.NewRelationService(t)

		policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID}).Return([]policy.Policy{
			{ID: "p1", PrincipalID: userA, PrincipalType: schema.UserPrincipal},
			{ID: "p2", PrincipalID: userA, PrincipalType: schema.UserPrincipal},
			{ID: "p3", PrincipalID: userB, PrincipalType: schema.UserPrincipal},
		}, nil)
		policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
		policySvc.EXPECT().Delete(ctx, "p2").Return(nil)
		policySvc.EXPECT().Delete(ctx, "p3").Return(nil)
		relSvc.EXPECT().Delete(ctx, relFor(schema.OwnerRelationName, userA)).Return(relation.ErrNotExist)
		relSvc.EXPECT().Delete(ctx, relFor(schema.MemberRelationName, userA)).Return(nil)
		relSvc.EXPECT().Delete(ctx, relFor(schema.OwnerRelationName, userB)).Return(nil)
		relSvc.EXPECT().Delete(ctx, relFor(schema.MemberRelationName, userB)).Return(relation.ErrNotExist)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), policySvc, relSvc,
			mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t),
			mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t),
			mocks.NewAuditRecordRepository(t))

		assert.NoError(t, svc.RemoveAllGroupMembers(ctx, groupID))
	})

	t.Run("joins errors when a policy delete fails", func(t *testing.T) {
		policySvc := mocks.NewPolicyService(t)
		relSvc := mocks.NewRelationService(t)

		policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID}).Return([]policy.Policy{
			{ID: "p1", PrincipalID: userA, PrincipalType: schema.UserPrincipal},
			{ID: "p2", PrincipalID: userB, PrincipalType: schema.UserPrincipal},
		}, nil)
		policySvc.EXPECT().Delete(ctx, "p1").Return(errors.New("db down"))
		policySvc.EXPECT().Delete(ctx, "p2").Return(nil)
		relSvc.EXPECT().Delete(ctx, relFor(schema.OwnerRelationName, userB)).Return(nil)
		relSvc.EXPECT().Delete(ctx, relFor(schema.MemberRelationName, userB)).Return(nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), policySvc, relSvc,
			mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t),
			mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t),
			mocks.NewAuditRecordRepository(t))

		err := svc.RemoveAllGroupMembers(ctx, groupID)
		assert.ErrorContains(t, err, "db down")
	})
}

func TestService_OnGroupDeleted(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	groupID := uuid.New().String()
	grp := group.Group{ID: groupID, OrganizationID: orgID, Title: "T"}

	t.Run("removes members, group-as-principal policies, and unlinks from org", func(t *testing.T) {
		policySvc := mocks.NewPolicyService(t)
		relSvc := mocks.NewRelationService(t)
		grpSvc := mocks.NewGroupService(t)

		grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
		// RemoveAllGroupMembers — no member policies
		policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID}).Return([]policy.Policy{}, nil)
		// removeGroupAsPrincipalPolicies — one policy granting this group access elsewhere
		policySvc.EXPECT().List(ctx, policy.Filter{
			PrincipalType: schema.GroupPrincipal,
			PrincipalID:   groupID,
		}).Return([]policy.Policy{{ID: "principal-p1"}}, nil)
		policySvc.EXPECT().Delete(ctx, "principal-p1").Return(nil)

		// unlinkGroupFromOrg: both hierarchy relations
		relSvc.EXPECT().Delete(ctx, relation.Relation{
			Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
			Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
			RelationName: schema.OrganizationRelationName,
		}).Return(nil)
		relSvc.EXPECT().Delete(ctx, relation.Relation{
			Object: relation.Object{ID: orgID, Namespace: schema.OrganizationNamespace},
			Subject: relation.Subject{
				ID:              groupID,
				Namespace:       schema.GroupNamespace,
				SubRelationName: schema.MemberRelationName,
			},
			RelationName: schema.MemberRelationName,
		}).Return(nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), policySvc, relSvc,
			mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t),
			mocks.NewProjectService(t), grpSvc, mocks.NewServiceuserService(t),
			mocks.NewAuditRecordRepository(t))

		assert.NoError(t, svc.OnGroupDeleted(ctx, groupID))
	})

	t.Run("returns error if group not found", func(t *testing.T) {
		grpSvc := mocks.NewGroupService(t)
		grpSvc.EXPECT().Get(ctx, groupID).Return(group.Group{}, group.ErrNotExist)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)),
			mocks.NewPolicyService(t), mocks.NewRelationService(t),
			mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t),
			mocks.NewProjectService(t), grpSvc, mocks.NewServiceuserService(t),
			mocks.NewAuditRecordRepository(t))

		assert.ErrorIs(t, svc.OnGroupDeleted(ctx, groupID), group.ErrNotExist)
	})
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
