package permission

import (
	"context"
	"fmt"

	"github.com/odpf/shield/model"
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

	resource.Id = fmt.Sprintf("%s/%s", resource.NamespaceId, resource.Name)

	return c.PermissionsService.CheckPermission(ctx, user, resource, prmsn)
}
