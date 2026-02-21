package v1beta1connect

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateCurrentUserPersonalToken(ctx context.Context, request *connect.Request[frontierv1beta1.CreateCurrentUserPersonalTokenRequest]) (*connect.Response[frontierv1beta1.CreateCurrentUserPersonalTokenResponse], error) {
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

	expiresAt := request.Msg.GetExpiresAt().AsTime()
	if !expiresAt.After(time.Now()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, userpat.ErrExpiryInPast)
	}
	if expiresAt.After(time.Now().Add(h.patConfig.MaxExpiry())) {
		return nil, connect.NewError(connect.CodeInvalidArgument, userpat.ErrExpiryExceeded)
	}

	created, tokenValue, err := h.userPATService.Create(ctx, userpat.CreateRequest{
		UserID:     principal.User.ID,
		OrgID:      request.Msg.GetOrgId(),
		Title:      request.Msg.GetTitle(),
		Roles:      request.Msg.GetRoles(),
		ProjectIDs: request.Msg.GetProjectIds(),
		ExpiresAt:  request.Msg.GetExpiresAt().AsTime(),
		Metadata:   metadata.BuildFromProto(request.Msg.GetMetadata()),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateCurrentUserPersonalToken", err,
			zap.String("user_id", principal.User.ID),
			zap.String("org_id", request.Msg.GetOrgId()))

		switch {
		case errors.Is(err, userpat.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, userpat.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		case errors.Is(err, userpat.ErrLimitExceeded):
			return nil, connect.NewError(connect.CodeResourceExhausted, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.CreateCurrentUserPersonalTokenResponse{
		Token: transformPATToPB(created, tokenValue),
	}), nil
}

func transformPATToPB(pat userpat.PersonalAccessToken, tokenValue string) *frontierv1beta1.PersonalAccessToken {
	pbPAT := &frontierv1beta1.PersonalAccessToken{
		Id:        pat.ID,
		Title:     pat.Title,
		UserId:    pat.UserID,
		OrgId:     pat.OrgID,
		ExpiresAt: timestamppb.New(pat.ExpiresAt),
		CreatedAt: timestamppb.New(pat.CreatedAt),
		UpdatedAt: timestamppb.New(pat.UpdatedAt),
	}
	if tokenValue != "" {
		pbPAT.Token = tokenValue
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
	return pbPAT
}
