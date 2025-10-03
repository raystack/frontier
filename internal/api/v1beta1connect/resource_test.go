package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/metadata"
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
		request *connect.Request[frontierv1beta1.ListResourcesRequest]
		want    *connect.Response[frontierv1beta1.ListResourcesResponse]
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{}).Return([]resource.Resource{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListResourcesRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return internal error if transformation fails",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{}).Return([]resource.Resource{
					{
						Metadata: metadata.Metadata{
							"key": map[int]any{},
						},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListResourcesRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return resources if resource service return nil error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{}).Return([]resource.Resource{
					testResource,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListResourcesRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListResourcesResponse{
				Resources: []*frontierv1beta1.Resource{
					testResourcePB,
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			mockDep := ConnectHandler{resourceService: mockResourceSrv}
			resp, err := mockDep.ListResources(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestConnectHandler_ListProjectResources(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *connect.Request[frontierv1beta1.ListProjectResourcesRequest]
		want    *connect.Response[frontierv1beta1.ListProjectResourcesResponse]
		wantErr error
	}{
		{
			name: "should return internal error if resource service returns error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"),
					resource.Filter{ProjectID: testProjectID}).Return(nil, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListProjectResourcesRequest{
				ProjectId: testProjectID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return empty list if resource service returns empty slice",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{ProjectID: testProjectID}).Return([]resource.Resource{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListProjectResourcesRequest{
				ProjectId: testProjectID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListProjectResourcesResponse{
				Resources: nil,
			}),
			wantErr: nil,
		},
		{
			name: "should return resources if resource service returns resources",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{ProjectID: testProjectID}).Return([]resource.Resource{testResource}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListProjectResourcesRequest{
				ProjectId: testProjectID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListProjectResourcesResponse{
				Resources: []*frontierv1beta1.Resource{testResourcePB},
			}),
			wantErr: nil,
		},
		{
			name: "should handle namespace parameter correctly",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{
					NamespaceID: "test-namespace",
					ProjectID:   testProjectID,
				}).Return([]resource.Resource{testResource}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListProjectResourcesRequest{
				ProjectId: testProjectID,
				Namespace: "test-namespace",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListProjectResourcesResponse{
				Resources: []*frontierv1beta1.Resource{testResourcePB},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			h := ConnectHandler{resourceService: mockResourceSrv}
			resp, err := h.ListProjectResources(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, resp)
			}
		})
	}
}

func TestConnectHandler_CreateProjectResource(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService, ps *mocks.ProjectService)
		request *connect.Request[frontierv1beta1.CreateProjectResourceRequest]
		want    *connect.Response[frontierv1beta1.CreateProjectResourceResponse]
		wantErr error
	}{
		{
			name: "should return error if request body is nil",
			request: connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testProjectID,
				Body:      nil,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return internal error if project service returns error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return internal error if resource service returns error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrInvalidDetail)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return conflict error if resource already exists",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return success if resource service returns nil",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(testResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateProjectResourceResponse{
				Resource: testResourcePB,
			}),
			wantErr: nil,
		},
		{
			name: "should handle metadata correctly",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(res resource.Resource) bool {
					return res.Name == testResource.Name && len(res.Metadata) > 0
				})).Return(testResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateProjectResourceRequest{
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"key": structpb.NewStringValue("value"),
						},
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateProjectResourceResponse{
				Resource: testResourcePB,
			}),
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
			h := ConnectHandler{resourceService: mockResourceSrv, projectService: mockProjectSrv}
			resp, err := h.CreateProjectResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, resp)
			}
		})
	}
}

func TestConnectHandler_GetProjectResource(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *connect.Request[frontierv1beta1.GetProjectResourceRequest]
		want    *connect.Response[frontierv1beta1.GetProjectResourceResponse]
		wantErr error
	}{
		{
			name: "should return internal error if resource service returns error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(resource.Resource{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(resource.Resource{}, resource.ErrInvalidID)
			},
			request: connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-id").Return(resource.Resource{}, resource.ErrInvalidUUID)
			},
			request: connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{
				Id: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return not found error if resource not exist",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(resource.Resource{}, resource.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return success if resource service returns resource",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(testResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			}),
			want: connect.NewResponse(&frontierv1beta1.GetProjectResourceResponse{
				Resource: testResourcePB,
			}),
			wantErr: nil,
		},
		{
			name: "should return internal error if transform fails",
			setup: func(rs *mocks.ResourceService) {
				invalidResource := testResource
				invalidResource.Metadata = map[string]interface{}{
					"invalid": func() {}, // functions can't be marshaled
				}
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(invalidResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			h := ConnectHandler{resourceService: mockResourceSrv}
			resp, err := h.GetProjectResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, resp)
			}
		})
	}
}

func TestConnectHandler_UpdateProjectResource(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService, ps *mocks.ProjectService)
		request *connect.Request[frontierv1beta1.UpdateProjectResourceRequest]
		want    *connect.Response[frontierv1beta1.UpdateProjectResourceResponse]
		wantErr error
	}{
		{
			name: "should return error if request body is nil",
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				ProjectId: testProjectID,
				Body:      nil,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return internal error if project service returns error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return internal error if resource service returns error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
					NamespaceID:   testResource.NamespaceID,
				}).Return(resource.Resource{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if resource not exist",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return not found error if id is invalid",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					ID:            "some-id",
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrInvalidUUID)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        "some-id",
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return bad request error if field value is invalid",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, resource.ErrInvalidDetail)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return conflict error if resource service returns conflict",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
				}, nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
					NamespaceID:   testResource.NamespaceID,
				}).Return(resource.Resource{}, resource.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return success if resource service returns updated resource",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					ID:            testResourceID,
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
				}).Return(testResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateProjectResourceResponse{
				Resource: testResourcePB,
			}),
			wantErr: nil,
		},
		{
			name: "should handle metadata correctly",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(res resource.Resource) bool {
					return res.Name == testResource.Name && len(res.Metadata) > 0
				})).Return(testResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"updated": structpb.NewStringValue("value"),
						},
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateProjectResourceResponse{
				Resource: testResourcePB,
			}),
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
			h := ConnectHandler{resourceService: mockResourceSrv, projectService: mockProjectSrv}
			resp, err := h.UpdateProjectResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, resp)
			}
		})
	}
}

func TestConnectHandler_DeleteProjectResource(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService, ps *mocks.ProjectService)
		request *connect.Request[frontierv1beta1.DeleteProjectResourceRequest]
		want    *connect.Response[frontierv1beta1.DeleteProjectResourceResponse]
		wantErr error
	}{
		{
			name: "should return internal error if resource service Get returns error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(resource.Resource{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if resource not exist",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(resource.Resource{}, resource.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return not found error if id is invalid",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-id").Return(resource.Resource{}, resource.ErrInvalidUUID)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(resource.Resource{}, resource.ErrInvalidID)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return internal error if project service returns error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(testResource, nil)
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return internal error if resource service Delete returns error",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(testResource, nil)
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testResource.NamespaceID, testResource.ID).Return(errors.New("delete error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success if resource is deleted successfully",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(testResource, nil)
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testResource.NamespaceID, testResource.ID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteProjectResourceResponse{}),
			wantErr: nil,
		},
		{
			name: "should handle deletion with different namespace correctly",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				resourceWithDiffNS := testResource
				resourceWithDiffNS.NamespaceID = "different-namespace"
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(resourceWithDiffNS, nil)
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), "different-namespace", testResource.ID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteProjectResourceResponse{}),
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
			h := ConnectHandler{resourceService: mockResourceSrv, projectService: mockProjectSrv}
			resp, err := h.DeleteProjectResource(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, resp)
			}
		})
	}
}
