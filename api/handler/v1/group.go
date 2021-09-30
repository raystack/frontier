package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) GetAllGroups(ctx context.Context, request *shieldv1.GetAllGroupsRequest) (*shieldv1.GetAllGroupsResponse, error) {
	panic("implement me")
}

func (v Dep) CreateGroup(ctx context.Context, request *shieldv1.CreateGroupRequest) (*shieldv1.GroupResponse, error) {
	panic("implement me")
}

func (v Dep) GetGroupByID(ctx context.Context, request *shieldv1.GetGroupRequest) (*shieldv1.GroupResponse, error) {
	panic("implement me")
}

func (v Dep) GetGroupUsers(ctx context.Context, request *shieldv1.GetGroupRequest) (*shieldv1.GetGroupUsersResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateGroupByID(ctx context.Context, request *shieldv1.UpdateGroupRequest) (*shieldv1.GroupResponse, error) {
	panic("implement me")
}
