package reconcile

import (
	"context"
	"fmt"
	"sort"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
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

func (r *RoleReconciler) Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error) {
	var specs []RoleSpec
	if err := yaml.Unmarshal(spec, &specs); err != nil {
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

// Export returns the current platform roles as a desired-state spec. Custom
// roles are exported in full. Predefined roles are exported only when they
// differ from their shipped definition (title, permissions, or scopes), and
// only with the differing fields, so converged ones stay out of the file.
// The comparisons mirror diffPredefinedRole exactly — including skipping
// fields that are empty on the server, which a file cannot represent — so
// reconciling an export's output always plans no changes.
func (r *RoleReconciler) Export(ctx context.Context) (any, error) {
	current, err := r.fetchCurrent(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(current, func(i, j int) bool { return current[i].Name < current[j].Name })

	specs := make([]RoleSpec, 0, len(current))
	for _, c := range current {
		def, isPredefined := predefinedRole(c.Name)
		if !isPredefined {
			specs = append(specs, RoleSpec{
				Name:        c.Name,
				Title:       c.Title,
				Description: c.Description,
				Permissions: sortedCopy(c.Permissions),
				Scopes:      sortedCopy(c.Scopes),
			})
			continue
		}
		entry := RoleSpec{Name: c.Name}
		changed := false
		if c.Title != "" && c.Title != def.Title {
			entry.Title = c.Title
			changed = true
		}
		if c.Description != "" && c.Description != def.Description {
			entry.Description = c.Description
			changed = true
		}
		if len(c.Permissions) > 0 && !stringSetsEqual(normalizePermissions(c.Permissions), normalizePermissions(def.Permissions)) {
			entry.Permissions = sortedCopy(c.Permissions)
			changed = true
		}
		if len(c.Scopes) > 0 && !stringSetsEqual(sortedCopy(c.Scopes), sortedCopy(def.Scopes)) {
			entry.Scopes = sortedCopy(c.Scopes)
			changed = true
		}
		if changed {
			specs = append(specs, entry)
		}
	}
	return specs, nil
}

func (r *RoleReconciler) fetchCurrent(ctx context.Context) ([]currentRole, error) {
	resp, err := r.client.ListRoles(ctx, authReq(&frontierv1beta1.ListRolesRequest{}, r.header))
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	var current []currentRole
	for _, ro := range resp.Msg.GetRoles() {
		current = append(current, currentRole{
			ID:          ro.GetId(),
			Name:        ro.GetName(),
			Title:       ro.GetTitle(),
			Description: metadataString(ro.GetMetadata(), descriptionKey),
			Permissions: ro.GetPermissions(),
			Scopes:      ro.GetScopes(),
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
		body, err := roleBody(op.spec)
		if err != nil {
			return err
		}
		_, err = r.client.CreateRole(ctx, authReq(&frontierv1beta1.CreateRoleRequest{Body: body}, r.header))
		return err
	case opUpdate:
		body, err := roleBody(op.spec)
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

func roleBody(spec RoleSpec) (*frontierv1beta1.RoleRequestBody, error) {
	fields := map[string]any{managedByKey: managedByValue}
	if spec.Description != "" {
		fields[descriptionKey] = spec.Description
	}
	md, err := structpb.NewStruct(fields)
	if err != nil {
		return nil, fmt.Errorf("build role metadata: %w", err)
	}
	return &frontierv1beta1.RoleRequestBody{
		Name:        spec.Name,
		Title:       spec.Title,
		Permissions: spec.Permissions,
		Scopes:      spec.Scopes,
		Metadata:    md,
	}, nil
}
