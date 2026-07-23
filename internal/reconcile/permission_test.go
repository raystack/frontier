package reconcile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffPermissions(t *testing.T) {
	current := []currentPermission{
		{ID: "p1", Namespace: "compute/order", Name: "get"},
		{ID: "p2", Namespace: "compute/order", Name: "legacy"},
	}

	t.Run("adds missing and deletes flagged, adds first", func(t *testing.T) {
		ops, err := diffPermissions([]PermissionSpec{
			{Namespace: "compute/order", Name: "get"},
			{Namespace: "compute/order", Name: "legacy", Delete: true},
			{Namespace: "compute/disk", Name: "mount"},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 2) {
			assert.Equal(t, "add permission compute/disk:mount", ops[0].String())
			assert.Equal(t, "delete permission compute/order:legacy", ops[1].String())
			assert.Equal(t, "p2", ops[1].id)
		}
	})

	t.Run("no changes when converged", func(t *testing.T) {
		ops, err := diffPermissions([]PermissionSpec{
			{Namespace: "compute/order", Name: "get"},
			{Namespace: "compute/order", Name: "legacy"},
		}, current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("delete of an absent permission is a no-op", func(t *testing.T) {
		ops, err := diffPermissions([]PermissionSpec{
			{Namespace: "compute/order", Name: "get"},
			{Namespace: "compute/order", Name: "legacy"},
			{Namespace: "compute/order", Name: "gone", Delete: true},
		}, current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("a server permission missing from the file fails the plan", func(t *testing.T) {
		_, err := diffPermissions([]PermissionSpec{
			{Namespace: "compute/order", Name: "get"},
		}, current)
		assert.ErrorContains(t, err, "compute/order:legacy")
		assert.ErrorContains(t, err, "delete: true")
	})

	t.Run("conflicting delete flags for the same permission fail", func(t *testing.T) {
		_, err := diffPermissions([]PermissionSpec{
			{Namespace: "compute/order", Name: "get"},
			{Namespace: "compute/order", Name: "get", Delete: true},
			{Namespace: "compute/order", Name: "legacy"},
		}, current)
		assert.ErrorContains(t, err, "listed both with and without delete")
	})

	t.Run("a file namespace that would collide with a server slug is rejected, not absorbed", func(t *testing.T) {
		// The server stores resource/order_item; the file lists a genuinely different
		// namespace resource_order/item that flattens to the same slug. The diff used
		// to treat it as already present and plan zero ops (the rule 2 gap). The
		// ambiguous namespace is now rejected at validation, so it cannot be absorbed.
		server := []currentPermission{{ID: "p1", Namespace: "resource/order_item", Name: "get"}}
		_, err := diffPermissions([]PermissionSpec{{Namespace: "resource_order/item", Name: "get"}}, server)
		assert.ErrorContains(t, err, "resource_order/item")
	})

	t.Run("rejects a namespace with an underscore or uppercase in a part", func(t *testing.T) {
		// The slug joins service, resource, and verb with "_", so an underscore in a
		// part makes two namespaces flatten to one slug; uppercase cannot be a
		// SpiceDB object type. Both are rejected so the slug stays one-to-one.
		for _, ns := range []string{"resource_order/item", "resource/order_item", "Compute/order", "compute/Order"} {
			_, err := diffPermissions([]PermissionSpec{{Namespace: ns, Name: "get"}}, nil)
			if assert.Error(t, err, ns) {
				assert.ErrorContains(t, err, "namespace")
			}
		}
	})

	t.Run("accepts valid custom namespaces", func(t *testing.T) {
		for _, ns := range []string{"resource/aoi", "user/project", "org/user", "compute/disk"} {
			ops, err := diffPermissions([]PermissionSpec{{Namespace: ns, Name: "get"}}, nil)
			assert.NoError(t, err, ns)
			assert.Len(t, ops, 1) // a valid new permission plans an add
		}
	})

	t.Run("rejects base-schema namespaces and bad shapes", func(t *testing.T) {
		for _, s := range []PermissionSpec{
			{Namespace: "app/organization", Name: "hack"},
			{Namespace: "app", Name: "hack"},
			{Namespace: "compute", Name: "get"},
			{Namespace: "compute/order", Name: "not-alnum"},
			{Namespace: "compute/order", Name: ""},
		} {
			_, err := diffPermissions([]PermissionSpec{s}, nil)
			assert.Error(t, err, "%+v", s)
		}
	})
}
