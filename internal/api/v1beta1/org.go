package v1beta1

import (
	"context"

	"github.com/odpf/shield/internal/bootstrap/schema"

	"github.com/odpf/shield/core/project"

	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/metadata"
	suuid "github.com/odpf/shield/pkg/uuid"
	"github.com/pkg/errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/core/organization"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var grpcOrgNotFoundErr = status.Errorf(codes.NotFound, "org doesn't exist")

//go:generate mockery --name=OrganizationService -r --case underscore --with-expecter --structname OrganizationService --filename org_service.go --output=./mocks
type OrganizationService interface {
	Get(ctx context.Context, idOrSlug string) (organization.Organization, error)
	Create(ctx context.Context, org organization.Organization) (organization.Organization, error)
	List(ctx context.Context, f organization.Filter) ([]organization.Organization, error)
	Update(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error)
	ListByUser(ctx context.Context, userID string) ([]organization.Organization, error)
	AddUsers(ctx context.Context, orgID string, userID []string) error
	RemoveUsers(ctx context.Context, orgID string, userID []string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
}

func (h Handler) ListOrganizations(ctx context.Context, request *shieldv1beta1.ListOrganizationsRequest) (*shieldv1beta1.ListOrganizationsResponse, error) {
	logger := grpczap.Extract(ctx)
	var orgs []*shieldv1beta1.Organization
	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:  organization.State(request.GetState()),
		UserID: request.GetUserId(),
	})
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

		orgs = append(orgs, orgPB)
	}

	return &shieldv1beta1.ListOrganizationsResponse{
		Organizations: orgs,
	}, nil
}

func (h Handler) ListAllOrganizations(ctx context.Context, request *shieldv1beta1.ListAllOrganizationsRequest) (*shieldv1beta1.ListAllOrganizationsResponse, error) {
	logger := grpczap.Extract(ctx)

	var orgs []*shieldv1beta1.Organization
	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:  organization.State(request.GetState()),
		UserID: request.GetUserId(),
	})
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

		orgs = append(orgs, orgPB)
	}

	return &shieldv1beta1.ListAllOrganizationsResponse{
		Organizations: orgs,
	}, nil
}

func (h Handler) CreateOrganization(ctx context.Context, request *shieldv1beta1.CreateOrganizationRequest) (*shieldv1beta1.CreateOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	org := organization.Organization{
		Name:     request.GetBody().GetName(),
		Title:    request.GetBody().GetTitle(),
		Metadata: metaDataMap,
	}

	newOrg, err := h.orgService.Create(ctx, org)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, organization.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateOrganizationResponse{Organization: orgPB}, nil
}

func (h Handler) GetOrganization(ctx context.Context, request *shieldv1beta1.GetOrganizationRequest) (*shieldv1beta1.GetOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedOrg, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, grpcOrgNotFoundErr
		case errors.Is(err, organization.ErrInvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	orgPB, err := transformOrgToPB(fetchedOrg)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, ErrInternalServer.Error())
	}

	return &shieldv1beta1.GetOrganizationResponse{
		Organization: orgPB,
	}, nil
}

func (h Handler) UpdateOrganization(ctx context.Context, request *shieldv1beta1.UpdateOrganizationRequest) (*shieldv1beta1.UpdateOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	var updatedOrg organization.Organization
	if suuid.IsValid(request.GetId()) {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			ID:       request.GetId(),
			Name:     request.GetBody().GetName(),
			Title:    request.GetBody().GetTitle(),
			Metadata: metaDataMap,
		})
	} else {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			Name:     request.GetBody().GetName(),
			Metadata: metaDataMap,
		})
	}
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, grpcOrgNotFoundErr
		case errors.Is(err, organization.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	orgPB, err := transformOrgToPB(updatedOrg)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternalServer
	}

	return &shieldv1beta1.UpdateOrganizationResponse{Organization: orgPB}, nil
}

func (h Handler) ListOrganizationAdmins(ctx context.Context, request *shieldv1beta1.ListOrganizationAdminsRequest) (*shieldv1beta1.ListOrganizationAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	admins, err := h.userService.ListByOrg(ctx, request.GetId(), organization.AdminPermission)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	var adminsPB []*shieldv1beta1.User
	for _, user := range admins {
		u, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		adminsPB = append(adminsPB, u)
	}

	return &shieldv1beta1.ListOrganizationAdminsResponse{Users: adminsPB}, nil
}

func (h Handler) ListOrganizationUsers(ctx context.Context, request *shieldv1beta1.ListOrganizationUsersRequest) (*shieldv1beta1.ListOrganizationUsersResponse, error) {
	logger := grpczap.Extract(ctx)

	permissionFilter := schema.MembershipPermission
	if len(request.GetPermissionFilter()) > 0 {
		permissionFilter = request.GetPermissionFilter()
	}

	users, err := h.userService.ListByOrg(ctx, request.GetId(), permissionFilter)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	var usersPB []*shieldv1beta1.User
	for _, rel := range users {
		u, err := transformUserToPB(rel)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		usersPB = append(usersPB, u)
	}

	return &shieldv1beta1.ListOrganizationUsersResponse{Users: usersPB}, nil
}

func (h Handler) ListOrganizationProjects(ctx context.Context, request *shieldv1beta1.ListOrganizationProjectsRequest) (*shieldv1beta1.ListOrganizationProjectsResponse, error) {
	logger := grpczap.Extract(ctx)

	projects, err := h.projectService.List(ctx, project.Filter{
		OrgID: request.GetId(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	var projectPB []*shieldv1beta1.Project
	for _, rel := range projects {
		u, err := transformProjectToPB(rel)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		projectPB = append(projectPB, u)
	}

	return &shieldv1beta1.ListOrganizationProjectsResponse{Projects: projectPB}, nil
}

func (h Handler) AddOrganizationUsers(ctx context.Context, request *shieldv1beta1.AddOrganizationUsersRequest) (*shieldv1beta1.AddOrganizationUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.orgService.AddUsers(ctx, request.GetId(), request.GetUserIds()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.AddOrganizationUsersResponse{}, nil
}

func (h Handler) RemoveOrganizationUser(ctx context.Context, request *shieldv1beta1.RemoveOrganizationUserRequest) (*shieldv1beta1.RemoveOrganizationUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.orgService.RemoveUsers(ctx, request.GetId(), []string{request.GetUserId()}); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.RemoveOrganizationUserResponse{}, nil
}

func (h Handler) EnableOrganization(ctx context.Context, request *shieldv1beta1.EnableOrganizationRequest) (*shieldv1beta1.EnableOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.orgService.Enable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.EnableOrganizationResponse{}, nil
}

func (h Handler) DisableOrganization(ctx context.Context, request *shieldv1beta1.DisableOrganizationRequest) (*shieldv1beta1.DisableOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.orgService.Disable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.DisableOrganizationResponse{}, nil
}

func transformOrgToPB(org organization.Organization) (*shieldv1beta1.Organization, error) {
	metaData, err := org.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &shieldv1beta1.Organization{
		Id:        org.ID,
		Name:      org.Name,
		Title:     org.Title,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(org.CreatedAt),
		UpdatedAt: timestamppb.New(org.UpdatedAt),
	}, nil
}
