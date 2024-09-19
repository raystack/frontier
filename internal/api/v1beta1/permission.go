package v1beta1

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/raystack/frontier/core/permission"

	"github.com/raystack/frontier/core/namespace"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

var grpcPermissionNotFoundErr = status.Errorf(codes.NotFound, "permission doesn't exist")

func (h Handler) ListPermissions(ctx context.Context, request *frontierv1beta1.ListPermissionsRequest) (*frontierv1beta1.ListPermissionsResponse, error) {
	actionsList, err := h.permissionService.List(ctx, permission.Filter{})
	if err != nil {
		return nil, err
	}

	var perms []*frontierv1beta1.Permission
	for _, act := range actionsList {
		actPB, err := transformPermissionToPB(act)
		if err != nil {
			return nil, err
		}
		perms = append(perms, actPB)
	}
	return &frontierv1beta1.ListPermissionsResponse{Permissions: perms}, nil
}

func (h Handler) CreatePermission(ctx context.Context, request *frontierv1beta1.CreatePermissionRequest) (*frontierv1beta1.CreatePermissionResponse, error) {
	var err error

	definition := schema.ServiceDefinition{}
	var permissionSlugs []string
	for _, permBody := range request.GetBodies() {
		permNamespace, permName := schema.PermissionNamespaceAndNameFromKey(permBody.GetKey())
		if permNamespace == "" || permName == "" {
			permNamespace, permName = permBody.GetNamespace(), permBody.GetName()
		}
		if permName == "" || permNamespace == "" {
			return nil, grpcBadBodyError
		}
		if !schema.IsValidPermissionName(permName) {
			return nil, status.Errorf(codes.InvalidArgument, "permission name cannot contain special characters")
		}

		if permNamespace == schema.DefaultNamespace {
			return nil, status.Errorf(codes.InvalidArgument, "permission namespace cannot be %s", schema.DefaultNamespace)
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

	err = h.bootstrapService.AppendSchema(ctx, definition)
	if err != nil {
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail),
			errors.Is(err, permission.ErrInvalidID):
			return nil, grpcBadBodyError
		default:
			return nil, err
		}
	}

	permList, err := h.permissionService.List(ctx, permission.Filter{Slugs: permissionSlugs})
	if err != nil {
		return nil, err
	}

	var pbPerms []*frontierv1beta1.Permission
	for _, perm := range permList {
		permPB, err := transformPermissionToPB(perm)
		if err != nil {
			return nil, err
		}
		pbPerms = append(pbPerms, permPB)
	}
	return &frontierv1beta1.CreatePermissionResponse{Permissions: pbPerms}, nil
}

func (h Handler) GetPermission(ctx context.Context, request *frontierv1beta1.GetPermissionRequest) (*frontierv1beta1.GetPermissionResponse, error) {
	fetchedPermission, err := h.permissionService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, permission.ErrNotExist), errors.Is(err, permission.ErrInvalidID):
			return nil, grpcPermissionNotFoundErr
		default:
			return nil, err
		}
	}

	permissionPB, err := transformPermissionToPB(fetchedPermission)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetPermissionResponse{Permission: permissionPB}, nil
}

// UpdatePermission should only be used to update permission metadata at the moment
func (h Handler) UpdatePermission(ctx context.Context, request *frontierv1beta1.UpdatePermissionRequest) (*frontierv1beta1.UpdatePermissionResponse, error) {
	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())
	}

	permNamespace, permName := schema.PermissionNamespaceAndNameFromKey(request.GetBody().GetKey())
	if permNamespace == "" || permName == "" {
		permNamespace, permName = request.GetBody().GetNamespace(), request.GetBody().GetName()
	}
	updatedPermission, err := h.permissionService.Update(ctx, permission.Permission{
		ID:          request.GetId(),
		Name:        permName,
		NamespaceID: permNamespace,
		Metadata:    metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, permission.ErrNotExist),
			errors.Is(err, permission.ErrInvalidID):
			return nil, grpcPermissionNotFoundErr
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			return nil, err
		}
	}

	actionPB, err := transformPermissionToPB(updatedPermission)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.UpdatePermissionResponse{Permission: actionPB}, nil
}

func (h Handler) DeletePermission(ctx context.Context, in *frontierv1beta1.DeletePermissionRequest) (*frontierv1beta1.DeletePermissionResponse, error) {
	return nil, grpcOperationUnsupported
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
