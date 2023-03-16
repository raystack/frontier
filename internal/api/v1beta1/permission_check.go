package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/goto/shield/core/action"
	"github.com/goto/shield/core/resource"
	"github.com/goto/shield/core/user"
	shieldv1beta1 "github.com/goto/shield/proto/v1beta1"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h Handler) CheckResourcePermission(ctx context.Context, req *shieldv1beta1.CheckResourcePermissionRequest) (*shieldv1beta1.CheckResourcePermissionResponse, error) {
	logger := grpczap.Extract(ctx)
	//if err := req.ValidateAll(); err != nil {
	//	formattedErr := getValidationErrorMessage(err)
	//	logger.Error(formattedErr.Error())
	//	return nil, status.Errorf(codes.NotFound, formattedErr.Error())
	//}

	result, err := h.resourceService.CheckAuthz(ctx, resource.Resource{
		Name:        req.GetObjectId(),
		NamespaceID: req.GetObjectNamespace(),
	}, action.Action{ID: req.GetPermission()})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
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
