package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) GetAllUsers(ctx context.Context, request *shieldv1.GetAllUsersRequest) (*shieldv1.GetAllUsersResponse, error) {
	panic("implement me")
}

func (v Dep) CreateUser(ctx context.Context, request *shieldv1.CreateUserRequest) (*shieldv1.UserResponse, error) {
	panic("implement me")
}

func (v Dep) GetUserByID(ctx context.Context, request *shieldv1.GetUserRequest) (*shieldv1.UserResponse, error) {
	panic("implement me")
}

func (v Dep) GetCurrentUser(ctx context.Context, request *shieldv1.GetCurrentUserRequest) (*shieldv1.UserResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateUserByID(ctx context.Context, request *shieldv1.UpdateUserRequest) (*shieldv1.UserResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateCurrentUser(ctx context.Context, request *shieldv1.UpdateCurrentUserRequest) (*shieldv1.UserResponse, error) {
	panic("implement me")
}