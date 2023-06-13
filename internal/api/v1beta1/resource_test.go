package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/shield/pkg/utils"

	"github.com/raystack/shield/core/project"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/core/resource"
	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/internal/api/v1beta1/mocks"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testResourceID = utils.NewString()
	testResource   = resource.Resource{
		ID:          testResourceID,
		URN:         "res-urn",
		Name:        "a resource name",
		ProjectID:   testProjectID,
		NamespaceID: testNSID,
		UserID:      testUserID,
	}
	testResourcePB = &shieldv1beta1.Resource{
		Id:        testResource.ID,
		Name:      testResource.Name,
		Urn:       testResource.URN,
		ProjectId: testProjectID,
		Namespace: testNSID,
		UserId:    testUserID,
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
)

func TestHandler_ListResources(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *shieldv1beta1.ListResourcesRequest
		want    *shieldv1beta1.ListResourcesResponse
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), resource.Filter{}).Return([]resource.Resource{}, errors.New("some error"))
			},
			request: &shieldv1beta1.ListResourcesRequest{},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return resources if resource service return nil error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), resource.Filter{}).Return([]resource.Resource{
					testResource,
				}, nil)
			},
			request: &shieldv1beta1.ListResourcesRequest{},
			want: &shieldv1beta1.ListResourcesResponse{
				Resources: []*shieldv1beta1.Resource{
					testResourcePB,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			mockDep := Handler{resourceService: mockResourceSrv}
			resp, err := mockDep.ListResources(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateProjectResource(t *testing.T) {
	email := "user@raystack.org"
	tests := []struct {
		name    string
		setup   func(ctx context.Context, rs *mocks.ResourceService, ps *mocks.ProjectService, rls *mocks.RelationService) context.Context
		request *shieldv1beta1.CreateProjectResourceRequest
		want    *shieldv1beta1.CreateProjectResourceResponse
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(ctx context.Context, rs *mocks.ResourceService, ps *mocks.ProjectService, rls *mocks.RelationService) context.Context {
				rs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), resource.Resource{
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					NamespaceID: testResource.NamespaceID,
					UserID:      testUserID,
				}).Return(resource.Resource{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				}},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(ctx context.Context, rs *mocks.ResourceService, ps *mocks.ProjectService, rls *mocks.RelationService) context.Context {
				rs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), resource.Resource{
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					NamespaceID: testResource.NamespaceID,
					UserID:      testUserID,
				}).Return(resource.Resource{}, resource.ErrInvalidDetail)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if resource service return nil",
			setup: func(ctx context.Context, rs *mocks.ResourceService, ps *mocks.ProjectService, rls *mocks.RelationService) context.Context {
				rls.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), relation.Relation{
					Object: relation.Object{
						ID:        testResource.ID,
						Namespace: testResource.NamespaceID,
					},
					Subject: relation.Subject{
						SubRelationName: "owner",
						Namespace:       "user",
						ID:              testUserID,
					},
				}).Return(relation.Relation{}, nil)

				rs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), resource.Resource{
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					NamespaceID: testResource.NamespaceID,
					UserID:      testUserID,
				}).Return(testResource, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want: &shieldv1beta1.CreateProjectResourceResponse{
				Resource: testResourcePB,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			mockProjectSrv := new(mocks.ProjectService)
			mockRelationSrv := new(mocks.RelationService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockResourceSrv, mockProjectSrv, mockRelationSrv)
			}
			mockDep := Handler{resourceService: mockResourceSrv, projectService: mockProjectSrv, relationService: mockRelationSrv}
			resp, err := mockDep.CreateProjectResource(ctx, tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetProjectResource(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *shieldv1beta1.GetProjectResourceRequest
		want    *shieldv1beta1.GetProjectResourceResponse
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ID).Return(resource.Resource{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(resource.Resource{}, resource.ErrInvalidID)
			},
			request: &shieldv1beta1.GetProjectResourceRequest{},
			want:    nil,
			wantErr: grpcResourceNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(resource.Resource{}, resource.ErrInvalidUUID)
			},
			request: &shieldv1beta1.GetProjectResourceRequest{
				Id: "some-id",
			},
			want:    nil,
			wantErr: grpcResourceNotFoundErr,
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ID).Return(resource.Resource{}, resource.ErrNotExist)
			},
			request: &shieldv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			},
			want:    nil,
			wantErr: grpcResourceNotFoundErr,
		},
		{
			name: "should return success if resource service return nil error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ID).Return(testResource, nil)
			},
			request: &shieldv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			},
			want: &shieldv1beta1.GetProjectResourceResponse{
				Resource: testResourcePB,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			mockDep := Handler{resourceService: mockResourceSrv}
			resp, err := mockDep.GetProjectResource(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateProjectResource(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService, ps *mocks.ProjectService)
		request *shieldv1beta1.UpdateProjectResourceRequest
		want    *shieldv1beta1.UpdateProjectResourceResponse
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:          testResourceID,
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					UserID:      testResource.UserID,
					NamespaceID: testResource.NamespaceID,
				}).Return(resource.Resource{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:          "",
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					NamespaceID: testResource.NamespaceID,
					UserID:      testResource.UserID,
				}).Return(resource.Resource{}, resource.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want:    nil,
			wantErr: grpcResourceNotFoundErr,
		},
		{
			name: "should return not found error if id is not exist",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:          testResourceID,
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					UserID:      testResource.UserID,
					NamespaceID: testResource.NamespaceID,
				}).Return(resource.Resource{}, resource.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want:    nil,
			wantErr: grpcResourceNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:          "some-id",
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					UserID:      testResource.UserID,
					NamespaceID: testResource.NamespaceID,
				}).Return(resource.Resource{}, resource.ErrInvalidUUID)
			},
			request: &shieldv1beta1.UpdateProjectResourceRequest{
				Id:        "some-id",
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want:    nil,
			wantErr: grpcResourceNotFoundErr,
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:          testResourceID,
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					UserID:      testResource.UserID,
					NamespaceID: testResource.NamespaceID,
				}).Return(resource.Resource{}, resource.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return already exist error if resource service return err conflict",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:          testResourceID,
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					UserID:      testResource.UserID,
					NamespaceID: testResource.NamespaceID,
				}).Return(resource.Resource{}, resource.ErrConflict)
			},
			request: &shieldv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return success if resource service return nil",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:          testResourceID,
					Name:        testResource.Name,
					ProjectID:   testResource.ProjectID,
					NamespaceID: testResource.NamespaceID,
					UserID:      testResource.UserID,
				}).Return(testResource, nil)
			},
			request: &shieldv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &shieldv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					UserId:    testUserID,
				},
			},
			want: &shieldv1beta1.UpdateProjectResourceResponse{
				Resource: testResourcePB,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv, mockProjectSrv)
			}
			mockDep := Handler{resourceService: mockResourceSrv, projectService: mockProjectSrv}
			resp, err := mockDep.UpdateProjectResource(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
