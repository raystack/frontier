package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	"go.uber.org/zap"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListProjects(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectsRequest]) (*connect.Response[frontierv1beta1.ListProjectsResponse], error) {
	var projects []*frontierv1beta1.Project
	projectList, err := h.projectService.List(ctx, project.Filter{
		State: project.State(request.Msg.GetState()),
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range projectList {
		projectPB, err := transformProjectToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		projects = append(projects, projectPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectsResponse{Projects: projects}), nil
}

func (h *ConnectHandler) CreateProject(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProjectRequest]) (*connect.Response[frontierv1beta1.CreateProjectResponse], error) {
	auditor := audit.GetAuditor(ctx, request.Msg.GetBody().GetOrgId())

	metaDataMap := map[string]any{}
	var err error
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	}

	prj := project.Project{
		Name:         request.Msg.GetBody().GetName(),
		Title:        request.Msg.GetBody().GetTitle(),
		Metadata:     metaDataMap,
		Organization: organization.Organization{ID: request.Msg.GetBody().GetOrgId()},
	}
	newProject, err := h.projectService.Create(ctx, prj)
	if err != nil {
		return nil, translateProjectServiceError(err)
	}

	projectPB, err := transformProjectToPB(newProject)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	auditor.Log(audit.ProjectCreatedEvent, audit.ProjectTarget(newProject.ID))
	return connect.NewResponse(&frontierv1beta1.CreateProjectResponse{Project: projectPB}), nil
}

func (h *ConnectHandler) GetProject(ctx context.Context, request *connect.Request[frontierv1beta1.GetProjectRequest]) (*connect.Response[frontierv1beta1.GetProjectResponse], error) {
	fetchedProject, err := h.projectService.Get(ctx, request.Msg.GetId())
	if err != nil {
		return nil, translateProjectServiceError(err)
	}

	projectPB, err := transformProjectToPB(fetchedProject)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetProjectResponse{Project: projectPB}), nil
}

func (h *ConnectHandler) UpdateProject(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateProjectRequest]) (*connect.Response[frontierv1beta1.UpdateProjectResponse], error) {
	auditor := audit.GetAuditor(ctx, request.Msg.GetBody().GetOrgId())
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	updatedProject, err := h.projectService.Update(ctx, project.Project{
		ID:           request.Msg.GetId(),
		Name:         request.Msg.GetBody().GetName(),
		Title:        request.Msg.GetBody().GetTitle(),
		Organization: organization.Organization{ID: request.Msg.GetBody().GetOrgId()},
		Metadata:     metaDataMap,
	})
	if err != nil {
		return nil, translateProjectServiceError(err)
	}

	projectPB, err := transformProjectToPB(updatedProject)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	auditor.Log(audit.ProjectUpdatedEvent, audit.ProjectTarget(updatedProject.ID))
	return connect.NewResponse(&frontierv1beta1.UpdateProjectResponse{Project: projectPB}), nil
}

func (h *ConnectHandler) ListProjectAdmins(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectAdminsRequest]) (*connect.Response[frontierv1beta1.ListProjectAdminsResponse], error) {
	users, err := h.projectService.ListUsers(ctx, request.Msg.GetId(), project.AdminPermission)
	if err != nil {
		return nil, translateProjectServiceError(err)
	}

	var transformedAdmins []*frontierv1beta1.User
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		transformedAdmins = append(transformedAdmins, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectAdminsResponse{Users: transformedAdmins}), nil
}

func (h *ConnectHandler) ListProjectUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectUsersRequest]) (*connect.Response[frontierv1beta1.ListProjectUsersResponse], error) {
	logger := grpczap.Extract(ctx)

	permissionFilter := project.MemberPermission
	if len(request.Msg.GetPermissionFilter()) > 0 {
		permissionFilter = request.Msg.GetPermissionFilter()
	}

	users, err := h.projectService.ListUsers(ctx, request.Msg.GetId(), permissionFilter)
	if err != nil {
		return nil, translateProjectServiceError(err)
	}

	var transformedUsers []*frontierv1beta1.User
	var rolePairPBs []*frontierv1beta1.ListProjectUsersResponse_RolePair
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		transformedUsers = append(transformedUsers, u)
	}

	if request.Msg.GetWithRoles() {
		for _, user := range users {
			roles, err := h.policyService.ListRoles(ctx, schema.UserPrincipal, user.ID, schema.ProjectNamespace, request.Msg.GetId())
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
			rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListProjectUsersResponse_RolePair{
				UserId: user.ID,
				Roles:  rolesPb,
			})
		}
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectUsersResponse{
		Users:     transformedUsers,
		RolePairs: rolePairPBs,
	}), nil
}

func (h *ConnectHandler) ListProjectServiceUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListProjectServiceUsersResponse], error) {
	logger := grpczap.Extract(ctx)

	users, err := h.projectService.ListServiceUsers(ctx, request.Msg.GetId(), project.MemberPermission)
	if err != nil {
		return nil, translateProjectServiceError(err)
	}

	var transformedUsers []*frontierv1beta1.ServiceUser
	var rolePairPBs []*frontierv1beta1.ListProjectServiceUsersResponse_RolePair
	for _, a := range users {
		u, err := transformServiceUserToPB(a)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		transformedUsers = append(transformedUsers, u)
	}

	if request.Msg.GetWithRoles() {
		for _, user := range users {
			roles, err := h.policyService.ListRoles(ctx, schema.ServiceUserPrincipal, user.ID, schema.ProjectNamespace, request.Msg.GetId())
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
			rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListProjectServiceUsersResponse_RolePair{
				ServiceuserId: user.ID,
				Roles:         rolesPb,
			})
		}
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectServiceUsersResponse{
		Serviceusers: transformedUsers,
		RolePairs:    rolePairPBs,
	}), nil
}

func (h *ConnectHandler) ListProjectGroups(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectGroupsRequest]) (*connect.Response[frontierv1beta1.ListProjectGroupsResponse], error) {
	logger := grpczap.Extract(ctx)

	groups, err := h.projectService.ListGroups(ctx, request.Msg.GetId())
	if err != nil {
		return nil, translateProjectServiceError(err)
	}

	var groupsPB []*frontierv1beta1.Group
	var rolePairPBs []*frontierv1beta1.ListProjectGroupsResponse_RolePair
	for _, g := range groups {
		u, err := transformGroupToPB(g)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		groupsPB = append(groupsPB, &u)
	}

	if request.Msg.GetWithRoles() {
		for _, group := range groups {
			roles, err := h.policyService.ListRoles(ctx, schema.GroupPrincipal, group.ID,
				schema.ProjectNamespace, request.Msg.GetId())
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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

			rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListProjectGroupsResponse_RolePair{
				GroupId: group.ID,
				Roles:   rolesPb,
			})
		}
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectGroupsResponse{
		Groups:    groupsPB,
		RolePairs: rolePairPBs,
	}), nil
}

func (h *ConnectHandler) EnableProject(ctx context.Context, request *connect.Request[frontierv1beta1.EnableProjectRequest]) (*connect.Response[frontierv1beta1.EnableProjectResponse], error) {
	if err := h.projectService.Enable(ctx, request.Msg.GetId()); err != nil {
		return nil, translateProjectServiceError(err)
	}
	return connect.NewResponse(&frontierv1beta1.EnableProjectResponse{}), nil
}

func (h *ConnectHandler) DisableProject(ctx context.Context, request *connect.Request[frontierv1beta1.DisableProjectRequest]) (*connect.Response[frontierv1beta1.DisableProjectResponse], error) {
	if err := h.projectService.Disable(ctx, request.Msg.GetId()); err != nil {
		return nil, translateProjectServiceError(err)
	}
	return connect.NewResponse(&frontierv1beta1.DisableProjectResponse{}), nil
}

func transformProjectToPB(prj project.Project) (*frontierv1beta1.Project, error) {
	metaData, err := prj.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Project{
		Id:           prj.ID,
		Name:         prj.Name,
		Title:        prj.Title,
		OrgId:        prj.Organization.ID,
		Metadata:     metaData,
		CreatedAt:    timestamppb.New(prj.CreatedAt),
		UpdatedAt:    timestamppb.New(prj.UpdatedAt),
		MembersCount: int32(prj.MemberCount),
	}, nil
}

func translateProjectServiceError(err error) error {
	switch {
	case errors.Is(err, user.ErrInvalidEmail):
		return connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	case errors.Is(err, organization.ErrInvalidUUID), errors.Is(err, project.ErrInvalidDetail):
		return connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	case errors.Is(err, project.ErrConflict):
		return connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
	case errors.Is(err, project.ErrNotExist), errors.Is(err, project.ErrInvalidUUID), errors.Is(err, project.ErrInvalidID):
		return connect.NewError(connect.CodeNotFound, ErrNotFound)
	default:
		return connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
}
