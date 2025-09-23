package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	orgMetaSchema = "organization"
)

func (h *ConnectHandler) GetOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationRequest]) (*connect.Response[frontierv1beta1.GetOrganizationResponse], error) {
	fetchedOrg, err := h.orgService.GetRaw(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, err)
		case errors.Is(err, organization.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	orgPB, err := transformOrgToPB(fetchedOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&frontierv1beta1.GetOrganizationResponse{
		Organization: orgPB,
	}), nil
}

func (h *ConnectHandler) ListOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsResponse], error) {
	var orgs []*frontierv1beta1.Organization
	paginate := pagination.NewPagination(request.Msg.GetPageNum(), request.Msg.GetPageSize())

	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:      organization.State(request.Msg.GetState()),
		UserID:     request.Msg.GetUserId(),
		Pagination: paginate,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		orgs = append(orgs, orgPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationsResponse{
		Organizations: orgs,
	}), nil
}

func (h *ConnectHandler) ListAllOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListAllOrganizationsResponse], error) {
	var orgs []*frontierv1beta1.Organization
	paginate := pagination.NewPagination(request.Msg.GetPageNum(), request.Msg.GetPageSize())

	orgList, err := h.orgService.List(ctx, organization.Filter{
		State:      organization.State(request.Msg.GetState()),
		UserID:     request.Msg.GetUserId(),
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

	return connect.NewResponse(&frontierv1beta1.ListAllOrganizationsResponse{
		Organizations: orgs,
		Count:         paginate.Count,
	}), nil
}

func (h *ConnectHandler) CreateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationResponse], error) {
	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newOrg, err := h.orgService.Create(ctx, organization.Organization{
		Name:     request.Msg.GetBody().GetName(),
		Title:    request.Msg.GetBody().GetTitle(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, newOrg.ID).LogWithAttrs(audit.OrgCreatedEvent, audit.OrgTarget(newOrg.ID), map[string]string{
		"title": newOrg.Title,
		"name":  newOrg.Name,
	})
	return connect.NewResponse(&frontierv1beta1.CreateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) AdminCreateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.AdminCreateOrganizationRequest]) (*connect.Response[frontierv1beta1.AdminCreateOrganizationResponse], error) {
	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newOrg, err := h.orgService.AdminCreate(ctx, organization.Organization{
		Name:     request.Msg.GetBody().GetName(),
		Title:    request.Msg.GetBody().GetTitle(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	}, request.Msg.GetBody().GetOrgOwnerEmail())
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, newOrg.ID).LogWithAttrs(audit.OrgCreatedEvent, audit.OrgTarget(newOrg.ID), map[string]string{
		"title": newOrg.Title,
		"name":  newOrg.Name,
	})
	return connect.NewResponse(&frontierv1beta1.AdminCreateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) UpdateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateOrganizationRequest]) (*connect.Response[frontierv1beta1.UpdateOrganizationResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	var updatedOrg organization.Organization
	var err error
	if utils.IsValidUUID(request.Msg.GetId()) {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			ID:       request.Msg.GetId(),
			Name:     request.Msg.GetBody().GetName(),
			Title:    request.Msg.GetBody().GetTitle(),
			Avatar:   request.Msg.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	} else {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			Name:     request.Msg.GetBody().GetName(),
			Title:    request.Msg.GetBody().GetTitle(),
			Avatar:   request.Msg.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	}
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	orgPB, err := transformOrgToPB(updatedOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, updatedOrg.ID).Log(audit.OrgUpdatedEvent, audit.OrgTarget(updatedOrg.ID))
	return connect.NewResponse(&frontierv1beta1.UpdateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) ListOrganizationProjects(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationProjectsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationProjectsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	projects, err := h.projectService.List(ctx, project.Filter{
		OrgID:           orgResp.ID,
		WithMemberCount: request.Msg.GetWithMemberCount(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var projectPB []*frontierv1beta1.Project
	for _, rel := range projects {
		u, err := transformProjectToPB(rel)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		projectPB = append(projectPB, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationProjectsResponse{Projects: projectPB}), nil
}

func (h *ConnectHandler) ListOrganizationAdmins(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationAdminsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationAdminsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	admins, err := h.userService.ListByOrg(ctx, orgResp.ID, organization.AdminRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var adminsPB []*frontierv1beta1.User
	for _, user := range admins {
		u, err := transformUserToPB(user)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		adminsPB = append(adminsPB, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationAdminsResponse{Users: adminsPB}), nil
}

func (h *ConnectHandler) ListOrganizationUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationUsersRequest]) (*connect.Response[frontierv1beta1.ListOrganizationUsersResponse], error) {
	if len(request.Msg.GetRoleFilters()) > 0 && request.Msg.GetWithRoles() {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrRoleFilter)
	}

	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	var users []user.User
	var rolePairPBs []*frontierv1beta1.ListOrganizationUsersResponse_RolePair

	if len(request.Msg.GetRoleFilters()) > 0 {
		// convert role names to ids if needed
		roleIDs := request.Msg.GetRoleFilters()
		for i, roleFilter := range request.Msg.GetRoleFilters() {
			if !utils.IsValidUUID(roleFilter) {
				role, err := h.roleService.Get(ctx, roleFilter)
				if err != nil {
					return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
				}
				roleIDs[i] = role.ID
			}
		}

		// need to fetch users with roles assigned to them
		policies, err := h.policyService.List(ctx, policy.Filter{
			OrgID:         request.Msg.GetId(),
			PrincipalType: schema.UserPrincipal,
			ResourceType:  schema.OrganizationNamespace,
			RoleIDs:       roleIDs,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		users = utils.Filter(utils.Map(policies, func(pol policy.Policy) user.User {
			u, _ := h.userService.GetByID(ctx, pol.PrincipalID)
			return u
		}), func(u user.User) bool {
			return u.ID != ""
		})
	} else {
		// list all users
		users, err = h.userService.ListByOrg(ctx, orgResp.ID, request.Msg.GetPermissionFilter())
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		if request.Msg.GetWithRoles() {
			for _, user := range users {
				roles, err := h.policyService.ListRoles(ctx, schema.UserPrincipal, user.ID, schema.OrganizationNamespace, request.Msg.GetId())
				if err != nil {
					return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
				}

				rolesPb := utils.Filter(utils.Map(roles, func(role role.Role) *frontierv1beta1.Role {
					pb, err := transformRoleToPB(role)
					if err != nil {
						logger.Error("failed to transform role for user", zap.Error(err))
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
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		usersPB = append(usersPB, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationUsersResponse{
		Users:     usersPB,
		RolePairs: rolePairPBs,
	}), nil
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
