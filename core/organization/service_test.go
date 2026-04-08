package organization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/organization/mocks"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	pat "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Get(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockUserSvc := mocks.NewUserService(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockPolicySvc := mocks.NewPolicyService(t)
	mockPrefSvc := mocks.NewPreferencesService(t)
	mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

	t.Run("should return orgs when fetched by id (by calling repo.GetByID)", func(t *testing.T) {
		IDParam := uuid.New()
		expectedOrg := organization.Organization{
			ID:    IDParam.String(),
			Name:  "test-org",
			Title: "test organization",
			State: "enabled",
		}

		mockRepo.On("GetByID", mock.Anything, IDParam.String()).Return(expectedOrg, nil).Once()
		org, err := svc.Get(context.Background(), IDParam.String())
		assert.Nil(t, err)
		assert.Equal(t, expectedOrg, org)
	})

	t.Run("should return orgs when fetched by name (by calling repo.GetByName)", func(t *testing.T) {
		nameParam := "test-org"
		expectedOrg := organization.Organization{
			ID:    uuid.New().String(),
			Name:  nameParam,
			Title: "test organization",
			State: "enabled",
		}

		mockRepo.On("GetByName", mock.Anything, nameParam).Return(expectedOrg, nil).Once()
		org, err := svc.Get(context.Background(), nameParam)
		assert.Nil(t, err)
		assert.Equal(t, expectedOrg, org)
	})

	t.Run("should return an error if org being fetched is disabled", func(t *testing.T) {
		nameParam := "test-org"
		orgFromRepository := organization.Organization{
			ID:    uuid.New().String(),
			Name:  nameParam,
			Title: "test organization",
			State: organization.Disabled,
		}

		mockRepo.On("GetByName", mock.Anything, nameParam).Return(orgFromRepository, nil).Once()
		org, err := svc.Get(context.Background(), nameParam)
		assert.NotNil(t, err)
		assert.Equal(t, err, organization.ErrDisabled)
		assert.Equal(t, organization.Organization{}, org)
	})
}

func TestService_GetRaw(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockUserSvc := mocks.NewUserService(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockPolicySvc := mocks.NewPolicyService(t)
	mockPrefSvc := mocks.NewPreferencesService(t)
	mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

	t.Run("should return an org based on ID passed", func(t *testing.T) {
		IDParam := uuid.New()
		expectedOrg := organization.Organization{
			ID:    IDParam.String(),
			Name:  "test-org",
			Title: "test organization",
			State: "enabled",
		}

		mockRepo.On("GetByID", mock.Anything, IDParam.String()).Return(expectedOrg, nil).Once()
		org, err := svc.GetRaw(context.Background(), IDParam.String())
		assert.Nil(t, err)
		assert.Equal(t, expectedOrg, org)
	})

	t.Run("should return an org based on name passed", func(t *testing.T) {
		nameParam := "test-org"
		expectedOrg := organization.Organization{
			ID:    uuid.New().String(),
			Name:  nameParam,
			Title: "test organization",
			State: "enabled",
		}

		mockRepo.On("GetByName", mock.Anything, nameParam).Return(expectedOrg, nil).Once()
		org, err := svc.GetRaw(context.Background(), nameParam)
		assert.Nil(t, err)
		assert.Equal(t, expectedOrg, org)
	})

	t.Run("should return an org even if it is disabled", func(t *testing.T) {
		nameParam := "test-org"
		expectedOrg := organization.Organization{
			ID:    uuid.New().String(),
			Name:  nameParam,
			Title: "test organization",
			State: organization.Disabled,
		}

		mockRepo.On("GetByName", mock.Anything, nameParam).Return(expectedOrg, nil).Once()
		org, err := svc.GetRaw(context.Background(), nameParam)
		assert.Nil(t, err)
		assert.Equal(t, expectedOrg, org)
	})
}

func TestService_GetDefaultOrgStateOnCreate(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockUserSvc := mocks.NewUserService(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockPolicySvc := mocks.NewPolicyService(t)
	mockPrefSvc := mocks.NewPreferencesService(t)
	mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

	t.Run("should return org state to be set on creation, as per preferences", func(t *testing.T) {
		expectedPrefs := map[string]string{
			preference.PlatformDisableOrgsOnCreate: "true",
		}
		mockPrefSvc.On("LoadPlatformPreferences", mock.Anything).Return(expectedPrefs, nil).Once()
		state, err := svc.GetDefaultOrgStateOnCreate(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, organization.Disabled, state)
	})

	t.Run("should return org state as enabled + error if preferences cannot be fetched", func(t *testing.T) {
		expectedPrefs := map[string]string{}
		mockPrefSvc.On("LoadPlatformPreferences", mock.Anything).Return(expectedPrefs, errors.New("an error occurred")).Once()
		state, err := svc.GetDefaultOrgStateOnCreate(context.Background())
		assert.NotNil(t, err)
		assert.Equal(t, "an error occurred", errors.Unwrap(err).Error())
		assert.Equal(t, organization.Enabled, state)
	})
}

func TestService_AddMember(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockUserSvc := mocks.NewUserService(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockPolicySvc := mocks.NewPolicyService(t)
	mockPrefSvc := mocks.NewPreferencesService(t)
	mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

	t.Run("should create policy and relation for member as per role", func(t *testing.T) {
		inputOrgID := "test-id"
		inputRelationName := schema.MemberRelationName
		inputPrincipal := authenticate.Principal{
			ID:   "test-principal-id",
			Type: schema.UserPrincipal,
		}

		expectedOrg := organization.Organization{
			ID:    inputOrgID,
			Name:  "test-org",
			Title: "Test Organization",
			State: organization.Enabled,
		}

		policyToBeCreated := policy.Policy{
			RoleID:        organization.MemberRole,
			ResourceID:    inputOrgID,
			ResourceType:  schema.OrganizationNamespace,
			PrincipalID:   inputPrincipal.ID,
			PrincipalType: inputPrincipal.Type,
		}

		relationToBeCreated := relation.Relation{
			Object: relation.Object{
				ID:        inputOrgID,
				Namespace: schema.OrganizationNamespace,
			},
			Subject: relation.Subject{
				ID:        inputPrincipal.ID,
				Namespace: inputPrincipal.Type,
			},
			RelationName: schema.MemberRelationName,
		}

		mockPolicySvc.On("Create", mock.Anything, policyToBeCreated).Return(policy.Policy{}, nil)
		mockRelationSvc.On("Create", mock.Anything, relationToBeCreated).Return(relation.Relation{}, nil)
		mockRepo.On("GetByID", mock.Anything, inputOrgID).Return(expectedOrg, nil).Once()
		mockAuditRecordRepo.On("Create", mock.Anything, mock.AnythingOfType("models.AuditRecord")).Return(auditrecord.AuditRecord{}, nil).Once()

		err := svc.AddMember(context.Background(), inputOrgID, inputRelationName, inputPrincipal)
		assert.Nil(t, err)
	})
}

func TestService_AttachToPlatform(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockUserSvc := mocks.NewUserService(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockPolicySvc := mocks.NewPolicyService(t)
	mockPrefSvc := mocks.NewPreferencesService(t)
	mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

	inputOrgID := "some-org-id"
	relationToBeCreated := relation.Relation{
		Object: relation.Object{
			ID:        inputOrgID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		RelationName: schema.PlatformRelationName,
	}

	t.Run("should create a relation for org in platform namespace", func(t *testing.T) {
		mockRelationSvc.On("Create", mock.Anything, relationToBeCreated).Return(relation.Relation{}, nil).Once()
		err := svc.AttachToPlatform(context.Background(), inputOrgID)
		assert.Nil(t, err)
	})

	t.Run("should return an error if relation creation fails", func(t *testing.T) {
		expectedErr := errors.New("Internal error")
		mockRelationSvc.On("Create", mock.Anything, relationToBeCreated).Return(relation.Relation{}, expectedErr).Once()
		err := svc.AttachToPlatform(context.Background(), inputOrgID)
		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestService_ListByUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should resolve PAT to user and intersect with PAT org", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockUserSvc := mocks.NewUserService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockPrefSvc := mocks.NewPreferencesService(t)
		mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

		mockRoleSvc := mocks.NewRoleService(t)
		svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

		// LookupResources should be called with user ID/type, not PAT
		mockRelationSvc.On("LookupResources", ctx, relation.Relation{
			Object:       relation.Object{Namespace: schema.OrganizationNamespace},
			Subject:      relation.Subject{ID: "user-123", Namespace: schema.UserPrincipal},
			RelationName: schema.MembershipPermission,
		}).Return([]string{"org-1", "org-2"}, nil).Once()

		// Repo should only be called with the PAT's org (intersection result)
		mockRepo.On("List", ctx, organization.Filter{
			IDs: []string{"org-1"},
		}).Return([]organization.Organization{
			{ID: "org-1", Name: "org-one"},
		}, nil).Once()

		result, err := svc.ListByUser(ctx, authenticate.Principal{
			ID:   "pat-456",
			Type: schema.PATPrincipal,
			PAT:  &pat.PAT{ID: "pat-456", UserID: "user-123", OrgID: "org-1"},
		}, organization.Filter{})

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "org-1", result[0].ID)
	})

	t.Run("should return empty when PAT org not in user memberships", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockUserSvc := mocks.NewUserService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockPrefSvc := mocks.NewPreferencesService(t)
		mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

		mockRoleSvc := mocks.NewRoleService(t)
		svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

		mockRelationSvc.On("LookupResources", ctx, relation.Relation{
			Object:       relation.Object{Namespace: schema.OrganizationNamespace},
			Subject:      relation.Subject{ID: "user-123", Namespace: schema.UserPrincipal},
			RelationName: schema.MembershipPermission,
		}).Return([]string{"org-1", "org-2"}, nil).Once()

		result, err := svc.ListByUser(ctx, authenticate.Principal{
			ID:   "pat-456",
			Type: schema.PATPrincipal,
			PAT:  &pat.PAT{ID: "pat-456", UserID: "user-123", OrgID: "org-999"},
		}, organization.Filter{})

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("should pass through for regular user principal", func(t *testing.T) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockUserSvc := mocks.NewUserService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockPrefSvc := mocks.NewPreferencesService(t)
		mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)

		mockRoleSvc := mocks.NewRoleService(t)
		svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

		mockRelationSvc.On("LookupResources", ctx, relation.Relation{
			Object:       relation.Object{Namespace: schema.OrganizationNamespace},
			Subject:      relation.Subject{ID: "user-123", Namespace: schema.UserPrincipal},
			RelationName: schema.MembershipPermission,
		}).Return([]string{"org-1", "org-2"}, nil).Once()

		mockRepo.On("List", ctx, organization.Filter{
			IDs: []string{"org-1", "org-2"},
		}).Return([]organization.Organization{
			{ID: "org-1", Name: "org-one"},
			{ID: "org-2", Name: "org-two"},
		}, nil).Once()

		result, err := svc.ListByUser(ctx, authenticate.Principal{
			ID:   "user-123",
			Type: schema.UserPrincipal,
		}, organization.Filter{})

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestService_SetMemberRole(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	ownerRoleID := uuid.New().String()
	memberRoleID := uuid.New().String()

	tests := []struct {
		name      string
		setup     func(*mocks.Repository, *mocks.UserService, *mocks.RoleService, *mocks.PolicyService, *mocks.AuditRecordRepository)
		orgID     string
		userID    string
		newRoleID string
		wantErr   error
	}{
		{
			name: "should return error if org does not exist",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, _ *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   organization.ErrNotExist,
		},
		{
			name: "should return error if org is disabled",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, _ *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{State: organization.Disabled}, nil)
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   organization.ErrDisabled,
		},
		{
			name: "should return error if user does not exist",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, _ *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{ID: orgID, State: organization.Enabled}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{}, user.ErrNotExist)
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   user.ErrNotExist,
		},
		{
			name: "should return error if role does not exist",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, _ *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{ID: orgID, State: organization.Enabled}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{}, role.ErrNotExist)
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   role.ErrNotExist,
		},
		{
			name: "should return error if role is not valid for org scope",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, _ *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{ID: orgID, State: organization.Enabled}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				// role exists but has project scope, not org scope
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: "project-role", Scopes: []string{schema.ProjectNamespace}}, nil)
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   organization.ErrInvalidOrgRole,
		},
		{
			name: "should return error if user is not a member of the org",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, _ *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{ID: orgID, State: organization.Enabled}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: "member", Scopes: []string{schema.OrganizationNamespace}}, nil)
				// get user's existing policies - empty, user is not a member
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{}, nil)
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   organization.ErrNotMember,
		},
		{
			name: "should return error if demoting last owner",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, _ *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{ID: orgID, State: organization.Enabled}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: "member", Scopes: []string{schema.OrganizationNamespace}}, nil)
				// get user's existing policies - user is owner
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "policy-1", RoleID: ownerRoleID}}, nil)
				// get owner role for comparison
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace}}, nil)
				// count owners - only 1
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:  orgID,
					RoleID: ownerRoleID,
				}).Return([]policy.Policy{{ID: "policy-1", RoleID: ownerRoleID}}, nil)
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   organization.ErrLastOwnerRole,
		},
		{
			name: "should succeed when changing role with multiple owners",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, auditRepo *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{ID: orgID, State: organization.Enabled}, nil).Times(2)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{
					ID:    userID,
					Title: "test-user",
					Email: "test-user@acme.dev",
				}, nil).Times(2)
				roleSvc.EXPECT().Get(ctx, memberRoleID).Return(role.Role{ID: memberRoleID, Name: "member", Scopes: []string{schema.OrganizationNamespace}}, nil)
				// get user's existing policies - user is owner
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "policy-1", RoleID: ownerRoleID}}, nil)
				// get owner role for comparison
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace}}, nil)
				// count owners - 2 owners exist
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:  orgID,
					RoleID: ownerRoleID,
				}).Return([]policy.Policy{
					{ID: "policy-1", RoleID: ownerRoleID},
					{ID: "policy-2", RoleID: ownerRoleID},
				}, nil)
				// delete existing policy
				policySvc.EXPECT().Delete(ctx, "policy-1").Return(nil)
				// create new policy
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID:        memberRoleID,
					ResourceID:    orgID,
					ResourceType:  schema.OrganizationNamespace,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				// audit logging
				auditRepo.EXPECT().Create(ctx, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
					if ar.Target == nil {
						return false
					}
					return ar.Event == pkgAuditRecord.OrganizationMemberRoleChangedEvent &&
						ar.Resource.ID == orgID &&
						ar.Target.ID == userID &&
						ar.Target.Metadata["email"] == "test-user@acme.dev" &&
						ar.Target.Metadata["role_id"] == memberRoleID &&
						ar.OrgID == orgID
				})).Return(auditrecord.AuditRecord{}, nil).Once()
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: memberRoleID,
			wantErr:   nil,
		},
		{
			name: "should succeed when promoting to owner",
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, roleSvc *mocks.RoleService, policySvc *mocks.PolicyService, auditRepo *mocks.AuditRecordRepository) {
				repo.EXPECT().GetByID(ctx, orgID).Return(organization.Organization{ID: orgID, State: organization.Enabled}, nil).Times(2)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{
					ID:    userID,
					Title: "test-user",
					Email: "test-user@acme.dev",
				}, nil).Times(2)
				roleSvc.EXPECT().Get(ctx, ownerRoleID).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace}}, nil)
				// get user's existing policies - user is member
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "policy-1", RoleID: memberRoleID}}, nil)
				// get owner role for comparison
				roleSvc.EXPECT().Get(ctx, schema.RoleOrganizationOwner).Return(role.Role{ID: ownerRoleID, Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace}}, nil)
				// no owner count check needed when promoting to owner
				// delete existing policy
				policySvc.EXPECT().Delete(ctx, "policy-1").Return(nil)
				// create new policy
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID:        ownerRoleID,
					ResourceID:    orgID,
					ResourceType:  schema.OrganizationNamespace,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				// audit logging
				auditRepo.EXPECT().Create(ctx, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
					if ar.Target == nil {
						return false
					}
					return ar.Event == pkgAuditRecord.OrganizationMemberRoleChangedEvent &&
						ar.Resource.ID == orgID &&
						ar.Target.ID == userID &&
						ar.Target.Metadata["email"] == "test-user@acme.dev" &&
						ar.Target.Metadata["role_id"] == ownerRoleID &&
						ar.OrgID == orgID
				})).Return(auditrecord.AuditRecord{}, nil).Once()
			},
			orgID:     orgID,
			userID:    userID,
			newRoleID: ownerRoleID,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepository(t)
			mockRelationSvc := mocks.NewRelationService(t)
			mockUserSvc := mocks.NewUserService(t)
			mockAuthnSvc := mocks.NewAuthnService(t)
			mockPolicySvc := mocks.NewPolicyService(t)
			mockPrefSvc := mocks.NewPreferencesService(t)
			mockAuditRecordRepo := mocks.NewAuditRecordRepository(t)
			mockRoleSvc := mocks.NewRoleService(t)

			if tt.setup != nil {
				tt.setup(mockRepo, mockUserSvc, mockRoleSvc, mockPolicySvc, mockAuditRecordRepo)
			}

			svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockAuditRecordRepo, mockRoleSvc)

			err := svc.SetMemberRole(ctx, tt.orgID, tt.userID, tt.newRoleID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
