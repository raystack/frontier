package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var grpcInvitationNotFoundError = status.Error(codes.NotFound, "invitation not found")

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
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	invite, err := h.invitationService.List(ctx, invitation.Filter{
		OrgID:  orgResp.ID,
		UserID: request.GetUserId(),
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

func (h Handler) ListCurrentUserInvitations(ctx context.Context, request *frontierv1beta1.ListCurrentUserInvitationsRequest) (*frontierv1beta1.ListCurrentUserInvitationsResponse, error) {
	logger := grpczap.Extract(ctx)
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if principal.User == nil {
		return nil, status.Errorf(codes.Internal, "invalid user")
	}

	invites, err := h.invitationService.ListByUser(ctx, principal.User.Email)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var invPBs []*frontierv1beta1.Invitation
	var orgIds []string
	for _, inv := range invites {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		invPBs = append(invPBs, pbInv)
		orgIds = append(orgIds, inv.OrgID)
	}

	var orgPBs []*frontierv1beta1.Organization
	for _, org := range orgIds {
		orgResp, err := h.orgService.Get(ctx, org)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, fmt.Errorf("failed to get org: %w", err).Error())
		}
		orgPB, err := transformOrgToPB(orgResp)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, fmt.Errorf("failed to transform org to pb: %w", err).Error())
		}
		orgPBs = append(orgPBs, orgPB)
	}
	return &frontierv1beta1.ListCurrentUserInvitationsResponse{
		Invitations: invPBs,
		Orgs:        orgPBs,
	}, nil
}

func (h Handler) ListUserInvitations(ctx context.Context, request *frontierv1beta1.ListUserInvitationsRequest) (*frontierv1beta1.ListUserInvitationsResponse, error) {
	logger := grpczap.Extract(ctx)
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
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	for _, userID := range request.GetUserIds() {
		if !isValidEmail(userID) {
			logger.Error("invalid email")
			return nil, status.Errorf(codes.InvalidArgument, "invalid email")
		}
	}

	createdInvitations := []invitation.Invitation{}
	for _, userID := range request.GetUserIds() {
		inv, err := h.invitationService.Create(ctx, invitation.Invitation{
			UserID:   userID,
			RoleIDs:  request.GetRoleIds(),
			OrgID:    orgResp.ID,
			GroupIDs: request.GetGroupIds(),
		})
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		createdInvitations = append(createdInvitations, inv)
	}

	var pbInvs []*frontierv1beta1.Invitation
	for _, inv := range createdInvitations {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			logger.Error(err.Error())
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		pbInvs = append(pbInvs, pbInv)
	}
	return &frontierv1beta1.CreateOrganizationInvitationResponse{
		Invitations: pbInvs,
	}, nil
}

func (h Handler) GetOrganizationInvitation(ctx context.Context, request *frontierv1beta1.GetOrganizationInvitationRequest) (*frontierv1beta1.GetOrganizationInvitationResponse, error) {
	logger := grpczap.Extract(ctx)
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

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
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

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
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

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
		RoleIds:   inv.RoleIDs,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(inv.CreatedAt),
		ExpiresAt: timestamppb.New(inv.ExpiresAt),
	}, nil
}
