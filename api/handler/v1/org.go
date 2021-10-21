package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

func (v Dep) ListOrganizations(ctx context.Context, request *shieldv1.ListOrganizationsRequest) (*shieldv1.ListOrganizationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) CreateOrganization(ctx context.Context, request *shieldv1.CreateOrganizationRequest) (*shieldv1.CreateOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) GetOrganization(ctx context.Context, request *shieldv1.GetOrganizationRequest) (*shieldv1.GetOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) UpdateOrganization(ctx context.Context, request *shieldv1.UpdateOrganizationRequest) (*shieldv1.UpdateOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
