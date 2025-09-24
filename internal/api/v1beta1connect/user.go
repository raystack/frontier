package v1beta1connect

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"connectrpc.com/connect"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/str"
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

func (h *ConnectHandler) CreateUser(ctx context.Context, request *connect.Request[frontierv1beta1.CreateUserRequest]) (*connect.Response[frontierv1beta1.CreateUserResponse], error) {
	logger := grpczap.Extract(ctx)
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	email := strings.TrimSpace(request.Msg.GetBody().GetEmail())
	if email == "" {
		currentUserEmail, ok := authenticate.GetEmailFromContext(ctx)
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		currentUserEmail = strings.TrimSpace(currentUserEmail)
		if currentUserEmail == "" {
			logger.Error(ErrEmptyEmailID.Error())
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		email = currentUserEmail
	}
	title := request.Msg.GetBody().GetTitle()
	name := strings.TrimSpace(request.Msg.GetBody().GetName())
	if name == "" {
		name = str.GenerateUserSlug(email)
	}
	var metaDataMap metadata.Metadata
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
		if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
		}
	}
	// TODO might need to check the valid email form
	newUser, err := h.userService.Create(ctx, user.User{
		Title:    title,
		Email:    email,
		Name:     name,
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		case errors.Is(errors.Unwrap(err), user.ErrKeyDoesNotExists):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	transformedUser, err := transformUserToPB(newUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).
		LogWithAttrs(audit.UserCreatedEvent, audit.UserTarget(newUser.ID), map[string]string{
			"email":  newUser.Email,
			"name":   newUser.Name,
			"title":  newUser.Title,
			"avatar": newUser.Avatar,
		})
	return connect.NewResponse(&frontierv1beta1.CreateUserResponse{User: transformedUser}), nil
}

func (h *ConnectHandler) GetUser(ctx context.Context, request *connect.Request[frontierv1beta1.GetUserRequest]) (*connect.Response[frontierv1beta1.GetUserResponse], error) {
	fetchedUser, err := h.userService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidUUID), errors.Is(err, user.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	userPB, err := transformUserToPB(fetchedUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetUserResponse{
		User: userPB,
	}), nil
}

func (h *ConnectHandler) GetCurrentUser(ctx context.Context, request *connect.Request[frontierv1beta1.GetCurrentUserRequest]) (*connect.Response[frontierv1beta1.GetCurrentUserResponse], error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if principal.Type == schema.ServiceUserPrincipal {
		serviceUserPB, err := transformServiceUserToPB(*principal.ServiceUser)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		return connect.NewResponse(&frontierv1beta1.GetCurrentUserResponse{
			Serviceuser: serviceUserPB,
		}), nil
	}

	userPB, err := transformUserToPB(*principal.User)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.GetCurrentUserResponse{
		User: userPB,
	}), nil
}

func (h *ConnectHandler) UpdateUser(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateUserRequest]) (*connect.Response[frontierv1beta1.UpdateUserResponse], error) {
	auditor := audit.GetAuditor(ctx, schema.PlatformOrgID.String())
	var updatedUser user.User

	if strings.TrimSpace(request.Msg.GetId()) == "" {
		return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
	}

	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	email := strings.TrimSpace(request.Msg.GetBody().GetEmail())
	if email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	var err error
	id := request.Msg.GetId()
	// upsert by email
	if isValidEmail(id) {
		_, err = h.userService.GetByEmail(ctx, id)
		if err != nil {
			if errors.Is(err, user.ErrNotExist) {
				createUserResponse, err := h.CreateUser(ctx, connect.NewRequest(&frontierv1beta1.CreateUserRequest{Body: request.Msg.GetBody()}))
				if err != nil {
					return nil, err
				}
				return connect.NewResponse(&frontierv1beta1.UpdateUserResponse{User: createUserResponse.Msg.GetUser()}), nil
			} else {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}
		}
		// if email in request body is different from that of user getting updated
		if email != id {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrEmailConflict)
		}
	}

	updatedUser, err = h.userService.Update(ctx, user.User{
		ID:       request.Msg.GetId(),
		Title:    request.Msg.GetBody().GetTitle(),
		Email:    request.Msg.GetBody().GetEmail(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Name:     request.Msg.GetBody().GetName(),
		Metadata: metaDataMap,
	})

	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, user.ErrInvalidDetails):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, user.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	auditor.LogWithAttrs(audit.UserUpdatedEvent, audit.UserTarget(updatedUser.ID), map[string]string{
		"email":  updatedUser.Email,
		"name":   updatedUser.Name,
		"title":  updatedUser.Title,
		"avatar": updatedUser.Avatar,
	})
	return connect.NewResponse(&frontierv1beta1.UpdateUserResponse{User: userPB}), nil
}

func (h *ConnectHandler) UpdateCurrentUser(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateCurrentUserRequest]) (*connect.Response[frontierv1beta1.UpdateCurrentUserResponse], error) {
	auditor := audit.GetAuditor(ctx, schema.PlatformOrgID.String())
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}

	// if email in request body is different from the email in the header
	if principal.User != nil && principal.User.Email != request.Msg.GetBody().GetEmail() {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	var metaDataMap metadata.Metadata
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
		if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
		}
	}

	updatedUser, err := h.userService.Update(ctx, user.User{
		ID:       principal.ID,
		Title:    request.Msg.GetBody().GetTitle(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Name:     request.Msg.GetBody().GetName(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, user.ErrInvalidDetails):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	auditor.LogWithAttrs(audit.UserUpdatedEvent, audit.UserTarget(updatedUser.ID), map[string]string{
		"email":  updatedUser.Email,
		"name":   updatedUser.Name,
		"title":  updatedUser.Title,
		"avatar": updatedUser.Avatar,
	})
	return connect.NewResponse(&frontierv1beta1.UpdateCurrentUserResponse{User: userPB}), nil
}

func (h *ConnectHandler) EnableUser(ctx context.Context, request *connect.Request[frontierv1beta1.EnableUserRequest]) (*connect.Response[frontierv1beta1.EnableUserResponse], error) {
	if err := h.userService.Enable(ctx, request.Msg.GetId()); err != nil {
		switch {
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, user.ErrInvalidID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	return connect.NewResponse(&frontierv1beta1.EnableUserResponse{}), nil
}

func (h *ConnectHandler) ListUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListUsersRequest]) (*connect.Response[frontierv1beta1.ListUsersResponse], error) {
	auditor := audit.GetAuditor(ctx, request.Msg.GetOrgId())

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

	auditor.Log(audit.UserListedEvent, audit.OrgTarget(request.Msg.GetOrgId()))
	return connect.NewResponse(&frontierv1beta1.ListUsersResponse{
		Count: int32(len(users)),
		Users: users,
	}), nil
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
