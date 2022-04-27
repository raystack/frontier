package bootstrap

import (
	"testing"

	"github.com/odpf/shield/model"
	"github.com/stretchr/testify/assert"
)

func TestGetResourceRole(t *testing.T) {
	t.Run("should create role for team", func(t *testing.T) {
		output := getResourceRole("admin", model.Namespace{Id: "team"})
		expected := model.Role{
			Id:        "team_admin",
			Name:      "team_admin",
			Namespace: model.Namespace{Id: "team", Name: "Team"},
			Types:     []string{"user"},
		}
		assert.EqualValues(t, expected, output)
	})

	t.Run("should create role for resource", func(t *testing.T) {
		output := getResourceRole("admin", model.Namespace{Id: "kafka"})
		expected := model.Role{
			Id:        "kafka_admin",
			Name:      "kafka_admin",
			Namespace: model.Namespace{Id: "kafka"},
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
