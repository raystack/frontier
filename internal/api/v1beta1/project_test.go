package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/odpf/shield/pkg/metadata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/user"

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
		Metadata: metadata.Metadata{
			"email": "org1@org1.com",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
		ID:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
		Name: "Prj 2",
		Slug: "prj-2",
		Metadata: metadata.Metadata{
			"email": "org1@org2.com",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestCreateProject(t *testing.T) {
	email := "user@odpf.io"
	table := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProjectService) context.Context
		req   *shieldv1beta1.CreateProjectRequest
		want  *shieldv1beta1.CreateProjectResponse
		err   error
	}{
		{
			title: "should return forbidden error if auth email in context is empty and project service return invalid user email",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "odpf 1",
					Slug: "odpf-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, user.ErrInvalidEmail)
				return user.SetContextWithEmail(ctx, email)
			},
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcPermissionDenied,
		},
		{
			title: "should return internal error if project service return some error",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "odpf 1",
					Slug: "odpf-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcInternalServerError,
		},
		{
			title: "should return bad request error if org id is not uuid",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "odpf 1",
					Slug: "odpf-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, organization.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcBadBodyError,
		},
		{
			title: "should return bad request error if org id is not uuid",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "odpf 1",
					Slug: "odpf-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, organization.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcBadBodyError,
		},
		{
			title: "should return already exist error if group service return error conflict",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "odpf 1",
					Slug: "odpf-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, project.ErrConflict)
				return user.SetContextWithEmail(ctx, email)
			},
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: "odpf 1",
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcConflictError,
		},
		{
			title: "should return bad request error if name is empty",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Slug: "odpf-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, project.ErrInvalidDetail)
				return user.SetContextWithEmail(ctx, email)
			},
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Slug: "odpf-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcBadBodyError,
		},
		{
			title: "should return success if group service return nil",
			req: &shieldv1beta1.CreateProjectRequest{Body: &shieldv1beta1.ProjectRequestBody{
				Name: testProjectMap[testProjectID].Name,
				Slug: testProjectMap[testProjectID].Slug,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("org1@org1.com"),
					},
				},
			}},
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: testProjectMap[testProjectID].Name,
					Slug: testProjectMap[testProjectID].Slug,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
					},
				}).Return(project.Project{
					ID:   testProjectMap[testProjectID].ID,
					Name: testProjectMap[testProjectID].Name,
					Slug: testProjectMap[testProjectID].Slug,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
					},
				}, nil)
				return user.SetContextWithEmail(ctx, email)
			},
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
			mockProjectSrv := new(mocks.ProjectService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.CreateProject(ctx, tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestListProjects(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *shieldv1beta1.ListProjectsRequest
		want  *shieldv1beta1.ListProjectsResponse
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req:   &shieldv1beta1.ListProjectsRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return([]project.Project{}, errors.New("some error"))
			},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return success if project return nil error",
			req:   &shieldv1beta1.ListProjectsRequest{},
			setup: func(ps *mocks.ProjectService) {
				var prjs []project.Project

				for _, projectID := range testProjectIDList {
					prjs = append(prjs, testProjectMap[projectID])
				}

				ps.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return(prjs, nil)
			},
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
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.ListProjects(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestGetProject(t *testing.T) {
	t.Parallel()

	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *shieldv1beta1.GetProjectRequest
		want  *shieldv1beta1.GetProjectResponse
		err   error
	}{
		{
			title: "project doesnt exist",
			req:   &shieldv1beta1.GetProjectRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.Anything, mock.Anything).Return(project.Project{}, project.ErrNotExist)
			},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "uuid syntax error",
			req:   &shieldv1beta1.GetProjectRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.Anything, mock.Anything).Return(
					project.Project{}, project.ErrInvalidUUID)
			},
			err: grpcBadBodyError,
		},
		{
			title: "service error",
			req:   &shieldv1beta1.GetProjectRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.Anything, mock.Anything).Return(
					project.Project{}, errors.New("some error"))
			},
			err: grpcInternalServerError,
		},
		{
			title: "success",
			req:   &shieldv1beta1.GetProjectRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.Anything, mock.Anything).Return(
					testProjectMap[testProjectID], nil)
			},
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

			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.GetProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}
