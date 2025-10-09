package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

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

func (h *ConnectHandler) getPermissionName(ctx context.Context, ns, name string) (string, error) {
	if ns == schema.PlatformNamespace && schema.IsPlatformPermission(name) {
		return name, nil
	}
	perm, err := h.permissionService.Get(ctx, permission.AddNamespaceIfRequired(ns, name))
	if err != nil {
		switch {
		case errors.Is(err, permission.ErrNotExist):
			return "", connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return "", connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	// if the permission is on the same namespace as the object, use the name
	if perm.NamespaceID == ns {
		return perm.Name, nil
	}
	// else use fully qualified name(slug)
	return perm.Slug, nil
}

func (h *ConnectHandler) CheckFederatedResourcePermission(ctx context.Context, req *connect.Request[frontierv1beta1.CheckFederatedResourcePermissionRequest]) (*connect.Response[frontierv1beta1.CheckFederatedResourcePermissionResponse], error) {
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(req.Msg.GetResource())
	if err != nil || objectNamespace == "" || objectID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	principalNamespace, principalID, err := schema.SplitNamespaceAndResourceID(req.Msg.GetSubject())
	if err != nil || principalNamespace == "" || principalID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	permissionName, err := h.getPermissionName(ctx, objectNamespace, req.Msg.GetPermission())
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
		return nil, handleAuthErr(err)
	}

	logAuditForCheck(ctx, result, objectID, objectNamespace)
	if !result {
		return connect.NewResponse(&frontierv1beta1.CheckFederatedResourcePermissionResponse{Status: false}), nil
	}
	return connect.NewResponse(&frontierv1beta1.CheckFederatedResourcePermissionResponse{Status: true}), nil
}

func (h *ConnectHandler) fetchAccessPairsOnResource(ctx context.Context, objectNamespace string, ids, permissions []string) ([]relation.CheckPair, error) {
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	// remove all the failed checks
	return utils.Filter(checkPairs, func(pair relation.CheckPair) bool {
		return pair.Status
	}), nil
}

func (h *ConnectHandler) CheckResourcePermission(ctx context.Context, req *connect.Request[frontierv1beta1.CheckResourcePermissionRequest]) (*connect.Response[frontierv1beta1.CheckResourcePermissionResponse], error) {
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(req.Msg.GetResource())
	if len(req.Msg.GetResource()) == 0 || err != nil {
		objectNamespace = schema.ParseNamespaceAliasIfRequired(req.Msg.GetObjectNamespace())
		objectID = req.Msg.GetObjectId()
	}
	if objectNamespace == "" || objectID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	permissionName, err := h.getPermissionName(ctx, objectNamespace, req.Msg.GetPermission())
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
		return nil, handleAuthErr(err)
	}

	logAuditForCheck(ctx, result, objectID, objectNamespace)
	if !result {
		return connect.NewResponse(&frontierv1beta1.CheckResourcePermissionResponse{Status: false}), nil
	}
	return connect.NewResponse(&frontierv1beta1.CheckResourcePermissionResponse{Status: true}), nil
}

func (h *ConnectHandler) BatchCheckPermission(ctx context.Context, req *connect.Request[frontierv1beta1.BatchCheckPermissionRequest]) (*connect.Response[frontierv1beta1.BatchCheckPermissionResponse], error) {
	checks := make([]resource.Check, 0, len(req.Msg.GetBodies()))
	for _, body := range req.Msg.GetBodies() {
		objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(body.GetResource())
		if len(body.GetResource()) == 0 || err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
	return connect.NewResponse(&frontierv1beta1.BatchCheckPermissionResponse{
		Pairs: pairs,
	}), nil
}
