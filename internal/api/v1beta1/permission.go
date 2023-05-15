package v1beta1

import (
	"context"
	"errors"

	"github.com/odpf/shield/pkg/metadata"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/odpf/shield/core/permission"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/core/namespace"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockery --name=PermissionService -r --case underscore --with-expecter --structname PermissionService --filename action_service.go --output=./mocks
type PermissionService interface {
	Get(ctx context.Context, id string) (permission.Permission, error)
	List(ctx context.Context) ([]permission.Permission, error)
	Upsert(ctx context.Context, perm permission.Permission) (permission.Permission, error)
	Update(ctx context.Context, perm permission.Permission) (permission.Permission, error)
}

var grpcPermissionNotFoundErr = status.Errorf(codes.NotFound, "action doesn't exist")

func (h Handler) ListPermissions(ctx context.Context, request *shieldv1beta1.ListPermissionsRequest) (*shieldv1beta1.ListPermissionsResponse, error) {
	logger := grpczap.Extract(ctx)

	actionsList, err := h.permissionService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var actions []*shieldv1beta1.Permission
	for _, act := range actionsList {
		actPB, err := transformPermissionToPB(act)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		actions = append(actions, actPB)
	}

	return &shieldv1beta1.ListPermissionsResponse{Permissions: actions}, nil
}

func (h Handler) CreatePermission(ctx context.Context, request *shieldv1beta1.CreatePermissionRequest) (*shieldv1beta1.CreatePermissionResponse, error) {
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

	newPermission, err := h.permissionService.Upsert(ctx, permission.Permission{
		ID:          request.GetBody().GetId(),
		Name:        request.GetBody().GetName(),
		NamespaceID: request.GetBody().GetNamespaceId(),
		Metadata:    metaDataMap,
	})
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

	actionPB, err := transformPermissionToPB(newPermission)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreatePermissionResponse{Permission: actionPB}, nil
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

	actionPB, err := transformPermissionToPB(fetchedPermission)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetPermissionResponse{Permission: actionPB}, nil
}

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
		NamespaceID: request.GetBody().GetNamespaceId(),
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
		Id:          perm.ID,
		Name:        perm.Name,
		Slug:        perm.Slug,
		NamespaceId: perm.NamespaceID,
		Metadata:    metadata,
	}, nil
}
