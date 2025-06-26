package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h *ConnectHandler) ListUsers(context.Context, *connect.Request[frontierv1beta1.ListUsersRequest]) (*connect.Response[frontierv1beta1.ListUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
