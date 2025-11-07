package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PermissionService interface {
	Get(ctx context.Context, id string) (permission.Permission, error)
	List(ctx context.Context, filter permission.Filter) ([]permission.Permission, error)
	Upsert(ctx context.Context, perm permission.Permission) (permission.Permission, error)
	Update(ctx context.Context, perm permission.Permission) (permission.Permission, error)
}

type BootstrapService interface {
	AppendSchema(ctx context.Context, definition schema.ServiceDefinition) error
}

func (h *ConnectHandler) CreatePermission(ctx context.Context, request *connect.Request[frontierv1beta1.CreatePermissionRequest]) (*connect.Response[frontierv1beta1.CreatePermissionResponse], error) {
	errorLogger := NewErrorLogger()

	definition := schema.ServiceDefinition{}
	var permissionSlugs []string
	for _, permBody := range request.Msg.GetBodies() {
		permNamespace, permName := schema.PermissionNamespaceAndNameFromKey(permBody.GetKey())
		if permNamespace == "" || permName == "" {
			permNamespace, permName = permBody.GetNamespace(), permBody.GetName()
		}
		if permName == "" || permNamespace == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		if !schema.IsValidPermissionName(permName) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("permission name cannot contain special characters"))
		}

		if permNamespace == schema.DefaultNamespace {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("permission namespace cannot be "+schema.DefaultNamespace))
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
			zap.Strings("permission_slugs", permissionSlugs))

		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail),
			errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "CreatePermission.AppendSchema", err,
				zap.Strings("permission_slugs", permissionSlugs))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	permList, err := h.permissionService.List(ctx, permission.Filter{Slugs: permissionSlugs})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreatePermission.List", err,
			zap.Strings("permission_slugs", permissionSlugs))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var pbPerms []*frontierv1beta1.Permission
	for _, perm := range permList {
		permPB, err := transformPermissionToPB(perm)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "CreatePermission", perm.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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

	return &frontierv1beta1.Permission{
		Id:        perm.ID,
		Name:      perm.Name,
		CreatedAt: timestamppb.New(perm.CreatedAt),
		UpdatedAt: timestamppb.New(perm.UpdatedAt),
		Namespace: perm.NamespaceID,
		Metadata:  metadata,
		Key:       schema.PermissionKeyFromNamespaceAndName(perm.NamespaceID, perm.Name),
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
		permNamespace, permName = request.Msg.GetBody().GetNamespace(), request.Msg.GetBody().GetName()
	}
	updatedPermission, err := h.permissionService.Update(ctx, permission.Permission{
		ID:          request.Msg.GetId(),
		Name:        permName,
		NamespaceID: permNamespace,
		Metadata:    metaDataMap,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdatePermission", err,
			zap.String("permission_id", request.Msg.GetId()),
			zap.String("permission_name", permName),
			zap.String("permission_namespace", permNamespace))

		switch {
		case errors.Is(err, permission.ErrNotExist),
			errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "UpdatePermission", err,
				zap.String("permission_id", request.Msg.GetId()),
				zap.String("permission_name", permName),
				zap.String("permission_namespace", permNamespace))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	permissionPB, err := transformPermissionToPB(updatedPermission)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "UpdatePermission", updatedPermission.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.UpdatePermissionResponse{Permission: permissionPB}), nil
}

func (h *ConnectHandler) ListPermissions(ctx context.Context, request *connect.Request[frontierv1beta1.ListPermissionsRequest]) (*connect.Response[frontierv1beta1.ListPermissionsResponse], error) {
	errorLogger := NewErrorLogger()

	actionsList, err := h.permissionService.List(ctx, permission.Filter{})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListPermissions", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var perms []*frontierv1beta1.Permission
	for _, act := range actionsList {
		actPB, err := transformPermissionToPB(act)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListPermissions", act.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
			zap.String("permission_id", permissionID))

		switch {
		case errors.Is(err, permission.ErrNotExist), errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "GetPermission", err,
				zap.String("permission_id", permissionID))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	permissionPB, err := transformPermissionToPB(fetchedPermission)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetPermission", fetchedPermission.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetPermissionResponse{Permission: permissionPB}), nil
}
