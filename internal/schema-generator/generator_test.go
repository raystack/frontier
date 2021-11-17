package schema_generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSchema(t *testing.T) {
	t.Run("Generate Empty schema with name ", func(t *testing.T) {
		d := definition{
			Name: "Test",
		}
		assert.Equal(t, "definition Test {}", build_schema(d))
	})

	t.Run("Generate Empty schema with name and role ", func(t *testing.T) {
		d := definition{
			Name:  "Test",
			Roles: []role{{Name: "Admin", Types: []string{"User"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User
}`, build_schema(d))
	})

	t.Run("Generate Empty schema with name, role and permission ", func(t *testing.T) {
		d := definition{
			Name:  "Test",
			Roles: []role{{Name: "Admin", Types: []string{"User"}, Permission: []string{"read"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User
	permission read = Admin
}`, build_schema(d))
	})

	t.Run("Add role name and children", func(t *testing.T) {
		d := definition{
			Name: "Test",
			Roles: []role{
				{Name: "Admin", Types: []string{"User"}, Permission: []string{"read"}, Namespace: "Project"},
				{Name: "Member", Types: []string{"User"}, Namespace: "Group", Permission: []string{"read"}},
			},
		}
		assert.Equal(t, `definition Test {
	relation Project: Project
	relation Group: Group
	permission read = Project->Admin + Group->Member
}`, build_schema(d))
	})

	t.Run("Should add role subtype", func(t *testing.T) {
		d := definition{
			Name:  "Test",
			Roles: []role{{Name: "Admin", Types: []string{"User#member"}, Permission: []string{"read"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User#member
	permission read = Admin
}`, build_schema(d))
	})

	t.Run("Should add multiple role types", func(t *testing.T) {
		d := definition{
			Name:  "Test",
			Roles: []role{{Name: "Admin", Types: []string{"User", "Team#member"}, Permission: []string{"read"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User | Team#member
	permission read = Admin
}`, build_schema(d))
	})
}

func TestBuildPolicyDefinitions(t *testing.T) {
	t.Run("return policy definitions", func(t *testing.T) {
		policies := []Policy{
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "read",
			},
		}
		def := build_policy_definitions(policies)
		expected_def := []definition{
			{
				Name: "Project",
				Roles: []role{
					{
						Name:       "Admin",
						Types:      []string{"User"},
						Namespace:  "Project",
						Permission: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expected_def, def)
	})

	t.Run("merge roles in policy definitions", func(t *testing.T) {
		policies := []Policy{
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "read",
			},
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "write",
			},
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "delete",
			},
		}
		def := build_policy_definitions(policies)
		expected_def := []definition{
			{
				Name: "Project",
				Roles: []role{
					{
						Name:       "Admin",
						Types:      []string{"User"},
						Namespace:  "Project",
						Permission: []string{"read", "write", "delete"},
					},
				},
			},
		}

		assert.Equal(t, expected_def, def)
	})

	t.Run("create multiple roles in policy definitions", func(t *testing.T) {
		policies := []Policy{
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "read",
			},
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "write",
			},
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "delete",
			},
			{
				Namespace:  "Project",
				Role:       "Reader",
				RoleTypes:  []string{"User"},
				Permission: "read",
			},
		}
		def := build_policy_definitions(policies)
		expected_def := []definition{
			{
				Name: "Project",
				Roles: []role{
					{
						Name:       "Admin",
						Types:      []string{"User"},
						Namespace:  "Project",
						Permission: []string{"read", "write", "delete"},
					},
					{
						Name:       "Reader",
						Types:      []string{"User"},
						Namespace:  "Project",
						Permission: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expected_def, def)
	})

	t.Run("should add roles namespace", func(t *testing.T) {
		policies := []Policy{
			{
				Namespace:     "Project",
				RoleNamespace: "Org",
				Role:          "Admin",
				RoleTypes:     []string{"User"},
				Permission:    "read",
			},
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "write",
			},
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User"},
				Permission: "delete",
			},
			{
				Namespace:  "Project",
				Role:       "Reader",
				RoleTypes:  []string{"User"},
				Permission: "read",
			},
		}
		def := build_policy_definitions(policies)
		expected_def := []definition{
			{
				Name: "Project",
				Roles: []role{
					{
						Name:       "Admin",
						Types:      []string{"User"},
						Namespace:  "Org",
						Permission: []string{"read"},
					},
					{
						Name:       "Admin",
						Types:      []string{"User"},
						Namespace:  "Project",
						Permission: []string{"write", "delete"},
					},
					{
						Name:       "Reader",
						Types:      []string{"User"},
						Namespace:  "Project",
						Permission: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expected_def, def)
	})

	t.Run("should support multiple role types", func(t *testing.T) {
		policies := []Policy{
			{
				Namespace:  "Project",
				Role:       "Admin",
				RoleTypes:  []string{"User", "Team#members"},
				Permission: "read",
			},
		}
		def := build_policy_definitions(policies)
		expected_def := []definition{
			{
				Name: "Project",
				Roles: []role{
					{
						Name:       "Admin",
						Types:      []string{"User", "Team#members"},
						Namespace:  "Project",
						Permission: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expected_def, def)
	})

}
