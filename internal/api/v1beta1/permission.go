package v1beta1

import (
	"context"
	"errors"

	"github.com/raystack/shield/internal/bootstrap/schema"

	"github.com/raystack/shield/pkg/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/raystack/shield/core/permission"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/shield/core/namespace"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockery --name=PermissionService -r --case underscore --with-expecter --structname PermissionService --filename permission_service.go --output=./mocks
type PermissionService interface {
	Get(ctx context.Context, id string) (permission.Permission, error)
	List(ctx context.Context, filter permission.Filter) ([]permission.Permission, error)
	Upsert(ctx context.Context, perm permission.Permission) (permission.Permission, error)
	Update(ctx context.Context, perm permission.Permission) (permission.Permission, error)
}

//go:generate mockery --name=BootstrapService -r --case underscore --with-expecter --structname BootstrapService --filename bootstrap_service.go --output=./mocks
type BootstrapService interface {
	AppendSchema(ctx context.Context, definition schema.ServiceDefinition) error
}

var grpcPermissionNotFoundErr = status.Errorf(codes.NotFound, "permission doesn't exist")

func (h Handler) ListPermissions(ctx context.Context, request *shieldv1beta1.ListPermissionsRequest) (*shieldv1beta1.ListPermissionsResponse, error) {
	logger := grpczap.Extract(ctx)

	actionsList, err := h.permissionService.List(ctx, permission.Filter{})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var perms []*shieldv1beta1.Permission
	for _, act := range actionsList {
		actPB, err := transformPermissionToPB(act)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		perms = append(perms, actPB)
	}
	return &shieldv1beta1.ListPermissionsResponse{Permissions: perms}, nil
}

func (h Handler) CreatePermission(ctx context.Context, request *shieldv1beta1.CreatePermissionRequest) (*shieldv1beta1.CreatePermissionResponse, error) {
	logger := grpczap.Extract(ctx)
	var err error

	definition := schema.ServiceDefinition{}
	var permissionSlugs []string
	for _, permBody := range request.GetBodies() {
		if permBody.GetName() == "" || permBody.GetNamespace() == "" {
			return nil, grpcBadBodyError
		}
		permissionSlugs = append(permissionSlugs, schema.FQPermissionNameFromNamespace(permBody.GetNamespace(), permBody.GetName()))

		metaDataMap := metadata.Metadata{}
		if permBody.GetMetadata() != nil {
			metaDataMap, err = metadata.Build(permBody.GetMetadata().AsMap())
			if err != nil {
				logger.Error(err.Error())
				return nil, grpcBadBodyError
			}
		}
		if _, ok := metaDataMap["description"]; !ok {
			metaDataMap["description"] = ""
		}

		definition.Permissions = append(definition.Permissions, schema.ResourcePermission{
			Name:        permBody.Name,
			Namespace:   permBody.Namespace,
			Description: metaDataMap["description"].(string),
		})
	}

	err = h.bootstrapService.AppendSchema(ctx, definition)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail),
			errors.Is(err, permission.ErrInvalidID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	permList, err := h.permissionService.List(ctx, permission.Filter{Slugs: permissionSlugs})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var pbPerms []*shieldv1beta1.Permission
	for _, perm := range permList {
		permPB, err := transformPermissionToPB(perm)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		pbPerms = append(pbPerms, permPB)
	}
	return &shieldv1beta1.CreatePermissionResponse{Permissions: pbPerms}, nil
}

func (h Handler) GetPermission(ctx context.Context, request *shieldv1beta1.GetPermissionRequest) (*shieldv1beta1.GetPermissionResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedPermission, err := h.permissionService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, permission.ErrNotExist), errors.Is(err, permission.ErrInvalidID):
			return nil, grpcPermissionNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	permissionPB, err := transformPermissionToPB(fetchedPermission)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetPermissionResponse{Permission: permissionPB}, nil
}

// UpdatePermission should only be used to update permission metadata at the moment
func (h Handler) UpdatePermission(ctx context.Context, request *shieldv1beta1.UpdatePermissionRequest) (*shieldv1beta1.UpdatePermissionResponse, error) {
	logger := grpczap.Extract(ctx)

	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap, err = metadata.Build(request.GetBody().GetMetadata().AsMap())
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcBadBodyError
		}
	}

	updatedPermission, err := h.permissionService.Update(ctx, permission.Permission{
		ID:          request.GetId(),
		Name:        request.GetBody().GetName(),
		NamespaceID: request.GetBody().GetNamespace(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, permission.ErrNotExist),
			errors.Is(err, permission.ErrInvalidID):
			return nil, grpcPermissionNotFoundErr
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	actionPB, err := transformPermissionToPB(updatedPermission)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdatePermissionResponse{Permission: actionPB}, nil
}

func transformPermissionToPB(perm permission.Permission) (*shieldv1beta1.Permission, error) {
	var metadata *structpb.Struct
	var err error
	if len(perm.Metadata) > 0 {
		metadata, err = structpb.NewStruct(perm.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return &shieldv1beta1.Permission{
		Id:        perm.ID,
		Name:      perm.Name,
		Namespace: perm.NamespaceID,
		Metadata:  metadata,
	}, nil
}
