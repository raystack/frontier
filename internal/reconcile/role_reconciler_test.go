package reconcile

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeRoleAPI struct {
	roles     []*frontierv1beta1.Role
	created   []*frontierv1beta1.RoleRequestBody
	updated   map[string]*frontierv1beta1.RoleRequestBody
	deleted   []string
	createErr error
	updateErr error
}

func (f *fakeRoleAPI) ListRoles(_ context.Context, _ *connect.Request[frontierv1beta1.ListRolesRequest]) (*connect.Response[frontierv1beta1.ListRolesResponse], error) {
	return connect.NewResponse(&frontierv1beta1.ListRolesResponse{Roles: f.roles}), nil
}

func (f *fakeRoleAPI) CreateRole(_ context.Context, req *connect.Request[frontierv1beta1.CreateRoleRequest]) (*connect.Response[frontierv1beta1.CreateRoleResponse], error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = append(f.created, req.Msg.GetBody())
	return connect.NewResponse(&frontierv1beta1.CreateRoleResponse{}), nil
}

func (f *fakeRoleAPI) UpdateRole(_ context.Context, req *connect.Request[frontierv1beta1.UpdateRoleRequest]) (*connect.Response[frontierv1beta1.UpdateRoleResponse], error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	if f.updated == nil {
		f.updated = map[string]*frontierv1beta1.RoleRequestBody{}
	}
	f.updated[req.Msg.GetId()] = req.Msg.GetBody()
	return connect.NewResponse(&frontierv1beta1.UpdateRoleResponse{}), nil
}

func (f *fakeRoleAPI) DeleteRole(_ context.Context, req *connect.Request[frontierv1beta1.DeleteRoleRequest]) (*connect.Response[frontierv1beta1.DeleteRoleResponse], error) {
	f.deleted = append(f.deleted, req.Msg.GetId())
	return connect.NewResponse(&frontierv1beta1.DeleteRoleResponse{}), nil
}

func rolePB(id, name, title string, permissions []string, scopes ...string) *frontierv1beta1.Role {
	return &frontierv1beta1.Role{Id: id, Name: name, Title: title, Permissions: permissions, Scopes: scopes}
}

func rolePBWithDesc(id, name, title, description string, permissions []string, scopes ...string) *frontierv1beta1.Role {
	r := rolePB(id, name, title, permissions, scopes...)
	md, _ := structpb.NewStruct(map[string]any{descriptionKey: description})
	r.Metadata = md
	return r
}

func ownerDefaultPB() *frontierv1beta1.Role {
	return rolePB("r-owner", schema.RoleOrganizationOwner, "Owner",
		[]string{"app_organization_administer"}, schema.OrganizationNamespace)
}

func TestRoleReconciler(t *testing.T) {
	t.Run("applies add, update, and delete with the managed-by marker", func(t *testing.T) {
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			ownerDefaultPB(),
			rolePB("r1", "compute_manager", "Compute Manager", []string{"compute_order_get"}),
			rolePB("r2", "old_role", "Old", []string{"compute_order_get"}),
		}}
		spec := []byte(`
- {name: compute_manager, title: Compute Admin, permissions: [compute_order_get]}
- {name: old_role, delete: true}
- {name: new_role, title: New, permissions: [compute_order_get]}
`)
		rep, err := NewRoleReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, 3, rep.Applied)
		if assert.Len(t, api.created, 1) {
			assert.Equal(t, "new_role", api.created[0].GetName())
			assert.Equal(t, managedByValue, api.created[0].GetMetadata().GetFields()[managedByKey].GetStringValue())
		}
		if assert.Len(t, api.updated, 1) {
			body := api.updated["r1"]
			if assert.NotNil(t, body) {
				assert.Equal(t, "compute_manager", body.GetName()) // identity never changes
				assert.Equal(t, "Compute Admin", body.GetTitle())
				assert.Equal(t, []string{"compute_order_get"}, body.GetPermissions())
				assert.Equal(t, managedByValue, body.GetMetadata().GetFields()[managedByKey].GetStringValue())
			}
		}
		assert.Equal(t, []string{"r2"}, api.deleted)
	})

	t.Run("an update keeps metadata keys the reconciler does not manage", func(t *testing.T) {
		md, _ := structpb.NewStruct(map[string]any{"team": "compute", descriptionKey: "old", managedByKey: managedByValue})
		role := rolePB("r1", "compute_manager", "Compute Manager", []string{"compute_order_get"})
		role.Metadata = md
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{ownerDefaultPB(), role}}
		// only permissions change; the file lists the fields it keeps, and the
		// unmanaged team key must survive the update either way.
		spec := []byte("- {name: compute_manager, title: Compute Manager, description: old, permissions: [compute_order_get, compute_order_update]}\n")

		rep, err := NewRoleReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, 1, rep.Applied)
		if body := api.updated["r1"]; assert.NotNil(t, body) {
			f := body.GetMetadata().GetFields()
			assert.Equal(t, "compute", f["team"].GetStringValue())            // preserved, not dropped
			assert.Equal(t, managedByValue, f[managedByKey].GetStringValue()) // still stamped
			assert.Equal(t, "old", f[descriptionKey].GetStringValue())        // kept because the file lists it
		}
	})

	t.Run("an unknown field in the spec fails the plan", func(t *testing.T) {
		api := &fakeRoleAPI{}
		spec := []byte("- {name: new_role, permissions: [compute_order_get], delet: true}\n")

		_, err := NewRoleReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.ErrorContains(t, err, "parse Role spec")
		assert.Empty(t, api.created)
	})

	t.Run("updates a predefined role's title and permissions by id", func(t *testing.T) {
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{ownerDefaultPB()}}
		spec := []byte(`
- name: app_organization_owner
  title: Workspace Owner
  permissions: [app_organization_administer, app_organization_get]
`)
		rep, err := NewRoleReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, 1, rep.Applied)
		body := api.updated["r-owner"]
		if assert.NotNil(t, body) {
			assert.Equal(t, schema.RoleOrganizationOwner, body.GetName())
			assert.Equal(t, "Workspace Owner", body.GetTitle())
			assert.ElementsMatch(t, []string{"app_organization_administer", "app_organization_get"}, body.GetPermissions())
			assert.Equal(t, []string{schema.OrganizationNamespace}, body.GetScopes()) // omitted in the entry: the definition's value
		}
	})

	t.Run("an unlisted drifted predefined role resets to its definition", func(t *testing.T) {
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			rolePB("r-owner", schema.RoleOrganizationOwner, "Renamed Owner",
				[]string{"app_organization_administer", "app_organization_get"}, schema.OrganizationNamespace),
		}}

		rep, err := NewRoleReconciler(api, "").Reconcile(context.Background(), []byte("[]\n"), false)

		assert.NoError(t, err)
		assert.Equal(t, []string{"update role app_organization_owner (title, permissions; not in file, reset to default)"}, rep.Planned)
		assert.Equal(t, 1, rep.Applied)
		body := api.updated["r-owner"]
		if assert.NotNil(t, body) {
			assert.Equal(t, "Owner", body.GetTitle())
			assert.Equal(t, []string{"app_organization_administer"}, body.GetPermissions())
		}
	})

	t.Run("a predefined role with an empty legacy field round-trips", func(t *testing.T) {
		// A row from before the scopes column: empty scopes on the server, non-empty
		// in the definition. Presence-tracking pointers let a file hold an empty
		// value, so export writes `scopes: []` and reconcile keeps it as-is instead
		// of fighting the definition forever.
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			rolePB("r-owner", schema.RoleOrganizationOwner, "Owner", []string{"app_organization_administer"}),
		}}
		registry := map[string]Reconciler{KindRole: NewRoleReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindRole)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "scopes: []") // the empty value written explicitly

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("a custom role with no permissions round-trips", func(t *testing.T) {
		// A custom role whose permissions were emptied. Export writes an entry so
		// the role is kept, and reconciling that entry plans no changes.
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			ownerDefaultPB(),
			rolePB("r1", "compute_manager", "Compute Manager", nil), // no permissions
		}}
		registry := map[string]Reconciler{KindRole: NewRoleReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindRole)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "compute_manager") // the custom role is kept in the file

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("applying a role with a description writes it to metadata", func(t *testing.T) {
		api := &fakeRoleAPI{}
		spec := []byte("- {name: new_role, description: Runs compute orders, permissions: [compute_order_get]}\n")

		rep, err := NewRoleReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, 1, rep.Applied)
		if assert.Len(t, api.created, 1) {
			md := api.created[0].GetMetadata().GetFields()
			assert.Equal(t, "Runs compute orders", md[descriptionKey].GetStringValue())
			assert.Equal(t, managedByValue, md[managedByKey].GetStringValue())
		}
	})

	t.Run("a custom role's description round-trips", func(t *testing.T) {
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			ownerDefaultPB(),
			rolePBWithDesc("r1", "compute_manager", "Compute Manager", "Runs compute orders", []string{"compute_order_get"}),
		}}
		registry := map[string]Reconciler{KindRole: NewRoleReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindRole)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "Runs compute orders")

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("reconciling an exported document plans no changes", func(t *testing.T) {
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			ownerDefaultPB(), // matches defaults: not exported, skipped when unlisted
			rolePB("r-viewer", "app_organization_viewer", "Everyone", []string{"app_organization_get"}), // retitled predefined
			rolePB("r1", "compute_manager", "Compute Manager", []string{"compute_order_update", "compute_order_get"}, "compute/order"),
		}}
		registry := map[string]Reconciler{KindRole: NewRoleReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindRole)
		assert.NoError(t, err)
		assert.NotContains(t, string(out), schema.RoleOrganizationOwner)
		assert.Contains(t, string(out), "Everyone")

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("a role referencing a missing permission fails at the API call", func(t *testing.T) {
		// Kinds are independent: the reconciler does not resolve permission
		// references itself. The server rejects them, and that failure must
		// surface with the failing op named and the run stopped.
		api := &fakeRoleAPI{
			createErr: connect.NewError(connect.CodeInvalidArgument,
				errors.New("compute_order_missing: permission doesn't exist")),
		}
		spec := []byte("- {name: new_role, permissions: [compute_order_missing]}\n")

		rep, err := NewRoleReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.ErrorContains(t, err, "apply [add role new_role]")
		assert.ErrorContains(t, err, "permission doesn't exist")
		assert.Zero(t, rep.Applied)
		assert.Equal(t, []string{"add role new_role"}, rep.Planned) // the plan survives for the report
	})

	t.Run("export includes a predefined role whose permissions were narrowed", func(t *testing.T) {
		// default title, changed permission set: exactly the review scenario.
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			rolePB("r-owner", schema.RoleOrganizationOwner, "Owner",
				[]string{"app_organization_get", "app_organization_update", "app_organization_policymanage"},
				schema.OrganizationNamespace),
		}}
		registry := map[string]Reconciler{KindRole: NewRoleReconciler(api, "")}

		spec, err := NewRoleReconciler(api, "").Export(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []RoleSpec{{
			Name:        schema.RoleOrganizationOwner,
			Permissions: ptr([]string{"app_organization_get", "app_organization_policymanage", "app_organization_update"}),
		}}, spec)

		out, err := Export(context.Background(), registry, KindRole)
		assert.NoError(t, err)
		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("export includes a predefined role whose scopes changed", func(t *testing.T) {
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{
			rolePB("r-owner", schema.RoleOrganizationOwner, "Owner",
				[]string{"app_organization_administer"}, schema.OrganizationNamespace, "compute/order"),
		}}
		registry := map[string]Reconciler{KindRole: NewRoleReconciler(api, "")}

		spec, err := NewRoleReconciler(api, "").Export(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []RoleSpec{{
			Name:   schema.RoleOrganizationOwner,
			Scopes: ptr([]string{schema.OrganizationNamespace, "compute/order"}),
		}}, spec)

		out, err := Export(context.Background(), registry, KindRole)
		assert.NoError(t, err)
		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("export with only default predefined roles yields an empty list", func(t *testing.T) {
		api := &fakeRoleAPI{roles: []*frontierv1beta1.Role{ownerDefaultPB()}}
		registry := map[string]Reconciler{KindRole: NewRoleReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindRole)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "spec: []")
	})
}
