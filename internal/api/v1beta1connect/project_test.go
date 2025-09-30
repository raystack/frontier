package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testProjectID     = "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71"
	testProjectIDList = []string{"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71", "c7772c63-fca4-4c7c-bf93-c8f85115de4b"}
	testProjectMap    = map[string]project.Project{
		"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
			ID:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
			Name: "prj-1",
			Metadata: metadata.Metadata{
				"email": "org1@org1.com",
			},
			Organization: organization.Organization{
				ID: testOrgID,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
			ID:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
			Name: "prj-2",
			Metadata: metadata.Metadata{
				"email": "org1@org2.com",
			},
			Organization: organization.Organization{
				ID: testOrgID,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
)

func TestHandler_ListProjects(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *connect.Request[frontierv1beta1.ListProjectsRequest]
		want  *connect.Response[frontierv1beta1.ListProjectsResponse]
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req:   connect.NewRequest(&frontierv1beta1.ListProjectsRequest{}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{}).Return([]project.Project{}, errors.New("test error"))
			},
			want: nil,
			err:  connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return success if project return nil error",
			req:   connect.NewRequest(&frontierv1beta1.ListProjectsRequest{}),
			setup: func(ps *mocks.ProjectService) {
				var prjs []project.Project

				for _, projectID := range testProjectIDList {
					prjs = append(prjs, testProjectMap[projectID])
				}

				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{}).Return(prjs, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListProjectsResponse{Projects: []*frontierv1beta1.Project{
				{
					Id:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
					Name: "prj-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
						},
					},
					OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
					Name: "prj-2",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org2.com"),
						},
					},
					OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}}),
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.ListProjects(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_CreateProject(t *testing.T) {
	email := "user@raystack.org"
	table := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProjectService) context.Context
		req   *connect.Request[frontierv1beta1.CreateProjectRequest]
		want  *connect.Response[frontierv1beta1.CreateProjectResponse]
		err   error
	}{
		{
			title: "should return unauthenticated error if auth email in context is empty and project service return invalid user email",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, user.ErrInvalidEmail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}}),
			err: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			title: "should return internal error if project service return some error",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, errors.New("test error"))
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}}),
			err: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return bad request error if org id is not uuid",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, organization.ErrInvalidUUID)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}}),
			err: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			title: "should return already exist error if project service return error conflict",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, project.ErrConflict)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}}),
			err: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			title: "should return bad request error if name is empty",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, project.ErrInvalidDetail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}}),
			err: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			title: "should return success if project service return nil",
			req: connect.NewRequest(&frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: testProjectMap[testProjectID].Name,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("org1@org1.com"),
					},
				},
			}}),
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: testProjectMap[testProjectID].Name,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
					},
				}).Return(project.Project{
					ID:   testProjectMap[testProjectID].ID,
					Name: testProjectMap[testProjectID].Name,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
					},
				}, nil)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			want: connect.NewResponse(&frontierv1beta1.CreateProjectResponse{Project: &frontierv1beta1.Project{
				Id:   testProjectMap[testProjectID].ID,
				Name: testProjectMap[testProjectID].Name,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("org1@org1.com"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}}),
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
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.CreateProject(ctx, tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_GetProject(t *testing.T) {
	someProjectID := utils.NewString()
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *connect.Request[frontierv1beta1.GetProjectRequest]
		want  *connect.Response[frontierv1beta1.GetProjectResponse]
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req: connect.NewRequest(&frontierv1beta1.GetProjectRequest{
				Id: someProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), someProjectID).Return(project.Project{}, errors.New("test error"))
			},
			err: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return not found error if project doesnt exist",
			req: connect.NewRequest(&frontierv1beta1.GetProjectRequest{
				Id: someProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), someProjectID).Return(project.Project{}, project.ErrNotExist)
			},
			err: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			title: "should return project not found if project id is not uuid",
			req: connect.NewRequest(&frontierv1beta1.GetProjectRequest{
				Id: "some-id",
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-id").Return(project.Project{}, project.ErrInvalidUUID)
			},
			err: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			title: "should return project not found if project id is empty",
			req:   connect.NewRequest(&frontierv1beta1.GetProjectRequest{}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(project.Project{}, project.ErrInvalidUUID)
			},
			err: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			title: "should return success if project service return nil error",
			req: connect.NewRequest(&frontierv1beta1.GetProjectRequest{
				Id: someProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), someProjectID).Return(
					testProjectMap[testProjectID], nil)
			},
			want: connect.NewResponse(&frontierv1beta1.GetProjectResponse{Project: &frontierv1beta1.Project{
				Id:    testProjectMap[testProjectID].ID,
				Name:  testProjectMap[testProjectID].Name,
				OrgId: testProjectMap[testProjectID].Organization.ID,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("org1@org1.com"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}}),
			err: nil,
		},
	}
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.GetProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_UpdateProject(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ps *mocks.ProjectService)
		request *connect.Request[frontierv1beta1.UpdateProjectRequest]
		want    *connect.Response[frontierv1beta1.UpdateProjectResponse]
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &frontierv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return bad request error if org id is not uuid",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, organization.ErrInvalidUUID)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &frontierv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return not found error if project not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &frontierv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return conflict error if project service return err conflict",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &frontierv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return success if project service return nil error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(testProjectMap[testProjectID], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &frontierv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateProjectResponse{
				Project: &frontierv1beta1.Project{
					Id:    testProjectMap[testProjectID].ID,
					Name:  testProjectMap[testProjectID].Name,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.UpdateProject(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListProjectAdmins(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *connect.Request[frontierv1beta1.ListProjectAdminsRequest]
		want  *connect.Response[frontierv1beta1.ListProjectAdminsResponse]
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req: connect.NewRequest(&frontierv1beta1.ListProjectAdminsRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, project.AdminPermission).Return([]user.User{}, errors.New("test error"))
			},
			want: nil,
			err:  connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return success if project service return nil error",
			req: connect.NewRequest(&frontierv1beta1.ListProjectAdminsRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, project.AdminPermission).Return([]user.User{
					{
						ID:        "user-1",
						Name:      "User One",
						Email:     "user1@example.com",
						Metadata:  metadata.Metadata{},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListProjectAdminsResponse{
				Users: []*frontierv1beta1.User{
					{
						Id:        "user-1",
						Name:      "User One",
						Email:     "user1@example.com",
						Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			err: nil,
		},
	}
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.ListProjectAdmins(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_EnableProject(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *connect.Request[frontierv1beta1.EnableProjectRequest]
		want  *connect.Response[frontierv1beta1.EnableProjectResponse]
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req: connect.NewRequest(&frontierv1beta1.EnableProjectRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Enable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(errors.New("test error"))
			},
			want: nil,
			err:  connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return success if project service return nil error",
			req: connect.NewRequest(&frontierv1beta1.EnableProjectRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Enable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(nil)
			},
			want: connect.NewResponse(&frontierv1beta1.EnableProjectResponse{}),
			err:  nil,
		},
	}
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.EnableProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_DisableProject(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *connect.Request[frontierv1beta1.DisableProjectRequest]
		want  *connect.Response[frontierv1beta1.DisableProjectResponse]
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req: connect.NewRequest(&frontierv1beta1.DisableProjectRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Disable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(errors.New("test error"))
			},
			want: nil,
			err:  connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return success if project service return nil error",
			req: connect.NewRequest(&frontierv1beta1.DisableProjectRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Disable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(nil)
			},
			want: connect.NewResponse(&frontierv1beta1.DisableProjectResponse{}),
			err:  nil,
		},
	}
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.DisableProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_ListProjectUsers(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *connect.Request[frontierv1beta1.ListProjectUsersRequest]
		want  *connect.Response[frontierv1beta1.ListProjectUsersResponse]
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req: connect.NewRequest(&frontierv1beta1.ListProjectUsersRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, project.MemberPermission).Return([]user.User{}, errors.New("test error"))
			},
			want: nil,
			err:  connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return success if project service return nil error",
			req: connect.NewRequest(&frontierv1beta1.ListProjectUsersRequest{
				Id: testProjectID,
			}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, project.MemberPermission).Return([]user.User{
					{
						ID:        "user-1",
						Name:      "User One",
						Email:     "user1@example.com",
						Metadata:  metadata.Metadata{},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListProjectUsersResponse{
				Users: []*frontierv1beta1.User{
					{
						Id:        "user-1",
						Name:      "User One",
						Email:     "user1@example.com",
						Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
				RolePairs: []*frontierv1beta1.ListProjectUsersResponse_RolePair{},
			}),
			err: nil,
		},
	}
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.ListProjectUsers(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}
