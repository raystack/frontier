package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/odpf/shield/pkg/errors"
	"github.com/odpf/shield/pkg/metadata"
	"github.com/odpf/shield/pkg/uuid"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
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
			Name: "Prj 1",
			Slug: "prj-1",
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
			Name: "Prj 2",
			Slug: "prj-2",
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
			title: "should return already exist error if project service return error conflict",
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
			title: "should return success if project service return nil",
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
					OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
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
	someProjectID := uuid.NewString()

	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *shieldv1beta1.GetProjectRequest
		want  *shieldv1beta1.GetProjectResponse
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req: &shieldv1beta1.GetProjectRequest{
				Id: someProjectID,
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), someProjectID).Return(project.Project{}, errors.New("some error"))
			},
			err: grpcInternalServerError,
		},
		{
			title: "should return not found error if project doesnt exist",
			req: &shieldv1beta1.GetProjectRequest{
				Id: someProjectID,
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), someProjectID).Return(project.Project{}, project.ErrNotExist)
			},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "should return project not found if project id is not uuid",
			req: &shieldv1beta1.GetProjectRequest{
				Id: "some-id",
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(project.Project{}, project.ErrInvalidUUID)
			},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "should return project not found if project id is empty",
			req:   &shieldv1beta1.GetProjectRequest{},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(project.Project{}, project.ErrInvalidUUID)
			},
			err: grpcProjectNotFoundErr,
		},
		{
			title: "should return success if project service return nil error",
			req: &shieldv1beta1.GetProjectRequest{
				Id: someProjectID,
			},
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), someProjectID).Return(
					testProjectMap[testProjectID], nil)
			},
			want: &shieldv1beta1.GetProjectResponse{Project: &shieldv1beta1.Project{
				Id:    testProjectMap[testProjectID].ID,
				Name:  testProjectMap[testProjectID].Name,
				Slug:  testProjectMap[testProjectID].Slug,
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
		request *shieldv1beta1.UpdateProjectRequest
		want    *shieldv1beta1.UpdateProjectResponse
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testProjectMap[testProjectID]).Return(project.Project{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
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
			name: "should return not found error if org id is not uuid",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testProjectMap[testProjectID]).Return(project.Project{}, organization.ErrInvalidUUID)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
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
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
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
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
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
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testProjectMap[testProjectID]).Return(project.Project{}, project.ErrConflict)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
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
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), project.Project{
					ID:           testProjectID,
					Slug:         testProjectMap[testProjectID].Slug,
					Organization: testProjectMap[testProjectID].Organization,
					Metadata:     testProjectMap[testProjectID].Metadata,
				}).Return(project.Project{}, project.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
					Slug:  testProjectMap[testProjectID].Slug,
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
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), project.Project{
					ID:           testProjectID,
					Name:         testProjectMap[testProjectID].Name,
					Organization: testProjectMap[testProjectID].Organization,
					Metadata:     testProjectMap[testProjectID].Metadata,
				}).Return(project.Project{}, project.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
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
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), project.Project{
					Name:         testProjectMap[testProjectID].Name,
					Slug:         "", // consider it to update by slug and assigned empty to slug
					Organization: testProjectMap[testProjectID].Organization,
					Metadata:     testProjectMap[testProjectID].Metadata,
				}).Return(project.Project{}, project.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Body: &shieldv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
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
				ps.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testProjectMap[testProjectID]).Return(testProjectMap[testProjectID], nil)
			},
			request: &shieldv1beta1.UpdateProjectRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.ProjectRequestBody{
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
					OrgId: testProjectMap[testProjectID].Organization.ID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue(testProjectMap[testProjectID].Metadata["email"].(string)),
						},
					},
				},
			},
			want: &shieldv1beta1.UpdateProjectResponse{
				Project: &shieldv1beta1.Project{
					Id:    testProjectMap[testProjectID].ID,
					Name:  testProjectMap[testProjectID].Name,
					Slug:  testProjectMap[testProjectID].Slug,
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

func TestHandler_AddProjectAdmin(t *testing.T) {
	var (
		email       = "user@odpf.io"
		someUserIDs = []string{
			uuid.NewString(),
			uuid.NewString(),
			uuid.NewString(),
		}
		testUserList []user.User
		testUserIDs  []string
	)
	for _, u := range testUserMap {
		testUserList = append(testUserList, u)
		testUserIDs = append(testUserIDs, u.ID)
	}
	tests := []struct {
		name    string
		setup   func(ctx context.Context, ps *mocks.ProjectService) context.Context
		request *shieldv1beta1.AddProjectAdminRequest
		want    *shieldv1beta1.AddProjectAdminResponse
		wantErr error
	}{
		{
			name: "should return forbidden error if auth email in context is empty and project service return invalid user email",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().AddAdmins(mock.AnythingOfType("*context.emptyCtx"), testProjectID, someUserIDs).Return([]user.User{}, user.ErrInvalidEmail)
				return ctx
			},
			request: &shieldv1beta1.AddProjectAdminRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.AddProjectAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return forbidden error if caller is unauthorized",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserIDs).Return([]user.User{}, errors.Unauthorized)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddProjectAdminRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.AddProjectAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return internal error if project service return some error",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserIDs).Return([]user.User{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddProjectAdminRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.AddProjectAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if project not exist",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserIDs).Return([]user.User{}, project.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddProjectAdminRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.AddProjectAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return bad request error if one of user id is not uuid",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserIDs).Return([]user.User{}, user.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddProjectAdminRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.AddProjectAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if one of user id is empty",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserIDs).Return([]user.User{}, user.ErrInvalidID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddProjectAdminRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.AddProjectAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if project service return nil error",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), testProjectID, testUserIDs).Return(testUserList, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddProjectAdminRequest{
				Id: testProjectID,
				Body: &shieldv1beta1.AddProjectAdminRequestBody{
					UserIds: testUserIDs,
				},
			},
			want: &shieldv1beta1.AddProjectAdminResponse{
				Users: []*shieldv1beta1.User{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Name:  "User 1",
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
			ctx := context.Background()
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.AddProjectAdmin(ctx, tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListProjectAdmins(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ps *mocks.ProjectService)
		request *shieldv1beta1.ListProjectAdminsRequest
		want    *shieldv1beta1.ListProjectAdminsResponse
		wantErr error
	}{
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), testProjectID).Return([]user.User{}, errors.New("some error"))
			},
			request: &shieldv1beta1.ListProjectAdminsRequest{
				Id: testProjectID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if org id is not exist",
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), testProjectID).Return([]user.User{}, project.ErrNotExist)
			},
			request: &shieldv1beta1.ListProjectAdminsRequest{
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
				ps.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), testProjectID).Return(testUserList, nil)
			},
			request: &shieldv1beta1.ListProjectAdminsRequest{
				Id: testProjectID,
			},
			want: &shieldv1beta1.ListProjectAdminsResponse{
				Users: []*shieldv1beta1.User{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Name:  "User 1",
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

func TestHandler_RemoveProjectAdmin(t *testing.T) {
	var (
		email      = "user@odpf.io"
		someUserID = uuid.NewString()
	)
	tests := []struct {
		name    string
		setup   func(ctx context.Context, ps *mocks.ProjectService) context.Context
		request *shieldv1beta1.RemoveProjectAdminRequest
		want    *shieldv1beta1.RemoveProjectAdminResponse
		wantErr error
	}{
		{
			name: "should return forbidden error if auth email in context is empty and project service return invalid user email",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.emptyCtx"), testProjectID, someUserID).Return([]user.User{}, user.ErrInvalidEmail)
				return ctx
			},
			request: &shieldv1beta1.RemoveProjectAdminRequest{
				Id:     testProjectID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return forbidden error if caller is not an admin",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserID).Return([]user.User{}, errors.Unauthorized)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveProjectAdminRequest{
				Id:     testProjectID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return internal error if project service return some error",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserID).Return([]user.User{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveProjectAdminRequest{
				Id:     testProjectID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if project id not exist",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserID).Return([]user.User{}, project.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveProjectAdminRequest{
				Id:     testProjectID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcProjectNotFoundErr,
		},
		{
			name: "should return not found user error if user id is not uuid",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				ps.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), testProjectID, "test").Return([]user.User{}, user.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveProjectAdminRequest{
				Id:     testProjectID,
				UserId: "test",
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return success if project service return nil error",
			setup: func(ctx context.Context, ps *mocks.ProjectService) context.Context {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				ps.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), testProjectID, someUserID).Return(testUserList, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveProjectAdminRequest{
				Id:     testProjectID,
				UserId: someUserID,
			},
			want: &shieldv1beta1.RemoveProjectAdminResponse{
				Message: "Removed Admin from project",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockProjectSrv)
			}
			mockDep := Handler{projectService: mockProjectSrv}
			resp, err := mockDep.RemoveProjectAdmin(ctx, tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
