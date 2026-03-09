package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRoleService implements bootstrap.RoleService
type mockRoleService struct {
	mock.Mock
}

func (m *mockRoleService) Get(ctx context.Context, id string) (role.Role, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(role.Role), args.Error(1)
}

func (m *mockRoleService) List(ctx context.Context, f role.Filter) ([]role.Role, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]role.Role), args.Error(1)
}

func (m *mockRoleService) Upsert(ctx context.Context, toCreate role.Role) (role.Role, error) {
	args := m.Called(ctx, toCreate)
	return args.Get(0).(role.Role), args.Error(1)
}

// mockRelationService implements bootstrap.RelationService
type mockRelationService struct {
	mock.Mock
}

func (m *mockRelationService) Create(ctx context.Context, rel relation.Relation) (relation.Relation, error) {
	args := m.Called(ctx, rel)
	return args.Get(0).(relation.Relation), args.Error(1)
}

func (m *mockRelationService) Delete(ctx context.Context, rel relation.Relation) error {
	args := m.Called(ctx, rel)
	return args.Error(0)
}

func Test_migratePATRelations(t *testing.T) {
	t.Run("should create PAT wildcards for allowed permissions", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return([]role.Role{
			{ID: "role-1", Name: "viewer", Permissions: []string{"app_organization_get"}},
		}, nil)

		relSvc.On("Create", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.PATPrincipal},
			RelationName: "app_organization_get",
		}).Return(relation.Relation{}, nil).Once()

		svc := Service{roleService: roleSvc, relationService: relSvc}
		err := svc.migratePATRelations(context.Background())

		assert.NoError(t, err)
		relSvc.AssertExpectations(t)
	})

	t.Run("should delete PAT wildcards for denied permissions", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return([]role.Role{
			{ID: "role-1", Name: "admin", Permissions: []string{"app_organization_administer"}},
		}, nil)

		relSvc.On("Delete", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.PATPrincipal},
			RelationName: "app_organization_administer",
		}).Return(nil).Once()

		svc := Service{
			roleService:     roleSvc,
			relationService: relSvc,
			patDeniedPerms:  map[string]struct{}{"app_organization_administer": {}},
		}
		err := svc.migratePATRelations(context.Background())

		assert.NoError(t, err)
		relSvc.AssertExpectations(t)
	})

	t.Run("should handle mixed allowed and denied permissions across roles", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return([]role.Role{
			{ID: "role-1", Name: "manager", Permissions: []string{
				"app_organization_administer", // denied
				"app_organization_get",        // allowed
				"app_organization_update",     // allowed
			}},
			{ID: "role-2", Name: "viewer", Permissions: []string{
				"app_organization_get", // allowed
			}},
		}, nil)

		// role-1: delete denied, create allowed
		relSvc.On("Delete", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.PATPrincipal},
			RelationName: "app_organization_administer",
		}).Return(nil).Once()
		relSvc.On("Create", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.PATPrincipal},
			RelationName: "app_organization_get",
		}).Return(relation.Relation{}, nil).Once()
		relSvc.On("Create", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.PATPrincipal},
			RelationName: "app_organization_update",
		}).Return(relation.Relation{}, nil).Once()

		// role-2: create allowed
		relSvc.On("Create", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-2", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.PATPrincipal},
			RelationName: "app_organization_get",
		}).Return(relation.Relation{}, nil).Once()

		svc := Service{
			roleService:     roleSvc,
			relationService: relSvc,
			patDeniedPerms:  map[string]struct{}{"app_organization_administer": {}},
		}
		err := svc.migratePATRelations(context.Background())

		assert.NoError(t, err)
		relSvc.AssertExpectations(t)
	})

	t.Run("should be a no-op for empty roles list", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return([]role.Role{}, nil)

		svc := Service{roleService: roleSvc, relationService: relSvc}
		err := svc.migratePATRelations(context.Background())

		assert.NoError(t, err)
		relSvc.AssertNotCalled(t, "Create")
		relSvc.AssertNotCalled(t, "Delete")
	})

	t.Run("should return error when listing roles fails", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return(nil, errors.New("db error"))

		svc := Service{roleService: roleSvc, relationService: relSvc}
		err := svc.migratePATRelations(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "listing roles for PAT migration")
	})

	t.Run("should return error when creating relation fails", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return([]role.Role{
			{ID: "role-1", Name: "viewer", Permissions: []string{"app_organization_get"}},
		}, nil)

		relSvc.On("Create", mock.Anything, mock.Anything).
			Return(relation.Relation{}, errors.New("spicedb unavailable")).Once()

		svc := Service{roleService: roleSvc, relationService: relSvc}
		err := svc.migratePATRelations(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "creating PAT wildcard for role viewer")
	})

	t.Run("should return error when deleting denied relation fails", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return([]role.Role{
			{ID: "role-1", Name: "admin", Permissions: []string{"app_organization_administer"}},
		}, nil)

		relSvc.On("Delete", mock.Anything, mock.Anything).
			Return(errors.New("spicedb unavailable")).Once()

		svc := Service{
			roleService:     roleSvc,
			relationService: relSvc,
			patDeniedPerms:  map[string]struct{}{"app_organization_administer": {}},
		}
		err := svc.migratePATRelations(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deleting PAT wildcard for role admin denied permission")
	})

	t.Run("should handle nil denied permissions as all allowed", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		relSvc := new(mockRelationService)

		roleSvc.On("List", mock.Anything, role.Filter{}).Return([]role.Role{
			{ID: "role-1", Name: "admin", Permissions: []string{"app_organization_administer"}},
		}, nil)

		relSvc.On("Create", mock.Anything, relation.Relation{
			Object:       relation.Object{ID: "role-1", Namespace: schema.RoleNamespace},
			Subject:      relation.Subject{ID: "*", Namespace: schema.PATPrincipal},
			RelationName: "app_organization_administer",
		}).Return(relation.Relation{}, nil).Once()

		svc := Service{roleService: roleSvc, relationService: relSvc} // nil patDeniedPerms = all allowed
		err := svc.migratePATRelations(context.Background())

		assert.NoError(t, err)
		relSvc.AssertNotCalled(t, "Delete")
	})
}
