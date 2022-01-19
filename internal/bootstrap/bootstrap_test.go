package bootstrap

import (
	"github.com/odpf/shield/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetResourceRole(t *testing.T) {
	t.Run("should create role for resources", func(t *testing.T) {
		output := getResourceRole("admin", model.Namespace{Id: "team"})
		expected := model.Role{
			Id:        "team_admin",
			Name:      "team_admin",
			Namespace: model.Namespace{Id: "team"},
			Types:     []string{"user", "team#team_member"},
		}
		assert.EqualValues(t, expected, output)
	})

	t.Run("should assign role for other namespace for resources", func(t *testing.T) {
		output := getResourceRole("organization.organization_admin", model.Namespace{Id: "team"})
		expected := model.Role{
			Id:        "organization_admin",
			Name:      "organization_admin",
			Namespace: model.Namespace{Id: "organization"},
			Types:     []string{"user", "team#team_member"},
		}
		assert.EqualValues(t, expected, output)
	})
}
