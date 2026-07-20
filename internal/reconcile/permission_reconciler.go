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

// Validate checks every entry, and the in-file slug conflicts, without touching
// the server, so a bad entry stops the whole file before anything applies.
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
		slug := s.slug()
		if prev, dup := seen[slug]; dup {
			if prev.Namespace != s.Namespace || prev.Name != s.Name {
				return fmt.Errorf("permissions %s and %s collide on the same slug %q", prev, s, slug)
			}
			if prev.Delete != s.Delete {
				return fmt.Errorf("permission %s is listed both with and without delete", s)
			}
		}
		seen[slug] = s
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
		specs = append(specs, PermissionSpec{Namespace: c.Namespace, Name: c.Name})
	}
	sort.Slice(specs, func(i, j int) bool {
		a, b := specs[i], specs[j]
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		return a.Name < b.Name
	})
	return specs, nil
}

func (r *PermissionReconciler) fetchCurrent(ctx context.Context) ([]currentPermission, error) {
	resp, err := r.client.ListPermissions(ctx, authReq(&frontierv1beta1.ListPermissionsRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	var current []currentPermission
	for _, p := range resp.Msg.GetPermissions() {
		if isBaseNamespace(p.GetNamespace()) {
			continue
		}
		current = append(current, currentPermission{
			ID:        p.GetId(),
			Namespace: p.GetNamespace(),
			Name:      p.GetName(),
		})
	}
	return current, nil
}

func (r *PermissionReconciler) apply(ctx context.Context, op permissionOp) error {
	switch op.action {
	case opAdd:
		_, err := r.client.CreatePermission(ctx, authReq(&frontierv1beta1.CreatePermissionRequest{
			Bodies: []*frontierv1beta1.PermissionRequestBody{{
				Key: schema.PermissionKeyFromNamespaceAndName(op.spec.Namespace, op.spec.Name),
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
