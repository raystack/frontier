package membership_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: viewerRoleID, ResourceID: orgID, ResourceType: schema.OrganizationNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: nil,
		},
		{
			name: "should succeed even when the audit record write fails",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, errors.New("audit store unavailable"))
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  viewerRoleID,
			wantErr: nil,
		},
		{
			name: "should succeed adding a new member with owner role without any relation write",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, ownerRoleID).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: ownerRoleID, ResourceID: orgID, ResourceType: schema.OrganizationNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			orgID:   orgID,
			userID:  userID,
			roleID:  ownerRoleID,
			wantErr: nil,
		},
		{
			name: "should succeed with org-specific custom role",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
				orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(enabledUser, nil)
				roleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, OrgID: orgID, Scopes: []string{schema.OrganizationNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
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

	t.Run("should return error and cleanup policy if identity link creation fails", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockSuSvc := mocks.NewServiceuserService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockSuSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID, Title: "test-su", State: string(serviceuser.Enabled)}, nil)
		mockRoleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal}).Return([]policy.Policy{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{ID: "created-policy-1"}, nil)
		mockRelSvc.EXPECT().Create(ctx, mock.Anything).Return(relation.Relation{}, errors.New("spicedb unavailable"))
		// compensating delete should be called
		mockPolicySvc.EXPECT().Delete(ctx, "created-policy-1").Return(nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mockSuSvc, mocks.NewAuditRecordRepository(t))
		err := svc.AddOrganizationMember(ctx, orgID, suID, schema.ServiceUserPrincipal, viewerRoleID)
		assert.ErrorContains(t, err, "spicedb unavailable")
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

func TestService_AddOrganizationMember_PAT(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	patID := uuid.New().String()
	viewerRoleID := uuid.New().String()

	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}
	activePAT := pat.PAT{ID: patID, OrgID: orgID, Title: "test-pat", ExpiresAt: time.Now().Add(time.Hour)}

	t.Run("should add PAT without writing org member/owner relation", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.AddOrganizationMember(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.NoError(t, err)
	})

	t.Run("should reject PAT from different org", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(pat.PAT{ID: patID, OrgID: "other-org", ExpiresAt: time.Now().Add(time.Hour)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		svc.SetUserPATService(mockPATSvc)
		err := svc.AddOrganizationMember(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, membership.ErrPrincipalNotInOrg)
	})

	t.Run("should reject expired PAT", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(pat.PAT{ID: patID, OrgID: orgID, ExpiresAt: time.Now().Add(-time.Hour)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		svc.SetUserPATService(mockPATSvc)
		err := svc.AddOrganizationMember(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, membership.ErrPrincipalExpired)
	})

	t.Run("should reject PAT principal when userPATService is not wired", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		err := svc.AddOrganizationMember(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, membership.ErrInvalidPrincipal)
	})

	t.Run("should allow adding an org role to a PAT that has only a pat_granted policy", func(t *testing.T) {
		// PAT holds only an all-projects (pat_granted) policy. AddOrganizationMember
		// should not treat that as existing org membership.
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{
			{ID: "pat-granted-pol", RoleID: uuid.New().String(), GrantRelation: schema.PATGrantRelationName},
		}, nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.AddOrganizationMember(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.NoError(t, err)
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
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
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
				auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			roleID:  viewerRoleID,
			wantErr: nil,
		},
		{
			name: "should succeed promoting viewer to owner",
			setup: func(policySvc *mocks.PolicyService, _ *mocks.RelationService, roleSvc *mocks.RoleService, orgSvc *mocks.OrgService, userSvc *mocks.UserService, auditRepo *mocks.AuditRecordRepository) {
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
		mockPolicySvc.EXPECT().Delete(ctx, "p1").Return(nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
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

func TestService_SetOrganizationMemberRole_PAT(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	patID := uuid.New().String()
	viewerRoleID := uuid.New().String()
	oldRoleID := uuid.New().String()

	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}
	activePAT := pat.PAT{ID: patID, OrgID: orgID, Title: "test-pat", ExpiresAt: time.Now().Add(time.Hour)}

	t.Run("should replace PAT role without writing org member/owner relation", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{{ID: "p1", RoleID: oldRoleID}}, nil)
		mockPolicySvc.EXPECT().Delete(ctx, "p1").Return(nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetOrganizationMemberRole(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.NoError(t, err)
	})

	t.Run("should reject expired PAT", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(pat.PAT{ID: patID, OrgID: orgID, ExpiresAt: time.Now().Add(-time.Hour)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetOrganizationMemberRole(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.ErrorIs(t, err, membership.ErrPrincipalExpired)
	})

	t.Run("should leave the pat_granted policy untouched when only the granted role changes", func(t *testing.T) {
		// A PAT can hold both a granted org policy and a pat_granted all-projects
		// policy on the same org. SetOrganizationMemberRole should only replace
		// the granted one — the pat_granted policy is project-cascade scope and
		// must not be wiped as collateral.
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRelSvc := mocks.NewRelationService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, viewerRoleID).Return(role.Role{ID: viewerRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{
			{ID: "granted-pol", RoleID: oldRoleID, GrantRelation: schema.RoleGrantRelationName},
			{ID: "pat-granted-pol", RoleID: uuid.New().String(), GrantRelation: schema.PATGrantRelationName},
		}, nil)
		// Only the granted policy is deleted; pat-granted-pol stays.
		mockPolicySvc.EXPECT().Delete(ctx, "granted-pol").Return(nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.MatchedBy(func(p policy.Policy) bool {
			return p.RoleID == viewerRoleID && p.PrincipalID == patID
		})).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mockRelSvc, mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetOrganizationMemberRole(ctx, orgID, patID, schema.PATPrincipal, viewerRoleID)
		assert.NoError(t, err)
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

	grpObj := relation.Object{ID: groupID, Namespace: schema.GroupNamespace}
	userSub := relation.Subject{ID: userID, Namespace: schema.UserPrincipal}

	type testDeps struct {
		policySvc *mocks.PolicyService
		relSvc    *mocks.RelationService
		roleSvc   *mocks.RoleService
		orgSvc    *mocks.OrgService
		projSvc   *mocks.ProjectService
		grpSvc    *mocks.GroupService
		resSvc    *mocks.ResourceService
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
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{}).Return(nil)
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
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{projectID}).Return(nil)
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
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{}).Return(nil)
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
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{projectID}).Return(nil)
			},
			wantErrContain: "delete sub-resource policy",
		},
		{
			name: "should return error if group relation removal fails after org policies deleted",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{{ID: groupID}}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID, RoleID: viewerRoleID},
				}, nil)
				// org policy deleted first (viewer, plain Delete)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{}).Return(nil)
				// then relation removal fails
				d.relSvc.EXPECT().Delete(ctx, relation.Relation{Object: grpObj, Subject: userSub, RelationName: schema.MemberRelationName}).Return(errors.New("spicedb down"))
			},
			wantErrContain: "relations",
		},
		{
			name: "should remove custom-resource access across every project in the org",
			setup: func(d testDeps) {
				secondProjectID := uuid.New().String()
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{{ID: projectID}, {ID: secondProjectID}}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID},
				}, nil)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{projectID, secondProjectID}).Return(nil)
				d.auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
			wantErr: nil,
		},
		{
			name: "should return error if custom-resource cleanup fails",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{{ID: projectID}}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID},
				}, nil)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{projectID}).Return(errors.New("resource cleanup boom"))
			},
			wantErrContain: "remove custom-resource policies",
		},
		{
			name: "should not touch custom resources when the last-owner guard rejects mid-cascade",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: ownerRoleID}}, nil)
				d.roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, RoleID: ownerRoleID}).Return([]policy.Policy{{ID: "org-p1"}, {ID: "other-owner-p"}}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{{ID: projectID}}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID, RoleID: ownerRoleID},
				}, nil)
				// the other owner disappears between the check and the write
				d.policySvc.EXPECT().DeleteWithMinRoleGuard(ctx, "org-p1", ownerRoleID).Return(policy.ErrLastRoleGuard)
				// RemovePrincipalAccess must NOT be called — strict mock fails on unexpected call
			},
			wantErr: membership.ErrLastOwnerRole,
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
				resSvc:    mocks.NewResourceService(t),
				auditRepo: mocks.NewAuditRecordRepository(t),
			}
			if tt.setup != nil {
				tt.setup(d)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), d.policySvc, d.relSvc, d.roleSvc, d.orgSvc, mocks.NewUserService(t), d.projSvc, d.grpSvc, mocks.NewServiceuserService(t), d.auditRepo)
			svc.SetResourceService(d.resSvc)

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

func TestService_ForceRemoveOrganizationMember(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()

	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}

	type testDeps struct {
		policySvc *mocks.PolicyService
		relSvc    *mocks.RelationService
		roleSvc   *mocks.RoleService
		orgSvc    *mocks.OrgService
		projSvc   *mocks.ProjectService
		grpSvc    *mocks.GroupService
		resSvc    *mocks.ResourceService
		auditRepo *mocks.AuditRecordRepository
	}

	tests := []struct {
		name    string
		setup   func(d testDeps)
		wantErr error
	}{
		{
			name: "removes the org's last owner without the guard",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: ownerRoleID}}, nil)
				// no roleSvc.Get / owner-count query: the guard is skipped entirely
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID, RoleID: ownerRoleID},
				}, nil)
				// plain delete, not DeleteWithMinRoleGuard
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{}).Return(nil)
				d.auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
		},
		{
			name: "does not return ErrNotMember when principal has no org policies and still cascades",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{}, nil)
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{}).Return(nil)
				d.auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
		},
		{
			name: "proceeds with the cascade when the org is disabled",
			setup: func(d testDeps) {
				// Get discards the org payload on ErrDisabled; the force path
				// must reconstruct the org from the input ID and continue —
				// the deleter visits disabled orgs deliberately
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(organization.Organization{}, organization.ErrDisabled)
				d.policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: ownerRoleID}}, nil)
				d.projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
				d.grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
				d.policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
					{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID, RoleID: ownerRoleID},
				}, nil)
				d.policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
				d.resSvc.EXPECT().RemovePrincipalAccess(ctx, userID, schema.UserPrincipal, []string{}).Return(nil)
				d.auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)
			},
		},
		{
			name: "propagates org lookup failure",
			setup: func(d testDeps) {
				d.orgSvc.EXPECT().Get(ctx, orgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			wantErr: organization.ErrNotExist,
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
				resSvc:    mocks.NewResourceService(t),
				auditRepo: mocks.NewAuditRecordRepository(t),
			}
			if tt.setup != nil {
				tt.setup(d)
			}

			svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), d.policySvc, d.relSvc, d.roleSvc, d.orgSvc, mocks.NewUserService(t), d.projSvc, d.grpSvc, mocks.NewServiceuserService(t), d.auditRepo)
			svc.SetResourceService(d.resSvc)

			err := svc.ForceRemoveOrganizationMember(ctx, orgID, userID, schema.UserPrincipal)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// The resource service is injected after construction, so guard against a
// missed wiring taking down every member removal.
func TestService_RemoveOrganizationMember_WithoutResourceService(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()
	viewerRoleID := uuid.New().String()

	policySvc := mocks.NewPolicyService(t)
	relSvc := mocks.NewRelationService(t)
	roleSvc := mocks.NewRoleService(t)
	orgSvc := mocks.NewOrgService(t)
	projSvc := mocks.NewProjectService(t)
	grpSvc := mocks.NewGroupService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)

	orgSvc.EXPECT().Get(ctx, orgID).Return(organization.Organization{ID: orgID, Title: "Test Org"}, nil)
	policySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{{ID: "org-p1", RoleID: viewerRoleID}}, nil)
	roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner}, nil)
	projSvc.EXPECT().List(ctx, project.Filter{OrgID: orgID}).Return([]project.Project{}, nil)
	grpSvc.EXPECT().List(ctx, group.Filter{OrganizationID: orgID}).Return([]group.Group{}, nil)
	policySvc.EXPECT().List(ctx, policy.Filter{PrincipalID: userID, PrincipalType: schema.UserPrincipal}).Return([]policy.Policy{
		{ID: "org-p1", ResourceType: schema.OrganizationNamespace, ResourceID: orgID},
	}, nil)
	policySvc.EXPECT().Delete(ctx, "org-p1").Return(nil)
	auditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

	svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), policySvc, relSvc, roleSvc, orgSvc, mocks.NewUserService(t), projSvc, grpSvc, mocks.NewServiceuserService(t), auditRepo)

	assert.NoError(t, svc.RemoveOrganizationMember(ctx, orgID, userID, schema.UserPrincipal))
}

func TestService_SetPATAllProjectsRole(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	patID := uuid.New().String()
	projectRoleID := uuid.New().String()
	oldRoleID := uuid.New().String()

	enabledOrg := organization.Organization{ID: orgID, Title: "Test Org"}
	activePAT := pat.PAT{ID: patID, OrgID: orgID, Title: "test-pat", ExpiresAt: time.Now().Add(time.Hour)}
	projectRole := role.Role{ID: projectRoleID, Scopes: []string{schema.ProjectNamespace}}

	t.Run("should write pat_granted policy on the org", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, projectRoleID).Return(projectRole, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{}, nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.MatchedBy(func(p policy.Policy) bool {
			return p.RoleID == projectRoleID &&
				p.ResourceID == orgID &&
				p.ResourceType == schema.OrganizationNamespace &&
				p.PrincipalID == patID &&
				p.PrincipalType == schema.PATPrincipal &&
				p.GrantRelation == schema.PATGrantRelationName
		})).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetPATAllProjectsRole(ctx, orgID, patID, projectRoleID)
		assert.NoError(t, err)
	})

	t.Run("should be a no-op when the same pat_granted role is already set", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, projectRoleID).Return(projectRole, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{
			{ID: "p1", RoleID: projectRoleID, GrantRelation: schema.PATGrantRelationName},
		}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetPATAllProjectsRole(ctx, orgID, patID, projectRoleID)
		assert.NoError(t, err)
	})

	t.Run("should replace existing pat_granted policy with new role", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, projectRoleID).Return(projectRole, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{
			{ID: "p1", RoleID: oldRoleID, GrantRelation: schema.PATGrantRelationName},
		}, nil)
		mockPolicySvc.EXPECT().Delete(ctx, "p1").Return(nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetPATAllProjectsRole(ctx, orgID, patID, projectRoleID)
		assert.NoError(t, err)
	})

	t.Run("should ignore an existing granted policy and only replace pat_granted", func(t *testing.T) {
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, projectRoleID).Return(projectRole, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{
			{ID: "granted-pol", RoleID: oldRoleID, GrantRelation: schema.RoleGrantRelationName},
		}, nil)
		// no Delete on the granted policy; only Create for the new pat_granted
		mockPolicySvc.EXPECT().Create(ctx, mock.Anything).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetPATAllProjectsRole(ctx, orgID, patID, projectRoleID)
		assert.NoError(t, err)
	})

	t.Run("should replace only the pat_granted policy when both granted and pat_granted exist", func(t *testing.T) {
		// Granted policy's role matches the requested role — the function must
		// not treat this as a no-op (the role-match check is for pat_granted
		// only) and must not delete the granted policy.
		mockPolicySvc := mocks.NewPolicyService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)
		mockAuditRepo := mocks.NewAuditRecordRepository(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, projectRoleID).Return(projectRole, nil)
		mockPolicySvc.EXPECT().List(ctx, policy.Filter{OrgID: orgID, PrincipalID: patID, PrincipalType: schema.PATPrincipal}).Return([]policy.Policy{
			{ID: "granted-pol", RoleID: projectRoleID, GrantRelation: schema.RoleGrantRelationName},
			{ID: "pat-granted-pol", RoleID: oldRoleID, GrantRelation: schema.PATGrantRelationName},
		}, nil)
		mockPolicySvc.EXPECT().Delete(ctx, "pat-granted-pol").Return(nil)
		mockPolicySvc.EXPECT().Create(ctx, mock.MatchedBy(func(p policy.Policy) bool {
			return p.GrantRelation == schema.PATGrantRelationName && p.RoleID == projectRoleID
		})).Return(policy.Policy{}, nil)
		mockAuditRepo.EXPECT().Create(ctx, mock.Anything).Return(auditrecord.AuditRecord{}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockPolicySvc, mocks.NewRelationService(t), mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mockAuditRepo)
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetPATAllProjectsRole(ctx, orgID, patID, projectRoleID)
		assert.NoError(t, err)
	})

	t.Run("should reject role that is not project-scoped", func(t *testing.T) {
		mockRoleSvc := mocks.NewRoleService(t)
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(activePAT, nil)
		mockRoleSvc.EXPECT().Get(ctx, projectRoleID).Return(role.Role{ID: projectRoleID, Scopes: []string{schema.OrganizationNamespace}}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mockRoleSvc, mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetPATAllProjectsRole(ctx, orgID, patID, projectRoleID)
		assert.ErrorIs(t, err, membership.ErrInvalidProjectRole)
	})

	t.Run("should reject expired PAT", func(t *testing.T) {
		mockOrgSvc := mocks.NewOrgService(t)
		mockPATSvc := mocks.NewUserPATService(t)

		mockOrgSvc.EXPECT().Get(ctx, orgID).Return(enabledOrg, nil)
		mockPATSvc.EXPECT().GetByID(ctx, patID).Return(pat.PAT{ID: patID, OrgID: orgID, ExpiresAt: time.Now().Add(-time.Hour)}, nil)

		svc := membership.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mocks.NewPolicyService(t), mocks.NewRelationService(t), mocks.NewRoleService(t), mockOrgSvc, mocks.NewUserService(t), mocks.NewProjectService(t), mocks.NewGroupService(t), mocks.NewServiceuserService(t), mocks.NewAuditRecordRepository(t))
		svc.SetUserPATService(mockPATSvc)
		err := svc.SetPATAllProjectsRole(ctx, orgID, patID, projectRoleID)
		assert.ErrorIs(t, err, membership.ErrPrincipalExpired)
	})
}
