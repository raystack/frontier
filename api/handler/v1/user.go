package v1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/shield/internal/user"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

type UserService interface {
	GetUser(ctx context.Context, id string) (user.User, error)
	CreateUser(ctx context.Context, user user.User) (user.User, error)
	ListUsers(ctx context.Context) ([]user.User, error)
	UpdateUser(ctx context.Context, toUpdate user.User) (user.User, error)
}

func (v Dep) ListUsers(ctx context.Context, request *shieldv1.ListUsersRequest) (*shieldv1.ListUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*shieldv1.User
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

	return &shieldv1.ListUsersResponse{
		Users: users,
	}, nil
}

func (v Dep) CreateUser(ctx context.Context, request *shieldv1.CreateUserRequest) (*shieldv1.CreateUserResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	newUser, err := v.UserService.CreateUser(ctx, user.User{
		Name:     request.GetBody().Name,
		Email:    request.GetBody().Email,
		Metadata: metaDataMap,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	metaData, err := structpb.NewStruct(mapOfInterfaceValues(newUser.Metadata))
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.CreateUserResponse{User: &shieldv1.User{
		Id:        newUser.Id,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newUser.CreatedAt),
		UpdatedAt: timestamppb.New(newUser.UpdatedAt),
	}}, nil
}

func (v Dep) GetUser(ctx context.Context, request *shieldv1.GetUserRequest) (*shieldv1.GetUserResponse, error) {
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

	return &shieldv1.GetUserResponse{
		User: &userPB,
	}, nil
}

func (v Dep) GetCurrentUser(ctx context.Context, request *shieldv1.GetCurrentUserRequest) (*shieldv1.GetCurrentUserResponse, error) {
	panic("get CURRENT user was called")
}

func (v Dep) UpdateUser(ctx context.Context, request *shieldv1.UpdateUserRequest) (*shieldv1.UpdateUserResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedUser, err := v.UserService.UpdateUser(ctx, user.User{
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

	return &shieldv1.UpdateUserResponse{User: &userPB}, nil
}

func (v Dep) UpdateCurrentUser(ctx context.Context, request *shieldv1.UpdateCurrentUserRequest) (*shieldv1.UpdateCurrentUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func transformUserToPB(user user.User) (shieldv1.User, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(user.Metadata))
	if err != nil {
		return shieldv1.User{}, err
	}

	return shieldv1.User{
		Id:        user.Id,
		Name:      user.Name,
		Email:     user.Email,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}
