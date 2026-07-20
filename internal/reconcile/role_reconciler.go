package reconcile

import (
	"context"
	"fmt"
	"sort"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
)

// Roles the reconciler creates or updates are stamped with this metadata, so
// their origin is visible and other tooling can tell them apart.
const (
	managedByKey   = "managed_by"
	managedByValue = "frontier-reconcile"
	descriptionKey = "description"
)

// RoleAPI is the API subset the role reconciler needs. Reads live on
// FrontierService, writes on AdminService; the caller provides one value that
// serves both.
type RoleAPI interface {
	ListRoles(context.Context, *connect.Request[frontierv1beta1.ListRolesRequest]) (*connect.Response[frontierv1beta1.ListRolesResponse], error)
	CreateRole(context.Context, *connect.Request[frontierv1beta1.CreateRoleRequest]) (*connect.Response[frontierv1beta1.CreateRoleResponse], error)
	UpdateRole(context.Context, *connect.Request[frontierv1beta1.UpdateRoleRequest]) (*connect.Response[frontierv1beta1.UpdateRoleResponse], error)
	DeleteRole(context.Context, *connect.Request[frontierv1beta1.DeleteRoleRequest]) (*connect.Response[frontierv1beta1.DeleteRoleResponse], error)
}

// RoleReconciler makes platform-level roles match the desired spec. The role
// name is the identity; title, permissions, and scopes are the managed fields.
type RoleReconciler struct {
	client RoleAPI
	header string
}

func NewRoleReconciler(client RoleAPI, header string) *RoleReconciler {
	return &RoleReconciler{client: client, header: header}
}

func (r *RoleReconciler) Kind() string { return KindRole }

// Validate checks every entry, and rejects duplicate role names, without touching
// the server, so a bad entry stops the whole file before anything applies.
func (r *RoleReconciler) Validate(spec []byte) error {
	var specs []RoleSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return fmt.Errorf("parse %s spec: %w", KindRole, err)
	}
	seen := map[string]struct{}{}
	for _, s := range specs {
		_, isPredefined := roleDefault(s.Name)
		if err := validateRoleSpec(s, isPredefined); err != nil {
			return fmt.Errorf("invalid role spec %q: %w", s.Name, err)
		}
		if _, dup := seen[s.Name]; dup {
			return fmt.Errorf("role %q is listed more than once", s.Name)
		}
		seen[s.Name] = struct{}{}
	}
	return nil
}

func (r *RoleReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []RoleSpec
	if err := decodeSpec(spec, &specs); err != nil {
		return Report{}, fmt.Errorf("parse %s spec: %w", KindRole, err)
	}

	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return Report{}, err
	}

	ops, err := diffRoles(specs, current)
	if err != nil {
		return Report{}, err
	}

	rep := Report{Kind: KindRole, DryRun: dryRun}
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

// Export returns the current platform roles as a desired-state spec. Each
// server role is compared to its default (empty for a custom role, the shipped
// definition for a predefined one) and only the fields that differ are emitted.
// A custom role always emits an entry, because a custom role must appear in the
// file to be kept; a predefined role emits an entry only when some field
// differs, so converged ones stay out. An empty server value that differs from
// the default is written as an explicit empty (`title: ""`, `scopes: []`), so
// the value round-trips exactly and reconciling an export plans no changes.
func (r *RoleReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(current, func(i, j int) bool { return current[i].Name < current[j].Name })

	specs := make([]RoleSpec, 0, len(current))
	for _, c := range current {
		def, isPredefined := roleDefault(c.Name)
		have := c.fields()
		changes := diffFields(have, def)
		if isPredefined && len(changes) == 0 {
			continue
		}
		entry := RoleSpec{Name: c.Name}
		for _, field := range changes {
			switch field {
			case "title":
				entry.Title = strPtr(have.Title)
			case "description":
				entry.Description = strPtr(have.Description)
			case "permissions":
				entry.Permissions = slicePtr(have.Permissions)
			case "scopes":
				entry.Scopes = slicePtr(have.Scopes)
			}
		}
		specs = append(specs, entry)
	}
	return specs, nil
}

func strPtr(s string) *string { return &s }

func slicePtr(s []string) *[]string { return &s }

func (r *RoleReconciler) fetchCurrent(ctx context.Context) ([]currentRole, error) {
	resp, err := r.client.ListRoles(ctx, authReq(&frontierv1beta1.ListRolesRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	var current []currentRole
	for _, ro := range resp.Msg.GetRoles() {
		var metadata map[string]any
		if md := ro.GetMetadata(); md != nil {
			metadata = md.AsMap()
		}
		current = append(current, currentRole{
			ID:          ro.GetId(),
			Name:        ro.GetName(),
			Title:       ro.GetTitle(),
			Description: metadataString(ro.GetMetadata(), descriptionKey),
			Permissions: ro.GetPermissions(),
			Scopes:      ro.GetScopes(),
			Metadata:    metadata,
		})
	}
	return current, nil
}

// metadataString reads a string value from role metadata, or "" if absent.
func metadataString(md *structpb.Struct, key string) string {
	if md == nil {
		return ""
	}
	if v, ok := md.GetFields()[key]; ok {
		return v.GetStringValue()
	}
	return ""
}

func (r *RoleReconciler) apply(ctx context.Context, op roleOp) error {
	switch op.action {
	case opAdd:
		body, err := roleBody(op.name, op.fields, op.baseMetadata)
		if err != nil {
			return err
		}
		_, err = r.client.CreateRole(ctx, authReq(&frontierv1beta1.CreateRoleRequest{Body: body}, r.header))
		return err
	case opUpdate:
		body, err := roleBody(op.name, op.fields, op.baseMetadata)
		if err != nil {
			return err
		}
		_, err = r.client.UpdateRole(ctx, authReq(&frontierv1beta1.UpdateRoleRequest{Id: op.id, Body: body}, r.header))
		return err
	case opRemove:
		_, err := r.client.DeleteRole(ctx, authReq(&frontierv1beta1.DeleteRoleRequest{Id: op.id}, r.header))
		return err
	default:
		return fmt.Errorf("unknown op action %q", op.action)
	}
}

// roleBody builds the request body for an add or update. base is the role's
// current metadata (nil for a new role); the reconciler-managed keys are merged
// over a copy of it, so other metadata keys an operator set are kept, not
// dropped. An empty managed description clears the key, matching a reset.
func roleBody(name string, want roleFields, base map[string]any) (*frontierv1beta1.RoleRequestBody, error) {
	fields := map[string]any{}
	for k, v := range base {
		fields[k] = v
	}
	fields[managedByKey] = managedByValue
	if want.Description != "" {
		fields[descriptionKey] = want.Description
	} else {
		delete(fields, descriptionKey)
	}
	md, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, fmt.Errorf("build role metadata: %w", err)
	}
	return &frontierv1beta1.RoleRequestBody{
		Name:        name,
		Title:       want.Title,
		Permissions: want.Permissions,
		Scopes:      want.Scopes,
		Metadata:    md,
	}, nil
}
