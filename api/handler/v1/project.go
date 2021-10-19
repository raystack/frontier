package v1

import (
	"context"
	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

func (v Dep) ListProjects(ctx context.Context, request *shieldv1.ListProjectsRequest) (*shieldv1.ListProjectsResponse, error) {
	panic("implement me")
}

func (v Dep) CreateProject(ctx context.Context, request *shieldv1.CreateProjectRequest) (*shieldv1.CreateProjectResponse, error) {
	panic("implement me")
}

func (v Dep) GetProject(ctx context.Context, request *shieldv1.GetProjectRequest) (*shieldv1.GetProjectResponse, error) {
	panic("implement me")
}

func (v Dep) UpdateProject(ctx context.Context, request *shieldv1.UpdateProjectRequest) (*shieldv1.UpdateProjectResponse, error) {
	panic("implement me")
}
