package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
)

func (h *ConnectHandler) DeleteProject(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteProjectRequest]) (*connect.Response[frontierv1beta1.DeleteProjectResponse], error) {
	errorLogger := NewErrorLogger()

	if err := h.deleterService.DeleteProject(ctx, request.Msg.GetId()); err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteProject.DeleteProject", err,
			zap.String("project_id", request.Msg.GetId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.DeleteProjectResponse{}), nil
}

func (h *ConnectHandler) DeleteOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteOrganizationRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationResponse], error) {
	errorLogger := NewErrorLogger()

	if err := h.deleterService.DeleteOrganization(ctx, request.Msg.GetId()); err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteOrganization.DeleteOrganization", err,
			zap.String("organization_id", request.Msg.GetId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationResponse{}), nil
}
