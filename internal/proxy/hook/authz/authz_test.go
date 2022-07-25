package authz

import (
	"testing"

	"github.com/odpf/shield/core/resource"
	"github.com/stretchr/testify/assert"
)

func TestCreateResources(t *testing.T) {
	t.Run("should should throw error if project is missing", func(t *testing.T) {
		input := map[string]interface{}{
			"abc": "abc",
		}
		output, err := createResources(input)
		var expected []resource.Resource
		assert.EqualValues(t, expected, output)
		assert.Error(t, err)
	})

	t.Run("should should throw error if team is missing", func(t *testing.T) {
		input := map[string]interface{}{
			"project": "abc",
		}
		output, err := createResources(input)
		var expected []resource.Resource
		assert.EqualValues(t, expected, output)
		assert.Error(t, err)
	})

	t.Run("should return resource", func(t *testing.T) {
		input := map[string]interface{}{
			"project":       "project1",
			"team":          "team1",
			"organization":  "org1",
			"resource":      "res1",
			"namespace":     "ns1",
			"resource_type": "type",
		}
		output, err := createResources(input)
		expected := []resource.Resource{
			{
				ProjectID:      "project1",
				OrganizationID: "org1",
				GroupID:        "team1",
				Name:           "res1",
				NamespaceID:    "ns1_type",
			},
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})

	t.Run("should return multiple resource", func(t *testing.T) {
		input := map[string]interface{}{
			"project":       "project1",
			"team":          "team1",
			"organization":  "org1",
			"namespace":     "ns1",
			"resource":      []string{"res1", "res2", "res3"},
			"resource_type": "kind",
		}
		output, err := createResources(input)
		expected := []resource.Resource{
			{
				ProjectID:      "project1",
				OrganizationID: "org1",
				GroupID:        "team1",
				Name:           "res1",
				NamespaceID:    "ns1_kind",
			},
			{
				ProjectID:      "project1",
				OrganizationID: "org1",
				GroupID:        "team1",
				Name:           "res2",
				NamespaceID:    "ns1_kind",
			},
			{
				ProjectID:      "project1",
				OrganizationID: "org1",
				GroupID:        "team1",
				Name:           "res3",
				NamespaceID:    "ns1_kind",
			},
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})
}
