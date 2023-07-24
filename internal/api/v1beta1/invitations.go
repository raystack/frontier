package v1beta1

import (
	"context"
	"errors"

	"github.com/google/uuid"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/user"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var grpcInvitationNotFoundError = status.Error(codes.NotFound, "invitation not found")

//go:generate mockery --name=InvitationService -r --case underscore --with-expecter --structname InvitationService --filename invitation_service.go --output=./mocks
type InvitationService interface {
	Get(ctx context.Context, id uuid.UUID) (invitation.Invitation, error)
	List(ctx context.Context, filter invitation.Filter) ([]invitation.Invitation, error)
	ListByUser(ctx context.Context, userID string) ([]invitation.Invitation, error)
	Create(ctx context.Context, inv invitation.Invitation) (invitation.Invitation, error)
	Accept(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

func (h Handler) ListOrganizationInvitations(ctx context.Context, request *frontierv1beta1.ListOrganizationInvitationsRequest) (*frontierv1beta1.ListOrganizationInvitationsResponse, error) {
	logger := grpczap.Extract(ctx)
	userID := request.GetUserId()
	if userID != "" && !isValidEmail(request.GetUserId()) {
		logger.Error("invalid email")
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	orgID, err := uuid.Parse(request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	invite, err := h.invitationService.List(ctx, invitation.Filter{
		OrgID:  orgID.String(),
		UserID: userID,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var pbinvs []*frontierv1beta1.Invitation
	for _, inv := range invite {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		pbinvs = append(pbinvs, pbInv)
	}
	return &frontierv1beta1.ListOrganizationInvitationsResponse{
		Invitations: pbinvs,
	}, nil
}

func (h Handler) ListUserInvitations(ctx context.Context, request *frontierv1beta1.ListUserInvitationsRequest) (*frontierv1beta1.ListUserInvitationsResponse, error) {
	logger := grpczap.Extract(ctx)
	if !isValidEmail(request.GetId()) {
		logger.Error("invalid email")
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}
	invite, err := h.invitationService.ListByUser(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var pbinvs []*frontierv1beta1.Invitation
	for _, inv := range invite {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		pbinvs = append(pbinvs, pbInv)
	}
	return &frontierv1beta1.ListUserInvitationsResponse{
		Invitations: pbinvs,
	}, nil
}

func (h Handler) CreateOrganizationInvitation(ctx context.Context, request *frontierv1beta1.CreateOrganizationInvitationRequest) (*frontierv1beta1.CreateOrganizationInvitationResponse, error) {
	logger := grpczap.Extract(ctx)
	if !isValidEmail(request.GetUserId()) {
		logger.Error("invalid email")
		return nil, status.Errorf(codes.InvalidArgument, "invalid email")
	}

	inv, err := h.invitationService.Create(ctx, invitation.Invitation{
		UserID:   request.GetUserId(),
		OrgID:    request.GetOrgId(),
		GroupIDs: request.GetGroupIds(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	pbInv, err := transformInvitationToPB(inv)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.CreateOrganizationInvitationResponse{
		Invitation: pbInv,
	}, nil
}

func (h Handler) GetOrganizationInvitation(ctx context.Context, request *frontierv1beta1.GetOrganizationInvitationRequest) (*frontierv1beta1.GetOrganizationInvitationResponse, error) {
	logger := grpczap.Extract(ctx)
	inviteID, err := uuid.Parse(request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	inv, err := h.invitationService.Get(ctx, inviteID)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	pbInv, err := transformInvitationToPB(inv)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.GetOrganizationInvitationResponse{
		Invitation: pbInv,
	}, nil
}

func (h Handler) AcceptOrganizationInvitation(ctx context.Context, request *frontierv1beta1.AcceptOrganizationInvitationRequest) (*frontierv1beta1.AcceptOrganizationInvitationResponse, error) {
	logger := grpczap.Extract(ctx)
	inviteID, err := uuid.Parse(request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	if err := h.invitationService.Accept(ctx, inviteID); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, invitation.ErrNotFound):
			return nil, grpcInvitationNotFoundError
		case errors.Is(err, user.ErrNotExist):
			return nil, grpcUserNotFoundError
		default:
			return nil, grpcInternalServerError
		}
	}
	return &frontierv1beta1.AcceptOrganizationInvitationResponse{}, nil
}

func (h Handler) DeleteOrganizationInvitation(ctx context.Context, request *frontierv1beta1.DeleteOrganizationInvitationRequest) (*frontierv1beta1.DeleteOrganizationInvitationResponse, error) {
	logger := grpczap.Extract(ctx)
	inviteID, err := uuid.Parse(request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	if err := h.invitationService.Delete(ctx, inviteID); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.DeleteOrganizationInvitationResponse{}, nil
}

func transformInvitationToPB(inv invitation.Invitation) (*frontierv1beta1.Invitation, error) {
	metaData, err := inv.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Invitation{
		Id:        inv.ID.String(),
		UserId:    inv.UserID,
		OrgId:     inv.OrgID,
		GroupIds:  inv.GroupIDs,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(inv.CreatedAt),
		ExpiresAt: timestamppb.New(inv.ExpiresAt),
	}, nil
}
