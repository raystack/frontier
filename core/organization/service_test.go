package organization_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/organization/mocks"
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
