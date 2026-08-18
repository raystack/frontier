package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
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

func (m *mockRoleService) Update(ctx context.Context, toUpdate role.Role) (role.Role, error) {
	args := m.Called(ctx, toUpdate)
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

// mockPermissionService implements bootstrap.PermissionService
type mockPermissionService struct {
	mock.Mock
}

func (m *mockPermissionService) List(ctx context.Context, flt permission.Filter) ([]permission.Permission, error) {
	args := m.Called(ctx, flt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]permission.Permission), args.Error(1)
}

func (m *mockPermissionService) Upsert(ctx context.Context, action permission.Permission) (permission.Permission, error) {
	args := m.Called(ctx, action)
	return args.Get(0).(permission.Permission), args.Error(1)
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

func Test_migrateRole(t *testing.T) {
	def := schema.RoleDefinition{
		Title:       "Organization Manager",
		Name:        "app_organization_manager",
		Permissions: []string{"app_organization_get", "app_organization_update"},
		Scopes:      []string{schema.OrganizationNamespace},
	}

	t.Run("creates the role when it does not exist", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		roleSvc.On("Get", mock.Anything, def.Name).Return(role.Role{}, role.ErrNotExist)
		roleSvc.On("Upsert", mock.Anything, mock.MatchedBy(func(r role.Role) bool {
			return r.Name == def.Name && len(r.Permissions) == 2
		})).Return(role.Role{ID: "role-1"}, nil)

		svc := Service{roleService: roleSvc}
		assert.NoError(t, svc.migrateRole(context.Background(), "org-1", def))
		roleSvc.AssertNotCalled(t, "Update")
	})

	t.Run("leaves an existing role alone even when it differs from the definition", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		roleSvc.On("Get", mock.Anything, def.Name).Return(role.Role{
			ID:    "role-1",
			Name:  def.Name,
			Title: "Renamed By Operator",
			// differs from the definition; operators own existing roles, so
			// boot must not write it back
			Permissions: []string{"app_organization_get"},
		}, nil)

		svc := Service{roleService: roleSvc}
		assert.NoError(t, svc.migrateRole(context.Background(), "org-1", def))
		roleSvc.AssertNotCalled(t, "Upsert")
		roleSvc.AssertNotCalled(t, "Update")
	})

	t.Run("propagates a transient Get error instead of creating", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		// a non-ErrNotExist failure must not fall through to Upsert/Update
		roleSvc.On("Get", mock.Anything, def.Name).Return(role.Role{}, errors.New("db timeout"))

		svc := Service{roleService: roleSvc}
		err := svc.migrateRole(context.Background(), "org-1", def)
		assert.Error(t, err)
		roleSvc.AssertNotCalled(t, "Upsert")
		roleSvc.AssertNotCalled(t, "Update")
	})
}

func Test_AppendSchema(t *testing.T) {
	t.Run("returns the error when listing existing permissions fails", func(t *testing.T) {
		// The old code returned nil here, so a failed list skipped the schema
		// re-apply but still reported boot success. Boot must surface the error.
		permSvc := new(mockPermissionService)
		permSvc.On("List", mock.Anything, permission.Filter{}).
			Return(nil, errors.New("db timeout"))

		svc := Service{permissionService: permSvc}
		err := svc.AppendSchema(context.Background(), schema.ServiceDefinition{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db timeout")
	})
}

func Test_existingPermissionsAsServiceDefinition(t *testing.T) {
	perms := []permission.Permission{
		{Name: "get", NamespaceID: "compute/order", Metadata: metadata.Metadata{"description": "read an order"}},
		{Name: "delete", NamespaceID: "compute/order"},
	}

	def := existingPermissionsAsServiceDefinition(perms)

	assert.Equal(t, []schema.ResourcePermission{
		{Name: "get", Namespace: "compute/order", Description: "read an order"},
		{Name: "delete", Namespace: "compute/order", Description: ""},
	}, def.Permissions)
}

func Test_permissionDescription(t *testing.T) {
	cases := []struct {
		name string
		meta metadata.Metadata
		want string
	}{
		{"nil metadata", nil, ""},
		// a present key must be read, not ignored
		{"string description", metadata.Metadata{"description": "read access"}, "read access"},
		// a missing key must not panic; it used to assert nil to string
		{"missing description key", metadata.Metadata{"other": "x"}, ""},
		// a non-string value must not panic either
		{"non-string description", metadata.Metadata{"description": 42}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := permissionDescription(tc.meta)
			assert.Equal(t, tc.want, got)
		})
	}
}
