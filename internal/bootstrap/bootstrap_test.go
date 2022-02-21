package bootstrap

import (
	"testing"

	"github.com/odpf/shield/structs"

	"github.com/odpf/shield/model"
	"github.com/stretchr/testify/assert"
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

func TestGetOwnerRole(t *testing.T) {
	t.Run("should get owner role for namespace", func(t *testing.T) {
		ns := model.Namespace{Id: "organization", Name: "Organization"}
		output := GetOwnerRole(ns)
		expected := model.Role{
			Id:        "organization_owner",
			Name:      "Organization_Owner",
			Namespace: ns,
			Types:     []string{"user"},
		}
		assert.EqualValues(t, expected, output)
	})
}

func TestGetResourceAction(t *testing.T) {
	t.Run("should get resource action", func(t *testing.T) {
		output := getResourceAction("admin", model.Namespace{Id: "team"})
		expected := model.Action{
			Id:          "team_admin",
			Name:        "Team Admin",
			NamespaceId: "team",
			Namespace:   model.Namespace{Id: "team"},
		}
		assert.EqualValues(t, expected, output)
	})
}

func TestGetResourceNamespace(t *testing.T) {
	t.Run("should get resource namespace", func(t *testing.T) {
		output := getResourceNamespace(structs.Resource{
			Name:    "foo bar",
			Actions: nil,
		})
		expected := model.Namespace{
			Id:   "foo_bar",
			Name: "foo bar",
		}
		assert.EqualValues(t, expected, output)
	})
}
