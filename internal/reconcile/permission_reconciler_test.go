package reconcile

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
)

type fakePermissionAPI struct {
	perms   []*frontierv1beta1.Permission
	created []*frontierv1beta1.CreatePermissionRequest
	deleted []string
}

func (f *fakePermissionAPI) ListPermissions(_ context.Context, _ *connect.Request[frontierv1beta1.ListPermissionsRequest]) (*connect.Response[frontierv1beta1.ListPermissionsResponse], error) {
	return connect.NewResponse(&frontierv1beta1.ListPermissionsResponse{Permissions: f.perms}), nil
}

func (f *fakePermissionAPI) CreatePermission(_ context.Context, req *connect.Request[frontierv1beta1.CreatePermissionRequest]) (*connect.Response[frontierv1beta1.CreatePermissionResponse], error) {
	f.created = append(f.created, req.Msg)
	return connect.NewResponse(&frontierv1beta1.CreatePermissionResponse{}), nil
}

func (f *fakePermissionAPI) DeletePermission(_ context.Context, req *connect.Request[frontierv1beta1.DeletePermissionRequest]) (*connect.Response[frontierv1beta1.DeletePermissionResponse], error) {
	f.deleted = append(f.deleted, req.Msg.GetId())
	return connect.NewResponse(&frontierv1beta1.DeletePermissionResponse{}), nil
}

func permissionPB(id, namespace, name string) *frontierv1beta1.Permission {
	return &frontierv1beta1.Permission{Id: id, Namespace: namespace, Name: name}
}

func TestPermissionReconciler(t *testing.T) {
	t.Run("applies adds and deletes, base namespaces invisible", func(t *testing.T) {
		api := &fakePermissionAPI{perms: []*frontierv1beta1.Permission{
			permissionPB("b1", "app/organization", "administer"), // base: must be ignored
			permissionPB("p1", "compute/order", "legacy"),
		}}
		spec := []byte("- {namespace: compute/order, name: legacy, delete: true}\n- {namespace: compute/order, name: get}\n")

		rep, err := NewPermissionReconciler(api, "").Reconcile(context.Background(), spec, false)

		assert.NoError(t, err)
		assert.Equal(t, 2, rep.Applied)
		if assert.Len(t, api.created, 1) {
			bodies := api.created[0].GetBodies()
			if assert.Len(t, bodies, 1) {
				assert.Equal(t, "compute.order.get", bodies[0].GetKey())
			}
		}
		assert.Equal(t, []string{"p1"}, api.deleted)
	})

	t.Run("dry-run plans without applying", func(t *testing.T) {
		api := &fakePermissionAPI{}
		spec := []byte("- {namespace: compute/order, name: get}\n")

		rep, err := NewPermissionReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.NoError(t, err)
		assert.Equal(t, []string{"add permission compute/order:get"}, rep.Planned)
		assert.Zero(t, rep.Applied)
		assert.Empty(t, api.created)
	})

	t.Run("an unknown field in the spec fails the plan", func(t *testing.T) {
		api := &fakePermissionAPI{}
		// `delet` instead of `delete`: must fail, not silently ignore the delete
		spec := []byte("- {namespace: compute/order, name: get, delet: true}\n")

		_, err := NewPermissionReconciler(api, "").Reconcile(context.Background(), spec, true)

		assert.ErrorContains(t, err, "parse Permission spec")
		assert.Empty(t, api.created)
	})

	t.Run("reconciling an exported document plans no changes", func(t *testing.T) {
		api := &fakePermissionAPI{perms: []*frontierv1beta1.Permission{
			permissionPB("b1", "app/project", "get"),
			permissionPB("p2", "compute/order", "update"),
			permissionPB("p1", "compute/disk", "mount"),
		}}
		registry := map[string]Reconciler{KindPermission: NewPermissionReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindPermission)
		assert.NoError(t, err)
		assert.NotContains(t, string(out), "app/project")

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("export of no custom permissions yields an empty list", func(t *testing.T) {
		api := &fakePermissionAPI{perms: []*frontierv1beta1.Permission{
			permissionPB("b1", "app/organization", "administer"),
		}}
		registry := map[string]Reconciler{KindPermission: NewPermissionReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindPermission)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "spec: []")
	})
}
