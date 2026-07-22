package v1beta1connect

import (
	"context"
	"fmt"
	"log/slog"

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
	if err := audit.GetAuditor(ctx, schema.PlatformOrgID.String()).LogWithAttrs(audit.PermissionCheckedEvent, audit.Target{
		ID:   objectID,
		Type: objectNamespace,
	}, map[string]string{
		"status": auditStatus,
	}); err != nil {
		slog.ErrorContext(ctx, "PermissionCheck.AuditLog operation failed",
			"error", err,
			"object_id", objectID,
			"object_namespace", objectNamespace)
	}
}

func (h *ConnectHandler) getPermissionName(ctx context.Context, ns, name string) (string, error) {
	resolved, ok, err := h.resolvePermissionName(ctx, ns, name)
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, fmt.Errorf("getPermissionName: %w", err))
	}
	if !ok {
		return "", connect.NewError(connect.CodeNotFound, ErrNotFound)
	}
	return resolved, nil
}

// resolvePermissionName looks up the canonical permission name for (namespace, name).
// ok=false means the permission is not defined; err is reserved for genuine lookup
// failures. Callers that want to treat an unknown permission as "no result"
// should use this helper; callers that want to reject the request should use
// getPermissionName which maps unknown permissions to CodeNotFound.
func (h *ConnectHandler) resolvePermissionName(ctx context.Context, ns, name string) (string, bool, error) {
	if ns == schema.PlatformNamespace && schema.IsPlatformPermission(name) {
		return name, true, nil
	}
	perm, err := h.permissionService.Get(ctx, permission.AddNamespaceIfRequired(ns, name))
	if err != nil {
		if errors.Is(err, permission.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if perm.NamespaceID == ns {
		return perm.Name, true, nil
	}
	return perm.Slug, true, nil
}

func (h *ConnectHandler) CheckFederatedResourcePermission(ctx context.Context, req *connect.Request[frontierv1beta1.CheckFederatedResourcePermissionRequest]) (*connect.Response[frontierv1beta1.CheckFederatedResourcePermissionResponse], error) {
	errorLogger := NewErrorLogger()

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
		errorLogger.LogServiceError(ctx, req, "CheckFederatedResourcePermission", err,
			"object_id", objectID,
			"object_namespace", objectNamespace,
			"subject_id", principalID,
			"subject_namespace", principalNamespace,
			"permission", permissionName)
		return nil, handleAuthErr(err)
	}

	logAuditForCheck(ctx, result, objectID, objectNamespace)
	if !result {
		return connect.NewResponse(&frontierv1beta1.CheckFederatedResourcePermissionResponse{Status: false}), nil
	}
	return connect.NewResponse(&frontierv1beta1.CheckFederatedResourcePermissionResponse{Status: true}), nil
}

func (h *ConnectHandler) fetchAccessPairsOnResource(ctx context.Context, objectNamespace string, ids, permissions []string) ([]relation.CheckPair, error) {
	// Resolve each requested permission once, dropping unknown names and
	// duplicate inputs. Unknown names produce an empty result rather than
	// 4xx/5xx — see the contract on resolvePermissionName.
	resolvedPerms := make([]string, 0, len(permissions))
	seen := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		resolved, ok, err := h.resolvePermissionName(ctx, objectNamespace, p)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("fetchAccessPairsOnResource: %w", err))
		}
		if !ok {
			continue
		}
		if _, dup := seen[resolved]; dup {
			continue
		}
		seen[resolved] = struct{}{}
		resolvedPerms = append(resolvedPerms, resolved)
	}
	if len(resolvedPerms) == 0 || len(ids) == 0 {
		return []relation.CheckPair{}, nil
	}

	checks := make([]resource.Check, 0, len(ids)*len(resolvedPerms))
	for _, id := range ids {
		for _, p := range resolvedPerms {
			checks = append(checks, resource.Check{
				Object: relation.Object{
					ID:        id,
					Namespace: objectNamespace,
				},
				Permission: p,
			})
		}
	}
	checkPairs, err := h.resourceService.BatchCheck(ctx, checks)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("fetchAccessPairsOnResource: %w", err))
	}
	// remove all the failed checks
	return utils.Filter(checkPairs, func(pair relation.CheckPair) bool {
		return pair.Status
	}), nil
}

func (h *ConnectHandler) CheckResourcePermission(ctx context.Context, req *connect.Request[frontierv1beta1.CheckResourcePermissionRequest]) (*connect.Response[frontierv1beta1.CheckResourcePermissionResponse], error) {
	errorLogger := NewErrorLogger()

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
		errorLogger.LogServiceError(ctx, req, "CheckResourcePermission", err,
			"object_id", objectID,
			"object_namespace", objectNamespace,
			"permission", permissionName)
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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("BatchCheckPermission: batch_size=%d: %w", len(checks), err))
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
