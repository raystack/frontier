package authz

import (
	"github.com/odpf/shield/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateResources(t *testing.T) {
	t.Run("should return empty list", func(t *testing.T) {
		input := map[string]interface{}{}
		output, err := createResources(input)
		var expected []model.Resource
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})

	t.Run("should should throw error if project is missing", func(t *testing.T) {
		input := map[string]interface{}{
			"abc": "abc",
		}
		output, err := createResources(input)
		var expected []model.Resource
		assert.EqualValues(t, expected, output)
		assert.Error(t, err)
	})

	t.Run("should should throw error if team is missing", func(t *testing.T) {
		input := map[string]interface{}{
			"project": "abc",
		}
		output, err := createResources(input)
		var expected []model.Resource
		assert.EqualValues(t, expected, output)
		assert.Error(t, err)
	})

	t.Run("should should return resource", func(t *testing.T) {
		input := map[string]interface{}{
			"project":      "project1",
			"team":         "team1",
			"organization": "org1",
			"resource":     "res1",
		}
		output, err := createResources(input)
		expected := []model.Resource{
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res1",
			},
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})

	t.Run("should should return multiple resource", func(t *testing.T) {
		input := map[string]interface{}{
			"project":      "project1",
			"team":         "team1",
			"organization": "org1",
			"resource":     []string{"res1", "res2", "res3"},
		}
		output, err := createResources(input)
		expected := []model.Resource{
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res1",
			},
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res2",
			},
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res3",
			},
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})
}
