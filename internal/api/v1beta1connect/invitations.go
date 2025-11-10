package v1beta1connect

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListOrganizationInvitations(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationInvitationsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationInvitationsResponse], error) {
	errorLogger := NewErrorLogger()

	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogServiceError(ctx, request, "ListOrganizationInvitations.Get", err,
				zap.String("org_id", request.Msg.GetOrgId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	invite, err := h.invitationService.List(ctx, invitation.Filter{
		OrgID:  orgResp.ID,
		UserID: request.Msg.GetUserId(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListOrganizationInvitations.List", err,
			zap.String("org_id", orgResp.ID),
			zap.String("user_id", request.Msg.GetUserId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var pbinvs []*frontierv1beta1.Invitation
	for _, inv := range invite {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListOrganizationInvitations", inv.ID.String(), err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		pbinvs = append(pbinvs, pbInv)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationInvitationsResponse{
		Invitations: pbinvs,
	}), nil
}

func (h *ConnectHandler) ListCurrentUserInvitations(ctx context.Context, request *connect.Request[frontierv1beta1.ListCurrentUserInvitationsRequest]) (*connect.Response[frontierv1beta1.ListCurrentUserInvitationsResponse], error) {
	errorLogger := NewErrorLogger()

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListCurrentUserInvitations.GetLoggedInPrincipal", err)
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}
	if principal.User == nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	invites, err := h.invitationService.ListByUser(ctx, principal.User.Email)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListCurrentUserInvitations.ListByUser", err,
			zap.String("user_email", principal.User.Email))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var invPBs []*frontierv1beta1.Invitation
	var orgIds []string
	for _, inv := range invites {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListCurrentUserInvitations", inv.ID.String(), err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		invPBs = append(invPBs, pbInv)
		orgIds = append(orgIds, inv.OrgID)
	}

	var orgPBs []*frontierv1beta1.Organization
	for _, org := range orgIds {
		orgResp, err := h.orgService.Get(ctx, org)
		if err != nil {
			errorLogger.LogServiceError(ctx, request, "ListCurrentUserInvitations.Get", err,
				zap.String("org_id", org))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		orgPB, err := transformOrgToPB(orgResp)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListCurrentUserInvitations.transformOrgToPB", orgResp.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		orgPBs = append(orgPBs, orgPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListCurrentUserInvitationsResponse{
		Invitations: invPBs,
		Orgs:        orgPBs,
	}), nil
}

func (h *ConnectHandler) ListUserInvitations(ctx context.Context, request *connect.Request[frontierv1beta1.ListUserInvitationsRequest]) (*connect.Response[frontierv1beta1.ListUserInvitationsResponse], error) {
	errorLogger := NewErrorLogger()

	invite, err := h.invitationService.ListByUser(ctx, request.Msg.GetId())
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListUserInvitations.ListByUser", err,
			zap.String("user_id", request.Msg.GetId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var pbinvs []*frontierv1beta1.Invitation
	for _, inv := range invite {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListUserInvitations", inv.ID.String(), err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		pbinvs = append(pbinvs, pbInv)
	}

	return connect.NewResponse(&frontierv1beta1.ListUserInvitationsResponse{
		Invitations: pbinvs,
	}), nil
}

func (h *ConnectHandler) CreateOrganizationInvitation(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationInvitationResponse], error) {
	errorLogger := NewErrorLogger()

	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogServiceError(ctx, request, "CreateOrganizationInvitation.Get", err,
				zap.String("org_id", request.Msg.GetOrgId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	for _, userID := range request.Msg.GetUserIds() {
		if !isValidEmail(userID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidEmail)
		}
	}

	createdInvitations := make([]invitation.Invitation, 0, len(request.Msg.GetUserIds()))
	for _, userID := range request.Msg.GetUserIds() {
		inv, err := h.invitationService.Create(ctx, invitation.Invitation{
			UserEmailID: strings.ToLower(userID),
			RoleIDs:     request.Msg.GetRoleIds(),
			OrgID:       orgResp.ID,
			GroupIDs:    request.Msg.GetGroupIds(),
		})
		if err != nil {
			if errors.Is(err, invitation.ErrAlreadyMember) {
				return nil, connect.NewError(connect.CodeAlreadyExists, ErrAlreadyMember)
			}
			errorLogger.LogServiceError(ctx, request, "CreateOrganizationInvitation.Create", err,
				zap.String("user_email", userID),
				zap.String("org_id", orgResp.ID),
				zap.Strings("role_ids", request.Msg.GetRoleIds()),
				zap.Strings("group_ids", request.Msg.GetGroupIds()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		createdInvitations = append(createdInvitations, inv)
	}

	var pbInvs []*frontierv1beta1.Invitation
	for _, inv := range createdInvitations {
		pbInv, err := transformInvitationToPB(inv)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "CreateOrganizationInvitation", inv.ID.String(), err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		pbInvs = append(pbInvs, pbInv)
	}

	return connect.NewResponse(&frontierv1beta1.CreateOrganizationInvitationResponse{
		Invitations: pbInvs,
	}), nil
}

func (h *ConnectHandler) GetOrganizationInvitation(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.GetOrganizationInvitationResponse], error) {
	errorLogger := NewErrorLogger()

	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogServiceError(ctx, request, "GetOrganizationInvitation.Get", err,
				zap.String("org_id", request.Msg.GetOrgId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	inviteID, err := uuid.Parse(request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	inv, err := h.invitationService.Get(ctx, inviteID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetOrganizationInvitation.Get", err,
			zap.String("invitation_id", request.Msg.GetId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	pbInv, err := transformInvitationToPB(inv)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetOrganizationInvitation", inv.ID.String(), err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetOrganizationInvitationResponse{
		Invitation: pbInv,
	}), nil
}

func (h *ConnectHandler) AcceptOrganizationInvitation(ctx context.Context, request *connect.Request[frontierv1beta1.AcceptOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.AcceptOrganizationInvitationResponse], error) {
	errorLogger := NewErrorLogger()

	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogServiceError(ctx, request, "AcceptOrganizationInvitation.Get", err,
				zap.String("org_id", request.Msg.GetOrgId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	inviteID, err := uuid.Parse(request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	if err := h.invitationService.Accept(ctx, inviteID); err != nil {
		switch {
		case errors.Is(err, invitation.ErrInviteExpired):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvitationExpired)
		case errors.Is(err, invitation.ErrNotFound):
			return nil, connect.NewError(connect.CodeNotFound, ErrInvitationNotFound)
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		default:
			errorLogger.LogServiceError(ctx, request, "AcceptOrganizationInvitation.Accept", err,
				zap.String("invitation_id", request.Msg.GetId()),
				zap.String("org_id", request.Msg.GetOrgId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.AcceptOrganizationInvitationResponse{}), nil
}

func (h *ConnectHandler) DeleteOrganizationInvitation(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationInvitationResponse], error) {
	errorLogger := NewErrorLogger()

	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogServiceError(ctx, request, "DeleteOrganizationInvitation.Get", err,
				zap.String("org_id", request.Msg.GetOrgId()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	inviteID, err := uuid.Parse(request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	if err := h.invitationService.Delete(ctx, inviteID); err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteOrganizationInvitation.Delete", err,
			zap.String("invitation_id", request.Msg.GetId()),
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationInvitationResponse{}), nil
}

func transformInvitationToPB(inv invitation.Invitation) (*frontierv1beta1.Invitation, error) {
	metaData, err := inv.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Invitation{
		Id:        inv.ID.String(),
		UserId:    inv.UserEmailID,
		OrgId:     inv.OrgID,
		GroupIds:  inv.GroupIDs,
		RoleIds:   inv.RoleIDs,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(inv.CreatedAt),
		ExpiresAt: timestamppb.New(inv.ExpiresAt),
	}, nil
}
