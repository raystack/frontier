package reconcile

import (
	"context"
	"testing"

	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
)

func TestDiffRoles(t *testing.T) {
	current := []currentRole{
		{ID: "r1", Name: "compute_manager", Title: "Compute Manager",
			Permissions: []string{"compute_order_get", "compute_order_update"},
			Scopes:      []string{"compute/order"}},
		{ID: "r2", Name: "old_role", Title: "Old", Permissions: []string{"compute_order_get"}},
		{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Owner",
			Permissions: []string{"app_organization_administer"},
			Scopes:      []string{schema.OrganizationNamespace}},
	}

	keepCustom := []RoleSpec{
		{Name: "compute_manager", Permissions: []string{"compute_order_get", "compute_order_update"}},
		{Name: "old_role", Permissions: []string{"compute_order_get"}},
	}

	t.Run("no changes when converged, unlisted predefined roles at defaults", func(t *testing.T) {
		ops, err := diffRoles(keepCustom, current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("an unlisted predefined role that drifted resets to its definition", func(t *testing.T) {
		drifted := append([]currentRole{}, current[:2]...)
		drifted = append(drifted, currentRole{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Renamed Owner",
			Permissions: []string{"app_organization_administer", "app_organization_get"},
			Scopes:      []string{schema.OrganizationNamespace}})

		ops, err := diffRoles(keepCustom, drifted)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (title, permissions; not in file, reset to default)", ops[0].String())
			assert.Equal(t, "r3", ops[0].id)
			assert.Equal(t, "Owner", ops[0].spec.Title)
			assert.Equal(t, []string{"app_organization_administer"}, ops[0].spec.Permissions)
		}
	})

	t.Run("a listed predefined role resets the fields it omits to the definition", func(t *testing.T) {
		drifted := append([]currentRole{}, current[:2]...)
		drifted = append(drifted, currentRole{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Renamed Owner",
			Permissions: []string{"app_organization_administer"},
			Scopes:      []string{schema.OrganizationNamespace}})

		ops, err := diffRoles(append(keepCustom, RoleSpec{
			Name:        schema.RoleOrganizationOwner,
			Permissions: []string{"app_organization_administer", "app_organization_get"},
		}), drifted)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (title, permissions)", ops[0].String())
			assert.Equal(t, "Owner", ops[0].spec.Title) // omitted in the entry: back to the definition, not the server value
			assert.ElementsMatch(t, []string{"app_organization_administer", "app_organization_get"}, ops[0].spec.Permissions)
		}
	})

	t.Run("empty server values converge only when listed", func(t *testing.T) {
		legacy := append([]currentRole{}, current[:2]...)
		legacy = append(legacy, currentRole{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Owner",
			Permissions: []string{"app_organization_administer"}}) // scopes never recorded on this row

		ops, err := diffRoles(keepCustom, legacy)
		assert.NoError(t, err)
		assert.Empty(t, ops) // unlisted: an empty value cannot round-trip through a file, so it is settled

		ops, err = diffRoles(append(keepCustom, RoleSpec{
			Name: schema.RoleOrganizationOwner, Scopes: []string{schema.OrganizationNamespace},
		}), legacy)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (scopes)", ops[0].String())
		}
	})

	t.Run("permission references in any form match slugs", func(t *testing.T) {
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: []string{"compute/order:get", "compute.order.update"}},
			{Name: "old_role", Permissions: []string{"compute_order_get"}},
		}, current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("adds, updates, and deletes in that order", func(t *testing.T) {
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Title: "Compute Admin",
				Permissions: []string{"compute_order_get", "compute_order_update"}},
			{Name: "old_role", Delete: true},
			{Name: "new_role", Title: "New", Permissions: []string{"compute_order_get"}},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 3) {
			assert.Equal(t, "add role new_role", ops[0].String())
			assert.Equal(t, "update role compute_manager (title)", ops[1].String())
			assert.Equal(t, "delete role old_role", ops[2].String())
			assert.Equal(t, "r2", ops[2].id)
		}
	})

	t.Run("update merges managed fields over current values", func(t *testing.T) {
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: []string{"compute_order_get"}}, // narrow the set, no title in spec
			{Name: "old_role", Permissions: []string{"compute_order_get"}},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			op := ops[0]
			assert.Equal(t, opUpdate, op.action)
			assert.Equal(t, "permissions", op.detail)
			assert.Equal(t, "Compute Manager", op.spec.Title) // kept from server
			assert.Equal(t, []string{"compute_order_get"}, op.spec.Permissions)
			assert.Equal(t, []string{"compute/order"}, op.spec.Scopes) // kept from server
		}
	})

	t.Run("a custom role's description is managed like its title", func(t *testing.T) {
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Description: "Runs compute orders",
				Permissions: []string{"compute_order_get", "compute_order_update"}},
			{Name: "old_role", Permissions: []string{"compute_order_get"}},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role compute_manager (description)", ops[0].String())
			assert.Equal(t, "Runs compute orders", ops[0].spec.Description)
		}
	})

	t.Run("an unlisted predefined role's drifted description resets to the definition", func(t *testing.T) {
		drifted := append([]currentRole{}, current[:2]...)
		drifted = append(drifted, currentRole{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Owner",
			Description: "hand-edited note",
			Permissions: []string{"app_organization_administer"},
			Scopes:      []string{schema.OrganizationNamespace}})

		ops, err := diffRoles(keepCustom, drifted)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (description; not in file, reset to default)", ops[0].String())
			assert.Equal(t, "", ops[0].spec.Description) // the definition sets no description
		}
	})

	t.Run("predefined role title and permissions can be managed", func(t *testing.T) {
		specs := append(keepCustom, RoleSpec{
			Name:        schema.RoleOrganizationOwner,
			Title:       "Workspace Owner",
			Permissions: []string{"app_organization_administer", "app_organization_get"},
		})
		ops, err := diffRoles(specs, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (title, permissions)", ops[0].String())
			assert.Equal(t, "r3", ops[0].id)
		}
	})

	t.Run("deleting a predefined role is rejected", func(t *testing.T) {
		_, err := diffRoles(append(keepCustom, RoleSpec{Name: schema.RoleOrganizationOwner, Delete: true}), current)
		assert.ErrorContains(t, err, "predefined role")
	})

	t.Run("a listed predefined role missing on the server fails the plan", func(t *testing.T) {
		_, err := diffRoles(append(keepCustom, RoleSpec{Name: schema.GroupOwnerRole, Title: "X"}), current)
		assert.ErrorContains(t, err, "not found on the server")
	})

	t.Run("a custom server role missing from the file fails the plan", func(t *testing.T) {
		_, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: []string{"compute_order_get", "compute_order_update"}},
		}, current)
		assert.ErrorContains(t, err, "old_role")
		assert.ErrorContains(t, err, "delete: true")
	})

	t.Run("a custom role must list at least one permission", func(t *testing.T) {
		_, err := diffRoles([]RoleSpec{{Name: "compute_manager", Title: "X"}}, current)
		assert.ErrorContains(t, err, "must list at least one permission")

		_, err = diffRoles([]RoleSpec{{Name: "compute_manager", Permissions: []string{}}}, current)
		assert.ErrorContains(t, err, "must list at least one permission")
	})

	t.Run("emptying a predefined role's permissions or scopes is rejected", func(t *testing.T) {
		_, err := diffRoles(append(keepCustom, RoleSpec{
			Name: schema.RoleOrganizationOwner, Permissions: []string{},
		}), current)
		assert.ErrorContains(t, err, "permissions to an empty list")

		_, err = diffRoles(append(keepCustom, RoleSpec{
			Name: schema.RoleOrganizationOwner, Scopes: []string{},
		}), current)
		assert.ErrorContains(t, err, "scopes to an empty list")
	})

	t.Run("duplicate names fail", func(t *testing.T) {
		_, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: []string{"a"}},
			{Name: "compute_manager", Permissions: []string{"b"}},
		}, current)
		assert.ErrorContains(t, err, "listed more than once")
	})

	t.Run("delete of an absent custom role is a no-op", func(t *testing.T) {
		ops, err := diffRoles(append(keepCustom, RoleSpec{Name: "already_gone", Delete: true}), current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
}

func TestRun_RoleBeforePermissionFailsFast(t *testing.T) {
	roleAPI := &fakeRoleAPI{}
	permAPI := &fakePermissionAPI{}
	registry := map[string]Reconciler{
		KindRole:       NewRoleReconciler(roleAPI, ""),
		KindPermission: NewPermissionReconciler(permAPI, ""),
	}
	data := []byte("kind: Role\nspec: []\n---\nkind: Permission\nspec: []\n")

	_, err := Run(context.Background(), registry, data, false)

	assert.ErrorContains(t, err, `kind "Permission" must come before kind "Role"`)
	assert.Empty(t, roleAPI.created) // nothing dispatched
	assert.Empty(t, permAPI.created)

	// the right order passes the check
	_, err = Run(context.Background(), registry, []byte("kind: Permission\nspec: []\n---\nkind: Role\nspec: []\n"), true)
	assert.NoError(t, err)
}
