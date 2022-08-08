package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/organization"
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

var testGroupMap = map[string]group.Group{
	"9f256f86-31a3-11ec-8d3d-0242ac130003": {
		ID:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
		Name: "Group 1",
		Slug: "group-1",
		Organization: organization.Organization{
			ID:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
			Name: "organization 1",
			Slug: "org-1",
		},
		Metadata: metadata.Metadata{
			"foo": "bar",
		},
		OrganizationID: "9f256f86-31a3-11ec-8d3d-0242ac130003",
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
	},
}

func TestHandler_ListGroups(t *testing.T) {
	randomID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *shieldv1beta1.ListGroupsRequest
		want    *shieldv1beta1.ListGroupsResponse
		wantErr error
	}{
		{
			name: "should return empty groups if query param org_id is not uuid",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), group.Filter{
					OrganizationID: "some-id",
				}).Return([]group.Group{}, nil)
			},
			request: &shieldv1beta1.ListGroupsRequest{
				OrgId: "some-id",
			},
			want: &shieldv1beta1.ListGroupsResponse{
				Groups: nil,
			},
			wantErr: nil,
		},
		{
			name: "should return empty groups if query param org_id is not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), group.Filter{
					OrganizationID: randomID,
				}).Return([]group.Group{}, nil)
			},
			request: &shieldv1beta1.ListGroupsRequest{
				OrgId: randomID,
			},
			want: &shieldv1beta1.ListGroupsResponse{
				Groups: nil,
			},
			wantErr: nil,
		},
		{
			name: "should return all groups if no query param filter exist",
			setup: func(gs *mocks.GroupService) {
				var testGroupList []group.Group
				for _, u := range testGroupMap {
					testGroupList = append(testGroupList, u)
				}
				gs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), group.Filter{}).Return(testGroupList, nil)
			},
			request: &shieldv1beta1.ListGroupsRequest{},
			want: &shieldv1beta1.ListGroupsResponse{
				Groups: []*shieldv1beta1.Group{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Name:  "Group 1",
						Slug:  "group-1",
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
			},
			wantErr: nil,
		},
		{
			name: "should return filtered groups if query param org_id exist",
			setup: func(gs *mocks.GroupService) {
				var testGroupList []group.Group
				for _, u := range testGroupMap {
					testGroupList = append(testGroupList, u)
				}
				gs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), group.Filter{
					OrganizationID: "9f256f86-31a3-11ec-8d3d-0242ac130003",
				}).Return(testGroupList, nil)
			},
			request: &shieldv1beta1.ListGroupsRequest{
				OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
			},
			want: &shieldv1beta1.ListGroupsResponse{
				Groups: []*shieldv1beta1.Group{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Name:  "Group 1",
						Slug:  "group-1",
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
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.ListGroups(context.Background(), tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_CreateGroup(t *testing.T) {
	email := "user@odpf.io"
	someOrgID := uuid.NewString()
	someGroupID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(ctx context.Context, gs *mocks.GroupService) context.Context
		request *shieldv1beta1.CreateGroupRequest
		want    *shieldv1beta1.CreateGroupResponse
		wantErr error
	}{
		{
			name: "should return forbidden error if no auth email header in context and group service return error invalid user email",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",
					Organization: organization.Organization{
						ID: someOrgID,
					},
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, user.ErrInvalidEmail)
				return ctx
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
				Slug:     "some-group",
				OrgId:    someOrgID,
				Metadata: &structpb.Struct{},
			}},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",
					Organization: organization.Organization{
						ID: someOrgID,
					},
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
				Slug:     "some-group",
				OrgId:    someOrgID,
				Metadata: &structpb.Struct{},
			}},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return already exist error if group service return error conflict",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",
					Organization: organization.Organization{
						ID: someOrgID,
					},
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrConflict)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
				Slug:     "some-group",
				OrgId:    someOrgID,
				Metadata: &structpb.Struct{},
			}},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return bad request error if name empty",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Slug: "some-group",
					Organization: organization.Organization{
						ID: someOrgID,
					},
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Slug:     "some-group",
				OrgId:    someOrgID,
				Metadata: &structpb.Struct{},
			}},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if org id is not uuid",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",
					Organization: organization.Organization{
						ID: "some-org-id",
					},
					OrganizationID: "some-org-id",
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
				Slug:     "some-group",
				OrgId:    "some-org-id",
				Metadata: &structpb.Struct{},
			}},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if org id not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",
					Organization: organization.Organization{
						ID: someOrgID,
					},
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
				Slug:     "some-group",
				OrgId:    someOrgID,
				Metadata: &structpb.Struct{},
			}},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name:    "should return bad request error if body is empty",
			request: &shieldv1beta1.CreateGroupRequest{Body: nil},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if group service return nil",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",
					Organization: organization.Organization{
						ID: someOrgID,
					},
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{
					ID:   someGroupID,
					Name: "some group",
					Slug: "some-group",
					Organization: organization.Organization{
						ID: someOrgID,
					},
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
				Slug:     "some-group",
				OrgId:    someOrgID,
				Metadata: &structpb.Struct{},
			}},
			want: &shieldv1beta1.CreateGroupResponse{
				Group: &shieldv1beta1.Group{
					Id:    someGroupID,
					Name:  "some group",
					Slug:  "some-group",
					OrgId: someOrgID,
					Metadata: &structpb.Struct{
						Fields: make(map[string]*structpb.Value),
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
			mockGroupSvc := new(mocks.GroupService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.CreateGroup(ctx, tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_GetGroup(t *testing.T) {
	someGroupID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *shieldv1beta1.GetGroupRequest
		want    *shieldv1beta1.GetGroupResponse
		wantErr error
	}{
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return(group.Group{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetGroupRequest{Id: someGroupID},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is invalid",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(group.Group{}, group.ErrInvalidID)
			},
			request: &shieldv1beta1.GetGroupRequest{},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found error if group not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(group.Group{}, group.ErrNotExist)
			},
			request: &shieldv1beta1.GetGroupRequest{},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return success if group service return nil",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "9f256f86-31a3-11ec-8d3d-0242ac130003").Return(testGroupMap["9f256f86-31a3-11ec-8d3d-0242ac130003"], nil)
			},
			request: &shieldv1beta1.GetGroupRequest{Id: "9f256f86-31a3-11ec-8d3d-0242ac130003"},
			want: &shieldv1beta1.GetGroupResponse{
				Group: &shieldv1beta1.Group{
					Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
					Name:  "Group 1",
					Slug:  "group-1",
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
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.GetGroup(context.Background(), tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_AddGroupUser(t *testing.T) {
	email := "user@odpf.io"
	someGroupID := uuid.NewString()
	someUserIDs := []string{
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
	}
	var testUserList []user.User
	var testUserIDs []string
	for _, u := range testUserMap {
		testUserList = append(testUserList, u)
		testUserIDs = append(testUserIDs, u.ID)
	}
	tests := []struct {
		name    string
		setup   func(ctx context.Context, gs *mocks.GroupService) context.Context
		request *shieldv1beta1.AddGroupUserRequest
		want    *shieldv1beta1.AddGroupUserResponse
		wantErr error
	}{
		{
			name: "should return forbidden error if group service return invalid user email",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddUsers(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, user.ErrInvalidEmail)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return forbidden error if caller is not an admin",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddUsers(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, errors.Unauthorized)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddUsers(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if group id not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddUsers(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, group.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return bad request error if one of user ids is not uuid",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddUsers(mock.AnythingOfType("*context.valueCtx"), someGroupID, append([]string{"some-id"}, someUserIDs...)).Return([]user.User{}, user.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: append([]string{"some-id"}, someUserIDs...),
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if user ids is empty",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddUsers(mock.AnythingOfType("*context.valueCtx"), someGroupID, []string{}).Return([]user.User{}, user.ErrInvalidID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: []string{},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if body is empty",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id:   someGroupID,
				Body: nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if group service return nil",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddUsers(mock.AnythingOfType("*context.valueCtx"), someGroupID, testUserIDs).Return(testUserList, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupUserRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupUserRequestBody{
					UserIds: testUserIDs,
				},
			},
			want: &shieldv1beta1.AddGroupUserResponse{
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
			mockGroupSvc := new(mocks.GroupService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.AddGroupUser(ctx, tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_RemoveGroupUser(t *testing.T) {
	email := "user@odpf.io"
	someGroupID := uuid.NewString()
	someUserID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(ctx context.Context, gs *mocks.GroupService) context.Context
		request *shieldv1beta1.RemoveGroupUserRequest
		want    *shieldv1beta1.RemoveGroupUserResponse
		wantErr error
	}{
		{
			name: "should return forbidden error if group service return invalid user email",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, user.ErrInvalidEmail)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return forbidden error if caller is not an admin",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, errors.Unauthorized)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if group does not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, group.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found user error if user id is not uuid",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, "some-id").Return([]user.User{}, user.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id:     someGroupID,
				UserId: "some-id",
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return not found user error if user id is uuid but not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, user.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return not found user error if user id is empty",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, "").Return([]user.User{}, user.ErrInvalidID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id: someGroupID,
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return success if group service return nil error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				gs.EXPECT().RemoveUser(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return(testUserList, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupUserRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want: &shieldv1beta1.RemoveGroupUserResponse{
				Message: "Removed User from group",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.RemoveGroupUser(ctx, tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_UpdateGroup(t *testing.T) {
	someGroupID := uuid.NewString()
	someOrgID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *shieldv1beta1.UpdateGroupRequest
		want    *shieldv1beta1.UpdateGroupResponse
		wantErr error
	}{

		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "new-group",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if body is empty",
			request: &shieldv1beta1.UpdateGroupRequest{
				Id:   someGroupID,
				Body: nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return not found error if group id is not uuid (slug) and does not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					Name:           "new group",
					Slug:           "some-id",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: "some-id",
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found error if group id is uuid and does not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "new-group",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return already exist error if group service return error conflict",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "new-group",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrConflict)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return bad request error if org id does not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "new-group",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if org id is not uuid",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "new-group",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrInvalidUUID)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if name is empty",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Slug:           "new-group",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if slug is empty",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new group",
					OrganizationID: someOrgID,
					Organization: organization.Organization{
						ID: someOrgID,
					},
					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					OrgId: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.UpdateGroup(context.Background(), tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_ListGroupAdmins(t *testing.T) {
	someGroupID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *shieldv1beta1.ListGroupAdminsRequest
		want    *shieldv1beta1.ListGroupAdminsResponse
		wantErr error
	}{
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return([]user.User{}, errors.New("some error"))
			},
			request: &shieldv1beta1.ListGroupAdminsRequest{
				Id: someGroupID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if group id is not uuid (slug) and group not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), "group-slug").Return([]user.User{}, group.ErrNotExist)
			},
			request: &shieldv1beta1.ListGroupAdminsRequest{Id: "group-slug"},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found error if group id is invalid or empty",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), "").Return([]user.User{}, group.ErrNotExist)
			},
			request: &shieldv1beta1.ListGroupAdminsRequest{},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return empty list if group id is uuid and group not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return([]user.User{}, nil)
			},
			request: &shieldv1beta1.ListGroupAdminsRequest{Id: someGroupID},
			want: &shieldv1beta1.ListGroupAdminsResponse{
				Users: nil,
			},
			wantErr: nil,
		},
		{
			name: "should return list of admins if group id is uuid and group exists",
			setup: func(gs *mocks.GroupService) {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				gs.EXPECT().ListAdmins(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return(testUserList, nil)
			},
			request: &shieldv1beta1.ListGroupAdminsRequest{Id: someGroupID},
			want: &shieldv1beta1.ListGroupAdminsResponse{
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
			mockGroupSvc := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.ListGroupAdmins(context.Background(), tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_AddGroupAdmin(t *testing.T) {
	email := "user@odpf.io"
	someGroupID := uuid.NewString()
	someUserIDs := []string{
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
	}
	var testUserList []user.User
	var testUserIDs []string
	for _, u := range testUserMap {
		testUserList = append(testUserList, u)
		testUserIDs = append(testUserIDs, u.ID)
	}
	tests := []struct {
		name    string
		setup   func(ctx context.Context, gs *mocks.GroupService) context.Context
		request *shieldv1beta1.AddGroupAdminRequest
		want    *shieldv1beta1.AddGroupAdminResponse
		wantErr error
	}{

		{
			name: "should return forbidden error if group service return invalid user email",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, user.ErrInvalidEmail)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return forbidden error if caller is not an admin",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, errors.Unauthorized)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return not found error if group id not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserIDs).Return([]user.User{}, group.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupAdminRequestBody{
					UserIds: someUserIDs,
				},
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return bad request error if one of user ids is not uuid",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), someGroupID, append([]string{"some-id"}, someUserIDs...)).Return([]user.User{}, user.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupAdminRequestBody{
					UserIds: append([]string{"some-id"}, someUserIDs...),
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if user ids is empty",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), someGroupID, []string{}).Return([]user.User{}, user.ErrInvalidID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupAdminRequestBody{
					UserIds: []string{},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if body is empty",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id:   someGroupID,
				Body: nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if group service return nil",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().AddAdmins(mock.AnythingOfType("*context.valueCtx"), someGroupID, testUserIDs).Return(testUserList, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.AddGroupAdminRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.AddGroupAdminRequestBody{
					UserIds: testUserIDs,
				},
			},
			want: &shieldv1beta1.AddGroupAdminResponse{
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
			mockGroupSvc := new(mocks.GroupService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.AddGroupAdmin(ctx, tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_RemoveGroupAdmin(t *testing.T) {
	email := "user@odpf.io"
	someGroupID := uuid.NewString()
	someUserID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(ctx context.Context, gs *mocks.GroupService) context.Context
		request *shieldv1beta1.RemoveGroupAdminRequest
		want    *shieldv1beta1.RemoveGroupAdminResponse
		wantErr error
	}{
		{
			name: "should return forbidden error if group service return invalid user email",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, user.ErrInvalidEmail)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return forbidden error if caller is not an admin",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, errors.Unauthorized)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcPermissionDenied,
		},
		{
			name: "should return not found error if group does not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, group.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found user error if user id is not uuid",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, "some-id").Return([]user.User{}, user.ErrInvalidUUID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id:     someGroupID,
				UserId: "some-id",
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return not found user error if user id is uuid but not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return([]user.User{}, user.ErrNotExist)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return not found user error if user id is empty",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, "").Return([]user.User{}, user.ErrInvalidID)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id: someGroupID,
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return success if group service return nil error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				gs.EXPECT().RemoveAdmin(mock.AnythingOfType("*context.valueCtx"), someGroupID, someUserID).Return(testUserList, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.RemoveGroupAdminRequest{
				Id:     someGroupID,
				UserId: someUserID,
			},
			want: &shieldv1beta1.RemoveGroupAdminResponse{
				Message: "Removed Admin from group",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSvc := new(mocks.GroupService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.RemoveGroupAdmin(ctx, tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_ListGroupUsers(t *testing.T) {
	someGroupID := uuid.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *shieldv1beta1.ListGroupUsersRequest
		want    *shieldv1beta1.ListGroupUsersResponse
		wantErr error
	}{
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListUsers(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return([]user.User{}, errors.New("some error"))
			},
			request: &shieldv1beta1.ListGroupUsersRequest{Id: someGroupID},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if group id is not uuid (slug) nd group not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListUsers(mock.AnythingOfType("*context.emptyCtx"), "group-slug").Return([]user.User{}, group.ErrNotExist)
			},
			request: &shieldv1beta1.ListGroupUsersRequest{Id: "group-slug"},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found error if group id is invalid or empty",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListUsers(mock.AnythingOfType("*context.emptyCtx"), "").Return([]user.User{}, group.ErrNotExist)
			},
			request: &shieldv1beta1.ListGroupUsersRequest{},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return empty list if group id is uuid and group not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListUsers(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return([]user.User{}, nil)
			},
			request: &shieldv1beta1.ListGroupUsersRequest{Id: someGroupID},
			want: &shieldv1beta1.ListGroupUsersResponse{
				Users: nil,
			},
			wantErr: nil,
		},
		{
			name: "should return list of users if group id is uuid and group exists",
			setup: func(gs *mocks.GroupService) {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				gs.EXPECT().ListUsers(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return(testUserList, nil)
			},
			request: &shieldv1beta1.ListGroupUsersRequest{Id: someGroupID},
			want: &shieldv1beta1.ListGroupUsersResponse{
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
			mockGroupSvc := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc)
			}
			h := Handler{
				groupService: mockGroupSvc,
			}
			got, err := h.ListGroupUsers(context.Background(), tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}
