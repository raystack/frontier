package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/shield/internal/project"
	modelv1 "github.com/odpf/shield/model/v1"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testProjectID = "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71"

var testProjectMap = map[string]modelv1.Project{
	"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
		Id:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
		Name: "Prj 1",
		Slug: "prj-1",
		Metadata: map[string]string{
			"email": "org1@org1.com",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
		Id:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
		Name: "Prj 2",
		Slug: "prj-2",
		Metadata: map[string]string{
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
		req            *shieldv1.CreateProjectRequest
		want           *shieldv1.CreateProjectResponse
		err            error
	}{
		{
			title: "error in metadata parsing",
			req: &shieldv1.CreateProjectRequest{Body: &shieldv1.ProjectRequestBody{
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
			req: &shieldv1.CreateProjectRequest{Body: &shieldv1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			mockProjectSrv: mockProject{CreateProjectFunc: func(ctx context.Context, prj modelv1.Project) (modelv1.Project, error) {
				return modelv1.Project{}, errors.New("some service error")
			}},
			err: grpcInternalServerError,
		},
		{
			title: "success",
			req: &shieldv1.CreateProjectRequest{Body: &shieldv1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			mockProjectSrv: mockProject{CreateProjectFunc: func(ctx context.Context, prj modelv1.Project) (modelv1.Project, error) {
				return testProjectMap[testProjectID], nil
			}},
			want: &shieldv1.CreateProjectResponse{Project: &shieldv1.Project{
				Id:   testProjectMap[testProjectID].Id,
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

			mockDep := Dep{ProjectService: tt.mockProjectSrv}
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
		req            *shieldv1.ListProjectsRequest
		want           *shieldv1.ListProjectsResponse
		err            error
	}{
		{
			title: "error in service",
			req:   &shieldv1.ListProjectsRequest{},
			mockProjectSrv: mockProject{ListProjectFunc: func(ctx context.Context) ([]modelv1.Project, error) {
				return []modelv1.Project{}, errors.New("some store error")
			}},
			err: grpcInternalServerError,
		},
		{
			title: "success",
			req:   &shieldv1.ListProjectsRequest{},
			mockProjectSrv: mockProject{ListProjectFunc: func(ctx context.Context) ([]modelv1.Project, error) {
				var prjs []modelv1.Project

				for _, v := range testProjectMap {
					prjs = append(prjs, v)
				}

				return prjs, nil
			}},
			want: &shieldv1.ListProjectsResponse{Projects: []*shieldv1.Project{
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

			mockDep := Dep{ProjectService: tt.mockProjectSrv}
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
		req            *shieldv1.GetProjectRequest
		want           *shieldv1.GetProjectResponse
		err            error
	}{
		{
			title: "project doesnt exist",
			req:   &shieldv1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetProjectFunc: func(ctx context.Context, id string) (modelv1.Project, error) {
				return modelv1.Project{}, project.ProjectDoesntExist
			}},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "uuid syntax error",
			req:   &shieldv1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetProjectFunc: func(ctx context.Context, id string) (modelv1.Project, error) {
				return modelv1.Project{}, project.InvalidUUID
			}},
			err: grpcBadBodyError,
		},
		{
			title: "service error",
			req:   &shieldv1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetProjectFunc: func(ctx context.Context, id string) (modelv1.Project, error) {
				return modelv1.Project{}, errors.New("some error")
			}},
			err: grpcInternalServerError,
		},
		{
			title: "success",
			req:   &shieldv1.GetProjectRequest{},
			mockProjectSrv: mockProject{GetProjectFunc: func(ctx context.Context, id string) (modelv1.Project, error) {
				return testProjectMap[testProjectID], nil
			}},
			want: &shieldv1.GetProjectResponse{Project: &shieldv1.Project{
				Id:   testProjectMap[testProjectID].Id,
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

			mockDep := Dep{ProjectService: tt.mockProjectSrv}
			resp, err := mockDep.GetProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockProject struct {
	GetProjectFunc    func(ctx context.Context, id string) (modelv1.Project, error)
	CreateProjectFunc func(ctx context.Context, project modelv1.Project) (modelv1.Project, error)
	ListProjectFunc   func(ctx context.Context) ([]modelv1.Project, error)
	UpdateProjectFunc func(ctx context.Context, toUpdate modelv1.Project) (modelv1.Project, error)
}

func (m mockProject) List(ctx context.Context) ([]modelv1.Project, error) {
	return m.ListProjectFunc(ctx)
}

func (m mockProject) Create(ctx context.Context, project modelv1.Project) (modelv1.Project, error) {
	return m.CreateProjectFunc(ctx, project)
}

func (m mockProject) Get(ctx context.Context, id string) (modelv1.Project, error) {
	return m.GetProjectFunc(ctx, id)
}

func (m mockProject) Update(ctx context.Context, toUpdate modelv1.Project) (modelv1.Project, error) {
	return m.UpdateProjectFunc(ctx, toUpdate)
}
