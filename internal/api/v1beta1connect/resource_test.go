package v1beta1connect

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
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
		name            string
		setup           func(rs *mocks.ResourceService)
		request         *connect.Request[frontierv1beta1.ListResourcesRequest]
		want            *connect.Response[frontierv1beta1.ListResourcesResponse]
		wantErr         error
		skipErrMsgCheck bool
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{}).Return([]resource.Resource{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListResourcesRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("ListResources: namespace= project_id=: %w", errors.New("test error"))),
		},
		{
			// proto error message is non-deterministic (randomly uses U+00a0 or space)
			// so we only check the error code
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
			request:         connect.NewRequest(&frontierv1beta1.ListResourcesRequest{}),
			want:            nil,
			wantErr:         connect.NewError(connect.CodeInternal, errors.New("proto error")),
			skipErrMsgCheck: true,
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
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, connect.CodeOf(tt.wantErr), connect.CodeOf(err))
				if !tt.skipErrMsgCheck {
					assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
				}
			} else {
				assert.NoError(t, err)
			}
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("ListProjectResources: namespace= project_id=%s: %w", testProjectID, errors.New("test error"))),
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
		setup   func(rs *mocks.ResourceService)
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
			name: "should return not found if the project does not exist",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					Name:          testResource.Name,
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testUserID,
					PrincipalType: testResource.PrincipalType,
				}).Return(resource.Resource{}, fmt.Errorf("failed to get project: %w", project.ErrNotExist))
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
			wantErr: connect.NewError(connect.CodeNotFound, ErrProjectNotFound),
		},
		{
			name: "should return internal error if resource service returns error",
			setup: func(rs *mocks.ResourceService) {
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("CreateProjectResource: resource_id= project_id=%s resource_name=%s namespace=%s principal=%s: %w", testResource.ProjectID, testResource.Name, testResource.NamespaceID, testUserID, errors.New("test error"))),
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.ResourceService) {
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
			setup: func(rs *mocks.ResourceService) {
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
			setup: func(rs *mocks.ResourceService) {
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
			setup: func(rs *mocks.ResourceService) {
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
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			h := ConnectHandler{resourceService: mockResourceSrv}
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

func TestConnectHandler_CreateOrganizationResource(t *testing.T) {
	orgResource := resource.Resource{
		ID:            utils.NewString(),
		URN:           "frn:" + testOrgID + ":" + testNSID + ":store-1",
		Name:          "store-1",
		OrgID:         testOrgID,
		NamespaceID:   testNSID,
		PrincipalID:   testUserID,
		PrincipalType: schema.UserPrincipal,
	}
	toCreate := resource.Resource{
		Name:          orgResource.Name,
		OrgID:         testOrgID,
		NamespaceID:   testNSID,
		PrincipalID:   testUserID,
		PrincipalType: schema.UserPrincipal,
	}
	validRequest := func() *connect.Request[frontierv1beta1.CreateOrganizationResourceRequest] {
		return connect.NewRequest(&frontierv1beta1.CreateOrganizationResourceRequest{
			OrgId: testOrgID,
			Body: &frontierv1beta1.ResourceRequestBody{
				Name:      orgResource.Name,
				Namespace: testNSID,
				Principal: testUserID,
			},
		})
	}

	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *connect.Request[frontierv1beta1.CreateOrganizationResourceRequest]
		want    *connect.Response[frontierv1beta1.CreateOrganizationResourceResponse]
		wantErr error
	}{
		{
			name: "should return error if request body is nil",
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationResourceRequest{
				OrgId: testOrgID,
			}),
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return not found if the org does not exist",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Create(mock.Anything, toCreate).Return(resource.Resource{}, fmt.Errorf("failed to get organization: %w", organization.ErrNotExist))
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name: "should return bad request if a field does not exist in its foreign reference",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Create(mock.Anything, toCreate).Return(resource.Resource{}, resource.ErrInvalidDetail)
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return conflict if the resource already exists",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Create(mock.Anything, toCreate).Return(resource.Resource{}, resource.ErrConflict)
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return internal error if the resource service fails",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Create(mock.Anything, toCreate).Return(resource.Resource{}, errors.New("test error"))
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("CreateOrganizationResource: resource_id= org_id=%s resource_name=%s namespace=%s principal=%s: %w", testOrgID, orgResource.Name, testNSID, testUserID, errors.New("test error"))),
		},
		{
			name: "should return the resource with its org and no project",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Create(mock.Anything, toCreate).Return(orgResource, nil)
			},
			request: validRequest(),
			want: connect.NewResponse(&frontierv1beta1.CreateOrganizationResourceResponse{
				Resource: &frontierv1beta1.Resource{
					Id:        orgResource.ID,
					Name:      orgResource.Name,
					Urn:       orgResource.URN,
					OrgId:     testOrgID,
					Namespace: testNSID,
					Principal: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, testUserID),
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := mocks.NewResourceService(t)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			h := ConnectHandler{resourceService: mockResourceSrv}
			resp, err := h.CreateOrganizationResource(context.Background(), tt.request)
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

func TestConnectHandler_UpdateOrganizationResource(t *testing.T) {
	orgResource := resource.Resource{
		ID:            utils.NewString(),
		URN:           "frn:" + testOrgID + ":" + testNSID + ":store-1",
		Name:          "store-1",
		Title:         "Store One",
		OrgID:         testOrgID,
		NamespaceID:   testNSID,
		PrincipalID:   testUserID,
		PrincipalType: schema.UserPrincipal,
	}
	toUpdate := resource.Resource{
		ID:            orgResource.ID,
		Name:          orgResource.Name,
		Title:         orgResource.Title,
		OrgID:         testOrgID,
		NamespaceID:   testNSID,
		PrincipalID:   testUserID,
		PrincipalType: schema.UserPrincipal,
	}
	validRequest := func() *connect.Request[frontierv1beta1.UpdateOrganizationResourceRequest] {
		return connect.NewRequest(&frontierv1beta1.UpdateOrganizationResourceRequest{
			Id:    orgResource.ID,
			OrgId: testOrgID,
			Body: &frontierv1beta1.ResourceRequestBody{
				Name:      orgResource.Name,
				Title:     orgResource.Title,
				Namespace: testNSID,
				Principal: testUserID,
			},
		})
	}

	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *connect.Request[frontierv1beta1.UpdateOrganizationResourceRequest]
		want    *connect.Response[frontierv1beta1.UpdateOrganizationResourceResponse]
		wantErr error
	}{
		{
			name: "should return error if request body is nil",
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationResourceRequest{
				Id:    orgResource.ID,
				OrgId: testOrgID,
			}),
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return not found if the resource does not exist",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Update(mock.Anything, toUpdate).Return(resource.Resource{}, resource.ErrNotExist)
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return bad request if a field is invalid",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Update(mock.Anything, toUpdate).Return(resource.Resource{}, resource.ErrInvalidDetail)
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return conflict if the update conflicts",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Update(mock.Anything, toUpdate).Return(resource.Resource{}, resource.ErrConflict)
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return internal error if the resource service fails",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Update(mock.Anything, toUpdate).Return(resource.Resource{}, errors.New("test error"))
			},
			request: validRequest(),
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateOrganizationResource: resource_id=%s org_id=%s resource_name=%s namespace=%s principal=%s: %w", orgResource.ID, testOrgID, orgResource.Name, testNSID, testUserID, errors.New("test error"))),
		},
		{
			name: "should return the updated resource with its org",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Update(mock.Anything, toUpdate).Return(orgResource, nil)
			},
			request: validRequest(),
			want: connect.NewResponse(&frontierv1beta1.UpdateOrganizationResourceResponse{
				Resource: &frontierv1beta1.Resource{
					Id:        orgResource.ID,
					Name:      orgResource.Name,
					Title:     orgResource.Title,
					Urn:       orgResource.URN,
					OrgId:     testOrgID,
					Namespace: testNSID,
					Principal: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, testUserID),
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := mocks.NewResourceService(t)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			h := ConnectHandler{resourceService: mockResourceSrv}
			resp, err := h.UpdateOrganizationResource(context.Background(), tt.request)
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

func TestConnectHandler_DeleteOrganizationResource(t *testing.T) {
	orgResource := resource.Resource{
		ID:            utils.NewString(),
		Name:          "store-1",
		OrgID:         testOrgID,
		NamespaceID:   testNSID,
		PrincipalID:   testUserID,
		PrincipalType: schema.UserPrincipal,
	}
	request := connect.NewRequest(&frontierv1beta1.DeleteOrganizationResourceRequest{
		Id:    orgResource.ID,
		OrgId: testOrgID,
	})

	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		wantErr error
	}{
		{
			name: "should return not found if the resource does not exist",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.Anything, orgResource.ID).Return(resource.Resource{}, resource.ErrNotExist)
			},
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should return internal error if the resource lookup fails",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.Anything, orgResource.ID).Return(resource.Resource{}, errors.New("test error"))
			},
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteOrganizationResource.GetResource: resource_id=%s: %w", orgResource.ID, errors.New("test error"))),
		},
		{
			name: "should return internal error if the delete fails",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.Anything, orgResource.ID).Return(orgResource, nil)
				rs.EXPECT().Delete(mock.Anything, testNSID, orgResource.ID).Return(errors.New("delete error"))
			},
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteOrganizationResource: resource_id=%s org_id=%s namespace=%s: %w", orgResource.ID, testOrgID, testNSID, errors.New("delete error"))),
		},
		{
			name: "should return not found if the resource is deleted after the lookup",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.Anything, orgResource.ID).Return(orgResource, nil)
				rs.EXPECT().Delete(mock.Anything, testNSID, orgResource.ID).Return(resource.ErrNotExist)
			},
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
		},
		{
			name: "should delete the resource",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().Get(mock.Anything, orgResource.ID).Return(orgResource, nil)
				rs.EXPECT().Delete(mock.Anything, testNSID, orgResource.ID).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := mocks.NewResourceService(t)
			tt.setup(mockResourceSrv)
			h := ConnectHandler{resourceService: mockResourceSrv}
			resp, err := h.DeleteOrganizationResource(context.Background(), request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, connect.NewResponse(&frontierv1beta1.DeleteOrganizationResourceResponse{}), resp)
			}
		})
	}
}

func TestConnectHandler_GetProjectResource(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(rs *mocks.ResourceService)
		request         *connect.Request[frontierv1beta1.GetProjectResourceRequest]
		want            *connect.Response[frontierv1beta1.GetProjectResourceResponse]
		wantErr         error
		skipErrMsgCheck bool
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("GetProjectResource: resource_id=%s: %w", testResource.ID, errors.New("test error"))),
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
			// proto error message is non-deterministic (randomly uses U+00a0 or space)
			// so we only check the error code
			name: "should return internal error if transform fails",
			setup: func(rs *mocks.ResourceService) {
				invalidResource := testResource
				invalidResource.Metadata = map[string]any{
					"invalid": func() {}, // functions can't be marshaled
				}
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(invalidResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:            nil,
			wantErr:         connect.NewError(connect.CodeInternal, errors.New("proto error")),
			skipErrMsgCheck: true,
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
				if !tt.skipErrMsgCheck {
					assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
				}
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateProjectResource.GetProject: project_id=%s: %w", testResource.ProjectID, errors.New("test error"))),
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateProjectResource: resource_id=%s project_id=%s resource_name=%s namespace=%s principal=%s: %w", testResourceID, testResource.ProjectID, testResource.Name, testResource.NamespaceID, testUserID, errors.New("test error"))),
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
			name: "should pass the title from the request body to the resource service",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				titledResource := testResource
				titledResource.Title = "a resource title"
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), resource.Resource{
					ID:            testResourceID,
					Name:          testResource.Name,
					Title:         "a resource title",
					ProjectID:     testResource.ProjectID,
					NamespaceID:   testResource.NamespaceID,
					PrincipalID:   testResource.PrincipalID,
					PrincipalType: testResource.PrincipalType,
				}).Return(titledResource, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateProjectResourceRequest{
				Id:        testResourceID,
				ProjectId: testResource.ProjectID,
				Body: &frontierv1beta1.ResourceRequestBody{
					Name:      testResource.Name,
					Title:     "a resource title",
					Namespace: testResource.NamespaceID,
					Principal: testUserID,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateProjectResourceResponse{
				Resource: &frontierv1beta1.Resource{
					Id:        testResource.ID,
					Name:      testResource.Name,
					Title:     "a resource title",
					Urn:       testResource.URN,
					ProjectId: testProjectID,
					Namespace: testNSID,
					Principal: schema.JoinNamespaceAndResourceID(testResource.PrincipalType, testResource.PrincipalID),
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteProjectResource.GetResource: resource_id=%s: %w", testResource.ID, errors.New("test error"))),
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteProjectResource.GetProject: resource_id=%s project_id=%s: %w", testResource.ID, testResource.ProjectID, errors.New("test error"))),
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
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteProjectResource: resource_id=%s project_id=%s namespace=%s: %w", testResource.ID, testResource.ProjectID, testResource.NamespaceID, errors.New("delete error"))),
		},
		{
			name: "should return not found if the resource is deleted after the lookup",
			setup: func(rs *mocks.ResourceService, ps *mocks.ProjectService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ID).Return(testResource, nil)
				ps.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testResource.ProjectID).Return(project.Project{
					ID: testResource.ProjectID,
					Organization: organization.Organization{
						ID: "test-org-id",
					},
				}, nil)
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testResource.NamespaceID, testResource.ID).Return(resource.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteProjectResourceRequest{
				Id: testResource.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrResourceNotFound),
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
