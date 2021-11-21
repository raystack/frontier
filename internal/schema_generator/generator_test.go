package schema_generator

import (
	"github.com/odpf/shield/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSchema(t *testing.T) {
	t.Run("Generate Empty schema with name ", func(t *testing.T) {
		d := definition{
			name: "Test",
		}
		assert.Equal(t, "definition Test {}", buildSchema(d))
	})

	t.Run("Generate Empty schema with name and role ", func(t *testing.T) {
		d := definition{
			name:  "Test",
			roles: []role{{name: "Admin", types: []string{"User"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User
}`, buildSchema(d))
	})

	t.Run("Generate Empty schema with name, role and permission ", func(t *testing.T) {
		d := definition{
			name:  "Test",
			roles: []role{{name: "Admin", types: []string{"User"}, permissions: []string{"read"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User
	permission read = Admin
}`, buildSchema(d))
	})

	t.Run("Add role name and children", func(t *testing.T) {
		d := definition{
			name: "Test",
			roles: []role{
				{name: "Admin", types: []string{"User"}, permissions: []string{"read"}, namespace: "Project"},
				{name: "Member", types: []string{"User"}, namespace: "Group", permissions: []string{"read"}},
			},
		}
		assert.Equal(t, `definition Test {
	relation Project: Project
	relation Group: Group
	permission read = Project->Admin + Group->Member
}`, buildSchema(d))
	})

	t.Run("Should add role subtype", func(t *testing.T) {
		d := definition{
			name:  "Test",
			roles: []role{{name: "Admin", types: []string{"User#member"}, permissions: []string{"read"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User#member
	permission read = Admin
}`, buildSchema(d))
	})

	t.Run("Should add multiple role types", func(t *testing.T) {
		d := definition{
			name:  "Test",
			roles: []role{{name: "Admin", types: []string{"User", "Team#member"}, permissions: []string{"read"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User | Team#member
	permission read = Admin
}`, buildSchema(d))
	})
}

func TestBuildPolicyDefinitions(t *testing.T) {
	t.Run("return policy definitions", func(t *testing.T) {
		policies := []model.Policy{
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Read", Id: "read"},
			},
		}
		def, _ := BuildPolicyDefinitions(policies)
		expectedDef := []definition{
			{
				name: "project",
				roles: []role{
					{
						name:        "admin",
						types:       []string{"User"},
						namespace:   "project",
						permissions: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expectedDef, def)
	})

	t.Run("merge roles in policy definitions", func(t *testing.T) {
		policies := []model.Policy{
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Read", Id: "read"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Write", Id: "write"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Delete", Id: "delete"},
			},
		}
		def, _ := BuildPolicyDefinitions(policies)
		expectedDef := []definition{
			{
				name: "project",
				roles: []role{
					{
						name:        "admin",
						types:       []string{"User"},
						namespace:   "project",
						permissions: []string{"read", "write", "delete"},
					},
				},
			},
		}

		assert.Equal(t, expectedDef, def)
	})

	t.Run("create multiple roles in policy definitions", func(t *testing.T) {
		policies := []model.Policy{
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Read", Id: "read"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Write", Id: "write"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Delete", Id: "delete"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Reader", Id: "reader", Types: []string{"User"}},
				Action:    model.Action{Name: "Read", Id: "read"},
			},
		}
		def, _ := BuildPolicyDefinitions(policies)
		expectedDef := []definition{
			{
				name: "project",
				roles: []role{
					{
						name:        "admin",
						types:       []string{"User"},
						namespace:   "project",
						permissions: []string{"read", "write", "delete"},
					},
					{
						name:        "reader",
						types:       []string{"User"},
						namespace:   "project",
						permissions: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expectedDef, def)
	})

	t.Run("should add roles namespace", func(t *testing.T) {
		policies := []model.Policy{
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", NamespaceId: "Org", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Read", Id: "read"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Write", Id: "write"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
				Action:    model.Action{Name: "Delete", Id: "delete"},
			},
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Reader", Id: "reader", Types: []string{"User"}},
				Action:    model.Action{Name: "Read", Id: "read"},
			},
		}
		def, _ := BuildPolicyDefinitions(policies)
		expectedDef := []definition{
			{
				name: "project",
				roles: []role{
					{
						name:        "admin",
						types:       []string{"User"},
						namespace:   "Org",
						permissions: []string{"read"},
					},
					{
						name:        "admin",
						types:       []string{"User"},
						namespace:   "project",
						permissions: []string{"write", "delete"},
					},
					{
						name:        "reader",
						types:       []string{"User"},
						namespace:   "project",
						permissions: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expectedDef, def)
	})

	t.Run("should support multiple role types", func(t *testing.T) {
		policies := []model.Policy{
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User", "Team#members"}},
				Action:    model.Action{Name: "Read", Id: "read"},
			},
		}
		def, _ := BuildPolicyDefinitions(policies)
		expectedDef := []definition{
			{
				name: "project",
				roles: []role{
					{
						name:        "admin",
						types:       []string{"User", "Team#members"},
						namespace:   "project",
						permissions: []string{"read"},
					},
				},
			},
		}

		assert.Equal(t, expectedDef, def)
	})

	t.Run("should throw error if action namespace is different", func(t *testing.T) {
		policies := []model.Policy{
			{
				Namespace: model.Namespace{Name: "Project", Id: "project"},
				Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User", "Team#members"}},
				Action:    model.Action{Name: "Read", Id: "read", NamespaceId: "org"},
			},
		}
		def, err := BuildPolicyDefinitions(policies)
		expectedDef := []definition{}

		assert.Equal(t, expectedDef, def)
		assert.Errorf(t, err, "actions namespace doesnt match")

	})

}
