package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
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
	testGroupID  = "9f256f86-31a3-11ec-8d3d-0242ac130003"
	testGroupMap = map[string]group.Group{
		"9f256f86-31a3-11ec-8d3d-0242ac130003": {
			ID:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
			Name: "group-1",
			Metadata: metadata.Metadata{
				"foo": "bar",
			},
			OrganizationID: "9f256f86-31a3-11ec-8d3d-0242ac130003",
			CreatedAt:      time.Time{},
			UpdatedAt:      time.Time{},
		},
	}
)

func TestHandler_ListGroups(t *testing.T) {
	randomID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *connect.Request[frontierv1beta1.ListGroupsRequest]
		want    *connect.Response[frontierv1beta1.ListGroupsResponse]
		wantErr error
	}{
		{
			name: "should return empty groups if query param org_id is not uuid",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().List(mock.Anything, group.Filter{
					SU:             true,
					OrganizationID: "some-id",
				}).Return([]group.Group{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupsRequest{
				OrgId: "some-id",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListGroupsResponse{
				Groups: nil,
			}),
			wantErr: nil,
		},
		{
			name: "should return empty groups if query param org_id is not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().List(mock.Anything, group.Filter{
					SU:             true,
					OrganizationID: randomID,
				}).Return([]group.Group{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupsRequest{
				OrgId: randomID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListGroupsResponse{
				Groups: nil,
			}),
			wantErr: nil,
		},
		{
			name: "should return all groups if no query param filter exist",
			setup: func(gs *mocks.GroupService) {
				var testGroupList []group.Group
				for _, u := range testGroupMap {
					testGroupList = append(testGroupList, u)
				}
				gs.EXPECT().List(mock.Anything, group.Filter{
					SU: true,
				}).Return(testGroupList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupsRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListGroupsResponse{
				Groups: []*frontierv1beta1.Group{
					{
						Id:    testGroupID,
						Name:  "group-1",
						OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"foo": structpb.NewStringValue("bar"),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return filtered groups if query param org_id exist",
			setup: func(gs *mocks.GroupService) {
				var testGroupList []group.Group
				for _, u := range testGroupMap {
					testGroupList = append(testGroupList, u)
				}
				gs.EXPECT().List(mock.Anything, group.Filter{
					SU:             true,
					OrganizationID: "9f256f86-31a3-11ec-8d3d-0242ac130003",
				}).Return(testGroupList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupsRequest{
				OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListGroupsResponse{
				Groups: []*frontierv1beta1.Group{
					{
						Id:    testGroupID,
						Name:  "group-1",
						OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"foo": structpb.NewStringValue("bar"),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return an error if Group service return some error ",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().List(mock.Anything, group.Filter{
					SU:             true,
					OrganizationID: "9f256f86-31a3-11ec-8d3d-0242ac130003",
				}).Return(nil, errors.New("test-error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupsRequest{
				OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return error while traversing group list if key is integer type",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().List(mock.Anything, group.Filter{
					SU:             true,
					OrganizationID: "some-id",
				}).Return([]group.Group{
					{
						Metadata: metadata.Metadata{
							"key": map[int]any{},
						},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupsRequest{
				OrgId: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc)
			}
			h := ConnectHandler{
				groupService: mockGroupSvc,
			}
			got, err := h.ListGroups(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.want == nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			}
		})
	}
}

func TestConnectHandler_CreateGroup(t *testing.T) {
	someGroupID := utils.NewString()
	tests := []struct {
		name        string
		setup       func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService)
		request     *connect.Request[frontierv1beta1.CreateGroupRequest]
		want        *connect.Response[frontierv1beta1.CreateGroupResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return error if request body is nil",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				Body: nil,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadRequest,
		},
		{
			name: "should return error if org does not exist",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgNotFound,
		},
		{
			name: "should return error if org is disabled",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgDisabled,
		},
		{
			name: "should return error if error in metadata validation",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(errors.New("some-error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadBodyMetaSchemaError,
		},
		{
			name: "should return unauthenticated error if auth email in context is empty and group service return invalid user email",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.Anything, group.Group{
					OrganizationID: testOrgID,
					Title:          "Test Group",
					Name:           "Test-Group",
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, user.ErrInvalidEmail)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Title:    "Test Group",
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeUnauthenticated,
			wantErrMsg:  ErrUnauthenticated,
		},
		{
			name: "should return already exist error if group service return error conflict",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.Anything, group.Group{
					Name:           "some-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeAlreadyExists,
			wantErrMsg:  ErrConflictRequest,
		},
		{
			name: "should return bad request error if name empty and group service return invalid detail error",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.Anything, group.Group{
					Name:           "some-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadRequest,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.Anything, group.Group{
					Name:           "some-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return success if group service return nil",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.Anything, group.Group{
					Name:           "some-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{
					ID:             someGroupID,
					Name:           "some-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateGroupResponse{
				Group: &frontierv1beta1.Group{
					Id:    someGroupID,
					Name:  "some-group",
					OrgId: testOrgID,
					Metadata: &structpb.Struct{
						Fields: make(map[string]*structpb.Value),
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockOrgSvc := new(mocks.OrganizationService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockMetaSchemaSvc, mockOrgSvc)
			}
			h := &ConnectHandler{
				groupService:      mockGroupSvc,
				orgService:        mockOrgSvc,
				metaSchemaService: mockMetaSchemaSvc,
			}
			got, err := h.CreateGroup(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				connectErr := &connect.Error{}
				assert.True(t, errors.As(err, &connectErr))
				assert.Equal(t, tt.wantErrCode, connectErr.Code())
				assert.Equal(t, tt.wantErrMsg.Error(), connectErr.Message())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}
		})
	}
}

func TestConnectHandler_GetGroup(t *testing.T) {
	someGroupID := utils.NewString()
	tests := []struct {
		name        string
		setup       func(gs *mocks.GroupService, os *mocks.OrganizationService)
		request     *connect.Request[frontierv1beta1.GetGroupRequest]
		want        *connect.Response[frontierv1beta1.GetGroupResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return error if org does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetGroupRequest{
				OrgId: testOrgID,
				Id:    someGroupID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgNotFound,
		},
		{
			name: "should return error if org is disabled",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.GetGroupRequest{
				OrgId: testOrgID,
				Id:    someGroupID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgDisabled,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().Get(mock.Anything, someGroupID).Return(group.Group{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return not found error if id is invalid",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().Get(mock.Anything, "").Return(group.Group{}, group.ErrInvalidID)
			},
			request: connect.NewRequest(&frontierv1beta1.GetGroupRequest{
				Id:    "",
				OrgId: testOrgID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrGroupNotFound,
		},
		{
			name: "should return not found error if group not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().Get(mock.Anything, someGroupID).Return(group.Group{}, group.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrGroupNotFound,
		},
		{
			name: "should return success if group service return nil",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().Get(mock.Anything, testGroupID).Return(testGroupMap[testGroupID], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetGroupRequest{
				Id:    testGroupID,
				OrgId: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.GetGroupResponse{
				Group: &frontierv1beta1.Group{
					Id:    testGroupID,
					Name:  "group-1",
					OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: false,
		},
		{
			name: "should return internal error if group service return key as integer type",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().Get(mock.Anything, testGroupID).Return(group.Group{
					Metadata: metadata.Metadata{
						"key": map[int]any{},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetGroupRequest{
				Id:    testGroupID,
				OrgId: testOrgID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSvc := new(mocks.OrganizationService)
			mockGroupSvc := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockOrgSvc)
			}
			h := &ConnectHandler{
				groupService: mockGroupSvc,
				orgService:   mockOrgSvc,
			}
			got, err := h.GetGroup(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				connectErr := &connect.Error{}
				assert.True(t, errors.As(err, &connectErr))
				assert.Equal(t, tt.wantErrCode, connectErr.Code())
				assert.Equal(t, tt.wantErrMsg.Error(), connectErr.Message())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}
		})
	}
}

func TestConnectHandler_UpdateGroup(t *testing.T) {
	someGroupID := utils.NewString()
	tests := []struct {
		name        string
		setup       func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService)
		request     *connect.Request[frontierv1beta1.UpdateGroupRequest]
		want        *connect.Response[frontierv1beta1.UpdateGroupResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return bad request error if body is empty",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:   someGroupID,
				Body: nil,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadRequest,
		},
		{
			name: "should return error if org does not exist",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgNotFound,
		},
		{
			name: "should return org is disabled",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgDisabled,
		},
		{
			name: "should return error if error in metadata validation",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(errors.New("some-error"))
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "new-group",
					Metadata: &structpb.Struct{},
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadBodyMetaSchemaError,
		},
		{
			name: "should return not found error if group id is not uuid (slug) and does not exist",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.Anything, group.Group{
					ID:             "some-id",
					Name:           "some-id",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    "some-id",
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "some-id",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrGroupNotFound,
		},
		{
			name: "should return not found error if group id is uuid and does not exist",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.Anything, group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrGroupNotFound,
		},
		{
			name: "should return already exist error if group service return error conflict",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.Anything, group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeAlreadyExists,
			wantErrMsg:  ErrConflictRequest,
		},
		{
			name: "should return bad request error if name is empty",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.Anything, group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
			wantErrMsg:  ErrBadRequest,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.Anything, group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
			wantErrMsg:  ErrInternalServerError,
		},
		{
			name: "should return success if updated by id and group service return nil error",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.Anything, group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: testOrgID,
					Metadata:       metadata.Metadata{},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateGroupResponse{
				Group: &frontierv1beta1.Group{
					Id:    someGroupID,
					Name:  "new-group",
					OrgId: testOrgID,
					Metadata: &structpb.Struct{
						Fields: make(map[string]*structpb.Value),
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockOrgSvc := new(mocks.OrganizationService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockMetaSchemaSvc, mockOrgSvc)
			}
			h := &ConnectHandler{
				groupService:      mockGroupSvc,
				orgService:        mockOrgSvc,
				metaSchemaService: mockMetaSchemaSvc,
			}
			got, err := h.UpdateGroup(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				connectErr := &connect.Error{}
				assert.True(t, errors.As(err, &connectErr))
				assert.Equal(t, tt.wantErrCode, connectErr.Code())
				assert.Equal(t, tt.wantErrMsg.Error(), connectErr.Message())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}
		})
	}
}

func TestConnectHandler_ListOrganizationGroups(t *testing.T) {
	var validGroupResponseWithUser = &frontierv1beta1.Group{
		Id:    testGroupID,
		Name:  "group-1",
		OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"foo": structpb.NewStringValue("bar"),
			},
		},
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
		Users: []*frontierv1beta1.User{
			{
				Id: testUserID,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			},
		},
	}

	tests := []struct {
		name        string
		setup       func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService)
		request     *connect.Request[frontierv1beta1.ListOrganizationGroupsRequest]
		want        *connect.Response[frontierv1beta1.ListOrganizationGroupsResponse]
		wantErr     bool
		wantErrCode connect.Code
		wantErrMsg  error
	}{
		{
			name: "should return error if org does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationGroupsRequest{
				OrgId: testOrgID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgNotFound,
		},
		{
			name: "should return error if org is disabled",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationGroupsRequest{
				OrgId: testOrgID,
			}),
			want:        nil,
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
			wantErrMsg:  ErrOrgDisabled,
		},
		{
			name: "should return empty groups list if organization with valid uuid is not found",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().List(mock.Anything, group.Filter{
					OrganizationID: testOrgID,
				}).Return([]group.Group{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationGroupsRequest{
				OrgId: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationGroupsResponse{
				Groups: nil,
			}),
			wantErr: false,
		},
		{
			name: "should return success if list organization groups and group service return nil error",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				var testGroupList []group.Group
				for _, u := range testGroupMap {
					testGroupList = append(testGroupList, u)
				}
				gs.EXPECT().List(mock.Anything, group.Filter{
					OrganizationID: testOrgID,
				}).Return(testGroupList, nil)
				us.EXPECT().ListByGroup(mock.Anything, testGroupID, "").Return([]user.User{
					{
						ID:       testUserID,
						Metadata: map[string]any{},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationGroupsRequest{
				OrgId:       testOrgID,
				WithMembers: true,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationGroupsResponse{
				Groups: []*frontierv1beta1.Group{
					validGroupResponseWithUser,
				},
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSvc := new(mocks.OrganizationService)
			mockGroupSvc := new(mocks.GroupService)
			mockUserSvc := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockOrgSvc, mockUserSvc)
			}
			h := &ConnectHandler{
				groupService: mockGroupSvc,
				orgService:   mockOrgSvc,
				userService:  mockUserSvc,
			}
			got, err := h.ListOrganizationGroups(context.Background(), tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				connectErr := &connect.Error{}
				assert.True(t, errors.As(err, &connectErr))
				assert.Equal(t, tt.wantErrCode, connectErr.Code())
				assert.Equal(t, tt.wantErrMsg.Error(), connectErr.Message())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}
		})
	}
}

func TestConnectHandler_ListGroupUsers(t *testing.T) {
	someGroupID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService)
		request *connect.Request[frontierv1beta1.ListGroupUsersRequest]
		want    *connect.Response[frontierv1beta1.ListGroupUsersResponse]
		wantErr error
	}{
		{
			name: "should return error if org does not exist",
			setup: func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupUsersRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name: "should error if org is disabled",
			setup: func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupUsersRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return internal server error if error in listing group users",
			setup: func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByGroup(mock.Anything, someGroupID, "").Return(nil, errors.New("some error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupUsersRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return error if metadata transformation fails in list of group users",
			setup: func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				testUserList := []user.User{
					{
						Metadata: metadata.Metadata{
							"key": map[int]string{},
						},
					},
				}

				us.EXPECT().ListByGroup(mock.Anything, someGroupID, "").Return(testUserList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupUsersRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success if list group users and group service return nil error",
			setup: func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().ListByGroup(mock.Anything, someGroupID, "").Return(testUserList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupUsersRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListGroupUsersResponse{
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
			}),
			wantErr: nil,
		},
		{
			name: "should return error if policy service fails when WithRoles is true",
			setup: func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().ListByGroup(mock.Anything, someGroupID, "").Return(testUserList, nil)
				ps.EXPECT().ListRoles(mock.Anything, schema.UserPrincipal, "9f256f86-31a3-11ec-8d3d-0242ac130003", schema.GroupNamespace, someGroupID).Return(nil, errors.New("policy error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupUsersRequest{
				Id:        someGroupID,
				OrgId:     testOrgID,
				WithRoles: true,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success with roles when WithRoles is true",
			setup: func(gs *mocks.GroupService, us *mocks.UserService, os *mocks.OrganizationService, ps *mocks.PolicyService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().ListByGroup(mock.Anything, someGroupID, "").Return(testUserList, nil)

				testRoles := []role.Role{
					{
						ID:       "test-role-id",
						Name:     "admin",
						Title:    "Administrator",
						Metadata: metadata.Metadata{},
					},
				}
				ps.EXPECT().ListRoles(mock.Anything, schema.UserPrincipal, "9f256f86-31a3-11ec-8d3d-0242ac130003", schema.GroupNamespace, someGroupID).Return(testRoles, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListGroupUsersRequest{
				Id:        someGroupID,
				OrgId:     testOrgID,
				WithRoles: true,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListGroupUsersResponse{
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
				RolePairs: []*frontierv1beta1.ListGroupUsersResponse_RolePair{
					{
						UserId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Roles: []*frontierv1beta1.Role{
							{
								Id:        "test-role-id",
								Name:      "admin",
								Title:     "Administrator",
								Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
								CreatedAt: timestamppb.New(time.Time{}),
								UpdatedAt: timestamppb.New(time.Time{}),
							},
						},
					},
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockUserSvc := new(mocks.UserService)
			mockOrgSvc := new(mocks.OrganizationService)
			mockPolicySvc := new(mocks.PolicyService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockUserSvc, mockOrgSvc, mockPolicySvc)
			}
			h := ConnectHandler{
				groupService:  mockGroupSvc,
				userService:   mockUserSvc,
				orgService:    mockOrgSvc,
				policyService: mockPolicySvc,
			}
			got, err := h.ListGroupUsers(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestConnectHandler_AddGroupUsers(t *testing.T) {
	someGroupID := utils.NewString()
	someUserID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.AddGroupUsersRequest]
		want    *connect.Response[frontierv1beta1.AddGroupUsersResponse]
		wantErr error
	}{
		{
			name: "should return error if org does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.AddGroupUsersRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name: "should return error if org is disabled",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.AddGroupUsersRequest{
				Id:    someGroupID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return internal server error if error in adding group users",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().AddUsers(mock.Anything, someGroupID, []string{someUserID}).Return(errors.New("some error"))
			},
			request: connect.NewRequest(&frontierv1beta1.AddGroupUsersRequest{
				Id:      someGroupID,
				OrgId:   testOrgID,
				UserIds: []string{someUserID},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success if add group users and group service return nil error",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(testOrgMap[testOrgID], nil)
				gs.EXPECT().AddUsers(mock.Anything, someGroupID, []string{someUserID}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.AddGroupUsersRequest{
				Id:      someGroupID,
				OrgId:   testOrgID,
				UserIds: []string{someUserID},
			}),
			want:    connect.NewResponse(&frontierv1beta1.AddGroupUsersResponse{}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockOrgSvc)
			}
			h := ConnectHandler{
				groupService: mockGroupSvc,
				orgService:   mockOrgSvc,
			}
			got, err := h.AddGroupUsers(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestConnectHandler_RemoveGroupUser(t *testing.T) {
	randomID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService)
		request *connect.Request[frontierv1beta1.RemoveGroupUserRequest]
		want    *connect.Response[frontierv1beta1.RemoveGroupUserResponse]
		wantErr error
	}{
		{
			name: "should return error if organization does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, randomID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveGroupUserRequest{
				Id:     randomID,
				OrgId:  randomID,
				UserId: randomID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name: "should return error if organization is disabled",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, randomID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveGroupUserRequest{
				Id:     randomID,
				OrgId:  randomID,
				UserId: randomID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return error if user service fails when checking owners",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, randomID).Return(organization.Organization{ID: randomID}, nil)
				us.EXPECT().ListByGroup(mock.Anything, randomID, group.AdminRole).Return([]user.User{}, errors.New("user service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveGroupUserRequest{
				Id:     randomID,
				OrgId:  randomID,
				UserId: randomID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return error if user is the only admin",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, randomID).Return(organization.Organization{ID: randomID}, nil)
				us.EXPECT().ListByGroup(mock.Anything, randomID, group.AdminRole).Return([]user.User{{ID: randomID}}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveGroupUserRequest{
				Id:     randomID,
				OrgId:  randomID,
				UserId: randomID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrGroupMinOwnerCount),
		},
		{
			name: "should return error if group service fails to remove user",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, randomID).Return(organization.Organization{ID: randomID}, nil)
				us.EXPECT().ListByGroup(mock.Anything, randomID, group.AdminRole).Return([]user.User{{ID: "other-admin"}, {ID: randomID}}, nil)
				gs.EXPECT().RemoveUsers(mock.Anything, randomID, []string{randomID}).Return(errors.New("group service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveGroupUserRequest{
				Id:     randomID,
				OrgId:  randomID,
				UserId: randomID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should remove user successfully",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService, us *mocks.UserService) {
				os.EXPECT().Get(mock.Anything, randomID).Return(organization.Organization{ID: randomID}, nil)
				us.EXPECT().ListByGroup(mock.Anything, randomID, group.AdminRole).Return([]user.User{{ID: "other-admin"}, {ID: randomID}}, nil)
				gs.EXPECT().RemoveUsers(mock.Anything, randomID, []string{randomID}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveGroupUserRequest{
				Id:     randomID,
				OrgId:  randomID,
				UserId: randomID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.RemoveGroupUserResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockOrgSvc := new(mocks.OrganizationService)
			mockUserSvc := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockOrgSvc, mockUserSvc)
			}
			h := ConnectHandler{
				groupService: mockGroupSvc,
				orgService:   mockOrgSvc,
				userService:  mockUserSvc,
			}
			got, err := h.RemoveGroupUser(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestConnectHandler_EnableGroup(t *testing.T) {
	randomID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.EnableGroupRequest]
		want    *connect.Response[frontierv1beta1.EnableGroupResponse]
		wantErr error
	}{
		{
			name: "should return error if organization does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.EnableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name: "should return error if organization is disabled",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.EnableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return error if group does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{ID: testOrgID}, nil)
				gs.EXPECT().Enable(mock.Anything, randomID).Return(group.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.EnableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrGroupNotFound),
		},
		{
			name: "should enable group successfully",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{ID: testOrgID}, nil)
				gs.EXPECT().Enable(mock.Anything, randomID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.EnableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.EnableGroupResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockOrgSvc)
			}
			h := ConnectHandler{
				groupService: mockGroupSvc,
				orgService:   mockOrgSvc,
			}
			got, err := h.EnableGroup(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestConnectHandler_DisableGroup(t *testing.T) {
	randomID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.DisableGroupRequest]
		want    *connect.Response[frontierv1beta1.DisableGroupResponse]
		wantErr error
	}{
		{
			name: "should return error if organization does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DisableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name: "should return error if organization is disabled",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.DisableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return error if group does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{ID: testOrgID}, nil)
				gs.EXPECT().Disable(mock.Anything, randomID).Return(group.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DisableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrGroupNotFound),
		},
		{
			name: "should disable group successfully",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{ID: testOrgID}, nil)
				gs.EXPECT().Disable(mock.Anything, randomID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DisableGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DisableGroupResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockOrgSvc)
			}
			h := ConnectHandler{
				groupService: mockGroupSvc,
				orgService:   mockOrgSvc,
			}
			got, err := h.DisableGroup(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestConnectHandler_DeleteGroup(t *testing.T) {
	randomID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.DeleteGroupRequest]
		want    *connect.Response[frontierv1beta1.DeleteGroupResponse]
		wantErr error
	}{
		{
			name: "should return error if organization does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgNotFound),
		},
		{
			name: "should return error if organization is disabled",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return error if group does not exist",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{ID: testOrgID}, nil)
				gs.EXPECT().Delete(mock.Anything, randomID).Return(group.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrGroupNotFound),
		},
		{
			name: "should delete group successfully",
			setup: func(gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, testOrgID).Return(organization.Organization{ID: testOrgID}, nil)
				gs.EXPECT().Delete(mock.Anything, randomID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteGroupRequest{
				Id:    randomID,
				OrgId: testOrgID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteGroupResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockOrgSvc)
			}
			h := ConnectHandler{
				groupService: mockGroupSvc,
				orgService:   mockOrgSvc,
			}
			got, err := h.DeleteGroup(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.(*connect.Error).Code(), err.(*connect.Error).Code())
				assert.Equal(t, tt.wantErr.(*connect.Error).Message(), err.(*connect.Error).Message())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}
