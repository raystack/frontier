package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

func (v Dep) ListRoles(ctx context.Context, request *shieldv1.ListRolesRequest) (*shieldv1.ListRolesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) CreateRole(ctx context.Context, request *shieldv1.CreateRoleRequest) (*shieldv1.CreateRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) GetRole(ctx context.Context, request *shieldv1.GetRoleRequest) (*shieldv1.GetRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) UpdateRole(ctx context.Context, request *shieldv1.UpdateRoleRequest) (*shieldv1.UpdateRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
