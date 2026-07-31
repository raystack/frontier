package membership_test

import (
	"context"
	"errors"
	"testing"

	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/membership/mocks"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
			name: "should succeed demoting owner to member with multiple owners",
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
			name: "should succeed promoting member to owner",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, ownerRoleID).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole, Scopes: []string{schema.GroupNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: memberRoleID}}, nil)
				roleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{ID: "new-p"}, nil)
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
	creatorMemberRelation := relation.Relation{
		Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
		Subject:      relation.Subject{ID: creatorID, Namespace: schema.UserPrincipal},
		RelationName: schema.MemberRelationName,
	}

	t.Run("should link group to org and add creator as owner", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockGrpSvc := mocks.NewGroupService(t)
		mockUserSvc := mocks.NewUserService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockRelSvc.EXPECT().Create(ctx, groupOrgRelation).Return(relation.Relation{}, nil)

		mockGrpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
		mockUserSvc.EXPECT().GetByID(ctx, creatorID).Return(enabledUser, nil)
		mockRoleSvc.EXPECT().Get(ctx, schema.GroupOwnerRole).Return(role.Role{ID: ownerRoleID, Name: schema.GroupOwnerRole, Scopes: []string{schema.GroupNamespace}}, nil)
		mockUserSvc.EXPECT().GetByID(ctx, creatorID).Return(enabledUser, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: creatorID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1"}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: creatorID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{ID: "new-p"}, nil)
		mockRelSvc.EXPECT().Create(ctx, creatorMemberRelation).Return(relation.Relation{}, nil)
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

	t.Run("should rollback hierarchy relations if owner add fails", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockGrpSvc := mocks.NewGroupService(t)
		mockUserSvc := mocks.NewUserService(t)

		// linkGroupToOrg succeeds
		mockRelSvc.EXPECT().Create(ctx, groupOrgRelation).Return(relation.Relation{}, nil)

		// AddGroupMember fails before policy creation (group fetch fails)
		mockGrpSvc.EXPECT().Get(ctx, groupID).Return(group.Group{}, errors.New("db down"))

		// unused mocks: only here for completeness, won't be called
		_ = mockRoleSvc
		_ = mockUserSvc

		// rollback: delete the identity link
		mockRelSvc.EXPECT().Delete(ctx, groupOrgRelation).Return(nil)

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
			name: "should remove a member (non-owner) and delete the member relation",
			setup: func(policySvc *mocks.PolicyService, relSvc *mocks.RelationService, roleSvc *mocks.RoleService, grpSvc *mocks.GroupService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				grpSvc.EXPECT().Get(ctx, groupID).Return(grp, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{GroupID: groupID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: memberRoleID}}, nil)
				expectOwnerRoleLookup(roleSvc)
				// member-role policy → plain Delete (no guard)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				obj := relation.Object{ID: groupID, Namespace: schema.GroupNamespace}
				sub := relation.Subject{ID: userID, Namespace: schema.UserPrincipal}
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
		relSvc.EXPECT().Delete(ctx, relFor(schema.MemberRelationName, userA)).Return(nil)
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

		// unlinkGroupFromOrg: the identity link
		relSvc.EXPECT().Delete(ctx, relation.Relation{
			Object:       relation.Object{ID: groupID, Namespace: schema.GroupNamespace},
			Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
			RelationName: schema.OrganizationRelationName,
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
