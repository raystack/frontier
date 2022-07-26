package authz

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/internal/api"
)

var testPermissionAttributesMap = map[string]any{
	"project":       "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
	"team":          "team1",
	"resource":      []string{"resc1", "resc2"},
	"namespace":     "ns1",
	"resource_type": "kind",
}

var testProjectMap = map[string]project.Project{
	"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
		ID:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
		Name: "Prj 1",
		Slug: "prj-1",
		Metadata: map[string]any{
			"email": "org1@org1.com",
		},
		Organization: organization.Organization{
			ID:   "org1",
			Name: "Org 1",
			Slug: "Org Slug 1",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
		ID:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
		Name: "Prj 2",
		Slug: "prj-2",
		Metadata: map[string]any{
			"email": "org1@org2.com",
		},
		Organization: organization.Organization{
			ID:   "org2",
			Name: "Org 2",
			Slug: "Org Slug 2",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"project-3-slug": {
		ID:   "c3772d61-faa1-4d8d-fff3-c8fa5a1fdc4b",
		Name: "Prj 3",
		Slug: "project-3-slug",
		Metadata: map[string]any{
			"email": "org1@org2.com",
		},
		Organization: organization.Organization{
			ID:   "org2",
			Name: "Org 2",
			Slug: "Org Slug 2",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

var expectedResources = []resource.Resource{
	{
		ProjectID:      "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
		OrganizationID: "org1",
		GroupID:        "team1",
		Name:           "resc1",
		NamespaceID:    "ns1_kind",
	}, {
		ProjectID:      "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
		OrganizationID: "org1",
		GroupID:        "team1",
		Name:           "resc2",
		NamespaceID:    "ns1_kind",
	},
}

func TestCreateResources(t *testing.T) {
	t.Parallel()

	table := []struct {
		title                string
		mockProjectServ      mockProject
		permissionAttributes map[string]any
		v                    api.Deps
		want                 []resource.Resource
		err                  error
	}{
		{
			title: "success/should return multiple resources",
			mockProjectServ: mockProject{
				GetProjectFunc: func(ctx context.Context, id string) (project.Project, error) {
					return testProjectMap[id], nil
				}},
			permissionAttributes: testPermissionAttributesMap,
			v:                    api.Deps{},
			want:                 expectedResources,
			err:                  nil,
		}, {
			title: "should should throw error if project is missing",
			mockProjectServ: mockProject{
				GetProjectFunc: func(ctx context.Context, id string) (project.Project, error) {
					return project.Project{}, fmt.Errorf("Project ID not found")
				},
			},
			permissionAttributes: map[string]any{
				"team":          "team1",
				"resource":      []string{"resc1", "resc2"},
				"namespace":     "ns1",
				"resource_type": "kind",
			},
			v:    api.Deps{},
			want: nil,
			err:  fmt.Errorf("namespace, resource type, projects, resource, and team are required"),
		}, {
			title: "should should throw error if team is missing",
			mockProjectServ: mockProject{
				GetProjectFunc: func(ctx context.Context, id string) (project.Project, error) {
					return testProjectMap[id], nil
				},
			},
			permissionAttributes: map[string]any{
				"project":       "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
				"resource":      []string{"resc1", "resc2"},
				"namespace":     "ns1",
				"resource_type": "kind",
			},
			v:    api.Deps{},
			want: nil,
			err:  fmt.Errorf("namespace, resource type, projects, resource, and team are required"),
		}, {
			title: "success/should return resource",
			mockProjectServ: mockProject{
				GetProjectFunc: func(ctx context.Context, id string) (project.Project, error) {
					return testProjectMap[id], nil
				}},
			permissionAttributes: map[string]any{
				"project":       "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
				"team":          "team1",
				"resource":      "res1",
				"namespace":     "ns1",
				"resource_type": "type",
			},
			v: api.Deps{},
			want: []resource.Resource{
				{
					ProjectID:      "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
					OrganizationID: "org2",
					GroupID:        "team1",
					Name:           "res1",
					NamespaceID:    "ns1_type",
				},
			},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			resp, err := createResources(context.Background(), tt.permissionAttributes, tt.mockProjectServ)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockProject struct {
	GetProjectFunc func(ctx context.Context, id string) (project.Project, error)
}

func (m mockProject) Get(ctx context.Context, id string) (project.Project, error) {
	return m.GetProjectFunc(ctx, id)
}
