package permission

import (
	"context"

	"github.com/odpf/shield/model"
	"github.com/odpf/shield/utils"
)

type CheckService struct {
	PermissionsService Permissions
}

func NewCheckService(permissionService Permissions) CheckService {
	return CheckService{PermissionsService: permissionService}
}

func (c CheckService) CheckAuthz(ctx context.Context, resource model.Resource, prmsn model.Permission) (bool, error) {
	user, err := c.PermissionsService.FetchCurrentUser(ctx)
	if err != nil {
		return false, err
	}

	resource.Id = utils.CreateResourceId(resource)
	return c.PermissionsService.CheckPermission(ctx, user, resource, prmsn)
}
