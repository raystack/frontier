package v1beta1connect

import (
	"context"
	"net/mail"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListAllUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllUsersRequest]) (*connect.Response[frontierv1beta1.ListAllUsersResponse], error) {
	var users []*frontierv1beta1.User
	usersList, err := h.userService.List(ctx, user.Filter{
		Limit:   request.Msg.GetPageSize(),
		Page:    request.Msg.GetPageNum(),
		Keyword: request.Msg.GetKeyword(),
		OrgID:   request.Msg.GetOrgId(),
		GroupID: request.Msg.GetGroupId(),
		State:   user.State(request.Msg.GetState()),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	for _, user := range usersList {
		userPB, err := transformUserToPB(user)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		users = append(users, userPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListAllUsersResponse{
		Count: int32(len(users)),
		Users: users,
	}), nil
}

func (h *ConnectHandler) GetCurrentAdminUser(ctx context.Context, request *connect.Request[frontierv1beta1.GetCurrentAdminUserRequest]) (*connect.Response[frontierv1beta1.GetCurrentAdminUserResponse], error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if principal.Type == schema.ServiceUserPrincipal {
		serviceUserPB, err := transformServiceUserToPB(*principal.ServiceUser)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		return connect.NewResponse(&frontierv1beta1.GetCurrentAdminUserResponse{
			ServiceUser: serviceUserPB,
		}), nil
	}

	userPB, err := transformUserToPB(*principal.User)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.GetCurrentAdminUserResponse{
		User: userPB,
	}), nil
}

func (h *ConnectHandler) ListUsers(context.Context, *connect.Request[frontierv1beta1.ListUsersRequest]) (*connect.Response[frontierv1beta1.ListUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func isValidEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}

func transformUserToPB(usr user.User) (*frontierv1beta1.User, error) {
	metaData, err := usr.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.User{
		Id:        usr.ID,
		Title:     usr.Title,
		Email:     usr.Email,
		Name:      usr.Name,
		Metadata:  metaData,
		Avatar:    usr.Avatar,
		State:     usr.State.String(),
		CreatedAt: timestamppb.New(usr.CreatedAt),
		UpdatedAt: timestamppb.New(usr.UpdatedAt),
	}, nil
}

func (h *ConnectHandler) ListAllServiceUsers(context.Context, *connect.Request[frontierv1beta1.ListAllServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListAllServiceUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
