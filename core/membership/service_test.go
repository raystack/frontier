package membership_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/membership/mocks"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/salt/log"
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
			name: "should return error if principal type is not user",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, _ *mocks.OrgService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				// no mocks needed — rejected before any service call
			},
			orgID:         orgID,
			userID:        userID,
			principalType: schema.ServiceUserPrincipal,
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

			svc := membership.NewService(log.NewNoop(), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mockUserSvc, mockAuditRepo)

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
