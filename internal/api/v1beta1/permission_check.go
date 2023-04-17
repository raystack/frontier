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

type resourcePermissionResult struct {
	objectId        string
	objectNamespace string
	permission      string
	allowed         bool
}

func (h Handler) CheckResourcePermission(ctx context.Context, req *shieldv1beta1.CheckResourcePermissionRequest) (*shieldv1beta1.CheckResourcePermissionResponse, error) {
	logger := grpczap.Extract(ctx)
	//if err := req.ValidateAll(); err != nil {
	//	formattedErr := getValidationErrorMessage(err)
	//	logger.Error(formattedErr.Error())
	//	return nil, status.Errorf(codes.NotFound, formattedErr.Error())
	//}
	//  To have backward compatibility
	if req.ObjectId != "" && len(req.ResourcePermissions) == 0 {
		return h.checkSingleResourcePermission(ctx, req)
	}

	if len(req.ResourcePermissions) > h.checkAPILimit || len(req.ResourcePermissions) == 0 {
		formattedErr := fmt.Errorf("%s: %s", ErrRequestBodyValidation, "resource_permissions")
		logger.Error(formattedErr.Error())
		return nil, formattedErr
	}

	var results []*shieldv1beta1.CheckResourcePermissionResponse_ResourcePermissionResponse
	checkCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultCh := make(chan resourcePermissionResult)
	defer close(resultCh)
	errorCh := make(chan error)
	defer close(errorCh)

	for _, permission := range req.ResourcePermissions {
		go func(checkCtx context.Context, resourcePermission *shieldv1beta1.ResourcePermission,
			resCh chan<- resourcePermissionResult, errCh chan<- error) {
			result, err := h.resourceService.CheckAuthz(ctx, resource.Resource{
				Name:        resourcePermission.GetObjectId(),
				NamespaceID: resourcePermission.GetObjectNamespace(),
			}, action.Action{ID: resourcePermission.GetPermission()})
			if err != nil {
				switch {
				case errors.Is(err, user.ErrInvalidEmail):
					errCh <- grpcUnauthenticated
					return
				default:
					formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
					logger.Error(formattedErr.Error())
					errCh <- status.Errorf(codes.Internal, ErrInternalServer.Error())
					return
				}
			}
			select {
			case <-checkCtx.Done():
				return
			case resCh <- resourcePermissionResult{
				objectId:        resourcePermission.GetObjectId(),
				objectNamespace: resourcePermission.GetObjectNamespace(),
				permission:      resourcePermission.GetPermission(),
				allowed:         result,
			}:
			}
		}(checkCtx, permission, resultCh, errorCh)
	}

	for i := 0; i < len(req.ResourcePermissions); i++ {
		select {
		case result, ok := <-resultCh:
			if !ok {
				break
			}
			results = append(results, &shieldv1beta1.CheckResourcePermissionResponse_ResourcePermissionResponse{
				ObjectId:        result.objectId,
				ObjectNamespace: result.objectNamespace,
				Permission:      result.permission,
				Allowed:         result.allowed,
			})
		case err := <-errorCh:
			cancel()
			return nil, err
		}
	}

	return &shieldv1beta1.CheckResourcePermissionResponse{ResourcePermissions: results}, nil
}

// Deprecated: checkSingleResourcePermission is used to check the single resource permission, this method is deprecated it needs to be cleaned up once the clients are onboarded
// with the new check API
func (h Handler) checkSingleResourcePermission(ctx context.Context, req *shieldv1beta1.CheckResourcePermissionRequest) (*shieldv1beta1.CheckResourcePermissionResponse, error) {
	logger := grpczap.Extract(ctx)

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
