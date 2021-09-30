package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) GetAllProjects(ctx context.Context, request *shieldv1.GetAllProjectsRequest) (*shieldv1.GetAllProjectsResponse, error) {
	panic("implement me")
}

func (v Dep) CreateProject(ctx context.Context, request *shieldv1.CreateProjectRequest) (*shieldv1.ProjectResponse, error) {
	panic("implement me")
}

func (v Dep) GetProjectByID(ctx context.Context, request *shieldv1.GetProjectRequest) (*shieldv1.ProjectResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateProjectByID(ctx context.Context, request *shieldv1.UpdateProjectRequest) (*shieldv1.ProjectResponse, error) {
	panic("implement me")
}
