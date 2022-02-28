package v1beta1

import (
	"context"
	"errors"
	"strings"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/internal/project"
	"github.com/odpf/shield/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var grpcProjectNotFoundErr = status.Errorf(codes.NotFound, "project doesn't exist")

type ProjectService interface {
	Get(ctx context.Context, id string) (model.Project, error)
	Create(ctx context.Context, project model.Project) (model.Project, error)
	List(ctx context.Context) ([]model.Project, error)
	Update(ctx context.Context, toUpdate model.Project) (model.Project, error)
	AddAdmin(ctx context.Context, id string, userIds []string) ([]model.User, error)
	ListAdmins(ctx context.Context, id string) ([]model.User, error)
	RemoveAdmin(ctx context.Context, id string, userId string) ([]model.User, error)
}

func (v Dep) ListProjects(ctx context.Context, request *shieldv1beta1.ListProjectsRequest) (*shieldv1beta1.ListProjectsResponse, error) {
	logger := grpczap.Extract(ctx)
	var projects []*shieldv1beta1.Project

	projectList, err := v.ProjectService.List(ctx)
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

func (v Dep) CreateProject(ctx context.Context, request *shieldv1beta1.CreateProjectRequest) (*shieldv1beta1.CreateProjectResponse, error) {
	logger := grpczap.Extract(ctx)
	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	slug := request.GetBody().Slug
	if strings.TrimSpace(slug) == "" {
		slug = generateSlug(request.GetBody().Name)
	}

	newProject, err := v.ProjectService.Create(ctx, model.Project{
		Name:         request.GetBody().Name,
		Slug:         slug,
		Metadata:     metaDataMap,
		Organization: model.Organization{Id: request.GetBody().OrgId},
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	metaData, err := structpb.NewStruct(mapOfInterfaceValues(newProject.Metadata))
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateProjectResponse{Project: &shieldv1beta1.Project{
		Id:        newProject.Id,
		Name:      newProject.Name,
		Slug:      newProject.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newProject.CreatedAt),
		UpdatedAt: timestamppb.New(newProject.UpdatedAt),
	}}, nil
}

func (v Dep) GetProject(ctx context.Context, request *shieldv1beta1.GetProjectRequest) (*shieldv1beta1.GetProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedProject, err := v.ProjectService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ProjectDoesntExist):
			return nil, grpcProjectNotFoundErr
		case errors.Is(err, project.InvalidUUID):
			return nil, grpcBadBodyError
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

func (v Dep) UpdateProject(ctx context.Context, request *shieldv1beta1.UpdateProjectRequest) (*shieldv1beta1.UpdateProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedProject, err := v.ProjectService.Update(ctx, model.Project{
		Id:           request.GetId(),
		Name:         request.GetBody().Name,
		Slug:         request.GetBody().Slug,
		Organization: model.Organization{Id: request.GetBody().OrgId},
		Metadata:     metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	projectPB, err := transformProjectToPB(updatedProject)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateProjectResponse{Project: &projectPB}, nil
}

func (v Dep) AddProjectAdmin(ctx context.Context, request *shieldv1beta1.AddProjectAdminRequest) (*shieldv1beta1.AddProjectAdminResponse, error) {
	logger := grpczap.Extract(ctx)
	userIds := request.GetBody().UserIds

	admins, err := v.ProjectService.AddAdmin(ctx, request.GetId(), userIds)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ProjectDoesntExist):
			return nil, status.Errorf(codes.NotFound, "project to be updated not found")
		default:
			return nil, grpcInternalServerError
		}
	}

	var transformedAdmins []*shieldv1beta1.User
	for _, a := range admins {
		userPB, err := transformUserToPB(a)
		if err != nil {
			logger.Error(err.Error())
			return nil, internalServerError
		}

		transformedAdmins = append(transformedAdmins, &userPB)
	}

	return &shieldv1beta1.AddProjectAdminResponse{Users: transformedAdmins}, nil
}

func (v Dep) ListProjectAdmins(ctx context.Context, request *shieldv1beta1.ListProjectAdminsRequest) (*shieldv1beta1.ListProjectAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	admins, err := v.ProjectService.ListAdmins(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ProjectDoesntExist):
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
			return nil, internalServerError
		}

		transformedAdmins = append(transformedAdmins, &u)
	}

	return &shieldv1beta1.ListProjectAdminsResponse{Users: transformedAdmins}, nil
}

func (v Dep) RemoveProjectAdmin(ctx context.Context, request *shieldv1beta1.RemoveProjectAdminRequest) (*shieldv1beta1.RemoveProjectAdminResponse, error) {
	logger := grpczap.Extract(ctx)

	_, err := v.ProjectService.RemoveAdmin(ctx, request.GetId(), request.GetUserId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, project.ProjectDoesntExist):
			return nil, status.Errorf(codes.NotFound, "project to be updated not found")
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.RemoveProjectAdminResponse{
		Message: "Removed Admin from org",
	}, nil
}

func transformProjectToPB(prj model.Project) (shieldv1beta1.Project, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(prj.Metadata))
	if err != nil {
		return shieldv1beta1.Project{}, err
	}

	return shieldv1beta1.Project{
		Id:        prj.Id,
		Name:      prj.Name,
		Slug:      prj.Slug,
		OrgId:     prj.Organization.Id,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(prj.CreatedAt),
		UpdatedAt: timestamppb.New(prj.UpdatedAt),
	}, nil
}
