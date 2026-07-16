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

	t.Run("distinct entries that flatten to the same slug fail", func(t *testing.T) {
		_, err := diffPermissions([]PermissionSpec{
			{Namespace: "compute/order", Name: "get"},
			{Namespace: "compute/order", Name: "legacy"},
			{Namespace: "resource/order_item", Name: "get"},
			{Namespace: "resource_order/item", Name: "get"},
		}, current)
		assert.ErrorContains(t, err, `collide on the same slug "resource_order_item_get"`)
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
