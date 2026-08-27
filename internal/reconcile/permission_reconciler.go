package reconcile

import (
	"context"
	"fmt"
	"sort"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

// PermissionAPI is the API subset the permission reconciler needs. Reads live
// on FrontierService, writes on AdminService; the caller provides one value
// that serves both.
type PermissionAPI interface {
	ListPermissions(context.Context, *connect.Request[frontierv1beta1.ListPermissionsRequest]) (*connect.Response[frontierv1beta1.ListPermissionsResponse], error)
	CreatePermission(context.Context, *connect.Request[frontierv1beta1.CreatePermissionRequest]) (*connect.Response[frontierv1beta1.CreatePermissionResponse], error)
	DeletePermission(context.Context, *connect.Request[frontierv1beta1.DeletePermissionRequest]) (*connect.Response[frontierv1beta1.DeletePermissionResponse], error)
}

// PermissionReconciler makes custom permissions match the desired spec.
// Base-schema permissions (app namespaces) are server-managed and ignored.
type PermissionReconciler struct {
	client PermissionAPI
	header string
}

func NewPermissionReconciler(client PermissionAPI, header string) *PermissionReconciler {
	return &PermissionReconciler{client: client, header: header}
}

func (r *PermissionReconciler) Kind() string { return KindPermission }

// Validate checks every entry without touching the server, so a bad entry stops
// the whole file before anything applies.
func (r *PermissionReconciler) Validate(spec []byte) error {
	var specs []PermissionSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return fmt.Errorf("parse %s spec: %w", KindPermission, err)
	}
	seen := map[string]PermissionSpec{}
	for _, s := range specs {
		if err := validatePermissionSpec(s); err != nil {
			return fmt.Errorf("invalid permission spec %s: %w", s, err)
		}
		if prev, dup := seen[s.Key]; dup {
			if prev.Delete != s.Delete {
				return fmt.Errorf("permission %s is listed both with and without delete", s)
			}
		}
		seen[s.Key] = s
	}
	return nil
}

func (r *PermissionReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []PermissionSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindPermission, err)
	}

	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffPermissions(specs, current)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindPermission, DryRun: dryRun}
	for _, op := range ops {
		rep.Planned = append(rep.Planned, op.String())
	}
	if dryRun {
		return rep, nil
	}
	for _, op := range ops {
		if err := r.apply(ctx, op); err != nil {
			return rep, fmt.Errorf("apply [%s]: %w", op, err)
		}
		rep.Applied++
	}
	return rep, nil
}

// Export returns the current custom permissions as a desired-state spec,
// sorted so repeated exports produce identical files.
func (r *PermissionReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	specs := make([]PermissionSpec, 0, len(current))
	for _, c := range current {
		specs = append(specs, PermissionSpec{Key: c.Key})
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Key < specs[j].Key })
	return specs, nil
}

func (r *PermissionReconciler) fetchCurrent(ctx context.Context) ([]currentPermission, error) {
	resp, err := r.client.ListPermissions(ctx, authReq(&frontierv1beta1.ListPermissionsRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	var current []currentPermission
	for _, p := range resp.Msg.GetPermissions() {
		ns, name := schema.PermissionNamespaceAndNameFromKey(p.GetKey())
		// A key that does not split into service.resource.verb is not a custom
		// permission this reconciler manages: it is a base or system permission, or
		// comes from an older server that does not set the key. Skip it instead of
		// failing the whole list. Export only emits keys it kept here, so a skipped
		// permission never shows up as missing from the file either.
		if ns == "" || name == "" || isBaseNamespace(ns) {
			continue
		}
		// Also skip a legacy row whose parsed parts break the grammar the reconciler
		// enforces: an underscore or uppercase in a part, a reserved verb, or an
		// over-long slug. No file entry could name such a row, because
		// validatePermissionSpec rejects the same key, so keeping it would wedge every
		// plan: it can be neither kept (it counts as unaccounted) nor deleted (a delete
		// spec fails validation first). Treat it as out of scope, like a base
		// permission. New rows cannot reach this state now that CreatePermission runs
		// the same check, but rows created before that tightening can.
		if schema.ValidateCustomPermission(ns, name) != nil {
			continue
		}
		current = append(current, currentPermission{ID: p.GetId(), Key: p.GetKey()})
	}
	return current, nil
}

func (r *PermissionReconciler) apply(ctx context.Context, op permissionOp) error {
	switch op.action {
	case opAdd:
		_, err := r.client.CreatePermission(ctx, authReq(&frontierv1beta1.CreatePermissionRequest{
			Bodies: []*frontierv1beta1.PermissionRequestBody{{
				Key: op.spec.Key,
			}},
		}, r.header))
		return err
	case opRemove:
		_, err := r.client.DeletePermission(ctx, authReq(&frontierv1beta1.DeletePermissionRequest{
			Id: op.id,
		}, r.header))
		return err
	default:
		return fmt.Errorf("unknown op action %q", op.action)
	}
}
