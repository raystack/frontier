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
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal, ResourceType: schema.ProjectNamespace}).Return([]policy.Policy{}, nil)
			},
			principalID:   userID,
			principalType: schema.UserPrincipal,
			wantErr:       membership.ErrNotMember,
		},
		{
			name: "should succeed removing a user",
			setup: func(policySvc *mocks.PolicyService, prjSvc *mocks.ProjectService, auditRepo *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal, ResourceType: schema.ProjectNamespace}).Return([]policy.Policy{{ID: "p1"}}, nil)
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
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal, ResourceType: schema.ProjectNamespace}).Return([]policy.Policy{{ID: "p1"}}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			principalID:   suID,
			principalType: schema.ServiceUserPrincipal,
		},
		{
			name: "should succeed removing a PAT",
			setup: func(policySvc *mocks.PolicyService, prjSvc *mocks.ProjectService, auditRepo *mocks.AuditRecordRepository) {
				prjSvc.EXPECT().Get(ctx, projectID).Return(prj, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.PATPrincipal, ResourceType: schema.ProjectNamespace}).Return([]policy.Policy{{ID: "p1"}}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			principalID:   userID,
			principalType: schema.PATPrincipal,
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

func TestService_OnProjectCreated(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	projectID := uuid.New().String()
	creatorID := uuid.New().String()

	projectOrgRelation := relation.Relation{
		Object:       relation.Object{ID: projectID, Namespace: schema.ProjectNamespace},
		Subject:      relation.Subject{ID: orgID, Namespace: schema.OrganizationNamespace},
		RelationName: schema.OrganizationRelationName,
	}
	ownerPolicy := policy.Policy{
		RoleID:        schema.RoleProjectOwner,
		ResourceID:    projectID,
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   creatorID,
		PrincipalType: schema.UserPrincipal,
	}

	t.Run("should link project to org and add creator as owner", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)

		mockRelSvc.EXPECT().Create(ctx, projectOrgRelation).Return(relation.Relation{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, ownerPolicy).Return(policy.Policy{ID: "new-p"}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

		err := svc.OnProjectCreated(ctx, projectID, orgID, creatorID, schema.UserPrincipal)
		assert.NoError(t, err)
	})

	t.Run("should return error if hierarchy relation creation fails", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)

		mockRelSvc.EXPECT().Create(ctx, projectOrgRelation).Return(relation.Relation{}, errors.New("spicedb unavailable"))

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

		err := svc.OnProjectCreated(ctx, projectID, orgID, creatorID, schema.UserPrincipal)
		assert.ErrorContains(t, err, "link project to org")
	})

	t.Run("should remove the org link if owner policy creation fails", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)

		mockRelSvc.EXPECT().Create(ctx, projectOrgRelation).Return(relation.Relation{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, ownerPolicy).Return(policy.Policy{}, errors.New("db down"))
		mockRelSvc.EXPECT().Delete(ctx, projectOrgRelation).Return(nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mocks.NewRoleService(t), mocks.NewOrgService(t), mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))

		err := svc.OnProjectCreated(ctx, projectID, orgID, creatorID, schema.UserPrincipal)
		assert.ErrorContains(t, err, "db down")
	})
}
