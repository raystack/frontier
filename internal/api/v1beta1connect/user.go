package v1beta1connect

import (
	"context"
	"fmt"
	"net/mail"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, user := range usersList {
		userPB, err := transformUserToPB(user)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if principal.Type == schema.ServiceUserPrincipal {
		serviceUserPB, err := transformServiceUserToPB(*principal.ServiceUser)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		return connect.NewResponse(&frontierv1beta1.GetCurrentAdminUserResponse{
			ServiceUser: serviceUserPB,
		}), nil
	}

	userPB, err := transformUserToPB(*principal.User)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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

func (h *ConnectHandler) ExportUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ExportUsersRequest], stream *connect.ServerStream[httpbody.HttpBody]) error {
	userDataBytes, contentType, err := h.userService.Export(ctx)
	if err != nil {
		if errors.Is(err, user.ErrNoUsersFound) {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("no data to export: %v", err))
		}
		return connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return streamBytesInChunks(userDataBytes, contentType, stream)
}

func (h *ConnectHandler) SearchUsers(ctx context.Context, request *connect.Request[frontierv1beta1.SearchUsersRequest]) (*connect.Response[frontierv1beta1.SearchUsersResponse], error) {
	var users []*frontierv1beta1.User

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), user.User{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, user.User{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	userData, err := h.userService.Search(ctx, rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range userData.Users {
		transformedUser, err := transformUserToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		users = append(users, transformedUser)
	}

	groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
	for _, groupItem := range userData.Group.Data {
		groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
			Name:  groupItem.Name,
			Count: uint32(groupItem.Count),
		})
	}

	return connect.NewResponse(&frontierv1beta1.SearchUsersResponse{
		Users: users,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(userData.Pagination.Offset),
			Limit:  uint32(userData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: userData.Group.Name,
			Data: groupResponse,
		},
	}), nil
}
