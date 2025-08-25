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
