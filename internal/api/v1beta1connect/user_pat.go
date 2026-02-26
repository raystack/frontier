package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
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
		Roles:      request.Msg.GetRoleIds(),
		ProjectIDs: request.Msg.GetProjectIds(),
		ExpiresAt:  request.Msg.GetExpiresAt().AsTime(),
		Metadata:   metadata.BuildFromProto(request.Msg.GetMetadata()),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateCurrentUserPAT", err,
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

	return connect.NewResponse(&frontierv1beta1.CreateCurrentUserPATResponse{
		Pat: transformPATToPB(created, patValue),
	}), nil
}

func transformPATToPB(pat userpat.PAT, patValue string) *frontierv1beta1.PAT {
	pbPAT := &frontierv1beta1.PAT{
		Id:        pat.ID,
		Title:     pat.Title,
		UserId:    pat.UserID,
		OrgId:     pat.OrgID,
		ExpiresAt: timestamppb.New(pat.ExpiresAt),
		CreatedAt: timestamppb.New(pat.CreatedAt),
		UpdatedAt: timestamppb.New(pat.UpdatedAt),
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
	return pbPAT
}
