package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/userpat"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.CreateCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.CreateCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Type != schema.UserPrincipal {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.userPATService.ValidateExpiry(request.Msg.GetExpiresAt().AsTime()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	created, patValue, err := h.userPATService.Create(ctx, userpat.CreateRequest{
		UserID:     principal.User.ID,
		OrgID:      request.Msg.GetOrgId(),
		Title:      request.Msg.GetTitle(),
		RoleIDs:    request.Msg.GetRoleIds(),
		ProjectIDs: request.Msg.GetProjectIds(),
		ExpiresAt:  request.Msg.GetExpiresAt().AsTime(),
		Metadata:   metadata.BuildFromProto(request.Msg.GetMetadata()),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateCurrentUserPAT", err,
			zap.String("user_id", principal.User.ID),
			zap.String("org_id", request.Msg.GetOrgId()))

		switch {
		case errors.Is(err, paterrors.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, paterrors.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		case errors.Is(err, paterrors.ErrLimitExceeded):
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		case errors.Is(err, paterrors.ErrRoleNotFound):
			return nil, connect.NewError(connect.CodeInvalidArgument, paterrors.ErrRoleNotFound)
		case errors.Is(err, paterrors.ErrDeniedRole):
			return nil, connect.NewError(connect.CodeInvalidArgument, paterrors.ErrDeniedRole)
		case errors.Is(err, paterrors.ErrUnsupportedScope):
			return nil, connect.NewError(connect.CodeInvalidArgument, paterrors.ErrUnsupportedScope)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.CreateCurrentUserPATResponse{
		Pat: transformPATToPB(created, patValue),
	}), nil
}

func (h *ConnectHandler) GetCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.GetCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.GetCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.User == nil {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	pat, err := h.userPATService.Get(ctx, principal.User.ID, request.Msg.GetId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetCurrentUserPAT", err,
			zap.String("user_id", principal.User.ID),
			zap.String("pat_id", request.Msg.GetId()))

		switch {
		case errors.Is(err, paterrors.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, paterrors.ErrNotFound):
			return nil, connect.NewError(connect.CodeNotFound, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.GetCurrentUserPATResponse{
		Pat: transformPATToPB(pat, ""),
	}), nil
}

func (h *ConnectHandler) ListRolesForPAT(ctx context.Context, request *connect.Request[frontierv1beta1.ListRolesForPATRequest]) (*connect.Response[frontierv1beta1.ListRolesForPATResponse], error) {
	errorLogger := NewErrorLogger()

	roleList, err := h.userPATService.ListAllowedRoles(ctx, request.Msg.GetScopes())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListRolesForPAT", err)
		switch {
		case errors.Is(err, paterrors.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, paterrors.ErrUnsupportedScope):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	var roles []*frontierv1beta1.Role
	for _, v := range roleList {
		rolePB, err := transformRoleToPB(v)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListRolesForPAT", v.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		roles = append(roles, &rolePB)
	}

	return connect.NewResponse(&frontierv1beta1.ListRolesForPATResponse{Roles: roles}), nil
}

func (h *ConnectHandler) DeleteCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.DeleteCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.User == nil {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.userPATService.Delete(ctx, principal.User.ID, request.Msg.GetId()); err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteCurrentUserPAT", err,
			zap.String("user_id", principal.User.ID),
			zap.String("pat_id", request.Msg.GetId()))

		switch {
		case errors.Is(err, paterrors.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, paterrors.ErrNotFound):
			return nil, connect.NewError(connect.CodeNotFound, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeleteCurrentUserPATResponse{}), nil
}

func (h *ConnectHandler) UpdateCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.UpdateCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.User == nil {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	updated, err := h.userPATService.Update(ctx, models.PAT{
		UserID:     principal.User.ID,
		ID:         request.Msg.GetId(),
		Title:      request.Msg.GetTitle(),
		RoleIDs:    request.Msg.GetRoleIds(),
		ProjectIDs: request.Msg.GetProjectIds(),
		Metadata:   metadata.BuildFromProto(request.Msg.GetMetadata()),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdateCurrentUserPAT", err,
			zap.String("user_id", principal.User.ID),
			zap.String("pat_id", request.Msg.GetId()))

		switch {
		case errors.Is(err, paterrors.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, paterrors.ErrNotFound):
			return nil, connect.NewError(connect.CodeNotFound, err)
		case errors.Is(err, paterrors.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		case errors.Is(err, paterrors.ErrRoleNotFound):
			return nil, connect.NewError(connect.CodeInvalidArgument, paterrors.ErrRoleNotFound)
		case errors.Is(err, paterrors.ErrDeniedRole):
			return nil, connect.NewError(connect.CodeInvalidArgument, paterrors.ErrDeniedRole)
		case errors.Is(err, paterrors.ErrUnsupportedScope):
			return nil, connect.NewError(connect.CodeInvalidArgument, paterrors.ErrUnsupportedScope)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.UpdateCurrentUserPATResponse{
		Pat: transformPATToPB(updated, ""),
	}), nil
}

func (h *ConnectHandler) RegenerateCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.RegenerateCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.RegenerateCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.User == nil {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated)
	}

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if request.Msg.GetExpiresAt() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("expires_at is required"))
	}

	regenerated, patValue, err := h.userPATService.Regenerate(ctx, principal.User.ID, request.Msg.GetId(), request.Msg.GetExpiresAt().AsTime())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "RegenerateCurrentUserPAT", err,
			zap.String("user_id", principal.User.ID),
			zap.String("pat_id", request.Msg.GetId()))

		switch {
		case errors.Is(err, paterrors.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, paterrors.ErrNotFound):
			return nil, connect.NewError(connect.CodeNotFound, err)
		case errors.Is(err, paterrors.ErrLimitExceeded):
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		case errors.Is(err, paterrors.ErrExpiryInPast):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, paterrors.ErrExpiryExceeded):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, paterrors.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.RegenerateCurrentUserPATResponse{
		Pat: transformPATToPB(regenerated, patValue),
	}), nil
}

func transformPATToPB(pat models.PAT, patValue string) *frontierv1beta1.PAT {
	pbPAT := &frontierv1beta1.PAT{
		Id:         pat.ID,
		Title:      pat.Title,
		UserId:     pat.UserID,
		OrgId:      pat.OrgID,
		RoleIds:    pat.RoleIDs,
		ProjectIds: pat.ProjectIDs,
		ExpiresAt:  timestamppb.New(pat.ExpiresAt),
		CreatedAt:  timestamppb.New(pat.CreatedAt),
		UpdatedAt:  timestamppb.New(pat.UpdatedAt),
	}
	if patValue != "" {
		pbPAT.Token = patValue
	}
	if pat.LastUsedAt != nil {
		pbPAT.LastUsedAt = timestamppb.New(*pat.LastUsedAt)
	}
	if pat.Metadata != nil {
		metaPB, err := pat.Metadata.ToStructPB()
		if err == nil {
			pbPAT.Metadata = metaPB
		}
	}
	pbPAT.RoleIds = pat.RoleIDs
	pbPAT.ProjectIds = pat.ProjectIDs
	return pbPAT
}

func (h *ConnectHandler) ListCurrentUserPATs(ctx context.Context, request *connect.Request[frontierv1beta1.ListCurrentUserPATsRequest]) (*connect.Response[frontierv1beta1.ListCurrentUserPATsResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	userID, _ := principal.ResolveSubject()

	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), models.PAT{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to transform rql query: %v", err))
	}
	if err = rql.ValidateQuery(rqlQuery, models.PAT{}); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	result, err := h.userPATService.List(ctx, userID, request.Msg.GetOrgId(), rqlQuery)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListCurrentUserPATs", err,
			zap.String("user_id", userID),
			zap.String("org_id", request.Msg.GetOrgId()))

		switch {
		case errors.Is(err, paterrors.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	pbPATs := make([]*frontierv1beta1.PAT, 0, len(result.PATs))
	for _, pat := range result.PATs {
		pbPATs = append(pbPATs, transformPATToPB(pat, ""))
	}

	return connect.NewResponse(&frontierv1beta1.ListCurrentUserPATsResponse{
		Pats: pbPATs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset:     uint32(result.Page.Offset),
			Limit:      uint32(result.Page.Limit),
			TotalCount: uint32(result.Page.TotalCount),
		},
	}), nil
}
