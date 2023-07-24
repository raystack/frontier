package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testResourceID = utils.NewString()
	testResource   = resource.Resource{
		ID:            testResourceID,
		URN:           "res-urn",
		Name:          "a resource name",
		ProjectID:     testProjectID,
		NamespaceID:   testNSID,
		PrincipalID:   testUserID,
		PrincipalType: schema.UserPrincipal,
	}
	testResourcePB = &frontierv1beta1.Resource{
		Id:        testResource.ID,
		Name:      testResource.Name,
		Urn:       testResource.URN,
		ProjectId: testProjectID,
		Namespace: testNSID,
		Principal: schema.JoinNamespaceAndResourceID(testResource.PrincipalType, testResource.PrincipalID),
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
)

func TestHandler_ListResources(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *frontierv1beta1.ListResourcesRequest
		want    *frontierv1beta1.ListResourcesResponse
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), resource.Filter{}).Return([]resource.Resource{}, errors.New("some error"))
			},
			request: &frontierv1beta1.ListResourcesRequest{},
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
			request: &frontierv1beta1.ListResourcesRequest{},
			want: &frontierv1beta1.ListResourcesResponse{
				Resources: []*frontierv1beta1.Resource{
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
		request *frontierv1beta1.CreateProjectResourceRequest
		want    *frontierv1beta1.CreateProjectResourceResponse
		wantErr error
	}{
		{
			name: "should return error if request body is nil",
			request: &frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testProjectID,
				Body:      nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad body error if unable to build metadata map",
			request: &frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"1": {},
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return internal error if resource service return some error",
			setup: func(ctx context.Context, rs *mocks.ResourceService, ps *mocks.ProjectService, rls *mocks.RelationService) context.Context {
				rs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), resource.Resource{
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, errors.New("some error"))
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				}},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(ctx context.Context, rs *mocks.ResourceService, ps *mocks.ProjectService, rls *mocks.RelationService) context.Context {
				rs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), resource.Resource{
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrInvalidDetail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
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
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(testResource, nil)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			},
			want: &frontierv1beta1.CreateProjectResourceResponse{
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
		request *frontierv1beta1.GetProjectResourceRequest
		want    *frontierv1beta1.GetProjectResourceResponse
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ID).Return(resource.Resource{}, errors.New("some error"))
			},
			request: &frontierv1beta1.GetProjectResourceRequest{
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
			request: &frontierv1beta1.GetProjectResourceRequest{},
			want:    nil,
			wantErr: grpcResourceNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(resource.Resource{}, resource.ErrInvalidUUID)
			},
			request: &frontierv1beta1.GetProjectResourceRequest{
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
			request: &frontierv1beta1.GetProjectResourceRequest{
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
			request: &frontierv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			},
			want: &frontierv1beta1.GetProjectResourceResponse{
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
		request *frontierv1beta1.UpdateProjectResourceRequest
		want    *frontierv1beta1.UpdateProjectResourceResponse
		wantErr error
	}{
		{
			name: "should return error if request body is nil",
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				ProjectId: testProjectID,
				Body:      nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad body error if unable to build metadata map",
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				ProjectId: testProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"1": {},
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResourceID,
				}, nil)

				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
					NamespaceID:   testResource.NamespaceID,
				}).Return(resource.Resource{}, errors.New("some error"))
			},
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
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
					ID:            "",
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrInvalidID)
			},
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
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
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrNotExist)
			},
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
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
					ID:            "some-id",
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrInvalidUUID)
			},
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				Id:        "some-id",
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
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
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
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
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
					NamespaceID:   testResource.NamespaceID,
				}).Return(resource.Resource{}, resource.ErrConflict)
			},
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
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
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
				}).Return(testResource, nil)
			},
			request: &frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			},
			want: &frontierv1beta1.UpdateProjectResourceResponse{
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

func TestHandler_ListProjectResources(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService, ps *mocks.ProjectService)
		request *frontierv1beta1.ListProjectResourcesRequest
		want    *frontierv1beta1.ListProjectResourcesResponse
		wantErr error
	}{
		{
			name: "should return internal error if resource service return error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"),
					resource.Filter{ProjectID: testProjectID}).Return(nil, errors.New("error"))
			},
			request: &frontierv1beta1.ListProjectResourcesRequest{
				ProjectId: testProjectID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return success if resource service return nil",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), resource.Filter{ProjectID: testProjectID}).Return([]resource.Resource{}, nil)
			},
			request: &frontierv1beta1.ListProjectResourcesRequest{
				ProjectId: testProjectID,
			},
			want: &frontierv1beta1.ListProjectResourcesResponse{
				Resources: nil,
			},
			wantErr: nil,
		},
		{
			name: "should return success if resource service return resources",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), resource.Filter{ProjectID: testProjectID}).Return([]resource.Resource{testResource}, nil)
			},
			request: &frontierv1beta1.ListProjectResourcesRequest{
				ProjectId: testProjectID,
			},
			want: &frontierv1beta1.ListProjectResourcesResponse{
				Resources: []*frontierv1beta1.Resource{testResourcePB},
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
			resp, err := mockDep.ListProjectResources(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
