package v1

import (
	"context"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

func (v Dep) ListOrganizations(ctx context.Context, request *shieldv1.ListOrganizationsRequest) (*shieldv1.ListOrganizationsResponse, error) {
	panic("implement me")
}

func (v Dep) CreateOrganization(ctx context.Context, request *shieldv1.CreateOrganizationRequest) (*shieldv1.CreateOrganizationResponse, error) {
	panic("implement me")
}

func (v Dep) GetOrganization(ctx context.Context, request *shieldv1.GetOrganizationRequest) (*shieldv1.GetOrganizationResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateOrganization(ctx context.Context, request *shieldv1.UpdateOrganizationRequest) (*shieldv1.UpdateOrganizationResponse, error) {
	panic("implement me")
}
