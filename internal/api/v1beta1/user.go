package v1beta1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/str"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var grpcUserNotFoundError = status.Errorf(codes.NotFound, "user doesn't exist")

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
	Create(ctx context.Context, user user.User) (user.User, error)
	List(ctx context.Context, flt user.Filter) (user.PagedUsers, error)
	UpdateByID(ctx context.Context, toUpdate user.User) (user.User, error)
	UpdateByEmail(ctx context.Context, toUpdate user.User) (user.User, error)
	FetchCurrentUser(ctx context.Context) (user.User, error)
}

var (
	emptyEmailId = errors.New("email id is empty")
)

func (h Handler) ListUsers(ctx context.Context, request *shieldv1beta1.ListUsersRequest) (*shieldv1beta1.ListUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*shieldv1beta1.User

	userResp, err := h.userService.List(ctx, user.Filter{
		Limit:   request.GetPageSize(),
		Page:    request.GetPageNum(),
		Keyword: request.GetKeyword(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	userList := userResp.Users
	for _, user := range userList {
		userPB, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		users = append(users, &userPB)
	}

	return &shieldv1beta1.ListUsersResponse{
		Count: userResp.Count,
		Users: users,
	}, nil
}

func (h Handler) CreateUser(ctx context.Context, request *shieldv1beta1.CreateUserRequest) (*shieldv1beta1.CreateUserResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	currentUserEmail, _ := fetchEmailFromMetadata(ctx, h.identityProxyHeader)
	if len(currentUserEmail) == 0 {
		logger.Error(emptyEmailId.Error())
		return nil, emptyEmailId
	}
	email := str.DefaultStringIfEmpty(request.GetBody().Email, currentUserEmail)
	userT := user.User{
		Name:     request.GetBody().Name,
		Email:    email,
		Metadata: metaDataMap,
	}
	newUser, err := h.userService.Create(ctx, userT)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	metaData, err := structpb.NewStruct(mapOfInterfaceValues(newUser.Metadata))
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateUserResponse{User: &shieldv1beta1.User{
		Id:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newUser.CreatedAt),
		UpdatedAt: timestamppb.New(newUser.UpdatedAt),
	}}, nil
}

func (h Handler) GetUser(ctx context.Context, request *shieldv1beta1.GetUserRequest) (*shieldv1beta1.GetUserResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedUser, err := h.userService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "user not found")
		case errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	userPB, err := transformUserToPB(fetchedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, internalServerError.Error())
	}

	return &shieldv1beta1.GetUserResponse{
		User: &userPB,
	}, nil
}

func (h Handler) GetCurrentUser(ctx context.Context, request *shieldv1beta1.GetCurrentUserRequest) (*shieldv1beta1.GetCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)

	email, err := fetchEmailFromMetadata(ctx, h.identityProxyHeader)
	if err != nil {
		return nil, grpcBadBodyError
	}
	if len(email) == 0 {
		logger.Error(emptyEmailId.Error())
		return nil, emptyEmailId
	}

	fetchedUser, err := h.userService.GetByEmail(ctx, email)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "user not found")
		case errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	userPB, err := transformUserToPB(fetchedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetCurrentUserResponse{
		User: &userPB,
	}, nil
}

func (h Handler) UpdateUser(ctx context.Context, request *shieldv1beta1.UpdateUserRequest) (*shieldv1beta1.UpdateUserResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedUser, err := h.userService.UpdateByID(ctx, user.User{
		ID:       request.GetId(),
		Name:     request.GetBody().Name,
		Email:    request.GetBody().Email,
		Metadata: metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist):
			return nil, grpcUserNotFoundError
		case errors.Is(err, user.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, internalServerError
	}

	return &shieldv1beta1.UpdateUserResponse{User: &userPB}, nil
}

func (h Handler) UpdateCurrentUser(ctx context.Context, request *shieldv1beta1.UpdateCurrentUserRequest) (*shieldv1beta1.UpdateCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)

	email, err := fetchEmailFromMetadata(ctx, h.identityProxyHeader)
	if err != nil {
		return nil, grpcBadBodyError
	}

	if request.Body == nil {
		return nil, grpcBadBodyError
	}
	if len(email) == 0 {
		logger.Error(emptyEmailId.Error())
		return nil, emptyEmailId
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	// if email in request body is different from the email in the header
	if request.GetBody().Email != email {
		return nil, grpcBadBodyError
	}

	updatedUser, err := h.userService.UpdateByEmail(ctx, user.User{
		Name:     request.GetBody().Name,
		Email:    email,
		Metadata: metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist):
			return nil, grpcUserNotFoundError
		default:
			return nil, grpcInternalServerError
		}
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateCurrentUserResponse{User: &userPB}, nil
}

func transformUserToPB(user user.User) (shieldv1beta1.User, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(user.Metadata))
	if err != nil {
		return shieldv1beta1.User{}, err
	}

	return shieldv1beta1.User{
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (h Handler) ListUserGroups(ctx context.Context, request *shieldv1beta1.ListUserGroupsRequest) (*shieldv1beta1.ListUserGroupsResponse, error) {
	logger := grpczap.Extract(ctx)
	var groups []*shieldv1beta1.Group
	groupsList, err := h.groupService.ListUserGroups(ctx, request.Id, request.Role)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, group := range groupsList {
		groupPB, err := transformGroupToPB(group)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		groups = append(groups, &groupPB)
	}

	return &shieldv1beta1.ListUserGroupsResponse{
		Groups: groups,
	}, nil
}

func fetchEmailFromMetadata(ctx context.Context, headerKey string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", grpcBadBodyError
	}

	var email string
	metadataValues := md.Get(headerKey)
	if len(metadataValues) > 0 {
		email = metadataValues[0]
	}
	return email, nil
}
