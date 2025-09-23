package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h *ConnectHandler) DeleteOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteOrganizationRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationResponse], error) {
	if err := h.deleterService.DeleteOrganization(ctx, request.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationResponse{}), nil
}