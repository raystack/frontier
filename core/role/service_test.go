package role_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	auditMocks "github.com/raystack/frontier/core/auditrecord/mocks"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/role/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Get(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPermissionSvc := mocks.NewPermissionService(t)
	mockAuditRecordRepo := auditMocks.NewRepository(t)

	t.Run("should fetch by id if id is passed", func(t *testing.T) {
		mockID := uuid.New().String()
		expectedRole := role.Role{
			ID:   "role id",
			Name: "role name",
		}

		mockRepository.On("Get", mock.Anything, mockID).Return(expectedRole, nil).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
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

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		res, err := svc.Get(context.Background(), mockSlug)

		assert.Equal(t, nil, err)
		assert.Equal(t, expectedRole, res)
	})

	t.Run("should return an error if fetching role fails", func(t *testing.T) {
		mockID := uuid.New().String()
		expectedErr := errors.New("an error occurred")

		mockRepository.On("Get", mock.Anything, mockID).Return(role.Role{}, expectedErr).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		_, err := svc.Get(context.Background(), mockID)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_List(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPermissionSvc := mocks.NewPermissionService(t)
	mockAuditRecordRepo := auditMocks.NewRepository(t)

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

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		res, err := svc.List(context.Background(), f)

		assert.Equal(t, nil, err)
		assert.Equal(t, expectedRoles, res)
	})

	t.Run("should return an error if fetching roles fails", func(t *testing.T) {
		expectedErr := errors.New("An error occurred")
		f := role.Filter{}
		mockRepository.On("List", mock.Anything, f).Return(nil, expectedErr).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		_, err := svc.List(context.Background(), f)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_Upsert(t *testing.T) {
	mockRepository := mocks.NewRepository(t)
	mockRelationSvc := mocks.NewRelationService(t)
	mockPermissionSvc := mocks.NewPermissionService(t)
	mockAuditRecordRepo := auditMocks.NewRepository(t)

	t.Run("should return an error if one of the permissions in role does not exist", func(t *testing.T) {
		nonExistentPermission := "non_existent_permission"
		roleToBeUpserted := role.Role{
			ID:          "id 1",
			Permissions: []string{"app_project_viewer", nonExistentPermission},
		}
		expectedErr := errors.New("Permission does not exist")

		mockPermissionSvc.On("Get", mock.Anything, "app_project_viewer").Return(permission.Permission{}, nil).Once()
		mockPermissionSvc.On("Get", mock.Anything, nonExistentPermission).Return(permission.Permission{}, expectedErr).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		_, err := svc.Upsert(context.Background(), roleToBeUpserted)

		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("%s: %w", nonExistentPermission, expectedErr), err)
	})

	t.Run("should return an error if upsert of role to repository fails", func(t *testing.T) {
		roleToBeUpserted := role.Role{
			ID:          "id 1",
			Permissions: []string{"app_project_viewer"},
		}
		permissionForRole := permission.Permission{
			ID:          "mock-permission",
			Name:        "mock-permission-name",
			NamespaceID: "project",
		}
		slugForPermission := permissionForRole.GenerateSlug()

		expectedErr := errors.New("Error upserting role")

		mockPermissionSvc.On("Get", mock.Anything, "app_project_viewer").Return(permissionForRole, nil).Once()
		mockRepository.On("Upsert", mock.Anything, role.Role{ID: roleToBeUpserted.ID, Permissions: []string{slugForPermission}}).Return(role.Role{}, expectedErr).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		_, err := svc.Upsert(context.Background(), roleToBeUpserted)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should return an error if role-perimssion relation creation fails", func(t *testing.T) {
		roleToBeUpserted := role.Role{
			ID:          "id 1",
			Permissions: []string{"app_project_viewer"},
		}
		permissionForRole := permission.Permission{
			ID:          "mock-permission",
			Name:        "mock-permission-name",
			NamespaceID: "project",
		}
		slugForPermission := permissionForRole.GenerateSlug()

		roleWithPermSlug := role.Role{
			ID:          roleToBeUpserted.ID,
			Permissions: []string{slugForPermission},
		}

		mockPermissionSvc.On("Get", mock.Anything, "app_project_viewer").Return(permissionForRole, nil).Once()
		mockRepository.On("Upsert", mock.Anything, roleWithPermSlug).Return(roleWithPermSlug, nil).Once()

		userRoleRelation := relation.Relation{
			Object: relation.Object{
				ID:        roleWithPermSlug.ID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*", // all principles who have role will have access
				Namespace: schema.UserPrincipal,
			},
			RelationName: slugForPermission,
		}
		expectedErr := errors.New("Error creating user role relation")
		mockRelationSvc.On("Create", mock.Anything, userRoleRelation).Return(relation.Relation{}, expectedErr).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		_, err := svc.Upsert(context.Background(), roleToBeUpserted)

		assert.NotNil(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should return the created role if all steps are successful", func(t *testing.T) {
		roleToBeUpserted := role.Role{
			ID:          "id 1",
			Permissions: []string{"app_project_viewer"},
		}
		permissionForRole := permission.Permission{
			ID:          "mock-permission",
			Name:        "mock-permission-name",
			NamespaceID: "project",
		}
		slugForPermission := permissionForRole.GenerateSlug()

		roleWithPermSlug := role.Role{
			ID:          roleToBeUpserted.ID,
			Permissions: []string{slugForPermission},
		}

		mockPermissionSvc.On("Get", mock.Anything, "app_project_viewer").Return(permissionForRole, nil).Once()
		mockRepository.On("Upsert", mock.Anything, roleWithPermSlug).Return(roleWithPermSlug, nil).Once()

		userRoleRelation := relation.Relation{
			Object: relation.Object{
				ID:        roleWithPermSlug.ID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*", // all principles who have role will have access
				Namespace: schema.UserPrincipal,
			},
			RelationName: slugForPermission,
		}
		mockRelationSvc.On("Create", mock.Anything, userRoleRelation).Return(relation.Relation{}, nil).Once()

		serviceUserRoleRelation := relation.Relation{
			Object: relation.Object{
				ID:        roleWithPermSlug.ID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*", // all principles who have role will have access
				Namespace: schema.ServiceUserPrincipal,
			},
			RelationName: slugForPermission,
		}
		mockRelationSvc.On("Create", mock.Anything, serviceUserRoleRelation).Return(relation.Relation{}, nil).Once()

		patRoleRelation := relation.Relation{
			Object: relation.Object{
				ID:        roleWithPermSlug.ID,
				Namespace: schema.RoleNamespace,
			},
			Subject: relation.Subject{
				ID:        "*",
				Namespace: schema.PATPrincipal,
			},
			RelationName: slugForPermission,
		}
		mockRelationSvc.On("Create", mock.Anything, patRoleRelation).Return(relation.Relation{}, nil).Once()

		// Mock audit record repository
		mockAuditRecordRepo.On("Create", mock.Anything, mock.Anything).Return(auditrecord.AuditRecord{}, nil).Once()

		svc := role.NewService(mockRepository, mockRelationSvc, mockPermissionSvc, mockAuditRecordRepo, nil)
		roleCreated, err := svc.Upsert(context.Background(), roleToBeUpserted)

		assert.Nil(t, err)
		assert.Equal(t, roleWithPermSlug, roleCreated)
	})

	t.Run("should skip PAT wildcard for denied permissions", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		relSvc := mocks.NewRelationService(t)
		permSvc := mocks.NewPermissionService(t)
		auditRepo := auditMocks.NewRepository(t)

		perm := permission.Permission{
			ID:          "perm-1",
			Name:        "administer",
			NamespaceID: "organization",
		}
		slug := perm.GenerateSlug()

		permSvc.On("Get", mock.Anything, "app_organization_administer").Return(perm, nil).Once()
		repo.On("Upsert", mock.Anything, mock.Anything).Return(role.Role{
			ID:          "role-1",
			Permissions: []string{slug},
		}, nil).Once()

		// only user and serviceuser wildcards — NO PAT wildcard
		relSvc.On("Create", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.UserPrincipal},
			RelationName: slug,
		}).Return(relation.Relation{}, nil).Once()
		relSvc.On("Create", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.ServiceUserPrincipal},
			RelationName: slug,
		}).Return(relation.Relation{}, nil).Once()

		auditRepo.On("Create", mock.Anything, mock.Anything).Return(auditrecord.AuditRecord{}, nil).Once()

		deniedPerms := map[string]struct{}{slug: {}}
		svc := role.NewService(repo, relSvc, permSvc, auditRepo, deniedPerms)
		_, err := svc.Upsert(context.Background(), role.Role{
			ID:          "role-1",
			Permissions: []string{"app_organization_administer"},
		})

		assert.Nil(t, err)
		relSvc.AssertExpectations(t) // ensures PAT Create was NOT called
	})
}
