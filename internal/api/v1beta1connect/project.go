package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListProjects(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectsRequest]) (*connect.Response[frontierv1beta1.ListProjectsResponse], error) {
	errorLogger := NewErrorLogger()
	var projects []*frontierv1beta1.Project

	projectList, err := h.projectService.List(ctx, project.Filter{
		State: project.State(request.Msg.GetState()),
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjects", err,
			"org_id", request.Msg.GetOrgId(),
			"state", request.Msg.GetState())
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range projectList {
		projectPB, err := transformProjectToPB(v)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProjects", v.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		projects = append(projects, projectPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectsResponse{Projects: projects}), nil
}

func (h *ConnectHandler) CreateProject(ctx context.Context, request *connect.Request[frontierv1beta1.CreateProjectRequest]) (*connect.Response[frontierv1beta1.CreateProjectResponse], error) {
	errorLogger := NewErrorLogger()
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
		errorLogger.LogServiceError(ctx, request, "CreateProject", err,
			"project_name", request.Msg.GetBody().GetName(),
			"org_id", request.Msg.GetBody().GetOrgId())
		return nil, translateProjectServiceError(err)
	}

	projectPB, err := transformProjectToPB(newProject)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "CreateProject", newProject.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	auditor.Log(audit.ProjectCreatedEvent, audit.ProjectTarget(newProject.ID))
	return connect.NewResponse(&frontierv1beta1.CreateProjectResponse{Project: projectPB}), nil
}

func (h *ConnectHandler) GetProject(ctx context.Context, request *connect.Request[frontierv1beta1.GetProjectRequest]) (*connect.Response[frontierv1beta1.GetProjectResponse], error) {
	errorLogger := NewErrorLogger()
	projectID := request.Msg.GetId()

	fetchedProject, err := h.projectService.Get(ctx, projectID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetProject", err,
			"project_id", projectID)
		return nil, translateProjectServiceError(err)
	}

	projectPB, err := transformProjectToPB(fetchedProject)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetProject", fetchedProject.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetProjectResponse{Project: projectPB}), nil
}

func (h *ConnectHandler) UpdateProject(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateProjectRequest]) (*connect.Response[frontierv1beta1.UpdateProjectResponse], error) {
	errorLogger := NewErrorLogger()
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
		errorLogger.LogServiceError(ctx, request, "UpdateProject", err,
			"project_id", request.Msg.GetId(),
			"project_name", request.Msg.GetBody().GetName(),
			"org_id", request.Msg.GetBody().GetOrgId())
		return nil, translateProjectServiceError(err)
	}

	projectPB, err := transformProjectToPB(updatedProject)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "UpdateProject", updatedProject.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	auditor.Log(audit.ProjectUpdatedEvent, audit.ProjectTarget(updatedProject.ID))
	return connect.NewResponse(&frontierv1beta1.UpdateProjectResponse{Project: projectPB}), nil
}

func (h *ConnectHandler) ListProjectAdmins(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectAdminsRequest]) (*connect.Response[frontierv1beta1.ListProjectAdminsResponse], error) {
	errorLogger := NewErrorLogger()
	projectID := request.Msg.GetId()

	prj, err := h.projectService.Get(ctx, projectID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectAdmins.Get", err,
			"project_id", projectID)
		return nil, translateProjectServiceError(err)
	}

	members, err := h.membershipService.ListPrincipalsByResource(ctx, prj.ID, schema.ProjectNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectAdmins.ListPrincipalsByResource", err,
			"project_id", prj.ID)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	userIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	users, err := h.userService.GetByIDs(ctx, userIDs)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectAdmins.GetByIDs", err,
			"project_id", prj.ID)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var transformedAdmins []*frontierv1beta1.User
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProjectAdmins", a.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		transformedAdmins = append(transformedAdmins, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectAdminsResponse{Users: transformedAdmins}), nil
}

func (h *ConnectHandler) ListProjectUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectUsersRequest]) (*connect.Response[frontierv1beta1.ListProjectUsersResponse], error) {
	errorLogger := NewErrorLogger()
	projectID := request.Msg.GetId()

	prj, err := h.projectService.Get(ctx, projectID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectUsers.Get", err,
			"project_id", projectID)
		return nil, translateProjectServiceError(err)
	}

	members, err := h.membershipService.ListPrincipalsByResource(ctx, prj.ID, schema.ProjectNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
		WithRoles:     request.Msg.GetWithRoles(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectUsers.ListPrincipalsByResource", err,
			"project_id", prj.ID)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	userIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	users, err := h.userService.GetByIDs(ctx, userIDs)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectUsers.GetByIDs", err,
			"project_id", prj.ID)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var transformedUsers []*frontierv1beta1.User
	rolePairPBs := []*frontierv1beta1.ListProjectUsersResponse_RolePair{}
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProjectUsers", a.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		transformedUsers = append(transformedUsers, u)
	}

	if request.Msg.GetWithRoles() {
		for _, m := range members {
			rolesPb := utils.Filter(utils.Map(m.Roles, func(r role.Role) *frontierv1beta1.Role {
				pb, err := transformRoleToPB(r)
				if err != nil {
					errorLogger.LogTransformError(ctx, request, "ListProjectUsers.TransformRole", r.ID, err)
					return nil
				}
				return &pb
			}), func(r *frontierv1beta1.Role) bool {
				return r != nil
			})
			rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListProjectUsersResponse_RolePair{
				UserId: m.PrincipalID,
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
	errorLogger := NewErrorLogger()
	projectID := request.Msg.GetId()

	prj, err := h.projectService.Get(ctx, projectID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectServiceUsers.Get", err,
			"project_id", projectID)
		return nil, translateProjectServiceError(err)
	}

	members, err := h.membershipService.ListPrincipalsByResource(ctx, prj.ID, schema.ProjectNamespace, membership.MemberFilter{
		PrincipalType: schema.ServiceUserPrincipal,
		WithRoles:     request.Msg.GetWithRoles(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectServiceUsers.ListPrincipalsByResource", err,
			"project_id", prj.ID)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	suIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	var users []serviceuser.ServiceUser
	if len(suIDs) > 0 {
		users, err = h.serviceUserService.GetByIDs(ctx, suIDs)
		if err != nil {
			errorLogger.LogServiceError(ctx, request, "ListProjectServiceUsers.GetByIDs", err,
				"project_id", prj.ID)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	var transformedUsers []*frontierv1beta1.ServiceUser
	rolePairPBs := []*frontierv1beta1.ListProjectServiceUsersResponse_RolePair{}
	for _, a := range users {
		u, err := transformServiceUserToPB(a)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProjectServiceUsers", a.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		transformedUsers = append(transformedUsers, u)
	}

	if request.Msg.GetWithRoles() {
		for _, m := range members {
			rolesPb := utils.Filter(utils.Map(m.Roles, func(r role.Role) *frontierv1beta1.Role {
				pb, err := transformRoleToPB(r)
				if err != nil {
					errorLogger.LogTransformError(ctx, request, "ListProjectServiceUsers.TransformRole", r.ID, err)
					return nil
				}
				return &pb
			}), func(r *frontierv1beta1.Role) bool {
				return r != nil
			})
			rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListProjectServiceUsersResponse_RolePair{
				ServiceuserId: m.PrincipalID,
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
	errorLogger := NewErrorLogger()
	projectID := request.Msg.GetId()

	prj, err := h.projectService.Get(ctx, projectID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectGroups.Get", err,
			"project_id", projectID)
		return nil, translateProjectServiceError(err)
	}

	members, err := h.membershipService.ListPrincipalsByResource(ctx, prj.ID, schema.ProjectNamespace, membership.MemberFilter{
		PrincipalType: schema.GroupPrincipal,
		WithRoles:     request.Msg.GetWithRoles(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListProjectGroups.ListPrincipalsByResource", err,
			"project_id", prj.ID)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	groupIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	var groups []group.Group
	if len(groupIDs) > 0 {
		groups, err = h.groupService.GetByIDs(ctx, groupIDs)
		if err != nil {
			errorLogger.LogServiceError(ctx, request, "ListProjectGroups.GetByIDs", err,
				"project_id", prj.ID)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	var groupsPB []*frontierv1beta1.Group
	rolePairPBs := []*frontierv1beta1.ListProjectGroupsResponse_RolePair{}
	for _, g := range groups {
		u, err := transformGroupToPB(g)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListProjectGroups", g.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		groupsPB = append(groupsPB, &u)
	}

	if request.Msg.GetWithRoles() {
		for _, m := range members {
			rolesPb := utils.Filter(utils.Map(m.Roles, func(r role.Role) *frontierv1beta1.Role {
				pb, err := transformRoleToPB(r)
				if err != nil {
					errorLogger.LogTransformError(ctx, request, "ListProjectGroups.TransformRole", r.ID, err)
					return nil
				}
				return &pb
			}), func(r *frontierv1beta1.Role) bool {
				return r != nil
			})

			rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListProjectGroupsResponse_RolePair{
				GroupId: m.PrincipalID,
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
	errorLogger := NewErrorLogger()
	projectID := request.Msg.GetId()

	if err := h.projectService.Enable(ctx, projectID); err != nil {
		errorLogger.LogServiceError(ctx, request, "EnableProject", err,
			"project_id", projectID)
		return nil, translateProjectServiceError(err)
	}
	return connect.NewResponse(&frontierv1beta1.EnableProjectResponse{}), nil
}

func (h *ConnectHandler) DisableProject(ctx context.Context, request *connect.Request[frontierv1beta1.DisableProjectRequest]) (*connect.Response[frontierv1beta1.DisableProjectResponse], error) {
	errorLogger := NewErrorLogger()
	projectID := request.Msg.GetId()

	if err := h.projectService.Disable(ctx, projectID); err != nil {
		errorLogger.LogServiceError(ctx, request, "DisableProject", err,
			"project_id", projectID)
		return nil, translateProjectServiceError(err)
	}
	return connect.NewResponse(&frontierv1beta1.DisableProjectResponse{}), nil
}

func (h *ConnectHandler) SetProjectMemberRole(ctx context.Context, request *connect.Request[frontierv1beta1.SetProjectMemberRoleRequest]) (*connect.Response[frontierv1beta1.SetProjectMemberRoleResponse], error) {
	errorLogger := NewErrorLogger()

	projectID := request.Msg.GetProjectId()
	principalID := request.Msg.GetPrincipalId()
	principalType := request.Msg.GetPrincipalType()
	roleID := request.Msg.GetRoleId()

	if err := h.membershipService.SetProjectMemberRole(ctx, projectID, principalID, principalType, roleID); err != nil {
		errorLogger.LogServiceError(ctx, request, "SetProjectMemberRole", err,
			"project_id", projectID,
			"principal_id", principalID,
			"principal_type", principalType,
			"role_id", roleID)

		switch {
		case errors.Is(err, project.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, serviceuser.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrServiceUserNotFound)
		case errors.Is(err, group.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		case errors.Is(err, membership.ErrNotOrgMember):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrNotMember)
		case errors.Is(err, user.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrBadRequest)
		case errors.Is(err, role.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrInvalidRoleID)
		case errors.Is(err, role.ErrInvalidID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidRoleID)
		case errors.Is(err, membership.ErrInvalidProjectRole):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidProjectRole)
		case errors.Is(err, membership.ErrInvalidPrincipalType):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	audit.GetAuditor(ctx, "").LogWithAttrs(audit.ProjectMemberRoleChangedEvent, audit.ProjectTarget(projectID), map[string]string{
		"principal_id":   principalID,
		"principal_type": principalType,
		"role_id":        roleID,
	})
	return connect.NewResponse(&frontierv1beta1.SetProjectMemberRoleResponse{}), nil
}

func (h *ConnectHandler) RemoveProjectMember(ctx context.Context, request *connect.Request[frontierv1beta1.RemoveProjectMemberRequest]) (*connect.Response[frontierv1beta1.RemoveProjectMemberResponse], error) {
	errorLogger := NewErrorLogger()

	projectID := request.Msg.GetProjectId()
	principalID := request.Msg.GetPrincipalId()
	principalType := request.Msg.GetPrincipalType()

	if err := h.membershipService.RemoveProjectMember(ctx, projectID, principalID, principalType); err != nil {
		errorLogger.LogServiceError(ctx, request, "RemoveProjectMember", err,
			"project_id", projectID,
			"principal_id", principalID,
			"principal_type", principalType)

		switch {
		case errors.Is(err, project.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrProjectNotFound)
		case errors.Is(err, membership.ErrNotMember):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotMember)
		case errors.Is(err, membership.ErrInvalidPrincipalType):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	audit.GetAuditor(ctx, "").LogWithAttrs(audit.ProjectMemberRemovedEvent, audit.ProjectTarget(projectID), map[string]string{
		"principal_id":   principalID,
		"principal_type": principalType,
	})
	return connect.NewResponse(&frontierv1beta1.RemoveProjectMemberResponse{}), nil
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
