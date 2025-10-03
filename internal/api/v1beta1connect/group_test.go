package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
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
			wantErrMsg:  ErrNotFound,
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
