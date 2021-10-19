package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) ListUsers(ctx context.Context, request *shieldv1.ListUsersRequest) (*shieldv1.ListUsersResponse, error) {
	panic("implement me")
}

func (v Dep) CreateUser(ctx context.Context, request *shieldv1.CreateUserRequest) (*shieldv1.CreateUserResponse, error) {
	panic("implement me")
}

func (v Dep) GetUser(ctx context.Context, request *shieldv1.GetUserRequest) (*shieldv1.GetUserResponse, error) {
	panic("get user was called")
}

func (v Dep) GetCurrentUser(ctx context.Context, request *shieldv1.GetCurrentUserRequest) (*shieldv1.GetCurrentUserResponse, error) {
	panic("get CURRENT user was called")
}

func (v Dep) UpdateUser(ctx context.Context, request *shieldv1.UpdateUserRequest) (*shieldv1.UpdateUserResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateCurrentUser(ctx context.Context, request *shieldv1.UpdateCurrentUserRequest) (*shieldv1.UpdateCurrentUserResponse, error) {
	panic("implement me")
}
