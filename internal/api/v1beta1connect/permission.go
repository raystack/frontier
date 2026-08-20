package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreatePermission(ctx context.Context, request *connect.Request[frontierv1beta1.CreatePermissionRequest]) (*connect.Response[frontierv1beta1.CreatePermissionResponse], error) {
	errorLogger := NewErrorLogger()

	definition := schema.ServiceDefinition{}
	var permissionSlugs []string
	for _, permBody := range request.Msg.GetBodies() {
		permNamespace, permName := schema.PermissionNamespaceAndNameFromKey(permBody.GetKey())
		if permName == "" || permNamespace == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrPermissionKeyNotation)
		}
		// One shared check for the verb grammar, the namespace grammar, the reserved
		// verbs, and the slug length, so the create API and the reconcile plan agree
		// and a permission that would fail when SpiceDB compiles the schema is
		// rejected up front with a message that says which rule it broke.
		if err := schema.ValidateCustomPermission(permNamespace, permName); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		permissionSlugs = append(permissionSlugs, schema.FQPermissionNameFromNamespace(permNamespace, permName))

		metaDataMap := metadata.Metadata{}
		if permBody.GetMetadata() != nil {
			metaDataMap = metadata.Build(permBody.GetMetadata().AsMap())
		}
		if _, ok := metaDataMap["description"]; !ok {
			metaDataMap["description"] = ""
		}

		definition.Permissions = append(definition.Permissions, schema.ResourcePermission{
			Name:        permName,
			Namespace:   permNamespace,
			Description: metaDataMap["description"].(string),
		})
	}

	err := h.bootstrapService.AppendSchema(ctx, definition)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreatePermission.AppendSchema", err,
			"permission_slugs", permissionSlugs)

		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail),
			errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreatePermission.AppendSchema: permission_slugs=%v: %w", permissionSlugs, err))
		}
	}

	permList, err := h.permissionService.List(ctx, permission.Filter{Slugs: permissionSlugs})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreatePermission.List: permission_slugs=%v: %w", permissionSlugs, err))
	}

	var pbPerms []*frontierv1beta1.Permission
	for _, perm := range permList {
		permPB, err := transformPermissionToPB(perm)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreatePermission: entity_id=%s: %w", perm.ID, err))
		}
		pbPerms = append(pbPerms, permPB)
	}
	return connect.NewResponse(&frontierv1beta1.CreatePermissionResponse{Permissions: pbPerms}), nil
}

func transformPermissionToPB(perm permission.Permission) (*frontierv1beta1.Permission, error) {
	var metadata *structpb.Struct
	var err error
	if len(perm.Metadata) > 0 {
		metadata, err = structpb.NewStruct(perm.Metadata)
		if err != nil {
			return nil, err
		}
	}

	// key is the only identity carried on the wire, so it must read back to
	// the exact stored pair. A row that cannot round-trip (namespace without
	// a slash, or with a dot in a part) fails loudly here instead of
	// returning a key that reads back as a different permission.
	key := schema.PermissionKeyFromNamespaceAndName(perm.NamespaceID, perm.Name)
	if ns, name := schema.PermissionNamespaceAndNameFromKey(key); ns != perm.NamespaceID || name != perm.Name {
		return nil, fmt.Errorf("permission namespace %q and name %q do not round-trip through key %q", perm.NamespaceID, perm.Name, key)
	}

	return &frontierv1beta1.Permission{
		Id:        perm.ID,
		CreatedAt: timestamppb.New(perm.CreatedAt),
		UpdatedAt: timestamppb.New(perm.UpdatedAt),
		Metadata:  metadata,
		Key:       key,
	}, nil
}

// UpdatePermission should only be used to update permission metadata at the moment
func (h *ConnectHandler) UpdatePermission(ctx context.Context, request *connect.Request[frontierv1beta1.UpdatePermissionRequest]) (*connect.Response[frontierv1beta1.UpdatePermissionResponse], error) {
	errorLogger := NewErrorLogger()

	var metaDataMap metadata.Metadata
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	}

	permNamespace, permName := schema.PermissionNamespaceAndNameFromKey(request.Msg.GetBody().GetKey())
	if permNamespace == "" || permName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrPermissionKeyNotation)
	}
	// Same shared check as create, so an update cannot rename a permission to a key
	// SpiceDB will reject and then break the next schema compile at boot.
	if err := schema.ValidateCustomPermission(permNamespace, permName); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	updatedPermission, err := h.permissionService.Update(ctx, permission.Permission{
		ID:          request.Msg.GetId(),
		Name:        permName,
		NamespaceID: permNamespace,
		Metadata:    metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdatePermission", err,
			"permission_id", request.Msg.GetId(),
			"permission_name", permName,
			"permission_namespace", permNamespace)

		switch {
		case errors.Is(err, permission.ErrNotExist),
			errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdatePermission: permission_id=%s permission_name=%s permission_namespace=%s: %w", request.Msg.GetId(), permName, permNamespace, err))
		}
	}

	permissionPB, err := transformPermissionToPB(updatedPermission)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdatePermission: entity_id=%s: %w", updatedPermission.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.UpdatePermissionResponse{Permission: permissionPB}), nil
}

// DeletePermission deletes a permission and the tuples that reference it.
// Built-in permissions (defined by the base schema) are rejected,
// because bootstrap recreates them on the next boot. So only permissions added
// through the API can be deleted.
func (h *ConnectHandler) DeletePermission(ctx context.Context, request *connect.Request[frontierv1beta1.DeletePermissionRequest]) (*connect.Response[frontierv1beta1.DeletePermissionResponse], error) {
	errorLogger := NewErrorLogger()
	permissionID := request.Msg.GetId()

	// load the permission so we can check it before deleting
	perm, err := h.permissionService.Get(ctx, permissionID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DeletePermission.Get", err, "permission_id", permissionID)
		if errors.Is(err, permission.ErrNotExist) || errors.Is(err, permission.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeletePermission.Get: permission_id=%s: %w", permissionID, err))
	}

	// refuse to delete a built-in permission — bootstrap would just recreate it
	builtin, err := h.bootstrapService.BuiltinPermissions(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeletePermission.BuiltinPermissions: permission_id=%s: %w", permissionID, err))
	}
	slug := perm.Slug
	if slug == "" {
		slug = perm.GenerateSlug()
	}
	if _, isBuiltin := builtin[slug]; isBuiltin {
		return nil, connect.NewError(connect.CodeFailedPrecondition,
			errors.New("cannot delete a built-in permission (defined by the base schema); it is recreated on the next boot"))
	}

	// A namespace exists in SpiceDB only as long as it has a permission. If this
	// is its last permission, the namespace disappears on the next boot — and
	// SpiceDB won't drop a type that still has relationships, which breaks
	// startup. So if nothing else keeps the namespace alive, refuse the delete
	// while any relationship of that type still exists.
	otherPerms, err := h.permissionService.List(ctx, permission.Filter{Namespace: perm.NamespaceID})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeletePermission.List: permission_id=%s: %w", permissionID, err))
	}
	namespaceSurvives := false
	for _, p := range otherPerms {
		if p.ID != perm.ID {
			namespaceSurvives = true
			break
		}
	}
	if !namespaceSurvives {
		// ask SpiceDB itself (the source of truth) whether anything of this type
		// still exists — resources, policy grants, or directly-created tuples.
		rels, err := h.relationService.ListRelations(ctx, relation.Relation{
			Object: relation.Object{Namespace: perm.NamespaceID},
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeletePermission.ListRelations: permission_id=%s: %w", permissionID, err))
		}
		if len(rels) > 0 {
			return nil, connect.NewError(connect.CodeFailedPrecondition,
				errors.New("cannot delete the last permission of namespace "+perm.NamespaceID+": relationships of this type still exist in SpiceDB (e.g. resources); remove them first"))
		}
	}

	err = h.permissionService.Delete(ctx, permissionID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DeletePermission", err, "permission_id", permissionID)

		switch {
		case errors.Is(err, permission.ErrNotExist), errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeletePermission: permission_id=%s: %w", permissionID, err))
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeletePermissionResponse{}), nil
}

func (h *ConnectHandler) ListPermissions(ctx context.Context, request *connect.Request[frontierv1beta1.ListPermissionsRequest]) (*connect.Response[frontierv1beta1.ListPermissionsResponse], error) {
	actionsList, err := h.permissionService.List(ctx, permission.Filter{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPermissions: %w", err))
	}

	var perms []*frontierv1beta1.Permission
	for _, act := range actionsList {
		actPB, err := transformPermissionToPB(act)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListPermissions: entity_id=%s: %w", act.ID, err))
		}
		perms = append(perms, actPB)
	}
	return connect.NewResponse(&frontierv1beta1.ListPermissionsResponse{Permissions: perms}), nil
}

func (h *ConnectHandler) GetPermission(ctx context.Context, request *connect.Request[frontierv1beta1.GetPermissionRequest]) (*connect.Response[frontierv1beta1.GetPermissionResponse], error) {
	errorLogger := NewErrorLogger()
	permissionID := request.Msg.GetId()

	fetchedPermission, err := h.permissionService.Get(ctx, permissionID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetPermission", err,
			"permission_id", permissionID)

		switch {
		case errors.Is(err, permission.ErrNotExist), errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetPermission: permission_id=%s: %w", permissionID, err))
		}
	}

	permissionPB, err := transformPermissionToPB(fetchedPermission)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetPermission: entity_id=%s: %w", fetchedPermission.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.GetPermissionResponse{Permission: permissionPB}), nil
}
