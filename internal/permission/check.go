package permission

import (
	"context"

	"github.com/odpf/shield/model"
	"github.com/odpf/shield/utils"
)

type CheckService struct {
	PermissionsService Permissions
	ResourceStore      ResourceStore
}

type ResourceStore interface {
	GetResourceByURN(ctx context.Context, urn string) (model.Resource, error)
}

func NewCheckService(permissionService Permissions, resourceStore ResourceStore) CheckService {
	return CheckService{PermissionsService: permissionService, ResourceStore: resourceStore}
}

func (c CheckService) CheckAuthz(ctx context.Context, resource model.Resource, action model.Action) (bool, error) {
	user, err := c.PermissionsService.FetchCurrentUser(ctx)
	if err != nil {
		return false, err
	}

	resource.Urn = utils.CreateResourceURN(resource)

	isSystemNS := utils.IsSystemNS(resource)
	fetchedResource := resource

	if isSystemNS {
		fetchedResource.Idxa = resource.Urn
	} else {
		fetchedResource, err = c.ResourceStore.GetResourceByURN(ctx, resource.Urn)
		if err != nil {
			return false, err
		}
	}

	return c.PermissionsService.CheckPermission(ctx, user, fetchedResource, action)
}
