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

		mockRepo.On("Get", mock.Anything, inputID).Return(expectedPermission, nil).Once()
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

		mockRepo.On("GetBySlug", mock.Anything, inputSlug).Return(expectedPermission, nil).Once()
		perm, err := svc.Get(context.Background(), inputSlug)

		assert.Nil(t, err)
		assert.Equal(t, expectedPermission, perm)
	})

	t.Run("should return an error if permission cannot be fetched", func(t *testing.T) {
		inputID := uuid.New().String()

		expectedError := errors.New("An error occurred")

		mockRepo.On("Get", mock.Anything, inputID).Return(permission.Permission{}, expectedError).Once()
		_, err := svc.Get(context.Background(), inputID)

		assert.NotNil(t, err)
		assert.Equal(t, expectedError, err)
	})
}

func TestService_Upsert(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := permission.NewService(mockRepo)

	t.Run("should upsert permission", func(t *testing.T) {
		inputPermission := permission.Permission{
			ID:   uuid.New().String(),
			Name: "permissionname",
			Slug: "app/permission_name",
		}

		mockRepo.On("Upsert", mock.Anything, inputPermission).Return(inputPermission, nil).Once()
		perm, err := svc.Upsert(context.Background(), inputPermission)

		assert.Nil(t, err)
		assert.Equal(t, inputPermission, perm)
	})

	t.Run("should generate slug if not present in request", func(t *testing.T) {
		inputPermission := permission.Permission{
			ID:   uuid.New().String(),
			Name: "permissionname",
		}

		expectedRepoInput := inputPermission
		expectedRepoInput.Slug = inputPermission.GenerateSlug()

		mockRepo.On("Upsert", mock.Anything, expectedRepoInput).Return(expectedRepoInput, nil).Once()
		perm, err := svc.Upsert(context.Background(), inputPermission)

		assert.Nil(t, err)
		assert.Equal(t, expectedRepoInput, perm)
	})

	t.Run("should return an error if permission cannot be upserted", func(t *testing.T) {
		inputPermission := permission.Permission{
			ID:   uuid.New().String(),
			Name: "permissionname",
			Slug: "app/permission_name",
		}

		expectedErr := errors.New("An error occurred")

		mockRepo.On("Upsert", mock.Anything, inputPermission).Return(permission.Permission{}, expectedErr).Once()
		_, err := svc.Upsert(context.Background(), inputPermission)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestService_List(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := permission.NewService(mockRepo)

	t.Run("should list permissions", func(t *testing.T) {
		filters := permission.Filter{
			Namespace: "app/organization",
		}

		expectedRes := []permission.Permission{
			{
				ID:   uuid.New().String(),
				Name: "firstpermissionname",
				Slug: "app/first_permission_name",
			},
			{
				ID:   uuid.New().String(),
				Name: "secondpermissionname",
				Slug: "app/second_permission_name",
			},
		}

		mockRepo.On("List", mock.Anything, filters).Return(expectedRes, nil).Once()

		perms, err := svc.List(context.Background(), filters)

		assert.Nil(t, err)
		assert.Equal(t, expectedRes, perms)
	})

	t.Run("should return an error if permissions cannot be list", func(t *testing.T) {
		filters := permission.Filter{
			Namespace: "app/organization",
		}

		expectedErr := errors.New("An error occurred")

		mockRepo.On("List", mock.Anything, filters).Return([]permission.Permission{}, expectedErr).Once()

		_, err := svc.List(context.Background(), filters)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
func TestService_Update(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := permission.NewService(mockRepo)

	t.Run("should update permission", func(t *testing.T) {
		inputPermission := permission.Permission{
			ID:   uuid.New().String(),
			Name: "permissionname",
			Slug: "app/permission_name",
		}

		mockRepo.On("Update", mock.Anything, inputPermission).Return(inputPermission, nil).Once()
		perm, err := svc.Update(context.Background(), inputPermission)

		assert.Nil(t, err)
		assert.Equal(t, inputPermission, perm)
	})

	t.Run("should generate slug if not present in request", func(t *testing.T) {
		inputPermission := permission.Permission{
			ID:   uuid.New().String(),
			Name: "permissionname",
		}

		expectedRepoInput := inputPermission
		expectedRepoInput.Slug = inputPermission.GenerateSlug()

		mockRepo.On("Update", mock.Anything, expectedRepoInput).Return(expectedRepoInput, nil).Once()
		perm, err := svc.Update(context.Background(), inputPermission)

		assert.Nil(t, err)
		assert.Equal(t, expectedRepoInput, perm)
	})

	t.Run("should return an error if permission cannot be updated", func(t *testing.T) {
		inputPermission := permission.Permission{
			ID:   uuid.New().String(),
			Name: "permissionname",
			Slug: "app/permission_name",
		}

		expectedErr := errors.New("An error occurred")

		mockRepo.On("Update", mock.Anything, inputPermission).Return(permission.Permission{}, expectedErr).Once()
		_, err := svc.Update(context.Background(), inputPermission)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestService_Delete(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	svc := permission.NewService(mockRepo)

	t.Run("should delete permission", func(t *testing.T) {
		permissionID := uuid.New().String()

		mockRepo.On("Delete", mock.Anything, permissionID).Return(nil).Once()

		err := svc.Delete(context.Background(), permissionID)

		assert.Nil(t, err)
	})

	t.Run("should return an error if permissions cannot be list", func(t *testing.T) {
		permissionID := uuid.New().String()
		expectedErr := errors.New("An error occurred")

		mockRepo.On("Delete", mock.Anything, permissionID).Return(expectedErr).Once()

		err := svc.Delete(context.Background(), permissionID)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
