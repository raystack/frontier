package organization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/organization/mocks"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
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

	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc)

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

	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc)

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

	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc)

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

	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc)

	t.Run("should create policy and relation for member as per role", func(t *testing.T) {
		inputOrgID := "test-id"
		inputRelationName := schema.MemberRelationName
		inputPrincipal := authenticate.Principal{
			ID:   "test-principal-id",
			Type: schema.UserPrincipal,
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

	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc)

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
