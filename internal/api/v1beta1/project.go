package v1beta1

import (
	"context"

	"github.com/raystack/shield/core/audit"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/pkg/errors"
	"github.com/raystack/shield/pkg/metadata"

	"github.com/raystack/shield/core/organization"
	"github.com/raystack/shield/core/project"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
)

var grpcProjectNotFoundErr = status.Errorf(codes.NotFound, "project doesn't exist")

//go:generate mockery --name=ProjectService -r --case underscore --with-expecter --structname ProjectService --filename project_service.go --output=./mocks
type ProjectService interface {
	Get(ctx context.Context, idOrName string) (project.Project, error)
	Create(ctx context.Context, prj project.Project) (project.Project, error)
	List(ctx context.Context, f project.Filter) ([]project.Project, error)
	Update(ctx context.Context, toUpdate project.Project) (project.Project, error)
	ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error)
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
}

func (h Handler) ListProjects(
	ctx context.Context,
	request *shieldv1beta1.ListProjectsRequest,
) (*shieldv1beta1.ListProjectsResponse, error) {
	logger := grpczap.Extract(ctx)

	var projects []*shieldv1beta1.Project
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

	return &shieldv1beta1.ListProjectsResponse{Projects: projects}, nil
}

func (h Handler) CreateProject(
	ctx context.Context,
	request *shieldv1beta1.CreateProjectRequest,
) (*shieldv1beta1.CreateProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	auditor := audit.GetAuditor(ctx, request.GetBody().GetOrgId())

	metaDataMap := map[string]any{}
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap, err = metadata.Build(request.GetBody().GetMetadata().AsMap())
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcBadBodyError
		}
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
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, organization.ErrInvalidUUID), errors.Is(err, project.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, project.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	projectPB, err := transformProjectToPB(newProject)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	auditor.Log(audit.ProjectCreatedEvent, audit.ProjectTarget(newProject.ID))
	return &shieldv1beta1.CreateProjectResponse{Project: projectPB}, nil
}

func (h Handler) GetProject(
	ctx context.Context,
	request *shieldv1beta1.GetProjectRequest,
) (*shieldv1beta1.GetProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedProject, err := h.projectService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ErrNotExist), errors.Is(err, project.ErrInvalidUUID), errors.Is(err, project.ErrInvalidID):
			return nil, grpcProjectNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	projectPB, err := transformProjectToPB(fetchedProject)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetProjectResponse{Project: projectPB}, nil
}

func (h Handler) UpdateProject(
	ctx context.Context,
	request *shieldv1beta1.UpdateProjectRequest,
) (*shieldv1beta1.UpdateProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	auditor := audit.GetAuditor(ctx, request.GetBody().GetOrgId())
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedProject, err := h.projectService.Update(ctx, project.Project{
		ID:           request.GetId(),
		Name:         request.GetBody().GetName(),
		Organization: organization.Organization{ID: request.GetBody().GetOrgId()},
		Metadata:     metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ErrNotExist),
			errors.Is(err, project.ErrInvalidUUID),
			errors.Is(err, project.ErrInvalidID),
			errors.Is(err, organization.ErrInvalidUUID):
			return nil, grpcProjectNotFoundErr
		case errors.Is(err, project.ErrConflict):
			return nil, grpcConflictError
		case errors.Is(err, project.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	projectPB, err := transformProjectToPB(updatedProject)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	auditor.Log(audit.ProjectUpdatedEvent, audit.ProjectTarget(updatedProject.ID))
	return &shieldv1beta1.UpdateProjectResponse{Project: projectPB}, nil
}

func (h Handler) ListProjectAdmins(
	ctx context.Context,
	request *shieldv1beta1.ListProjectAdminsRequest,
) (*shieldv1beta1.ListProjectAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	users, err := h.projectService.ListUsers(ctx, request.GetId(), project.AdminPermission)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ErrNotExist):
			return nil, grpcProjectNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	var transformedAdmins []*shieldv1beta1.User
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		transformedAdmins = append(transformedAdmins, u)
	}

	return &shieldv1beta1.ListProjectAdminsResponse{Users: transformedAdmins}, nil
}

func (h Handler) ListProjectUsers(
	ctx context.Context,
	request *shieldv1beta1.ListProjectUsersRequest,
) (*shieldv1beta1.ListProjectUsersResponse, error) {
	logger := grpczap.Extract(ctx)

	permissionFilter := project.MemberPermission
	if len(request.GetPermissionFilter()) > 0 {
		permissionFilter = request.GetPermissionFilter()
	}

	users, err := h.projectService.ListUsers(ctx, request.GetId(), permissionFilter)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ErrNotExist):
			return nil, grpcProjectNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	var transformedUsers []*shieldv1beta1.User
	for _, a := range users {
		u, err := transformUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		transformedUsers = append(transformedUsers, u)
	}

	return &shieldv1beta1.ListProjectUsersResponse{Users: transformedUsers}, nil
}

func (h Handler) EnableProject(ctx context.Context, request *shieldv1beta1.EnableProjectRequest) (*shieldv1beta1.EnableProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.projectService.Enable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.EnableProjectResponse{}, nil
}

func (h Handler) DisableProject(ctx context.Context, request *shieldv1beta1.DisableProjectRequest) (*shieldv1beta1.DisableProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.projectService.Disable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.DisableProjectResponse{}, nil
}

func transformProjectToPB(prj project.Project) (*shieldv1beta1.Project, error) {
	metaData, err := prj.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &shieldv1beta1.Project{
		Id:        prj.ID,
		Name:      prj.Name,
		Title:     prj.Title,
		OrgId:     prj.Organization.ID,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(prj.CreatedAt),
		UpdatedAt: timestamppb.New(prj.UpdatedAt),
	}, nil
}
