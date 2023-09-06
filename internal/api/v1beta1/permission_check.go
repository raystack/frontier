package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/resource"

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

	result, err := h.resourceService.CheckAuthz(ctx, resource.Check{
		Object: relation.Object{
			ID:        objectID,
			Namespace: objectNamespace,
		},
		Permission: req.GetPermission(),
	})
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

func (h Handler) BatchCheckPermission(ctx context.Context, req *frontierv1beta1.BatchCheckPermissionRequest) (*frontierv1beta1.BatchCheckPermissionResponse, error) {
	logger := grpczap.Extract(ctx)

	checks := []resource.Check{}
	for _, body := range req.GetBodies() {
		objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(body.Resource)
		if len(body.Resource) == 0 || err != nil {
			return nil, grpcBadBodyError
		}
		checks = append(checks, resource.Check{
			Object: relation.Object{
				ID:        objectID,
				Namespace: objectNamespace,
			},
			Permission: body.Permission,
		})
	}
	result, err := h.resourceService.BatchCheck(ctx, checks)
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

	pairs := []*frontierv1beta1.BatchCheckPermissionResponsePair{}
	for _, r := range result {
		pairs = append(pairs, &frontierv1beta1.BatchCheckPermissionResponsePair{
			Body: &frontierv1beta1.BatchCheckPermissionBody{
				Permission: r.Relation.RelationName,
				Resource:   schema.JoinNamespaceAndResourceID(r.Relation.Object.Namespace, r.Relation.Object.ID),
			},
			Status: r.Status,
		})
	}
	return &frontierv1beta1.BatchCheckPermissionResponse{
		Pairs: pairs,
	}, nil
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
	result, err := h.resourceService.CheckAuthz(ctx, resource.Check{
		Object: relation.Object{
			ID:        objectID,
			Namespace: objectNamespace,
		},
		Permission: permission,
	})
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
