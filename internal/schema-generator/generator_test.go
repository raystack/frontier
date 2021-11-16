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
			Roles: []role{{Name: "Admin", Type: "User"}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User
}`, build_schema(d))
	})

	t.Run("Generate Empty schema with name, role and permission ", func(t *testing.T) {
		d := definition{
			Name:  "Test",
			Roles: []role{{Name: "Admin", Type: "User", Permission: []string{"read"}}},
		}
		assert.Equal(t, `definition Test {
	relation Admin: User
	permission read = Admin
}`, build_schema(d))
	})
}
