package v1beta1

import (
	"context"
	"errors"
	"strings"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/internal/org"
	"github.com/odpf/shield/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

type OrganizationService interface {
	Get(ctx context.Context, id string) (model.Organization, error)
	Create(ctx context.Context, org model.Organization) (model.Organization, error)
	List(ctx context.Context) ([]model.Organization, error)
	Update(ctx context.Context, toUpdate model.Organization) (model.Organization, error)
	AddAdmin(ctx context.Context, id string, userIds []string) ([]model.User, error)
	ListAdmins(ctx context.Context, id string) ([]model.User, error)
	RemoveAdmin(ctx context.Context, id string, user_id string) (string, error)
}

func (v Dep) ListOrganizations(ctx context.Context, request *shieldv1beta1.ListOrganizationsRequest) (*shieldv1beta1.ListOrganizationsResponse, error) {
	logger := grpczap.Extract(ctx)
	var orgs []*shieldv1beta1.Organization

	orgList, err := v.OrgService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		orgs = append(orgs, &orgPB)
	}

	return &shieldv1beta1.ListOrganizationsResponse{
		Organizations: orgs,
	}, nil
}

func (v Dep) CreateOrganization(ctx context.Context, request *shieldv1beta1.CreateOrganizationRequest) (*shieldv1beta1.CreateOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	// TODO (@krtkvrm): Add validations using Proto
	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	slug := request.GetBody().Slug
	if strings.TrimSpace(slug) == "" {
		slug = generateSlug(request.GetBody().Name)
	}

	newOrg, err := v.OrgService.Create(ctx, model.Organization{
		Name:     request.GetBody().Name,
		Slug:     slug,
		Metadata: metaDataMap,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	metaData, err := structpb.NewStruct(mapOfInterfaceValues(newOrg.Metadata))
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateOrganizationResponse{Organization: &shieldv1beta1.Organization{
		Id:        newOrg.Id,
		Name:      newOrg.Name,
		Slug:      newOrg.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newOrg.CreatedAt),
		UpdatedAt: timestamppb.New(newOrg.UpdatedAt),
	}}, nil
}

func (v Dep) GetOrganization(ctx context.Context, request *shieldv1beta1.GetOrganizationRequest) (*shieldv1beta1.GetOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedOrg, err := v.OrgService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, org.OrgDoesntExist):
			return nil, status.Errorf(codes.NotFound, "organization not found")
		case errors.Is(err, org.InvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	orgPB, err := transformOrgToPB(fetchedOrg)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, internalServerError.Error())
	}

	return &shieldv1beta1.GetOrganizationResponse{
		Organization: &orgPB,
	}, nil
}

func (v Dep) UpdateOrganization(ctx context.Context, request *shieldv1beta1.UpdateOrganizationRequest) (*shieldv1beta1.UpdateOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedOrg, err := v.OrgService.Update(ctx, model.Organization{
		Id:       request.GetId(),
		Name:     request.GetBody().Name,
		Slug:     request.GetBody().Slug,
		Metadata: metaDataMap,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, internalServerError
	}

	orgPB, err := transformOrgToPB(updatedOrg)
	if err != nil {
		logger.Error(err.Error())
		return nil, internalServerError
	}

	return &shieldv1beta1.UpdateOrganizationResponse{Organization: &orgPB}, nil
}

func (v Dep) AddOrganizationAdmin(ctx context.Context, request *shieldv1beta1.AddOrganizationAdminRequest) (*shieldv1beta1.AddOrganizationAdminResponse, error) {
	logger := grpczap.Extract(ctx)
	userIds := request.GetBody().UserIds

	addedUsers, err := v.OrgService.AddAdmin(ctx, request.GetId(), userIds)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, org.OrgDoesntExist):
			return nil, status.Errorf(codes.NotFound, "org to be updated not found")
		default:
			return nil, grpcInternalServerError
		}
	}

	var addedUsersPB []*shieldv1beta1.User
	for _, u := range addedUsers {
		userPB, err := transformUserToPB(u)
		if err != nil {
			logger.Error(err.Error())
			return nil, internalServerError
		}

		addedUsersPB = append(addedUsersPB, &userPB)
	}

	return &shieldv1beta1.AddOrganizationAdminResponse{Users: addedUsersPB}, nil
}

func (v Dep) ListOrganizationAdmins(ctx context.Context, request *shieldv1beta1.ListOrganizationAdminsRequest) (*shieldv1beta1.ListOrganizationAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	admins, err := v.OrgService.ListAdmins(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, internalServerError
	}

	var adminsPB []*shieldv1beta1.User
	for _, user := range admins {
		u, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, internalServerError
		}

		adminsPB = append(adminsPB, &u)
	}

	return &shieldv1beta1.ListOrganizationAdminsResponse{Users: adminsPB}, nil
}

func (v Dep) RemoveOrganizationAdmin(ctx context.Context, request *shieldv1beta1.RemoveOrganizationAdminRequest) (*shieldv1beta1.RemoveOrganizationAdminResponse, error) {
	logger := grpczap.Extract(ctx)

	msg, err := v.OrgService.RemoveAdmin(ctx, request.GetId(), request.GetUserId())
	if err != nil {
		logger.Error(err.Error())
		return nil, internalServerError
	}

	return &shieldv1beta1.RemoveOrganizationAdminResponse{Message: msg}, nil
}

func transformOrgToPB(org model.Organization) (shieldv1beta1.Organization, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(org.Metadata))
	if err != nil {
		return shieldv1beta1.Organization{}, err
	}

	return shieldv1beta1.Organization{
		Id:        org.Id,
		Name:      org.Name,
		Slug:      org.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(org.CreatedAt),
		UpdatedAt: timestamppb.New(org.UpdatedAt),
	}, nil
}
