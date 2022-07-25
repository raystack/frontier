package authz

import (
	"context"
	"fmt"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"testing"
	"time"

	"github.com/odpf/shield/core/resource"
	"github.com/stretchr/testify/assert"

	"github.com/odpf/shield/api/handler/v1beta1"
)

var projectIdList = []string{"projcet1", "project2"}

var testPermissionAttributesMap = map[string]any{
	"project": "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
}

var testProjectMap = map[string]project.Project{
	"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
		Id:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
		Name: "Prj 1",
		Slug: "prj-1",
		Metadata: map[string]any{
			"email": "org1@org1.com",
		},
		Organization: organization.Organization{
			Id:   "org1",
			Name: "Org 1",
			Slug: "Org Slug 1",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
		Id:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
		Name: "Prj 2",
		Slug: "prj-2",
		Metadata: map[string]any{
			"email": "org1@org2.com",
		},
		Organization: organization.Organization{
			Id:   "org2",
			Name: "Org 2",
			Slug: "Org Slug 2",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestCreateResources(t *testing.T) {
	t.Parallel()

	table := []struct {
		title                string
		mockResourcesServ    mockResources
		permissionAttributes map[string]any
		v                    v1beta1.Dep
		want                 []resource.Resource
		err                  error
	}{
		{
			title: "success",
			mockResourcesServ: mockResources{
				GetProjectFunc: func(ctx context.Context, id string) (project.Project, error) {
					return testProjectMap[id], nil
				},
				createResourcesFunc: func(ctx context.Context, permissionAttributes map[string]interface{}, v v1beta1.Dep) ([]resource.Resource, error) {
					return nil, nil
				},
			},
		},
	}

	fmt.Println(table)

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
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res1",
				NamespaceId:    "ns1_type",
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
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res1",
				NamespaceId:    "ns1_kind",
			},
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res2",
				NamespaceId:    "ns1_kind",
			},
			{
				ProjectId:      "project1",
				OrganizationId: "org1",
				GroupId:        "team1",
				Name:           "res3",
				NamespaceId:    "ns1_kind",
			},
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)
	})
}

type mockResources struct {
	GetProjectFunc      func(ctx context.Context, id string) (project.Project, error)
	createResourcesFunc func(ctx context.Context, permissionAttributes map[string]interface{}, v v1beta1.Dep) ([]resource.Resource, error)
}

func (m mockResources) GetProject(ctx context.Context, id string) (project.Project, error) {
	return m.GetProjectFunc(ctx, id)
}

func (m mockResources) createResources(ctx context.Context, permissionAttributes map[string]interface{}, v v1beta1.Dep) ([]resource.Resource, error) {
	return m.createResourcesFunc(ctx, permissionAttributes, v)
}
