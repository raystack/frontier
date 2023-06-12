package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/shield/internal/bootstrap/schema"

	"github.com/raystack/shield/core/relation"

	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/pkg/errors"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h Handler) CheckResourcePermission(ctx context.Context, req *shieldv1beta1.CheckResourcePermissionRequest) (*shieldv1beta1.CheckResourcePermissionResponse, error) {
	logger := grpczap.Extract(ctx)
	objectNamespace := schema.ParseNamespaceAliasIfRequired(req.GetObjectNamespace())
	result, err := h.resourceService.CheckAuthz(ctx, relation.Object{
		ID:        req.GetObjectId(),
		Namespace: objectNamespace,
	}, req.GetPermission())
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail) || errors.Is(err, errors.ErrUnauthenticated):
			return nil, grpcUnauthenticated
		default:
			formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
			logger.Error(formattedErr.Error())
			return nil, status.Errorf(codes.Internal, ErrInternalServer.Error())
		}
	}

	if !result {
		return &shieldv1beta1.CheckResourcePermissionResponse{Status: false}, nil
	}

	return &shieldv1beta1.CheckResourcePermissionResponse{Status: true}, nil
}

func (h Handler) IsAuthorized(ctx context.Context, objectNamespace, objectID, permission string) error {
	logger := grpczap.Extract(ctx)
	result, err := h.resourceService.CheckAuthz(ctx, relation.Object{
		ID:        objectID,
		Namespace: objectNamespace,
	}, permission)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail) || errors.Is(err, errors.ErrUnauthenticated):
			return grpcUnauthenticated
		default:
			formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
			logger.Error(formattedErr.Error())
			return status.Errorf(codes.Internal, ErrInternalServer.Error())
		}
	}

	if !result {
		return grpcPermissionDenied
	}
	return nil
}

func (h Handler) IsSuperUser(ctx context.Context) error {
	logger := grpczap.Extract(ctx)
	currentUser, err := h.GetLoggedInUser(ctx)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if ok, err := h.userService.IsSudo(ctx, currentUser.ID); err != nil {
		return err
	} else if ok {
		return nil
	}
	return grpcPermissionDenied
}
