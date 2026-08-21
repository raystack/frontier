package v1beta1connect

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/deleter"
	"github.com/raystack/frontier/core/organization"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func (h *ConnectHandler) DeleteProject(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteProjectRequest]) (*connect.Response[frontierv1beta1.DeleteProjectResponse], error) {
	if err := h.deleterService.DeleteProject(ctx, request.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteProject.DeleteProject: project_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.DeleteProjectResponse{}), nil
}

func (h *ConnectHandler) DeleteOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteOrganizationRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationResponse], error) {
	if err := h.deleterService.DeleteOrganization(ctx, request.Msg.GetId()); err != nil {
		var blocked *deleter.BlockedError
		if errors.As(err, &blocked) {
			return nil, deleteBlockedError(ctx, blocked)
		}
		if errors.Is(err, organization.ErrNotExist) || errors.Is(err, organization.ErrInvalidUUID) || errors.Is(err, organization.ErrInvalidID) {
			return nil, connect.NewError(connect.CodeNotFound, organization.ErrNotExist)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteOrganization.DeleteOrganization: organization_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationResponse{}), nil
}

// deleteBlockedError turns the delete's blockers into a failed_precondition
// error carrying one PreconditionFailure violation per blocker, so a caller
// sees everything to fix in a single response.
func deleteBlockedError(ctx context.Context, blocked *deleter.BlockedError) *connect.Error {
	connectErr := connect.NewError(connect.CodeFailedPrecondition, blocked)
	failure := &errdetails.PreconditionFailure{}
	for _, b := range blocked.Blockers {
		failure.Violations = append(failure.Violations, &errdetails.PreconditionFailure_Violation{
			Type:        b.Type,
			Subject:     b.Subject,
			Description: b.Message,
		})
	}
	if detail, err := connect.NewErrorDetail(failure); err != nil {
		slog.WarnContext(ctx, "failed to attach precondition failure details", "error", err, "org_id", blocked.OrgID)
	} else {
		connectErr.AddDetail(detail)
	}
	slog.WarnContext(ctx, "organization delete blocked", "org_id", blocked.OrgID, "blockers", len(blocked.Blockers))
	return connectErr
}
