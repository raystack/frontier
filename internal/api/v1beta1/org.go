package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/frontier/pkg/utils"
	"go.uber.org/zap"

	"github.com/raystack/frontier/core/organization"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

var (
	grpcOrgNotFoundErr   = status.Errorf(codes.NotFound, "org doesn't exist")
	grpcOrgDisabledErr   = status.Errorf(codes.NotFound, "org is disabled. Please contact your administrator to enable it")
	grpcMinAdminCountErr = status.Errorf(codes.PermissionDenied, "org must have at least one admin, consider adding another admin before removing")
)

type OrganizationService interface {
	Get(ctx context.Context, idOrSlug string) (organization.Organization, error)
	GetRaw(ctx context.Context, idOrSlug string) (organization.Organization, error)
	Create(ctx context.Context, org organization.Organization) (organization.Organization, error)
	List(ctx context.Context, f organization.Filter) ([]organization.Organization, error)
	Update(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error)
	ListByUser(ctx context.Context, principal authenticate.Principal, flt organization.Filter) ([]organization.Organization, error)
	AddUsers(ctx context.Context, orgID string, userID []string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
}

func (h Handler) ListOrganizations(ctx context.Context, request *frontierv1beta1.ListOrganizationsRequest) (*frontierv1beta1.ListOrganizationsResponse, error) {
	var orgs []*frontierv1beta1.Organization
	paginate := pagination.NewPagination(request.GetPageNum(), request.GetPageSize())

	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:      organization.State(request.GetState()),
		UserID:     request.GetUserId(),
		Pagination: paginate,
	})
	if err != nil {
		return nil, err
	}

	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			return nil, err
		}

		orgs = append(orgs, orgPB)
	}

	return &frontierv1beta1.ListOrganizationsResponse{
		Organizations: orgs,
	}, nil
}

func (h Handler) ListAllOrganizations(ctx context.Context, request *frontierv1beta1.ListAllOrganizationsRequest) (*frontierv1beta1.ListAllOrganizationsResponse, error) {
	var orgs []*frontierv1beta1.Organization
	paginate := pagination.NewPagination(request.GetPageNum(), request.GetPageSize())

	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:      organization.State(request.GetState()),
		UserID:     request.GetUserId(),
		Pagination: paginate,
	})
	if err != nil {
		return nil, err
	}

	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			return nil, err
		}

		orgs = append(orgs, orgPB)
	}

	return &frontierv1beta1.ListAllOrganizationsResponse{
		Organizations: orgs,
		Count:         paginate.Count,
	}, nil
}

func (h Handler) CreateOrganization(ctx context.Context, request *frontierv1beta1.CreateOrganizationRequest) (*frontierv1beta1.CreateOrganizationResponse, error) {
	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, grpcBadBodyMetaSchemaError
	}

	newOrg, err := h.orgService.Create(ctx, organization.Organization{
		Name:     request.GetBody().GetName(),
		Title:    request.GetBody().GetTitle(),
		Avatar:   request.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, organization.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, err
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, newOrg.ID).LogWithAttrs(audit.OrgCreatedEvent, audit.OrgTarget(newOrg.ID), map[string]string{
		"title": newOrg.Title,
		"name":  newOrg.Name,
	})
	return &frontierv1beta1.CreateOrganizationResponse{Organization: orgPB}, nil
}

func (h Handler) GetOrganization(ctx context.Context, request *frontierv1beta1.GetOrganizationRequest) (*frontierv1beta1.GetOrganizationResponse, error) {
	fetchedOrg, err := h.orgService.GetRaw(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, grpcOrgNotFoundErr
		case errors.Is(err, organization.ErrInvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, err
		}
	}

	orgPB, err := transformOrgToPB(fetchedOrg)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetOrganizationResponse{
		Organization: orgPB,
	}, nil
}

func (h Handler) UpdateOrganization(ctx context.Context, request *frontierv1beta1.UpdateOrganizationRequest) (*frontierv1beta1.UpdateOrganizationResponse, error) {
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, grpcBadBodyMetaSchemaError
	}

	var updatedOrg organization.Organization
	var err error
	if utils.IsValidUUID(request.GetId()) {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			ID:       request.GetId(),
			Name:     request.GetBody().GetName(),
			Title:    request.GetBody().GetTitle(),
			Avatar:   request.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	} else {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			Name:     request.GetBody().GetName(),
			Title:    request.GetBody().GetTitle(),
			Avatar:   request.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	}
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, grpcOrgNotFoundErr
		case errors.Is(err, organization.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, err
		}
	}

	orgPB, err := transformOrgToPB(updatedOrg)
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, updatedOrg.ID).Log(audit.OrgUpdatedEvent, audit.OrgTarget(updatedOrg.ID))
	return &frontierv1beta1.UpdateOrganizationResponse{Organization: orgPB}, nil
}

func (h Handler) ListOrganizationAdmins(ctx context.Context, request *frontierv1beta1.ListOrganizationAdminsRequest) (*frontierv1beta1.ListOrganizationAdminsResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	admins, err := h.userService.ListByOrg(ctx, orgResp.ID, organization.AdminRole)
	if err != nil {
		return nil, err
	}

	var adminsPB []*frontierv1beta1.User
	for _, user := range admins {
		u, err := transformUserToPB(user)
		if err != nil {
			return nil, err
		}

		adminsPB = append(adminsPB, u)
	}

	return &frontierv1beta1.ListOrganizationAdminsResponse{Users: adminsPB}, nil
}

func (h Handler) ListOrganizationUsers(ctx context.Context, request *frontierv1beta1.ListOrganizationUsersRequest) (*frontierv1beta1.ListOrganizationUsersResponse, error) {
	if len(request.GetRoleFilters()) > 0 && request.GetWithRoles() {
		return nil, status.Errorf(codes.InvalidArgument, "cannot use role filters and with_roles together")
	}

	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	var users []user.User
	var rolePairPBs []*frontierv1beta1.ListOrganizationUsersResponse_RolePair

	if len(request.GetRoleFilters()) > 0 {
		// convert role names to ids if needed
		roleIDs := request.GetRoleFilters()
		for i, roleFilter := range request.GetRoleFilters() {
			if !utils.IsValidUUID(roleFilter) {
				role, err := h.roleService.Get(ctx, roleFilter)
				if err != nil {
					return nil, err
				}
				roleIDs[i] = role.ID
			}
		}

		// need to fetch users with roles assigned to them
		policies, err := h.policyService.List(ctx, policy.Filter{
			OrgID:         request.GetId(),
			PrincipalType: schema.UserPrincipal,
			ResourceType:  schema.OrganizationNamespace,
			RoleIDs:       roleIDs,
		})
		if err != nil {
			return nil, err
		}
		users = utils.Filter(utils.Map(policies, func(pol policy.Policy) user.User {
			u, _ := h.userService.GetByID(ctx, pol.PrincipalID)
			return u
		}), func(u user.User) bool {
			return u.ID != ""
		})
	} else {
		// list all users
		users, err = h.userService.ListByOrg(ctx, orgResp.ID, request.GetPermissionFilter())
		if err != nil {
			return nil, err
		}
		if request.GetWithRoles() {
			for _, user := range users {
				roles, err := h.policyService.ListRoles(ctx, schema.UserPrincipal, user.ID, schema.OrganizationNamespace, request.GetId())
				if err != nil {
					return nil, err
				}

				rolesPb := utils.Filter(utils.Map(roles, func(role role.Role) *frontierv1beta1.Role {
					pb, err := transformRoleToPB(role)
					if err != nil {
						logger.Error("failed to transform role for group", zap.Error(err))
						return nil
					}
					return &pb
				}), func(role *frontierv1beta1.Role) bool {
					return role != nil
				})
				rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListOrganizationUsersResponse_RolePair{
					UserId: user.ID,
					Roles:  rolesPb,
				})
			}
		}
	}

	var usersPB []*frontierv1beta1.User
	for _, rel := range users {
		u, err := transformUserToPB(rel)
		if err != nil {
			return nil, err
		}
		usersPB = append(usersPB, u)
	}
	return &frontierv1beta1.ListOrganizationUsersResponse{
		Users:     usersPB,
		RolePairs: rolePairPBs,
	}, nil
}

func (h Handler) ListOrganizationServiceUsers(ctx context.Context, request *frontierv1beta1.ListOrganizationServiceUsersRequest) (*frontierv1beta1.ListOrganizationServiceUsersResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	usersList, err := h.serviceUserService.List(ctx, serviceuser.Filter{
		OrgID: orgResp.ID,
	})
	if err != nil {
		return nil, err
	}

	var usersPB []*frontierv1beta1.ServiceUser
	for _, rel := range usersList {
		u, err := transformServiceUserToPB(rel)
		if err != nil {
			return nil, err
		}

		usersPB = append(usersPB, u)
	}
	return &frontierv1beta1.ListOrganizationServiceUsersResponse{Serviceusers: usersPB}, nil
}

func (h Handler) ListOrganizationProjects(ctx context.Context, request *frontierv1beta1.ListOrganizationProjectsRequest) (*frontierv1beta1.ListOrganizationProjectsResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	projects, err := h.projectService.List(ctx, project.Filter{
		OrgID:           orgResp.ID,
		WithMemberCount: request.GetWithMemberCount(),
	})
	if err != nil {
		return nil, err
	}

	var projectPB []*frontierv1beta1.Project
	for _, rel := range projects {
		u, err := transformProjectToPB(rel)
		if err != nil {
			return nil, err
		}

		projectPB = append(projectPB, u)
	}

	return &frontierv1beta1.ListOrganizationProjectsResponse{Projects: projectPB}, nil
}

func (h Handler) AddOrganizationUsers(ctx context.Context, request *frontierv1beta1.AddOrganizationUsersRequest) (*frontierv1beta1.AddOrganizationUsersResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	if err := h.orgService.AddUsers(ctx, orgResp.ID, request.GetUserIds()); err != nil {
		return nil, err
	}
	return &frontierv1beta1.AddOrganizationUsersResponse{}, nil
}

func (h Handler) RemoveOrganizationUser(ctx context.Context, request *frontierv1beta1.RemoveOrganizationUserRequest) (*frontierv1beta1.RemoveOrganizationUserResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	admins, err := h.userService.ListByOrg(ctx, orgResp.ID, organization.AdminRole)
	if err != nil {
		return nil, err
	}
	if len(admins) == 1 && admins[0].ID == request.GetUserId() {
		return nil, grpcMinAdminCountErr
	}

	if err := h.deleterService.RemoveUsersFromOrg(ctx, orgResp.ID, []string{request.GetUserId()}); err != nil {
		return nil, err
	}
	return &frontierv1beta1.RemoveOrganizationUserResponse{}, nil
}

func (h Handler) EnableOrganization(ctx context.Context, request *frontierv1beta1.EnableOrganizationRequest) (*frontierv1beta1.EnableOrganizationResponse, error) {
	if err := h.orgService.Enable(ctx, request.GetId()); err != nil {
		return nil, err
	}
	return &frontierv1beta1.EnableOrganizationResponse{}, nil
}

func (h Handler) DisableOrganization(ctx context.Context, request *frontierv1beta1.DisableOrganizationRequest) (*frontierv1beta1.DisableOrganizationResponse, error) {
	if err := h.orgService.Disable(ctx, request.GetId()); err != nil {
		return nil, err
	}
	return &frontierv1beta1.DisableOrganizationResponse{}, nil
}

func transformOrgToPB(org organization.Organization) (*frontierv1beta1.Organization, error) {
	metaData, err := org.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Organization{
		Id:        org.ID,
		Name:      org.Name,
		Title:     org.Title,
		Metadata:  metaData,
		State:     org.State.String(),
		Avatar:    org.Avatar,
		CreatedAt: timestamppb.New(org.CreatedAt),
		UpdatedAt: timestamppb.New(org.UpdatedAt),
	}, nil
}
