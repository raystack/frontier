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

// mockUserService implements bootstrap.UserService (additive promote only).
type mockUserService struct{ mock.Mock }

func (m *mockUserService) Sudo(ctx context.Context, id, relationName string) error {
	return m.Called(ctx, id, relationName).Error(0)
}

func TestMakeSuperUsers(t *testing.T) {
	t.Run("promotes each configured user as platform admin (trimmed)", func(t *testing.T) {
		userSvc := new(mockUserService)
		userSvc.On("Sudo", mock.Anything, "alice@x.com", schema.AdminRelationName).Return(nil)
		userSvc.On("Sudo", mock.Anything, "bob@x.com", schema.AdminRelationName).Return(nil)

		s := Service{
			adminConfig: AdminConfig{Users: []string{"alice@x.com", "  bob@x.com  "}},
			userService: userSvc,
		}
		assert.NoError(t, s.MakeSuperUsers(context.Background()))
		userSvc.AssertExpectations(t)
	})

	t.Run("returns the error when a promotion fails", func(t *testing.T) {
		userSvc := new(mockUserService)
		userSvc.On("Sudo", mock.Anything, "alice@x.com", schema.AdminRelationName).Return(errors.New("boom"))

		s := Service{
			adminConfig: AdminConfig{Users: []string{"alice@x.com"}},
			userService: userSvc,
		}
		assert.Error(t, s.MakeSuperUsers(context.Background()))
	})
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

	t.Run("skips reconcile when the permission set is unchanged", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		roleSvc.On("Get", mock.Anything, def.Name).Return(role.Role{
			ID:   "role-1",
			Name: def.Name,
			// same set, different order -> still equal
			Permissions: []string{"app_organization_update", "app_organization_get"},
		}, nil)

		svc := Service{roleService: roleSvc}
		assert.NoError(t, svc.migrateRole(context.Background(), "org-1", def))
		roleSvc.AssertNotCalled(t, "Upsert")
		roleSvc.AssertNotCalled(t, "Update")
	})

	t.Run("reconciles a drifted (over-granting) role to the definition", func(t *testing.T) {
		roleSvc := new(mockRoleService)
		roleSvc.On("Get", mock.Anything, def.Name).Return(role.Role{
			ID:   "role-1",
			Name: def.Name,
			// has an extra permission no longer in the definition
			Permissions: []string{"app_organization_get", "app_organization_update", "app_organization_delete"},
		}, nil)
		roleSvc.On("Update", mock.Anything, mock.MatchedBy(func(r role.Role) bool {
			return r.ID == "role-1" && len(r.Permissions) == 2 &&
				!contains(r.Permissions, "app_organization_delete")
		})).Return(role.Role{ID: "role-1"}, nil)

		svc := Service{roleService: roleSvc}
		assert.NoError(t, svc.migrateRole(context.Background(), "org-1", def))
		roleSvc.AssertNotCalled(t, "Upsert")
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

func contains(s []string, v string) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

func Test_permissionsEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"equal ignoring order", []string{"x", "y"}, []string{"y", "x"}, true},
		{"equal ignoring duplicates", []string{"x", "x"}, []string{"x"}, true},
		{"different members", []string{"x", "y"}, []string{"x", "z"}, false},
		{"superset", []string{"x"}, []string{"x", "y"}, false},
		{"both empty", nil, []string{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, permissionsEqual(c.a, c.b))
		})
	}
}
