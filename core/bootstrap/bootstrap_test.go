package bootstrap

import (
	"testing"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
	"github.com/stretchr/testify/assert"
)

func TestGetResourceRole(t *testing.T) {
	t.Run("should create role for team", func(t *testing.T) {
		output := getResourceRole("admin", namespace.Namespace{Id: "team"})
		expected := role.Role{
			Id:        "team_admin",
			Name:      "team_admin",
			Namespace: namespace.Namespace{Id: "team", Name: "Team"},
			Types:     []string{"user"},
		}
		assert.EqualValues(t, expected, output)
	})

	t.Run("should create role for resource", func(t *testing.T) {
		output := getResourceRole("admin", namespace.Namespace{Id: "kafka"})
		expected := role.Role{
			Id:        "kafka_admin",
			Name:      "kafka_admin",
			Namespace: namespace.Namespace{Id: "kafka"},
			Types:     []string{"user", "team#team_member"},
		}
		assert.EqualValues(t, expected, output)
	})

	t.Run("should assign role for other namespace for resources", func(t *testing.T) {
		output := getResourceRole("organization.organization_admin", namespace.Namespace{Id: "team"})
		expected := role.Role{
			Id:        "organization_admin",
			Name:      "organization_admin",
			Namespace: namespace.Namespace{Id: "organization"},
			Types:     []string{"user", "team#team_member"},
		}
		assert.EqualValues(t, expected, output)
	})
}
