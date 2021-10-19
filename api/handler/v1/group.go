package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) ListGroups(ctx context.Context, request *shieldv1.ListGroupsRequest) (*shieldv1.ListGroupsResponse, error) {
	panic("implement me")
}

func (v Dep) CreateGroup(ctx context.Context, request *shieldv1.CreateGroupRequest) (*shieldv1.CreateGroupResponse, error) {
	panic("implement me")
}

func (v Dep) GetGroup(ctx context.Context, request *shieldv1.GetGroupRequest) (*shieldv1.GetGroupResponse, error) {
	panic("implement me")
}

func (v Dep) ListGroupUsers(ctx context.Context, request *shieldv1.ListGroupUsersRequest) (*shieldv1.ListGroupUsersResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateGroup(ctx context.Context, request *shieldv1.UpdateGroupRequest) (*shieldv1.UpdateGroupResponse, error) {
	panic("implement me")
}
