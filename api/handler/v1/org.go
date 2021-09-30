package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) GetAllOrganizations(ctx context.Context, request *shieldv1.GetAllOrganizationsRequest) (*shieldv1.GetAllOrganizationsResponse, error) {
	panic("implement me")
}

func (v Dep) CreateOrganization(ctx context.Context, request *shieldv1.CreateOrganizationRequest) (*shieldv1.OrganizationResponse, error) {
	panic("implement me")
}

func (v Dep) GetOrganizationByID(ctx context.Context, request *shieldv1.GetOrganizationRequest) (*shieldv1.OrganizationResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateOrganizationByID(ctx context.Context, request *shieldv1.UpdateOrganizationRequest) (*shieldv1.OrganizationResponse, error) {
	panic("implement me")
}
