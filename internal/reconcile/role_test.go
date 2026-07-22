package reconcile

import (
	"context"
	"testing"

	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
)

// ptr and slicePtr build the presence-tracking pointers a RoleSpec now uses:
// a nil pointer is an omitted field, a non-nil one is a listed value.
func ptr[T any](v T) *T { return &v }

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

	// keepCustom lists every non-empty field of the two custom roles, since the
	// file is the whole desired state: an omitted field defaults to empty and
	// would clear the server value. This is what an export writes.
	keepCustom := []RoleSpec{
		{Name: "compute_manager", Title: ptr("Compute Manager"),
			Permissions: ptr([]string{"compute_order_get", "compute_order_update"}), Scopes: ptr([]string{"compute/order"})},
		{Name: "old_role", Title: ptr("Old"), Permissions: ptr([]string{"compute_order_get"})},
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
			assert.Equal(t, "Owner", ops[0].fields.Title)
			assert.Equal(t, []string{"app_organization_administer"}, ops[0].fields.Permissions)
		}
	})

	t.Run("a listed predefined role resets the fields it omits to the definition", func(t *testing.T) {
		drifted := append([]currentRole{}, current[:2]...)
		drifted = append(drifted, currentRole{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Renamed Owner",
			Permissions: []string{"app_organization_administer"},
			Scopes:      []string{schema.OrganizationNamespace}})

		ops, err := diffRoles(append(keepCustom, RoleSpec{
			Name:        schema.RoleOrganizationOwner,
			Permissions: ptr([]string{"app_organization_administer", "app_organization_get"}),
		}), drifted)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (title, permissions)", ops[0].String())
			assert.Equal(t, "Owner", ops[0].fields.Title) // omitted in the entry: back to the definition, not the server value
			assert.ElementsMatch(t, []string{"app_organization_administer", "app_organization_get"}, ops[0].fields.Permissions)
		}
	})

	t.Run("an unlisted predefined role with an empty legacy field resets to the definition", func(t *testing.T) {
		// A row from before the scopes column: empty scopes on the server, non-empty
		// in the definition. Unlisted, it now converges to the definition like any
		// other drift, instead of being left alone forever. Export writes the empty
		// value explicitly (see the reconciler round-trip tests), so a role that
		// should keep an empty field stays listed and never reaches this branch.
		legacy := append([]currentRole{}, current[:2]...)
		legacy = append(legacy, currentRole{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Owner",
			Permissions: []string{"app_organization_administer"}}) // scopes never recorded on this row

		ops, err := diffRoles(keepCustom, legacy)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (scopes; not in file, reset to default)", ops[0].String())
			assert.Equal(t, []string{schema.OrganizationNamespace}, ops[0].fields.Scopes)
		}
	})

	t.Run("a listed predefined role sets an empty server field", func(t *testing.T) {
		legacy := append([]currentRole{}, current[:2]...)
		legacy = append(legacy, currentRole{ID: "r3", Name: schema.RoleOrganizationOwner, Title: "Owner",
			Permissions: []string{"app_organization_administer"}}) // empty scopes

		ops, err := diffRoles(append(keepCustom, RoleSpec{
			Name: schema.RoleOrganizationOwner, Scopes: ptr([]string{schema.OrganizationNamespace}),
		}), legacy)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (scopes)", ops[0].String())
		}
	})

	t.Run("an explicit empty list clears a field", func(t *testing.T) {
		// `scopes: []` in the file is a listed value, not an omission, so it
		// overrides the definition's scopes with an empty set.
		ops, err := diffRoles(append(keepCustom, RoleSpec{
			Name: schema.RoleOrganizationOwner, Scopes: ptr([]string{}),
		}), current)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role app_organization_owner (scopes)", ops[0].String())
			assert.Empty(t, ops[0].fields.Scopes)
		}
	})

	t.Run("permission references in any form match slugs", func(t *testing.T) {
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Title: ptr("Compute Manager"),
				Permissions: ptr([]string{"compute/order:get", "compute.order.update"}), Scopes: ptr([]string{"compute/order"})},
			{Name: "old_role", Title: ptr("Old"), Permissions: ptr([]string{"compute_order_get"})},
		}, current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("adds, updates, and deletes in that order", func(t *testing.T) {
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Title: ptr("Compute Admin"),
				Permissions: ptr([]string{"compute_order_get", "compute_order_update"}), Scopes: ptr([]string{"compute/order"})},
			{Name: "old_role", Delete: true},
			{Name: "new_role", Title: ptr("New"), Permissions: ptr([]string{"compute_order_get"})},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 3) {
			assert.Equal(t, "add role new_role", ops[0].String())
			assert.Equal(t, "update role compute_manager (title)", ops[1].String())
			assert.Equal(t, "delete role old_role", ops[2].String())
			assert.Equal(t, "r2", ops[2].id)
		}
	})

	t.Run("a listed field is the whole desired value; omitted custom fields clear", func(t *testing.T) {
		// Under one field model, a custom role's omitted fields default to empty,
		// they are not kept from the server. So listing only permissions here also
		// clears the title and scopes, because the file is the whole desired state.
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: ptr([]string{"compute_order_get"})}, // narrow perms, title and scopes omitted
			{Name: "old_role", Title: ptr("Old"), Permissions: ptr([]string{"compute_order_get"})},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			op := ops[0]
			assert.Equal(t, opUpdate, op.action)
			assert.Equal(t, "title, permissions, scopes", op.detail)
			assert.Equal(t, "", op.fields.Title) // omitted: custom default is empty
			assert.Equal(t, []string{"compute_order_get"}, op.fields.Permissions)
			assert.Empty(t, op.fields.Scopes) // omitted: custom default is empty
		}
	})

	t.Run("a custom role keeps its fields when the file lists them", func(t *testing.T) {
		// To keep a custom role's title and scopes, the file lists them, since the
		// file is the whole desired state. Export writes them all out for exactly
		// this reason.
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Title: ptr("Compute Manager"),
				Permissions: ptr([]string{"compute_order_get"}), Scopes: ptr([]string{"compute/order"})},
			{Name: "old_role", Title: ptr("Old"), Permissions: ptr([]string{"compute_order_get"})},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			op := ops[0]
			assert.Equal(t, "permissions", op.detail) // only the narrowed set changed
			assert.Equal(t, "Compute Manager", op.fields.Title)
			assert.Equal(t, []string{"compute_order_get"}, op.fields.Permissions)
			assert.Equal(t, []string{"compute/order"}, op.fields.Scopes)
		}
	})

	t.Run("a custom role's description is managed like its title", func(t *testing.T) {
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Title: ptr("Compute Manager"), Description: ptr("Runs compute orders"),
				Permissions: ptr([]string{"compute_order_get", "compute_order_update"}), Scopes: ptr([]string{"compute/order"})},
			{Name: "old_role", Title: ptr("Old"), Permissions: ptr([]string{"compute_order_get"})},
		}, current)

		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role compute_manager (description)", ops[0].String())
			assert.Equal(t, "Runs compute orders", ops[0].fields.Description)
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
			assert.Equal(t, "", ops[0].fields.Description) // the definition sets no description
		}
	})

	t.Run("predefined role title and permissions can be managed", func(t *testing.T) {
		specs := append(keepCustom, RoleSpec{
			Name:        schema.RoleOrganizationOwner,
			Title:       ptr("Workspace Owner"),
			Permissions: ptr([]string{"app_organization_administer", "app_organization_get"}),
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
		_, err := diffRoles(append(keepCustom, RoleSpec{Name: schema.GroupOwnerRole, Title: ptr("X")}), current)
		assert.ErrorContains(t, err, "not found on the server")
	})

	t.Run("a custom server role missing from the file fails the plan", func(t *testing.T) {
		_, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: ptr([]string{"compute_order_get", "compute_order_update"})},
		}, current)
		assert.ErrorContains(t, err, "old_role")
		assert.ErrorContains(t, err, "delete: true")
	})

	t.Run("an omitted custom title clears it", func(t *testing.T) {
		// A custom role defaults to empty fields, so omitting the title means the
		// desired title is empty. A server role that still has a title drifts and
		// the title is cleared.
		ops, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: ptr([]string{"compute_order_get", "compute_order_update"}), Scopes: ptr([]string{"compute/order"})}, // title omitted
			{Name: "old_role", Title: ptr("Old"), Permissions: ptr([]string{"compute_order_get"})},
		}, current)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update role compute_manager (title)", ops[0].String())
			assert.Equal(t, "", ops[0].fields.Title)
		}
	})

	t.Run("duplicate names fail", func(t *testing.T) {
		_, err := diffRoles([]RoleSpec{
			{Name: "compute_manager", Permissions: ptr([]string{"a"})},
			{Name: "compute_manager", Permissions: ptr([]string{"b"})},
		}, current)
		assert.ErrorContains(t, err, "listed more than once")
	})

	t.Run("delete of an absent custom role is a no-op", func(t *testing.T) {
		ops, err := diffRoles(append(keepCustom, RoleSpec{Name: "already_gone", Delete: true}), current)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("a custom role that resolves to no permissions fails the plan", func(t *testing.T) {
		// The server rejects a role with no permissions, so a spec resolving to an
		// empty set is a plan that can never apply. It must fail at plan time. A
		// custom role defaults to empty, so both omitting permissions and writing
		// an explicit empty list resolve to empty.
		_, omitted := diffRoles([]RoleSpec{{Name: "empty_custom"}}, nil)
		assert.ErrorContains(t, omitted, "at least one permission")

		_, explicit := diffRoles([]RoleSpec{{Name: "empty_custom", Permissions: ptr([]string{})}}, nil)
		assert.ErrorContains(t, explicit, "at least one permission")
	})

	t.Run("a predefined role set to empty permissions fails the plan", func(t *testing.T) {
		cur := []currentRole{{ID: "r1", Name: schema.RoleOrganizationViewer, Permissions: []string{"app_organization_get"}}}
		_, err := diffRoles([]RoleSpec{{Name: schema.RoleOrganizationViewer, Permissions: ptr([]string{})}}, cur)
		assert.ErrorContains(t, err, "at least one permission")
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
