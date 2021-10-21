package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

func (v Dep) ListGroups(ctx context.Context, request *shieldv1.ListGroupsRequest) (*shieldv1.ListGroupsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) CreateGroup(ctx context.Context, request *shieldv1.CreateGroupRequest) (*shieldv1.CreateGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) GetGroup(ctx context.Context, request *shieldv1.GetGroupRequest) (*shieldv1.GetGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) ListGroupUsers(ctx context.Context, request *shieldv1.ListGroupUsersRequest) (*shieldv1.ListGroupUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) UpdateGroup(ctx context.Context, request *shieldv1.UpdateGroupRequest) (*shieldv1.UpdateGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
