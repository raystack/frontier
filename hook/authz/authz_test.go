package authz

import (
	"testing"

	"github.com/odpf/shield/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateResources(t *testing.T) {
	t.Run("should throw error if project is missing", func(t *testing.T) {
		input := map[string]interface{}{
			"abc": "abc",
		}
		output, err := createResources(input)
		var expected []model.Resource
		assert.EqualValues(t, expected, output)
		assert.Error(t, err)
	})

	t.Run("should throw error if team is missing", func(t *testing.T) {
		input := map[string]interface{}{
			"project": "abc",
		}
		output, err := createResources(input)
		var expected []model.Resource
		assert.EqualValues(t, expected, output)
		assert.Error(t, err)
	})

	t.Run("should return resource", func(t *testing.T) {
		input := map[string]interface{}{
			"project":      "project1",
			"team":         "team1",
			"organization": "org1",
			"resource":     "res1",
			"namespace":    "ns1",
		}
		output, err := createResources(input)
		expected := []model.Resource{
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res1",
				NamespaceId:    "ns1",
			},
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})

	t.Run("should return multiple resource", func(t *testing.T) {
		input := map[string]interface{}{
			"project":      "project1",
			"team":         "team1",
			"organization": "org1",
			"namespace":    "ns1",
			"resource":     []string{"res1", "res2", "res3"},
		}
		output, err := createResources(input)
		expected := []model.Resource{
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res1",
				NamespaceId:    "ns1",
			},
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res2",
				NamespaceId:    "ns1",
			},
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res3",
				NamespaceId:    "ns1",
			},
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})
}

func TestGetAttributesValues(t *testing.T) {
	t.Run("should return slice of strings for slice of strings", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		output, err := getAttributesValues(input)
		assert.Nil(t, err)
		expected := []string{"a", "b", "c"}
		assert.EqualValues(t, expected, output)
	})
	t.Run("should return slice of strings for slice of interface{}", func(t *testing.T) {
		input := []interface{}{"a", "b", "c"}
		output, err := getAttributesValues(input)
		assert.Nil(t, err)
		expected := []string{"a", "b", "c"}
		assert.EqualValues(t, expected, output)
	})
	t.Run("should return nil string slice for nil", func(t *testing.T) {
		output, err := getAttributesValues(nil)
		assert.Nil(t, err)
		var expected []string
		assert.EqualValues(t, expected, output)
	})

}
