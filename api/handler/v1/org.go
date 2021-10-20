package v1

import (
	"context"
	"errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/org"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

type OrganizationService interface {
	GetOrganization(ctx context.Context, id string) (org.Organization, error)
	CreateOrganization(ctx context.Context, org org.Organization) (org.Organization, error)
	ListOrganizations(ctx context.Context) ([]org.Organization, error)
}

// HTTP Codes defined here:
// https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go#L36

func (v Dep) ListOrganizations(ctx context.Context, request *shieldv1.ListOrganizationsRequest) (*shieldv1.ListOrganizationsResponse, error) {
	var orgs []*shieldv1.Organization
	orgList, err := v.OrgService.ListOrganizations(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, internalServerError.Error())
	}

	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		orgs = append(orgs, &orgPB)
	}

	return &shieldv1.ListOrganizationsResponse{
		Organizations: orgs,
	}, nil
}

func (v Dep) CreateOrganization(ctx context.Context, request *shieldv1.CreateOrganizationRequest) (*shieldv1.CreateOrganizationResponse, error) {
	newOrg, err := v.OrgService.CreateOrganization(ctx, org.Organization{
		Name:     request.GetBody().Name,
		Slug:     request.GetBody().Slug,
		Metadata: request.GetBody().Metadata.AsMap(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	metaData, err := structpb.NewStruct(newOrg.Metadata)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &shieldv1.CreateOrganizationResponse{Organization: &shieldv1.Organization{
		Id:        newOrg.Id,
		Name:      newOrg.Name,
		Slug:      newOrg.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newOrg.CreatedAt),
		UpdatedAt: timestamppb.New(newOrg.UpdatedAt),
	}}, nil
}

func (v Dep) GetOrganization(ctx context.Context, request *shieldv1.GetOrganizationRequest) (*shieldv1.GetOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)
	fetchedOrg, err := v.OrgService.GetOrganization(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, org.OrgDoesntExist):
			return nil, status.Errorf(codes.NotFound, "organization not found")
		//case errors.Is(err, org.InvalidUUID):
		//	return nil, status.Errorf(codes.Internal, "organization not found")
		default:
			return nil, status.Errorf(codes.Internal, internalServerError.Error())
		}
	}

	orgPB, err := transformOrgToPB(fetchedOrg)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, internalServerError.Error())
	}

	return &shieldv1.GetOrganizationResponse{
		Organization: &orgPB,
	}, nil
}

func (v Dep) UpdateOrganization(ctx context.Context, request *shieldv1.UpdateOrganizationRequest) (*shieldv1.UpdateOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func transformOrgToPB(org org.Organization) (shieldv1.Organization, error) {
	metaData, err := structpb.NewStruct(org.Metadata)
	if err != nil {
		return shieldv1.Organization{}, status.Errorf(codes.Internal, err.Error())
	}

	return shieldv1.Organization{
		Id:        org.Id,
		Name:      org.Name,
		Slug:      org.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(org.CreatedAt),
		UpdatedAt: timestamppb.New(org.UpdatedAt),
	}, nil
}
