package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

func (v Dep) ListProjects(ctx context.Context, request *shieldv1.ListProjectsRequest) (*shieldv1.ListProjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) CreateProject(ctx context.Context, request *shieldv1.CreateProjectRequest) (*shieldv1.CreateProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) GetProject(ctx context.Context, request *shieldv1.GetProjectRequest) (*shieldv1.GetProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) UpdateProject(ctx context.Context, request *shieldv1.UpdateProjectRequest) (*shieldv1.UpdateProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
