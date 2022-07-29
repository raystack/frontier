package v1beta1

import (
	"context"
	"strings"

	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/errors"
	"github.com/odpf/shield/pkg/uuid"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/core/organization"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

type OrganizationService interface {
	Get(ctx context.Context, idOrSlug string) (organization.Organization, error)
	Create(ctx context.Context, org organization.Organization) (organization.Organization, error)
	List(ctx context.Context) ([]organization.Organization, error)
	Update(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error)
	AddAdmins(ctx context.Context, idOrSlug string, userIds []string) ([]user.User, error)
	RemoveAdmin(ctx context.Context, idOrSlug string, userId string) ([]user.User, error)
	ListAdmins(ctx context.Context, id string) ([]user.User, error)
}

func (h Handler) ListOrganizations(ctx context.Context, request *shieldv1beta1.ListOrganizationsRequest) (*shieldv1beta1.ListOrganizationsResponse, error) {
	logger := grpczap.Extract(ctx)
	var orgs []*shieldv1beta1.Organization

	orgList, err := h.orgService.List(ctx)
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

func (h Handler) CreateOrganization(ctx context.Context, request *shieldv1beta1.CreateOrganizationRequest) (*shieldv1beta1.CreateOrganizationResponse, error) {
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

	newOrg, err := h.orgService.Create(ctx, organization.Organization{
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
		Id:        newOrg.ID,
		Name:      newOrg.Name,
		Slug:      newOrg.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newOrg.CreatedAt),
		UpdatedAt: timestamppb.New(newOrg.UpdatedAt),
	}}, nil
}

func (h Handler) GetOrganization(ctx context.Context, request *shieldv1beta1.GetOrganizationRequest) (*shieldv1beta1.GetOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedOrg, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "organization not found")
		case errors.Is(err, organization.ErrInvalidUUID):
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

func (h Handler) UpdateOrganization(ctx context.Context, request *shieldv1beta1.UpdateOrganizationRequest) (*shieldv1beta1.UpdateOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	var updatedOrg organization.Organization
	if uuid.IsValid(request.GetId()) {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			ID:       request.GetId(),
			Name:     request.GetBody().Name,
			Slug:     request.GetBody().Slug,
			Metadata: metaDataMap,
		})
	} else {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			Name:     request.GetBody().Name,
			Slug:     request.GetId(),
			Metadata: metaDataMap,
		})
	}
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

func (h Handler) AddOrganizationAdmin(ctx context.Context, request *shieldv1beta1.AddOrganizationAdminRequest) (*shieldv1beta1.AddOrganizationAdminResponse, error) {
	logger := grpczap.Extract(ctx)

	addedUsers, err := h.orgService.AddAdmins(ctx, request.GetId(), request.GetBody().GetUserIds())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "org to be updated not found")
		case errors.Is(err, errors.Unauthorized):
			return nil, grpcPermissionDenied
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

func (h Handler) ListOrganizationAdmins(ctx context.Context, request *shieldv1beta1.ListOrganizationAdminsRequest) (*shieldv1beta1.ListOrganizationAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	admins, err := h.orgService.ListAdmins(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "org to be updated not found")
		default:
			return nil, grpcInternalServerError
		}
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

func (h Handler) RemoveOrganizationAdmin(ctx context.Context, request *shieldv1beta1.RemoveOrganizationAdminRequest) (*shieldv1beta1.RemoveOrganizationAdminResponse, error) {
	logger := grpczap.Extract(ctx)

	if _, err := h.orgService.RemoveAdmin(ctx, request.GetId(), request.GetUserId()); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "org to be updated not found")
		case errors.Is(err, errors.Unauthorized):
			return nil, grpcPermissionDenied
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.RemoveOrganizationAdminResponse{
		Message: "Removed Admin from org",
	}, nil
}

func transformOrgToPB(org organization.Organization) (shieldv1beta1.Organization, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(org.Metadata))
	if err != nil {
		return shieldv1beta1.Organization{}, err
	}

	return shieldv1beta1.Organization{
		Id:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(org.CreatedAt),
		UpdatedAt: timestamppb.New(org.UpdatedAt),
	}, nil
}
