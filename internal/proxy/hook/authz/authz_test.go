package authz

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/shield/core/resource"
)

var testPermissionAttributesMap = map[string]any{
	"project":       "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
	"team":          "team1",
	"resource":      []string{"resc1", "resc2"},
	"organization":  "org1",
	"namespace":     "ns1",
	"resource_type": "kind",
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
		permissionAttributes map[string]any
		a                    Authz
		want                 []resource.Resource
		err                  error
	}{
		{
			title:                "success/should return multiple resources",
			permissionAttributes: testPermissionAttributesMap,
			a:                    Authz{},
			want:                 expectedResources,
			err:                  nil,
		}, {
			title: "should should throw error if project is missing",
			permissionAttributes: map[string]any{
				"team":          "team1",
				"resource":      []string{"resc1", "resc2"},
				"organization":  "org1",
				"namespace":     "ns1",
				"resource_type": "kind",
			},
			a:    Authz{},
			want: nil,
			err:  fmt.Errorf("namespace, resource type, projects, resource, and team are required"),
		}, {
			title: "should should throw error if team is missing",
			permissionAttributes: map[string]any{
				"project":       "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
				"resource":      []string{"resc1", "resc2"},
				"organization":  "org1",
				"namespace":     "ns1",
				"resource_type": "kind",
			},
			a:    Authz{},
			want: nil,
			err:  fmt.Errorf("namespace, resource type, projects, resource, and team are required"),
		}, {
			title: "success/should return resource",
			permissionAttributes: map[string]any{
				"project":       "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
				"team":          "team1",
				"organization":  "org2",
				"resource":      "res1",
				"namespace":     "ns1",
				"resource_type": "type",
			},
			a: Authz{},
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

			resp, err := tt.a.createResources(tt.permissionAttributes)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}
