package v1

import (
	"context"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

func (v Dep) ListRoles(ctx context.Context, request *shieldv1.ListRolesRequest) (*shieldv1.ListRolesResponse, error) {
	panic("implement me")
}

func (v Dep) CreateRole(ctx context.Context, request *shieldv1.CreateRoleRequest) (*shieldv1.CreateRoleResponse, error) {
	panic("implement me")
}

func (v Dep) GetRole(ctx context.Context, request *shieldv1.GetRoleRequest) (*shieldv1.GetRoleResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateRole(ctx context.Context, request *shieldv1.UpdateRoleRequest) (*shieldv1.UpdateRoleResponse, error) {
	panic("implement me")
}
