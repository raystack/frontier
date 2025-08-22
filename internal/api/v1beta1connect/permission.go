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
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail),
			errors.Is(err, permission.ErrInvalidID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	permList, err := h.permissionService.List(ctx, permission.Filter{Slugs: permissionSlugs})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var pbPerms []*frontierv1beta1.Permission
	for _, perm := range permList {
		permPB, err := transformPermissionToPB(perm)
		if err != nil {
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
