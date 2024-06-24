package v1beta1

import (
	"context"

	"github.com/raystack/frontier/core/group"

	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
	"go.uber.org/zap"

	"github.com/raystack/frontier/core/audit"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

var grpcProjectNotFoundErr = status.Errorf(codes.NotFound, "project doesn't exist")

type ProjectService interface {
	Get(ctx context.Context, idOrName string) (project.Project, error)
	Create(ctx context.Context, prj project.Project) (project.Project, error)
	List(ctx context.Context, f project.Filter) ([]project.Project, error)
	ListByUser(ctx context.Context, principalID, principalType string, flt project.Filter) ([]project.Project, error)
	Update(ctx context.Context, toUpdate project.Project) (project.Project, error)
	ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error)
	ListServiceUsers(ctx context.Context, id string, permissionFilter string) ([]serviceuser.ServiceUser, error)
	ListGroups(ctx context.Context, id string) ([]group.Group, error)
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	AddPrincipal(ctx context.Context, id, roleID string, principal project.Principal) error
}

func (h Handler) ListProjects(
	ctx context.Context,
	request *frontierv1beta1.ListProjectsRequest,
) (*frontierv1beta1.ListProjectsResponse, error) {
	logger := grpczap.Extract(ctx)

	var projects []*frontierv1beta1.Project
	projectList, err := h.projectService.List(ctx, project.Filter{
		State: project.State(request.GetState()),
		OrgID: request.GetOrgId(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, v := range projectList {
		projectPB, err := transformProjectToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		projects = append(projects, projectPB)
	}

	return &frontierv1beta1.ListProjectsResponse{Projects: projects}, nil
}

func (h Handler) CreateProject(
	ctx context.Context,
	request *frontierv1beta1.CreateProjectRequest,
) (*frontierv1beta1.CreateProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	auditor := audit.GetAuditor(ctx, request.GetBody().GetOrgId())

	metaDataMap := map[string]any{}
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())
	}

	prj := project.Project{
		Name:         request.GetBody().GetName(),
		Title:        request.GetBody().GetTitle(),
		Metadata:     metaDataMap,
		Organization: organization.Organization{ID: request.GetBody().GetOrgId()},
	}
	newProject, err := h.projectService.Create(ctx, prj)
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	projectPB, err := transformProjectToPB(newProject)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	auditor.Log(audit.ProjectCreatedEvent, audit.ProjectTarget(newProject.ID))
	return &frontierv1beta1.CreateProjectResponse{Project: projectPB}, nil
}

func (h Handler) GetProject(
	ctx context.Context,
	request *frontierv1beta1.GetProjectRequest,
) (*frontierv1beta1.GetProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedProject, err := h.projectService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	projectPB, err := transformProjectToPB(fetchedProject)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.GetProjectResponse{Project: projectPB}, nil
}

func (h Handler) UpdateProject(
	ctx context.Context,
	request *frontierv1beta1.UpdateProjectRequest,
) (*frontierv1beta1.UpdateProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	auditor := audit.GetAuditor(ctx, request.GetBody().GetOrgId())
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	updatedProject, err := h.projectService.Update(ctx, project.Project{
		ID:           request.GetId(),
		Name:         request.GetBody().GetName(),
		Title:        request.GetBody().GetTitle(),
		Organization: organization.Organization{ID: request.GetBody().GetOrgId()},
		Metadata:     metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	projectPB, err := transformProjectToPB(updatedProject)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	auditor.Log(audit.ProjectUpdatedEvent, audit.ProjectTarget(updatedProject.ID))
	return &frontierv1beta1.UpdateProjectResponse{Project: projectPB}, nil
}

func (h Handler) AddProjectPrincipal(ctx context.Context, request *frontierv1beta1.AddProjectPrincipalRequest) (*frontierv1beta1.AddProjectPrincipalResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetPrincipal() == "" || request.GetId() == "" || request.GetRoleId() == "" {
		return nil, grpcBadBodyError
	}

	principalType, principalId, err := schema.SplitNamespaceAndResourceID(request.GetPrincipal())
	if err != nil {
		return nil, ErrNamespaceSplitNotation
	}

	err = h.projectService.AddPrincipal(ctx, request.GetId(), request.GetRoleId(), project.Principal{
		Type: principalType,
		ID:   principalId,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	return &frontierv1beta1.AddProjectPrincipalResponse{}, nil
}

func (h Handler) ListProjectAdmins(
	ctx context.Context,
	request *frontierv1beta1.ListProjectAdminsRequest,
) (*frontierv1beta1.ListProjectAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	users, err := h.projectService.ListUsers(ctx, request.GetId(), project.AdminPermission)
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	var transformedAdmins []*frontierv1beta1.User
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		transformedAdmins = append(transformedAdmins, u)
	}

	return &frontierv1beta1.ListProjectAdminsResponse{Users: transformedAdmins}, nil
}

func (h Handler) ListProjectUsers(
	ctx context.Context,
	request *frontierv1beta1.ListProjectUsersRequest,
) (*frontierv1beta1.ListProjectUsersResponse, error) {
	logger := grpczap.Extract(ctx)

	permissionFilter := project.MemberPermission
	if len(request.GetPermissionFilter()) > 0 {
		permissionFilter = request.GetPermissionFilter()
	}

	users, err := h.projectService.ListUsers(ctx, request.GetId(), permissionFilter)
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	var transformedUsers []*frontierv1beta1.User
	var rolePairPBs []*frontierv1beta1.ListProjectUsersResponse_RolePair
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		transformedUsers = append(transformedUsers, u)
	}

	if request.GetWithRoles() {
		for _, user := range users {
			roles, err := h.policyService.ListRoles(ctx, schema.UserPrincipal, user.ID, schema.ProjectNamespace, request.GetId())
			if err != nil {
				logger.Error(err.Error())
				return nil, grpcInternalServerError
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

	return &frontierv1beta1.ListProjectUsersResponse{
		Users:     transformedUsers,
		RolePairs: rolePairPBs,
	}, nil
}

func (h Handler) ListProjectServiceUsers(ctx context.Context,
	request *frontierv1beta1.ListProjectServiceUsersRequest) (*frontierv1beta1.ListProjectServiceUsersResponse, error) {
	logger := grpczap.Extract(ctx)

	users, err := h.projectService.ListServiceUsers(ctx, request.GetId(), project.MemberPermission)
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	var transformedUsers []*frontierv1beta1.ServiceUser
	var rolePairPBs []*frontierv1beta1.ListProjectServiceUsersResponse_RolePair
	for _, a := range users {
		u, err := transformServiceUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		transformedUsers = append(transformedUsers, u)
	}

	if request.GetWithRoles() {
		for _, user := range users {
			roles, err := h.policyService.ListRoles(ctx, schema.ServiceUserPrincipal, user.ID, schema.ProjectNamespace, request.GetId())
			if err != nil {
				logger.Error(err.Error())
				return nil, grpcInternalServerError
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

	return &frontierv1beta1.ListProjectServiceUsersResponse{
		Serviceusers: transformedUsers,
		RolePairs:    rolePairPBs,
	}, nil
}
func (h Handler) ListProjectGroups(ctx context.Context, request *frontierv1beta1.ListProjectGroupsRequest) (*frontierv1beta1.ListProjectGroupsResponse, error) {
	logger := grpczap.Extract(ctx)

	groups, err := h.projectService.ListGroups(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}

	var groupsPB []*frontierv1beta1.Group
	var rolePairPBs []*frontierv1beta1.ListProjectGroupsResponse_RolePair
	for _, g := range groups {
		u, err := transformGroupToPB(g)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		groupsPB = append(groupsPB, &u)
	}

	if request.GetWithRoles() {
		for _, group := range groups {
			roles, err := h.policyService.ListRoles(ctx, schema.GroupPrincipal, group.ID,
				schema.ProjectNamespace, request.GetId())
			if err != nil {
				logger.Error(err.Error())
				return nil, grpcInternalServerError
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

	return &frontierv1beta1.ListProjectGroupsResponse{
		Groups:    groupsPB,
		RolePairs: rolePairPBs,
	}, nil
}

func (h Handler) EnableProject(ctx context.Context, request *frontierv1beta1.EnableProjectRequest) (*frontierv1beta1.EnableProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.projectService.Enable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}
	return &frontierv1beta1.EnableProjectResponse{}, nil
}

func (h Handler) DisableProject(ctx context.Context, request *frontierv1beta1.DisableProjectRequest) (*frontierv1beta1.DisableProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.projectService.Disable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, translateServiceError(err)
	}
	return &frontierv1beta1.DisableProjectResponse{}, nil
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

func translateServiceError(err error) error {
	switch {
	case errors.Is(err, user.ErrInvalidEmail):
		return grpcUnauthenticated
	case errors.Is(err, organization.ErrInvalidUUID), errors.Is(err, project.ErrInvalidDetail):
		return grpcBadBodyError
	case errors.Is(err, project.ErrConflict):
		return grpcConflictError
	case errors.Is(err, project.ErrNotExist), errors.Is(err, project.ErrInvalidUUID), errors.Is(err, project.ErrInvalidID):
		return grpcProjectNotFoundErr
	default:
		return grpcInternalServerError
	}
}
