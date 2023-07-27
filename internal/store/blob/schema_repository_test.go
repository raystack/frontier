package blob

import (
	"context"
	"sort"
	"testing"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/stretchr/testify/assert"
)

func TestGetSchema(t *testing.T) {
	testBucket, err := NewStore(context.Background(), "file://testdata", "")
	assert.NoError(t, err)

	s := SchemaRepository{
		bucket: testBucket,
	}
	def, err := s.GetDefinition(context.Background())
	assert.NoError(t, err)

	expectedMap := &schema.ServiceDefinition{
		Roles: []schema.RoleDefinition{
			{
				Name: "compute_order_manager",
				Permissions: []string{
					"compute/order:delete",
					"compute/order:update",
					"compute/order:get",
					"compute/order:list",
					"compute/order:create",
				},
			},
			{
				Name: "compute_order_viewer",
				Permissions: []string{
					"compute/order:list",
					"compute/order:get",
				},
			},
			{
				Name: "compute_order_owner",
				Permissions: []string{
					"compute/order:delete",
					"compute/order:update",
					"compute/order:get",
					"compute/order:create",
				},
			},
		},
		Permissions: []schema.ResourcePermission{
			{
				Name:      "delete",
				Namespace: "compute/order",
			},
			{
				Name:      "update",
				Namespace: "compute/order",
			},
			{
				Name:      "get",
				Namespace: "compute/order",
			},
			{
				Name:      "list",
				Namespace: "compute/order",
			},
			{
				Name:      "create",
				Namespace: "compute/order",
			},
			{
				Name:      "delete",
				Namespace: "database/order",
			},
			{
				Name:      "update",
				Namespace: "database/order",
			},
			{
				Name:      "get",
				Namespace: "database/order",
			},
		},
	}
	sort.Slice(def.Roles, func(i, j int) bool {
		return def.Roles[i].Name < def.Roles[j].Name
	})
	sort.Slice(expectedMap.Roles, func(i, j int) bool {
		return expectedMap.Roles[i].Name < expectedMap.Roles[j].Name
	})
	sort.Slice(def.Permissions, func(i, j int) bool {
		return schema.FQPermissionNameFromNamespace(def.Permissions[i].Namespace, def.Permissions[i].Name) < schema.FQPermissionNameFromNamespace(def.Permissions[j].Namespace, def.Permissions[j].Name)
	})
	sort.Slice(expectedMap.Permissions, func(i, j int) bool {
		return schema.FQPermissionNameFromNamespace(expectedMap.Permissions[i].Namespace, expectedMap.Permissions[i].Name) < schema.FQPermissionNameFromNamespace(expectedMap.Permissions[j].Namespace, expectedMap.Permissions[j].Name)
	})
	assert.Equal(t, expectedMap, def)
}
