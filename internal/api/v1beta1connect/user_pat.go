package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/userpat"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// getLoggedInPrincipalWithUser returns the authenticated principal, rejecting service users.
// PAT principals are allowed through since principal.User is populated for PATs.
// Use this for PAT management endpoints where the PAT owner's user is needed.
func (h *ConnectHandler) getLoggedInPrincipalWithUser(ctx context.Context) (*authenticate.Principal, error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.User == nil {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated)
	}
	return &principal, nil
}

// mapPATError maps PAT service errors to Connect RPC error codes.
func mapPATError(err error) *connect.Error {
	switch {
	case errors.Is(err, paterrors.ErrDisabled):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, paterrors.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, paterrors.ErrConflict):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, paterrors.ErrLimitExceeded):
		return connect.NewError(connect.CodeResourceExhausted, err)
	case errors.Is(err, paterrors.ErrRoleNotFound),
		errors.Is(err, paterrors.ErrDeniedRole),
		errors.Is(err, paterrors.ErrUnsupportedScope),
		errors.Is(err, paterrors.ErrScopeMismatch),
		errors.Is(err, paterrors.ErrProjectForbidden),
		errors.Is(err, paterrors.ErrExpiryInPast),
		errors.Is(err, paterrors.ErrExpiryExceeded):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
}

func (h *ConnectHandler) CreateCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.CreateCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.CreateCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.Type != schema.UserPrincipal {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated)
	}

	if err := h.userPATService.ValidateExpiry(request.Msg.GetExpiresAt().AsTime()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	created, patValue, err := h.userPATService.Create(ctx, userpat.CreateRequest{
		UserID:    principal.User.ID,
		OrgID:     request.Msg.GetOrgId(),
		Title:     request.Msg.GetTitle(),
		Scopes:    protoScopesToModel(request.Msg.GetScopes()),
		ExpiresAt: request.Msg.GetExpiresAt().AsTime(),
		Metadata:  metadata.BuildFromProto(request.Msg.GetMetadata()),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateCurrentUserPAT", err,
			"user_id", principal.User.ID,
			"org_id", request.Msg.GetOrgId())
		return nil, mapPATError(err)
	}

	return connect.NewResponse(&frontierv1beta1.CreateCurrentUserPATResponse{
		Pat: transformPATToPB(created, patValue),
	}), nil
}

func (h *ConnectHandler) GetCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.GetCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.GetCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.getLoggedInPrincipalWithUser(ctx)
	if err != nil {
		return nil, err
	}

	pat, err := h.userPATService.Get(ctx, principal.User.ID, request.Msg.GetId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetCurrentUserPAT", err,
			"user_id", principal.User.ID,
			"pat_id", request.Msg.GetId())
		return nil, mapPATError(err)
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
		return nil, mapPATError(err)
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

	principal, err := h.getLoggedInPrincipalWithUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.userPATService.Delete(ctx, principal.User.ID, request.Msg.GetId()); err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteCurrentUserPAT", err,
			"user_id", principal.User.ID,
			"pat_id", request.Msg.GetId())
		return nil, mapPATError(err)
	}

	return connect.NewResponse(&frontierv1beta1.DeleteCurrentUserPATResponse{}), nil
}

func (h *ConnectHandler) UpdateCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.UpdateCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.getLoggedInPrincipalWithUser(ctx)
	if err != nil {
		return nil, err
	}

	updated, err := h.userPATService.Update(ctx, models.PAT{
		UserID:   principal.User.ID,
		ID:       request.Msg.GetId(),
		Title:    request.Msg.GetTitle(),
		Scopes:   protoScopesToModel(request.Msg.GetScopes()),
		Metadata: metadata.BuildFromProto(request.Msg.GetMetadata()),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "UpdateCurrentUserPAT", err,
			"user_id", principal.User.ID,
			"pat_id", request.Msg.GetId())
		return nil, mapPATError(err)
	}

	return connect.NewResponse(&frontierv1beta1.UpdateCurrentUserPATResponse{
		Pat: transformPATToPB(updated, ""),
	}), nil
}

func (h *ConnectHandler) RegenerateCurrentUserPAT(ctx context.Context, request *connect.Request[frontierv1beta1.RegenerateCurrentUserPATRequest]) (*connect.Response[frontierv1beta1.RegenerateCurrentUserPATResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.getLoggedInPrincipalWithUser(ctx)
	if err != nil {
		return nil, err
	}

	if request.Msg.GetExpiresAt() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("expires_at is required"))
	}

	regenerated, patValue, err := h.userPATService.Regenerate(ctx, principal.User.ID, request.Msg.GetId(), request.Msg.GetExpiresAt().AsTime())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "RegenerateCurrentUserPAT", err,
			"user_id", principal.User.ID,
			"pat_id", request.Msg.GetId())
		return nil, mapPATError(err)
	}

	return connect.NewResponse(&frontierv1beta1.RegenerateCurrentUserPATResponse{
		Pat: transformPATToPB(regenerated, patValue),
	}), nil
}

func (h *ConnectHandler) CheckCurrentUserPATTitle(ctx context.Context, request *connect.Request[frontierv1beta1.CheckCurrentUserPATTitleRequest]) (*connect.Response[frontierv1beta1.CheckCurrentUserPATTitleResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.getLoggedInPrincipalWithUser(ctx)
	if err != nil {
		return nil, err
	}

	available, err := h.userPATService.IsTitleAvailable(ctx, principal.User.ID, request.Msg.GetOrgId(), request.Msg.GetTitle())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CheckCurrentUserPATTitle", err,
			"user_id", principal.User.ID,
			"org_id", request.Msg.GetOrgId())
		return nil, mapPATError(err)
	}

	return connect.NewResponse(&frontierv1beta1.CheckCurrentUserPATTitleResponse{
		Available: available,
	}), nil
}

func transformPATToPB(pat models.PAT, patValue string) *frontierv1beta1.PAT {
	pbPAT := &frontierv1beta1.PAT{
		Id:        pat.ID,
		Title:     pat.Title,
		UserId:    pat.UserID,
		OrgId:     pat.OrgID,
		Scopes:    modelScopesToProto(pat.Scopes),
		ExpiresAt: timestamppb.New(pat.ExpiresAt),
		CreatedAt: timestamppb.New(pat.CreatedAt),
		UpdatedAt: timestamppb.New(pat.UpdatedAt),
	}
	if patValue != "" {
		pbPAT.Token = patValue
	}
	if pat.UsedAt != nil {
		pbPAT.UsedAt = timestamppb.New(*pat.UsedAt)
	}
	if pat.RegeneratedAt != nil {
		pbPAT.RegeneratedAt = timestamppb.New(*pat.RegeneratedAt)
	}
	if pat.Metadata != nil {
		metaPB, err := pat.Metadata.ToStructPB()
		if err == nil {
			pbPAT.Metadata = metaPB
		}
	}
	return pbPAT
}

func protoScopesToModel(pbScopes []*frontierv1beta1.PATScope) []models.PATScope {
	scopes := make([]models.PATScope, 0, len(pbScopes))
	for _, s := range pbScopes {
		scopes = append(scopes, models.PATScope{
			RoleID:       s.GetRoleId(),
			ResourceType: s.GetResourceType(),
			ResourceIDs:  s.GetResourceIds(),
		})
	}
	return scopes
}

func modelScopesToProto(scopes []models.PATScope) []*frontierv1beta1.PATScope {
	pbScopes := make([]*frontierv1beta1.PATScope, 0, len(scopes))
	for _, s := range scopes {
		pbScopes = append(pbScopes, &frontierv1beta1.PATScope{
			RoleId:       s.RoleID,
			ResourceType: s.ResourceType,
			ResourceIds:  s.ResourceIDs,
		})
	}
	return pbScopes
}

func (h *ConnectHandler) SearchCurrentUserPATs(ctx context.Context, request *connect.Request[frontierv1beta1.SearchCurrentUserPATsRequest]) (*connect.Response[frontierv1beta1.SearchCurrentUserPATsResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	userID, _ := principal.ResolveSubject()

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), models.PAT{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to transform rql query: %v", err))
	}
	if err = rql.ValidateQuery(rqlQuery, models.PAT{}); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	result, err := h.userPATService.List(ctx, userID, request.Msg.GetOrgId(), rqlQuery)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "SearchCurrentUserPATs", err,
			"user_id", userID,
			"org_id", request.Msg.GetOrgId())

		return nil, mapPATError(err)
	}

	pbPATs := make([]*frontierv1beta1.PAT, 0, len(result.PATs))
	for _, pat := range result.PATs {
		pbPATs = append(pbPATs, transformPATToPB(pat, ""))
	}

	return connect.NewResponse(&frontierv1beta1.SearchCurrentUserPATsResponse{
		Pats: pbPATs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset:     uint32(result.Page.Offset),
			Limit:      uint32(result.Page.Limit),
			TotalCount: uint32(result.Page.TotalCount),
		},
	}), nil
}
