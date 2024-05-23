package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
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

func TestCreateProject(t *testing.T) {
	email := "user@raystack.org"
	table := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProjectService) context.Context
		req   *frontierv1beta1.CreateProjectRequest
		want  *frontierv1beta1.CreateProjectResponse
		err   error
	}{
		{
			title: "should return forbidden error if auth email in context is empty and project service return invalid user email",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, user.ErrInvalidEmail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcUnauthenticated,
		},
		{
			title: "should return internal error if project service return some error",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), project.Project{
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, errors.New("some error"))
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
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
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, organization.ErrInvalidUUID)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
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
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, organization.ErrInvalidUUID)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcBadBodyError,
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
			req: &frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
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
					Name: "raystack-1",
					Metadata: metadata.Metadata{
						"team": "Platforms",
					},
				}).Return(project.Project{}, project.ErrInvalidDetail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: "raystack-1",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"team": structpb.NewStringValue("Platforms"),
					},
				},
			}},
			err: grpcBadBodyError,
		},
		{
			title: "should return success if project service return nil",
			req: &frontierv1beta1.CreateProjectRequest{Body: &frontierv1beta1.ProjectRequestBody{
				Name: testProjectMap[testProjectID].Name,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("org1@org1.com"),
					},
				},
			}},
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
			want: &frontierv1beta1.CreateProjectResponse{Project: &frontierv1beta1.Project{
				Id:   testProjectMap[testProjectID].ID,
				Name: testProjectMap[testProjectID].Name,
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
		req   *frontierv1beta1.ListProjectsRequest
		want  *frontierv1beta1.ListProjectsResponse
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req:   &frontierv1beta1.ListProjectsRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{}).Return([]project.Project{}, errors.New("some error"))
			},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return success if project return nil error",
			req:   &frontierv1beta1.ListProjectsRequest{},
			setup: func(ps *mocks.ProjectService) {
				var prjs []project.Project

				for _, projectID := range testProjectIDList {
					prjs = append(prjs, testProjectMap[projectID])
				}

				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{}).Return(prjs, nil)
			},
			want: &frontierv1beta1.ListProjectsResponse{Projects: []*frontierv1beta1.Project{
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
	someProjectID := utils.NewString()

	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *frontierv1beta1.GetProjectRequest
		want  *frontierv1beta1.GetProjectResponse
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req: &frontierv1beta1.GetProjectRequest{
				Id: someProjectID,
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), someProjectID).Return(project.Project{}, errors.New("some error"))
			},
			err: grpcInternalServerError,
		},
		{
			title: "should return not found error if project doesnt exist",
			req: &frontierv1beta1.GetProjectRequest{
				Id: someProjectID,
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), someProjectID).Return(project.Project{}, project.ErrNotExist)
			},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "should return project not found if project id is not uuid",
			req: &frontierv1beta1.GetProjectRequest{
				Id: "some-id",
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-id").Return(project.Project{}, project.ErrInvalidUUID)
			},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "should return project not found if project id is empty",
			req:   &frontierv1beta1.GetProjectRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(project.Project{}, project.ErrInvalidUUID)
			},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "should return success if project service return nil error",
			req: &frontierv1beta1.GetProjectRequest{
				Id: someProjectID,
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), someProjectID).Return(
					testProjectMap[testProjectID], nil)
			},
			want: &frontierv1beta1.GetProjectResponse{Project: &frontierv1beta1.Project{
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
		request *frontierv1beta1.UpdateProjectRequest
		want    *frontierv1beta1.UpdateProjectResponse
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, errors.New("some error"))
			},
			request: &frontierv1beta1.UpdateProjectRequest{
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
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if org id is not uuid",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, organization.ErrInvalidUUID)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
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
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return not found error if project not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrNotExist)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
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
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return not found error if project not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrNotExist)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
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
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return already exist error if project service return err conflict",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrConflict)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
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
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return bad request error if update by id with empty name",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), project.Project{
					ID:           testProjectID,
					Organization: testProjectMap[testProjectID].Organization,
					Metadata:     testProjectMap[testProjectID].Metadata,
				}).Return(project.Project{}, project.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &frontierv1beta1.ProjectRequestBody{
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if update by id with empty slug",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), project.Project{
					ID:           testProjectID,
					Name:         testProjectMap[testProjectID].Name,
					Organization: testProjectMap[testProjectID].Organization,
					Metadata:     testProjectMap[testProjectID].Metadata,
				}).Return(project.Project{}, project.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
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
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return not found error if project id empty",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), project.Project{
					Name:         testProjectMap[testProjectID].Name,
					Organization: testProjectMap[testProjectID].Organization,
					Metadata:     testProjectMap[testProjectID].Metadata,
				}).Return(project.Project{}, project.ErrInvalidID)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
				Body: &frontierv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return success if project service return nil",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), testProjectMap[testProjectID]).Return(testProjectMap[testProjectID], nil)
			},
			request: &frontierv1beta1.UpdateProjectRequest{
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
			},
			want: &frontierv1beta1.UpdateProjectResponse{
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
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.UpdateProject(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListProjectAdmins(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ps *mocks.ProjectService)
		request *frontierv1beta1.ListProjectAdminsRequest
		want    *frontierv1beta1.ListProjectAdminsResponse
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, schema.DeletePermission).Return([]user.User{}, errors.New("some error"))
			},
			request: &frontierv1beta1.ListProjectAdminsRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if org id is not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, schema.DeletePermission).Return([]user.User{}, project.ErrNotExist)
			},
			request: &frontierv1beta1.ListProjectAdminsRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return success if org service return nil error",
			setup: func(ps *mocks.ProjectService) {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, schema.DeletePermission).Return(testUserList, nil)
			},
			request: &frontierv1beta1.ListProjectAdminsRequest{
				Id: testProjectID,
			},
			want: &frontierv1beta1.ListProjectAdminsResponse{
				Users: []*frontierv1beta1.User{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "User 1",
						Name:  "user1",
						Email: "test@test.com",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"foo":    structpb.NewStringValue("bar"),
								"age":    structpb.NewNumberValue(21),
								"intern": structpb.NewBoolValue(true),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.ListProjectAdmins(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_EnableProject(t *testing.T) {
	tests := []struct {
		name    string
		req     *frontierv1beta1.EnableProjectRequest
		setup   func(ps *mocks.ProjectService)
		want    *frontierv1beta1.EnableProjectResponse
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Enable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(errors.New("some error"))
			},
			req: &frontierv1beta1.EnableProjectRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if project id is not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Enable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(project.ErrNotExist)
			},
			req: &frontierv1beta1.EnableProjectRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return no error if project enabled successfully",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Enable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(nil)
			},
			req: &frontierv1beta1.EnableProjectRequest{
				Id: testProjectID,
			},
			want:    &frontierv1beta1.EnableProjectResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.EnableProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DisableProject(t *testing.T) {
	tests := []struct {
		name    string
		req     *frontierv1beta1.DisableProjectRequest
		setup   func(ps *mocks.ProjectService)
		want    *frontierv1beta1.DisableProjectResponse
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Disable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(errors.New("some error"))
			},
			req: &frontierv1beta1.DisableProjectRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if project id is not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Disable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(project.ErrNotExist)
			},
			req: &frontierv1beta1.DisableProjectRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return no error if project disabled successfully",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Disable(mock.AnythingOfType("context.backgroundCtx"), testProjectID).Return(nil)
			},
			req: &frontierv1beta1.DisableProjectRequest{
				Id: testProjectID,
			},
			want:    &frontierv1beta1.DisableProjectResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.DisableProject(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListProjectUsers(t *testing.T) {
	tests := []struct {
		name    string
		request *frontierv1beta1.ListProjectUsersRequest
		setup   func(ps *mocks.ProjectService)
		want    *frontierv1beta1.ListProjectUsersResponse
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, project.MemberPermission).Return(nil, errors.New("some error"))
			},
			request: &frontierv1beta1.ListProjectUsersRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if project id is not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, "get").Return(nil, project.ErrNotExist)
			},
			request: &frontierv1beta1.ListProjectUsersRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return project users list and no error on success",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListUsers(mock.AnythingOfType("context.backgroundCtx"), testProjectID, project.MemberPermission).Return([]user.User{{
					ID:        "user1",
					Name:      "user1",
					Title:     "user1",
					Email:     "user1@raystack.org",
					Metadata:  metadata.Metadata{},
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, {
					ID:        "user2",
					Name:      "user2",
					Title:     "user2",
					Email:     "user2@raystack.org",
					Metadata:  metadata.Metadata{},
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}}, nil)
			},
			request: &frontierv1beta1.ListProjectUsersRequest{
				Id: testProjectID,
			},
			want: &frontierv1beta1.ListProjectUsersResponse{
				Users: []*frontierv1beta1.User{
					{
						Id:    "user1",
						Title: "user1",
						Name:  "user1",
						Email: "user1@raystack.org",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					{
						Id:    "user2",
						Title: "user2",
						Name:  "user2",
						Email: "user2@raystack.org",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.ListProjectUsers(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
