package v1beta1

import (
	"context"
	"errors"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/metadata"
	"github.com/odpf/shield/pkg/str"
	"github.com/odpf/shield/pkg/uuid"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var grpcProjectNotFoundErr = status.Errorf(codes.NotFound, "project doesn't exist")

//go:generate mockery --name=ProjectService -r --case underscore --with-expecter --structname ProjectService --filename project_service.go --output=./mocks
type ProjectService interface {
	Get(ctx context.Context, idOrSlugd string) (project.Project, error)
	Create(ctx context.Context, prj project.Project) (project.Project, error)
	List(ctx context.Context) ([]project.Project, error)
	Update(ctx context.Context, toUpdate project.Project) (project.Project, error)
	AddAdmins(ctx context.Context, idOrSlug string, userIds []string) ([]user.User, error)
	RemoveAdmin(ctx context.Context, idOrSlug string, userId string) ([]user.User, error)
	ListAdmins(ctx context.Context, id string) ([]user.User, error)
}

func (h Handler) ListProjects(
	ctx context.Context,
	request *shieldv1beta1.ListProjectsRequest,
) (*shieldv1beta1.ListProjectsResponse, error) {
	logger := grpczap.Extract(ctx)
	var projects []*shieldv1beta1.Project

	projectList, err := h.projectService.List(ctx)
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

		projects = append(projects, &projectPB)
	}

	return &shieldv1beta1.ListProjectsResponse{Projects: projects}, nil
}

func (h Handler) CreateProject(
	ctx context.Context,
	request *shieldv1beta1.CreateProjectRequest,
) (*shieldv1beta1.CreateProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	prj := project.Project{
		Name:         request.GetBody().GetName(),
		Slug:         request.GetBody().GetSlug(),
		Metadata:     metaDataMap,
		Organization: organization.Organization{ID: request.GetBody().GetOrgId()},
	}

	if str.IsStringEmpty(prj.Slug) {
		prj.Slug = str.GenerateSlug(prj.Name)
	}

	newProject, err := h.projectService.Create(ctx, prj)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcPermissionDenied
		case errors.Is(err, organization.ErrInvalidUUID), errors.Is(err, project.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, project.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	metaData, err := newProject.Metadata.ToStructPB()
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateProjectResponse{Project: &shieldv1beta1.Project{
		Id:        newProject.ID,
		Name:      newProject.Name,
		Slug:      newProject.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newProject.CreatedAt),
		UpdatedAt: timestamppb.New(newProject.UpdatedAt),
	}}, nil
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

	return &shieldv1beta1.GetProjectResponse{Project: &projectPB}, nil
}

func (h Handler) UpdateProject(
	ctx context.Context,
	request *shieldv1beta1.UpdateProjectRequest,
) (*shieldv1beta1.UpdateProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	var updatedProject project.Project
	if uuid.IsValid(request.GetId()) {
		updatedProject, err = h.projectService.Update(ctx, project.Project{
			ID:           request.GetId(),
			Name:         request.GetBody().GetName(),
			Slug:         request.GetBody().GetSlug(),
			Organization: organization.Organization{ID: request.GetBody().GetOrgId()},
			Metadata:     metaDataMap,
		})
	} else {
		updatedProject, err = h.projectService.Update(ctx, project.Project{
			Name:         request.GetBody().GetName(),
			Slug:         request.GetId(),
			Organization: organization.Organization{ID: request.GetBody().GetOrgId()},
			Metadata:     metaDataMap,
		})
	}
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ErrNotExist),
			errors.Is(err, project.ErrInvalidUUID),
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

	return &shieldv1beta1.UpdateProjectResponse{Project: &projectPB}, nil
}

func (h Handler) AddProjectAdmin(
	ctx context.Context,
	request *shieldv1beta1.AddProjectAdminRequest,
) (*shieldv1beta1.AddProjectAdminResponse, error) {
	logger := grpczap.Extract(ctx)

	var admins []user.User
	admins, err := h.projectService.AddAdmins(ctx, request.GetId(), request.GetBody().GetUserIds())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcPermissionDenied
		case errors.Is(err, project.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "project to be updated not found")
		case errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	var transformedAdmins []*shieldv1beta1.User
	for _, a := range admins {
		userPB, err := transformUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		transformedAdmins = append(transformedAdmins, &userPB)
	}

	return &shieldv1beta1.AddProjectAdminResponse{Users: transformedAdmins}, nil
}

func (h Handler) ListProjectAdmins(
	ctx context.Context,
	request *shieldv1beta1.ListProjectAdminsRequest,
) (*shieldv1beta1.ListProjectAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	admins, err := h.projectService.ListAdmins(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "project to be updated not found")
		default:
			return nil, grpcInternalServerError
		}
	}

	var transformedAdmins []*shieldv1beta1.User
	for _, a := range admins {
		u, err := transformUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternalServer
		}

		transformedAdmins = append(transformedAdmins, &u)
	}

	return &shieldv1beta1.ListProjectAdminsResponse{Users: transformedAdmins}, nil
}

func (h Handler) RemoveProjectAdmin(
	ctx context.Context,
	request *shieldv1beta1.RemoveProjectAdminRequest,
) (*shieldv1beta1.RemoveProjectAdminResponse, error) {
	logger := grpczap.Extract(ctx)

	if _, err := h.projectService.RemoveAdmin(ctx, request.GetId(), request.GetUserId()); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcPermissionDenied
		case errors.Is(err, project.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "project to be updated not found")
		case errors.Is(err, user.ErrInvalidUUID),
			errors.Is(err, user.ErrNotExist):
			return nil, grpcUserNotFoundError
		case errors.Is(err, relation.ErrNotExist):
			return nil, grpcRelationNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.RemoveProjectAdminResponse{
		Message: "Removed Admin from project",
	}, nil
}

func transformProjectToPB(prj project.Project) (shieldv1beta1.Project, error) {
	metaData, err := prj.Metadata.ToStructPB()
	if err != nil {
		return shieldv1beta1.Project{}, err
	}

	return shieldv1beta1.Project{
		Id:        prj.ID,
		Name:      prj.Name,
		Slug:      prj.Slug,
		OrgId:     prj.Organization.ID,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(prj.CreatedAt),
		UpdatedAt: timestamppb.New(prj.UpdatedAt),
	}, nil
}
