package permission_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/permission/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Get(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := permission.NewService(mockRepo)

	t.Run("should get permission by id", func(t *testing.T) {
		inputID := uuid.New().String()
		expectedPermission := permission.Permission{
			ID:   inputID,
			Name: "permissionname",
			Slug: "app/permission_name",
		}

		mockRepo.On("Get", mock.Anything, inputID).Return(expectedPermission, nil)
		perm, err := svc.Get(context.Background(), inputID)

		assert.Nil(t, err)
		assert.Equal(t, expectedPermission, perm)
	})

	t.Run("should get permission by slug", func(t *testing.T) {
		inputSlug := "app/somepermission"
		expectedPermission := permission.Permission{
			ID:   uuid.New().String(),
			Name: "permissionname",
			Slug: inputSlug,
		}

		mockRepo.On("GetBySlug", mock.Anything, inputSlug).Return(expectedPermission, nil)
		perm, err := svc.Get(context.Background(), inputSlug)

		assert.Nil(t, err)
		assert.Equal(t, expectedPermission, perm)
	})

	t.Run("should return an error if permission cannot be fetched", func(t *testing.T) {
		inputID := uuid.New().String()

		expectedError := errors.New("An error occurred")

		mockRepo.On("Get", mock.Anything, inputID).Return(permission.Permission{}, expectedError)
		_, err := svc.Get(context.Background(), inputID)

		assert.NotNil(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestService_Upsert(t *testing.T) {
	t.Run("should upsert permission", func(t *testing.T) {})

	t.Run("should generate slug if not present in request", func(t *testing.T) {})

	t.Run("should return an error if permission cannot be upserted", func(t *testing.T) {})
}

func TestService_List(t *testing.T) {
	t.Run("should list permissions", func(t *testing.T) {})

	t.Run("should list after applying filters", func(t *testing.T) {})

	t.Run("should return an error if permissions cannot be list", func(t *testing.T) {})
}
