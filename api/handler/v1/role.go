package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) GetAllRoles(ctx context.Context, request *shieldv1.GetAllRolesRequest) (*shieldv1.GetAllRolesResponse, error) {
	panic("implement me")
}

func (v Dep) CreateRole(ctx context.Context, request *shieldv1.CreateRoleRequest) (*shieldv1.RoleResponse, error) {
	panic("implement me")
}

func (v Dep) GetRole(ctx context.Context, request *shieldv1.GetRoleRequest) (*shieldv1.RoleResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateRole(ctx context.Context, request *shieldv1.UpdateRoleRequest) (*shieldv1.RoleResponse, error) {
	panic("implement me")
}

