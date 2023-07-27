package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/relation"

	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h Handler) CheckResourcePermission(ctx context.Context, req *frontierv1beta1.CheckResourcePermissionRequest) (*frontierv1beta1.CheckResourcePermissionResponse, error) {
	logger := grpczap.Extract(ctx)
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(req.GetResource())
	if len(req.GetResource()) == 0 || err != nil {
		objectNamespace = schema.ParseNamespaceAliasIfRequired(req.GetObjectNamespace())
		objectID = req.GetObjectId()
	}
	if objectNamespace == "" || objectID == "" {
		return nil, grpcBadBodyError
	}

	result, err := h.resourceService.CheckAuthz(ctx, relation.Object{
		ID:        objectID,
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

	logAuditForCheck(ctx, result, objectID, objectNamespace)
	if !result {
		return &frontierv1beta1.CheckResourcePermissionResponse{Status: false}, nil
	}
	return &frontierv1beta1.CheckResourcePermissionResponse{Status: true}, nil
}

func logAuditForCheck(ctx context.Context, result bool, objectID string, objectNamespace string) {
	auditStatus := "success"
	if !result {
		auditStatus = "failure"
	}
	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).LogWithAttrs(audit.PermissionCheckedEvent, audit.Target{
		ID:   objectID,
		Type: objectNamespace,
	}, map[string]string{
		"status": auditStatus,
	})
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
	currentUser, err := h.GetLoggedInPrincipal(ctx)
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
