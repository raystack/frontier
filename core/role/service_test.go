package role_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/role/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Get(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPermissionSvc := mocks.NewPermissionService(t)

	t.Run("should fetch by id if id is passed", func(t *testing.T) {
		mockID := uuid.New().String()
		expectedRole := role.Role{
			ID:   "role id",
			Name: "role name",
		}

		mockRepository.On("Get", mock.Anything, mockID).Return(expectedRole, nil).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc)
		res, err := svc.Get(context.Background(), mockID)

		assert.Equal(t, nil, err)
		assert.Equal(t, expectedRole, res)
	})

	t.Run("should fetch by name if slug is passed", func(t *testing.T) {
		mockSlug := "some slug"
		expectedRole := role.Role{
			ID:   "role id",
			Name: "role name",
		}

		mockRepository.On("GetByName", mock.Anything, "", mockSlug).Return(expectedRole, nil).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc)
		res, err := svc.Get(context.Background(), mockSlug)

		assert.Equal(t, nil, err)
		assert.Equal(t, expectedRole, res)
	})

	t.Run("should return an error if fetching role fails", func(t *testing.T) {
		mockID := uuid.New().String()
		expectedErr := errors.New("an error occurred")

		mockRepository.On("Get", mock.Anything, mockID).Return(role.Role{}, expectedErr).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc)
		_, err := svc.Get(context.Background(), mockID)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_List(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPermissionSvc := mocks.NewPermissionService(t)

	t.Run("should return roles", func(t *testing.T) {
		expectedRoles := []role.Role{
			{
				ID:   "role 1",
				Name: "role 1 name",
			},
			{
				ID:   "role 2",
				Name: "role 2 name",
			},
		}

		f := role.Filter{}

		mockRepository.On("List", mock.Anything, f).Return(expectedRoles, nil).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc)
		res, err := svc.List(context.Background(), f)

		assert.Equal(t, nil, err)
		assert.Equal(t, expectedRoles, res)
	})

	t.Run("should return an error if fetching roles fails", func(t *testing.T) {
		expectedErr := errors.New("An error occurred")
		f := role.Filter{}
		mockRepository.On("List", mock.Anything, f).Return(nil, expectedErr).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc)
		_, err := svc.List(context.Background(), f)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
