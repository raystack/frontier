package reconcile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffPermissions(t *testing.T) {
	current := []currentPermission{
		{ID: "p1", Key: "compute.order.get"},
		{ID: "p2", Key: "compute.order.legacy"},
	}

	t.Run("adds missing and deletes flagged, adds first", func(t *testing.T) {
		ops, err := diffPermissions([]PermissionSpec{
			{Key: "compute.order.get"},
			{Key: "compute.order.legacy", Delete: true},
			{Key: "compute.disk.mount"},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 2) {
			assert.Equal(t, "add permission compute.disk.mount", ops[0].String())
			assert.Equal(t, "delete permission compute.order.legacy", ops[1].String())
			assert.Equal(t, "p2", ops[1].id)
		}
	})

	t.Run("no changes when converged", func(t *testing.T) {
		ops, err := diffPermissions([]PermissionSpec{
			{Key: "compute.order.get"},
			{Key: "compute.order.legacy"},
		}, current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("delete of an absent permission is a no-op", func(t *testing.T) {
		ops, err := diffPermissions([]PermissionSpec{
			{Key: "compute.order.get"},
			{Key: "compute.order.legacy"},
			{Key: "compute.order.gone", Delete: true},
		}, current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("a server permission missing from the file fails the plan", func(t *testing.T) {
		_, err := diffPermissions([]PermissionSpec{
			{Key: "compute.order.get"},
		}, current)
		assert.ErrorContains(t, err, "compute.order.legacy")
		assert.ErrorContains(t, err, "delete: true")
	})

	t.Run("conflicting delete flags for the same permission fail", func(t *testing.T) {
		_, err := diffPermissions([]PermissionSpec{
			{Key: "compute.order.get"},
			{Key: "compute.order.get", Delete: true},
			{Key: "compute.order.legacy"},
		}, current)
		assert.ErrorContains(t, err, "listed both with and without delete")
	})

	t.Run("rejects a key whose namespace part has an underscore or uppercase", func(t *testing.T) {
		// The slug joins service, resource, and verb with "_", so an underscore in a
		// part makes two keys flatten to one slug; uppercase cannot be a SpiceDB
		// object type. Both are rejected so the key stays one-to-one with the slug.
		for _, key := range []string{"resource_order.item.get", "resource.order_item.get", "Compute.order.get", "compute.Order.get"} {
			_, err := diffPermissions([]PermissionSpec{{Key: key}}, nil)
			if assert.Error(t, err, key) {
				assert.ErrorContains(t, err, "key")
			}
		}
	})

	t.Run("accepts valid custom keys", func(t *testing.T) {
		for _, key := range []string{"resource.aoi.get", "user.project.get", "org.user.get", "compute.disk.get"} {
			ops, err := diffPermissions([]PermissionSpec{{Key: key}}, nil)
			assert.NoError(t, err, key)
			assert.Len(t, ops, 1) // a valid new permission plans an add
		}
	})

	t.Run("rejects a key whose flattened slug is too long", func(t *testing.T) {
		// Each part is valid on its own, but the slug service_resource_verb overflows
		// SpiceDB's sixty-four character relation limit, so it must fail the plan
		// rather than pass and then fail when the schema compiles.
		key := strings.Repeat("a", 30) + "." + strings.Repeat("b", 30) + ".get"
		_, err := diffPermissions([]PermissionSpec{{Key: key}}, nil)
		if assert.Error(t, err) {
			assert.ErrorContains(t, err, "too long")
		}
	})

	t.Run("rejects a verb that collides with a generated relation", func(t *testing.T) {
		// The generator adds owner, project, and granted to every custom resource, so
		// a verb equal to one of them would declare the same relation twice and the
		// schema would fail to compile. The plan must reject it up front.
		for _, key := range []string{"compute.order.owner", "compute.order.project", "compute.order.granted"} {
			_, err := diffPermissions([]PermissionSpec{{Key: key}}, nil)
			if assert.Error(t, err, key) {
				assert.ErrorContains(t, err, "reserved")
			}
		}
	})

	t.Run("rejects base-schema keys and bad shapes", func(t *testing.T) {
		for _, key := range []string{
			"app.organization.hack", // base schema
			"compute.get",           // only two parts
			"compute.order.",        // empty verb
			"",                      // empty
		} {
			_, err := diffPermissions([]PermissionSpec{{Key: key}}, nil)
			assert.Error(t, err, key)
		}
	})
}
