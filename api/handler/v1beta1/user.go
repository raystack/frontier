package v1beta1

import (
	"context"
	"errors"
	"github.com/odpf/shield/pkg/utils"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/internal/user"
	"github.com/odpf/shield/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

type UserService interface {
	GetUser(ctx context.Context, id string) (model.User, error)
	GetCurrentUser(ctx context.Context, email string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	ListUsers(ctx context.Context) ([]model.User, error)
	UpdateUser(ctx context.Context, toUpdate model.User) (model.User, error)
	UpdateCurrentUser(ctx context.Context, toUpdate model.User) (model.User, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]model.Group, error)
}

func (v Dep) ListUsers(ctx context.Context, request *shieldv1beta1.ListUsersRequest) (*shieldv1beta1.ListUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*shieldv1beta1.User
	userList, err := v.UserService.ListUsers(ctx)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, user := range userList {
		userPB, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		users = append(users, &userPB)
	}

	return &shieldv1beta1.ListUsersResponse{
		Users: users,
	}, nil
}

func (v Dep) CreateUser(ctx context.Context, request *shieldv1beta1.CreateUserRequest) (*shieldv1beta1.CreateUserResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	currentUserEmail, _ := fetchEmailFromMetadata(ctx, v.IdentityProxyHeader)
	email := utils.DefaultStringIfEmpty(request.GetBody().Email, currentUserEmail)
	userT := model.User{
		Name:     request.GetBody().Name,
		Email:    email,
		Metadata: metaDataMap,
	}
	newUser, err := v.UserService.CreateUser(ctx, userT)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	metaData, err := structpb.NewStruct(mapOfInterfaceValues(newUser.Metadata))
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateUserResponse{User: &shieldv1beta1.User{
		Id:        newUser.Id,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newUser.CreatedAt),
		UpdatedAt: timestamppb.New(newUser.UpdatedAt),
	}}, nil
}

func (v Dep) GetUser(ctx context.Context, request *shieldv1beta1.GetUserRequest) (*shieldv1beta1.GetUserResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedUser, err := v.UserService.GetUser(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.UserDoesntExist):
			return nil, status.Errorf(codes.NotFound, "user not found")
		case errors.Is(err, user.InvalidUUID):
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

func (v Dep) GetCurrentUser(ctx context.Context, request *shieldv1beta1.GetCurrentUserRequest) (*shieldv1beta1.GetCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)

	email, err := fetchEmailFromMetadata(ctx, v.IdentityProxyHeader)
	if err != nil {
		return nil, grpcBadBodyError
	}

	fetchedUser, err := v.UserService.GetCurrentUser(ctx, email)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.UserDoesntExist):
			return nil, status.Errorf(codes.NotFound, "user not found")
		case errors.Is(err, user.InvalidUUID):
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

func (v Dep) UpdateUser(ctx context.Context, request *shieldv1beta1.UpdateUserRequest) (*shieldv1beta1.UpdateUserResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedUser, err := v.UserService.UpdateUser(ctx, model.User{
		Id:       request.GetId(),
		Name:     request.GetBody().Name,
		Email:    request.GetBody().Email,
		Metadata: metaDataMap,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, internalServerError
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, internalServerError
	}

	return &shieldv1beta1.UpdateUserResponse{User: &userPB}, nil
}

func (v Dep) UpdateCurrentUser(ctx context.Context, request *shieldv1beta1.UpdateCurrentUserRequest) (*shieldv1beta1.UpdateCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)

	email, err := fetchEmailFromMetadata(ctx, v.IdentityProxyHeader)
	if err != nil {
		return nil, grpcBadBodyError
	}

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	// if email in request body is different from the email in the header
	if request.GetBody().Email != email {
		return nil, grpcBadBodyError
	}

	updatedUser, err := v.UserService.UpdateCurrentUser(ctx, model.User{
		Name:     request.GetBody().Name,
		Metadata: metaDataMap,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateCurrentUserResponse{User: &userPB}, nil
}

func transformUserToPB(user model.User) (shieldv1beta1.User, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(user.Metadata))
	if err != nil {
		return shieldv1beta1.User{}, err
	}

	return shieldv1beta1.User{
		Id:        user.Id,
		Name:      user.Name,
		Email:     user.Email,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (v Dep) ListUserGroups(ctx context.Context, request *shieldv1beta1.ListUserGroupsRequest) (*shieldv1beta1.ListUserGroupsResponse, error) {
	logger := grpczap.Extract(ctx)
	var groups []*shieldv1beta1.Group
	groupsList, err := v.UserService.ListUserGroups(ctx, request.Id, request.Role)

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
