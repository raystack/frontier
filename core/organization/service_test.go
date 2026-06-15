package organization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/organization/mocks"
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

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockRoleSvc)

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

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockRoleSvc)

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

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockRoleSvc)

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

func TestService_AttachToPlatform(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockUserSvc := mocks.NewUserService(t)
	mockAuthnSvc := mocks.NewAuthnService(t)
	mockPolicySvc := mocks.NewPolicyService(t)
	mockPrefSvc := mocks.NewPreferencesService(t)

	mockRoleSvc := mocks.NewRoleService(t)
	svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc, mockPolicySvc, mockPrefSvc, mockRoleSvc)

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

func TestService_List(t *testing.T) {
	ctx := context.Background()

	newService := func() (*organization.Service, *mocks.Repository) {
		mockRepo := mocks.NewRepository(t)
		mockRelationSvc := mocks.NewRelationService(t)
		mockUserSvc := mocks.NewUserService(t)
		mockAuthnSvc := mocks.NewAuthnService(t)
		mockPolicySvc := mocks.NewPolicyService(t)
		mockPrefSvc := mocks.NewPreferencesService(t)
		mockRoleSvc := mocks.NewRoleService(t)
		svc := organization.NewService(mockRepo, mockRelationSvc, mockUserSvc, mockAuthnSvc,
			mockPolicySvc, mockPrefSvc, mockRoleSvc)
		return svc, mockRepo
	}

	t.Run("passes empty filter to repository unchanged", func(t *testing.T) {
		svc, mockRepo := newService()
		mockRepo.On("List", ctx, organization.Filter{}).
			Return([]organization.Organization{
				{ID: "org-1", Name: "org-one"},
				{ID: "org-2", Name: "org-two"},
			}, nil).Once()

		got, err := svc.List(ctx, organization.Filter{})
		assert.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("forwards IDs filter to repository", func(t *testing.T) {
		svc, mockRepo := newService()
		ids := []string{"org-1", "org-2"}
		mockRepo.On("List", ctx, organization.Filter{IDs: ids}).
			Return([]organization.Organization{
				{ID: "org-1", Name: "org-one"},
				{ID: "org-2", Name: "org-two"},
			}, nil).Once()

		got, err := svc.List(ctx, organization.Filter{IDs: ids})
		assert.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("forwards state filter to repository", func(t *testing.T) {
		svc, mockRepo := newService()
		mockRepo.On("List", ctx, organization.Filter{State: organization.Disabled}).
			Return([]organization.Organization{
				{ID: "org-3", Name: "org-three", State: organization.Disabled},
			}, nil).Once()

		got, err := svc.List(ctx, organization.Filter{State: organization.Disabled})
		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, organization.Disabled, got[0].State)
	})

	t.Run("forwards combined IDs and state filter to repository", func(t *testing.T) {
		svc, mockRepo := newService()
		f := organization.Filter{IDs: []string{"org-1"}, State: organization.Enabled}
		mockRepo.On("List", ctx, f).
			Return([]organization.Organization{
				{ID: "org-1", Name: "org-one", State: organization.Enabled},
			}, nil).Once()

		got, err := svc.List(ctx, f)
		assert.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("propagates repository errors unchanged", func(t *testing.T) {
		svc, mockRepo := newService()
		repoErr := errors.New("db down")
		mockRepo.On("List", ctx, organization.Filter{}).
			Return(nil, repoErr).Once()

		got, err := svc.List(ctx, organization.Filter{})
		assert.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
	})

	t.Run("returns empty slice when repository returns no rows", func(t *testing.T) {
		svc, mockRepo := newService()
		mockRepo.On("List", ctx, organization.Filter{IDs: []string{"org-nope"}}).
			Return([]organization.Organization{}, nil).Once()

		got, err := svc.List(ctx, organization.Filter{IDs: []string{"org-nope"}})
		assert.NoError(t, err)
		assert.Empty(t, got)
	})
}
