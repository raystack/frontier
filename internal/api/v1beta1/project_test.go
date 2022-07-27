package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/core/user"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/shield/core/project"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var testProjectID = "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71"

var testProjectIDList = []string{"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71", "c7772c63-fca4-4c7c-bf93-c8f85115de4b"}

var testProjectMap = map[string]project.Project{
	"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
		ID:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
		Name: "Prj 1",
		Slug: "prj-1",
		Metadata: map[string]any{
			"email": "org1@org1.com",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
		ID:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
		Name: "Prj 2",
		Slug: "prj-2",
		Metadata: map[string]any{
			"email": "org1@org2.com",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestCreateProject(t *testing.T) {
	t.Parallel()

	table := []struct {
		title          string
		mockProjectSrv mockProject
		req            *shieldv1beta1.CreateProjectRequest
		want           *shieldv1beta1.CreateProjectResponse
		err            error
	}{
		{
			title: "error in metadata parsing",
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"count": structpb.NewNumberValue(10),
					},
				},
			}},
			err: grpcBadBodyError,
		},
		{
			title: "error in service",
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			mockProjectSrv: mockProject{CreateFunc: func(ctx context.Context, prj project.Project) (project.Project, error) {
				return project.Project{}, errors.New("some service error")
			}},
			err: grpcInternalServerError,
		},
		{
			title: "success",
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			mockProjectSrv: mockProject{CreateFunc: func(ctx context.Context, prj project.Project) (project.Project, error) {
				return testProjectMap[testProjectID], nil
			}},
			want: &shieldv1beta1.CreateProjectResponse{Project: &shieldv1beta1.Project{
				Id:   testProjectMap[testProjectID].ID,
				Name: testProjectMap[testProjectID].Name,
				Slug: testProjectMap[testProjectID].Slug,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("org1@org1.com"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Handler{projectService: tt.mockProjectSrv}
			resp, err := mockDep.CreateProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestListProjects(t *testing.T) {
	t.Parallel()

	table := []struct {
		title          string
		mockProjectSrv mockProject
		req            *shieldv1beta1.ListProjectsRequest
		want           *shieldv1beta1.ListProjectsResponse
		err            error
	}{
		{
			title: "error in service",
			req:   &shieldv1beta1.ListProjectsRequest{},
			mockProjectSrv: mockProject{ListFunc: func(ctx context.Context) ([]project.Project, error) {
				return []project.Project{}, errors.New("some store error")
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "success",
			req:   &shieldv1beta1.ListProjectsRequest{},
			mockProjectSrv: mockProject{ListFunc: func(ctx context.Context) ([]project.Project, error) {
				var prjs []project.Project

				for _, projectID := range testProjectIDList {
					prjs = append(prjs, testProjectMap[projectID])
				}

				return prjs, nil
			}},
			want: &shieldv1beta1.ListProjectsResponse{Projects: []*shieldv1beta1.Project{
				{
					Id:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
					Name: "Prj 1",
					Slug: "prj-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
					Name: "Prj 2",
					Slug: "prj-2",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org2.com"),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Handler{projectService: tt.mockProjectSrv}
			resp, err := mockDep.ListProjects(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestGetProject(t *testing.T) {
	t.Parallel()

	table := []struct {
		title          string
		mockProjectSrv mockProject
		req            *shieldv1beta1.GetProjectRequest
		want           *shieldv1beta1.GetProjectResponse
		err            error
	}{
		{
			title: "project doesnt exist",
			req:   &shieldv1beta1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetByIDFunc: func(ctx context.Context, id string) (project.Project, error) {
				return project.Project{}, project.ErrNotExist
			}},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "uuid syntax error",
			req:   &shieldv1beta1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetByIDFunc: func(ctx context.Context, id string) (project.Project, error) {
				return project.Project{}, project.ErrInvalidUUID
			}},
			err: grpcBadBodyError,
		},
		{
			title: "service error",
			req:   &shieldv1beta1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetByIDFunc: func(ctx context.Context, id string) (project.Project, error) {
				return project.Project{}, errors.New("some error")
			}},
			err: grpcInternalServerError,
		},
		{
			title: "success",
			req:   &shieldv1beta1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetByIDFunc: func(ctx context.Context, id string) (project.Project, error) {
				return testProjectMap[testProjectID], nil
			}},
			want: &shieldv1beta1.GetProjectResponse{Project: &shieldv1beta1.Project{
				Id:   testProjectMap[testProjectID].ID,
				Name: testProjectMap[testProjectID].Name,
				Slug: testProjectMap[testProjectID].Slug,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("org1@org1.com"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Handler{projectService: tt.mockProjectSrv}
			resp, err := mockDep.GetProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockProject struct {
	GetByIDFunc           func(ctx context.Context, id string) (project.Project, error)
	GetBySlugFunc         func(ctx context.Context, slug string) (project.Project, error)
	CreateFunc            func(ctx context.Context, prj project.Project) (project.Project, error)
	ListFunc              func(ctx context.Context) ([]project.Project, error)
	UpdateByIDFunc        func(ctx context.Context, toUpdate project.Project) (project.Project, error)
	UpdateBySlugFunc      func(ctx context.Context, toUpdate project.Project) (project.Project, error)
	AddAdminByIDFunc      func(ctx context.Context, id string, userIds []string) ([]user.User, error)
	AddAdminBySlugFunc    func(ctx context.Context, slug string, userIds []string) ([]user.User, error)
	RemoveAdminByIDFunc   func(ctx context.Context, id string, userId string) ([]user.User, error)
	RemoveAdminBySlugFunc func(ctx context.Context, slug string, userId string) ([]user.User, error)
	ListAdminsFunc        func(ctx context.Context, id string) ([]user.User, error)
}

func (m mockProject) List(ctx context.Context) ([]project.Project, error) {
	return m.ListFunc(ctx)
}

func (m mockProject) Create(ctx context.Context, project project.Project) (project.Project, error) {
	return m.CreateFunc(ctx, project)
}

func (m mockProject) GetByID(ctx context.Context, id string) (project.Project, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m mockProject) GetBySlug(ctx context.Context, slug string) (project.Project, error) {
	return m.GetBySlugFunc(ctx, slug)
}

func (m mockProject) UpdateByID(ctx context.Context, toUpdate project.Project) (project.Project, error) {
	return m.UpdateByIDFunc(ctx, toUpdate)
}

func (m mockProject) UpdateBySlug(ctx context.Context, toUpdate project.Project) (project.Project, error) {
	return m.UpdateBySlugFunc(ctx, toUpdate)
}

func (m mockProject) AddAdminByID(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	return m.AddAdminBySlugFunc(ctx, id, userIds)
}

func (m mockProject) AddAdminBySlug(ctx context.Context, slug string, userIds []string) ([]user.User, error) {
	return m.AddAdminBySlugFunc(ctx, slug, userIds)
}
func (m mockProject) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return m.ListAdminsFunc(ctx, id)
}

func (m mockProject) RemoveAdminByID(ctx context.Context, id string, userId string) ([]user.User, error) {
	return m.RemoveAdminBySlugFunc(ctx, id, userId)
}
func (m mockProject) RemoveAdminBySlug(ctx context.Context, slug string, userId string) ([]user.User, error) {
	return m.RemoveAdminBySlugFunc(ctx, slug, userId)
}
