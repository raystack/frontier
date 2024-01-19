package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"

	"github.com/raystack/frontier/pkg/str"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/permission"

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

func (h Handler) getPermissionName(ctx context.Context, ns, name string) (string, error) {
	if ns == schema.PlatformNamespace && schema.IsPlatformPermission(name) {
		return name, nil
	}
	logger := grpczap.Extract(ctx)
	perm, err := h.permissionService.Get(ctx, permission.AddNamespaceIfRequired(ns, name))
	if err != nil {
		switch {
		case errors.Is(err, permission.ErrNotExist):
			return "", grpcPermissionNotFoundErr
		default:
			formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
			logger.Error(formattedErr.Error())
			return "", status.Errorf(codes.Internal, ErrInternalServer.Error())
		}
	}
	// if the permission is on the same namespace as the object, use the name
	if perm.NamespaceID == ns {
		return perm.Name, nil
	}
	// else use fully qualified name(slug)
	return perm.Slug, nil
}

func (h Handler) CheckResourcePermission(ctx context.Context, req *frontierv1beta1.CheckResourcePermissionRequest) (*frontierv1beta1.CheckResourcePermissionResponse, error) {
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(req.GetResource())
	if len(req.GetResource()) == 0 || err != nil {
		objectNamespace = schema.ParseNamespaceAliasIfRequired(req.GetObjectNamespace())
		objectID = req.GetObjectId()
	}
	if objectNamespace == "" || objectID == "" {
		return nil, grpcBadBodyError
	}

	permissionName, err := h.getPermissionName(ctx, objectNamespace, req.GetPermission())
	if err != nil {
		return nil, err
	}
	result, err := h.resourceService.CheckAuthz(ctx, resource.Check{
		Object: relation.Object{
			ID:        objectID,
			Namespace: objectNamespace,
		},
		Permission: permissionName,
	})
	if err != nil {
		return nil, handleAuthErr(ctx, err)
	}

	logAuditForCheck(ctx, result, objectID, objectNamespace)
	if !result {
		return &frontierv1beta1.CheckResourcePermissionResponse{Status: false}, nil
	}
	return &frontierv1beta1.CheckResourcePermissionResponse{Status: true}, nil
}

func (h Handler) BatchCheckPermission(ctx context.Context, req *frontierv1beta1.BatchCheckPermissionRequest) (*frontierv1beta1.BatchCheckPermissionResponse, error) {
	checks := make([]resource.Check, 0, len(req.GetBodies()))
	for _, body := range req.GetBodies() {
		objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(body.GetResource())
		if len(body.GetResource()) == 0 || err != nil {
			return nil, grpcBadBodyError
		}

		permissionName, err := h.getPermissionName(ctx, objectNamespace, body.GetPermission())
		if err != nil {
			return nil, err
		}
		checks = append(checks, resource.Check{
			Object: relation.Object{
				ID:        objectID,
				Namespace: objectNamespace,
			},
			Permission: permissionName,
		})
	}
	result, err := h.resourceService.BatchCheck(ctx, checks)
	if err != nil {
		return nil, handleAuthErr(ctx, err)
	}

	pairs := make([]*frontierv1beta1.BatchCheckPermissionResponsePair, 0, len(result))
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

func (h Handler) IsAuthorized(ctx context.Context, object relation.Object, permission string) error {
	if object.Namespace == "" || object.ID == "" {
		return grpcBadBodyError
	}

	logger := grpczap.Extract(ctx)
	currentUser, principalErr := h.GetLoggedInPrincipal(ctx)
	if principalErr != nil {
		logger.Error(principalErr.Error())
		return principalErr
	}
	result, err := h.resourceService.CheckAuthz(ctx, resource.Check{
		Object: object,
		Subject: relation.Subject{
			Namespace: currentUser.Type,
			ID:        currentUser.ID,
		},
		Permission: permission,
	})
	if err != nil {
		return handleAuthErr(ctx, err)
	}
	if result {
		return nil
	}

	// for invitation, we need to check if the user is the owner of the invitation by checking its email as well
	if object.Namespace == schema.InvitationNamespace &&
		currentUser.Type == schema.UserPrincipal {
		result2, checkErr := h.resourceService.CheckAuthz(ctx, resource.Check{
			Object: object,
			Subject: relation.Subject{
				Namespace: currentUser.Type,
				ID:        str.GenerateUserSlug(currentUser.User.Email),
			},
			Permission: permission,
		})
		if checkErr != nil {
			return handleAuthErr(ctx, checkErr)
		}
		if result2 {
			return nil
		}
	}

	return grpcPermissionDenied
}

func (h Handler) IsSuperUser(ctx context.Context) error {
	logger := grpczap.Extract(ctx)
	currentUser, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if currentUser.Type == schema.UserPrincipal {
		if ok, err := h.userService.IsSudo(ctx, currentUser.ID, schema.PlatformSudoPermission); err != nil {
			return err
		} else if ok {
			return nil
		}
	} else {
		if ok, err := h.serviceUserService.IsSudo(ctx, currentUser.ID, schema.PlatformSudoPermission); err != nil {
			return err
		} else if ok {
			return nil
		}
	}
	return grpcPermissionDenied
}

func (h Handler) fetchAccessPairsOnResource(ctx context.Context, objectNamespace string, ids, permissions []string) ([]relation.CheckPair, error) {
	checks := make([]resource.Check, 0, len(ids)*len(permissions))
	for _, id := range ids {
		for _, permission := range permissions {
			permissionName, err := h.getPermissionName(ctx, objectNamespace, permission)
			if err != nil {
				return nil, err
			}
			checks = append(checks, resource.Check{
				Object: relation.Object{
					ID:        id,
					Namespace: objectNamespace,
				},
				Permission: permissionName,
			})
		}
	}
	checkPairs, err := h.resourceService.BatchCheck(ctx, checks)
	if err != nil {
		return nil, err
	}
	// remove all the failed checks
	return utils.Filter(checkPairs, func(pair relation.CheckPair) bool {
		return pair.Status
	}), nil
}

func (h Handler) CheckFederatedResourcePermission(ctx context.Context, req *frontierv1beta1.CheckFederatedResourcePermissionRequest) (*frontierv1beta1.CheckFederatedResourcePermissionResponse, error) {
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(req.GetResource())
	if err != nil || objectNamespace == "" || objectID == "" {
		return nil, grpcBadBodyError
	}

	principalNamespace, principalID, err := schema.SplitNamespaceAndResourceID(req.GetSubject())
	if err != nil || principalNamespace == "" || principalID == "" {
		return nil, grpcBadBodyError
	}

	permissionName, err := h.getPermissionName(ctx, objectNamespace, req.GetPermission())
	if err != nil {
		return nil, err
	}
	result, err := h.resourceService.CheckAuthz(ctx, resource.Check{
		Object: relation.Object{
			ID:        objectID,
			Namespace: objectNamespace,
		},
		Subject: relation.Subject{
			ID:        principalID,
			Namespace: principalNamespace,
		},
		Permission: permissionName,
	})
	if err != nil {
		return nil, handleAuthErr(ctx, err)
	}

	logAuditForCheck(ctx, result, objectID, objectNamespace)
	if !result {
		return &frontierv1beta1.CheckFederatedResourcePermissionResponse{Status: false}, nil
	}
	return &frontierv1beta1.CheckFederatedResourcePermissionResponse{Status: true}, nil
}

func handleAuthErr(ctx context.Context, err error) error {
	logger := grpczap.Extract(ctx)
	switch {
	case errors.Is(err, user.ErrInvalidEmail) || errors.Is(err, errors.ErrUnauthenticated):
		return grpcUnauthenticated
	case errors.Is(err, organization.ErrNotExist):
		return status.Errorf(codes.NotFound, err.Error())
	case errors.Is(err, project.ErrNotExist):
		return status.Errorf(codes.NotFound, err.Error())
	default:
		formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
		logger.Error(formattedErr.Error())
		return status.Errorf(codes.Internal, ErrInternalServer.Error())
	}
}
