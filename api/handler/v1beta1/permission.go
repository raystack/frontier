package v1beta1

import (
	"context"

	"github.com/odpf/shield/model"
)

type PermissionService interface {
	CheckAuthz(ctx context.Context, resource model.Resource, prmsn model.Permission)
}
