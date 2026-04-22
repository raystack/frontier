package membership_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/membership/mocks"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
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
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, orgSvc *mocks.OrgService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
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

			svc := membership.NewService(log.NewNoop(), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mockUserSvc, mocks.NewProjectService(t), mocks.NewGroupService(t), mockAuditRepo)

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
			name: "should return error if principal type is not user",
			setup: func(_ *mocks.PolicyService, _ *mocks.RelationService, _ *mocks.RoleService, orgSvc *mocks.OrgService, _ *mocks.UserService, _ *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
			},
			principalType: schema.ServiceUserPrincipal,
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
			name: "should succeed demoting owner to viewer with multiple owners",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "p1", RoleID: ownerRoleID}, {ID: "p2", RoleID: ownerRoleID}}, nil)
				// replace policy
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
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
				// replace policy
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
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
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

			svc := membership.NewService(log.NewNoop(), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mockUserSvc, mocks.NewProjectService(t), mocks.NewGroupService(t), mockAuditRepo)

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
			wantErrContain: "delete project policy",
		},
		{
			name: "should return error if org relation removal fails without deleting org policies",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID},
				}, nil)
				// org policy Delete should NOT be called — relations fail first, org policies are last
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

			svc := membership.NewService(log.NewNoop(), d.policySvc, d.relSvc, d.roleSvc, d.orgSvc, mocks.NewUserService(t), d.projSvc, d.grpSvc, d.auditRepo)

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
