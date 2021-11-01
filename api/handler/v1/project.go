package v1

import (
	"context"
	"errors"
	"strings"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/internal/project"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var grpcProjectNotFoundErr = status.Errorf(codes.NotFound, "project doesn't exist")

type ProjectService interface {
	GetProject(ctx context.Context, id string) (project.Project, error)
	CreateProject(ctx context.Context, project project.Project) (project.Project, error)
	ListProject(ctx context.Context) ([]project.Project, error)
	UpdateProject(ctx context.Context, toUpdate project.Project) (project.Project, error)
}

func (v Dep) ListProjects(ctx context.Context, request *shieldv1.ListProjectsRequest) (*shieldv1.ListProjectsResponse, error) {
	logger := grpczap.Extract(ctx)
	var projects []*shieldv1.Project

	projectList, err := v.ProjectService.ListProject(ctx)
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

	return &shieldv1.ListProjectsResponse{Projects: projects}, nil
}

func (v Dep) CreateProject(ctx context.Context, request *shieldv1.CreateProjectRequest) (*shieldv1.CreateProjectResponse, error) {
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

	newProject, err := v.ProjectService.CreateProject(ctx, project.Project{
		Name:     request.GetBody().Name,
		Slug:     slug,
		Metadata: metaDataMap,
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

	return &shieldv1.CreateProjectResponse{Project: &shieldv1.Project{
		Id:        newProject.Id,
		Name:      newProject.Name,
		Slug:      newProject.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newProject.CreatedAt),
		UpdatedAt: timestamppb.New(newProject.UpdatedAt),
	}}, nil
}

func (v Dep) GetProject(ctx context.Context, request *shieldv1.GetProjectRequest) (*shieldv1.GetProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedProject, err := v.ProjectService.GetProject(ctx, request.GetId())
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

	return &shieldv1.GetProjectResponse{Project: &projectPB}, nil
}

func (v Dep) UpdateProject(ctx context.Context, request *shieldv1.UpdateProjectRequest) (*shieldv1.UpdateProjectResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedProject, err := v.ProjectService.UpdateProject(ctx, project.Project{
		Id:       request.GetId(),
		Name:     request.GetBody().Name,
		Slug:     request.GetBody().Slug,
		Metadata: metaDataMap,
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

	return &shieldv1.UpdateProjectResponse{Project: &projectPB}, nil
}

func transformProjectToPB(prj project.Project) (shieldv1.Project, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(prj.Metadata))
	if err != nil {
		return shieldv1.Project{}, err
	}

	return shieldv1.Project{
		Id:        prj.Id,
		Name:      prj.Name,
		Slug:      prj.Slug,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(prj.CreatedAt),
		UpdatedAt: timestamppb.New(prj.UpdatedAt),
	}, nil
}
