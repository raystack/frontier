package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/odpf/shield/internal/schema"
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
	testGroupID  = "9f256f86-31a3-11ec-8d3d-0242ac130003"
	testGroupMap = map[string]group.Group{
		"9f256f86-31a3-11ec-8d3d-0242ac130003": {
			ID:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
			Name: "Group 1",
			Slug: "group-1",
			Metadata: metadata.Metadata{
				"foo": "bar",
			},
			OrganizationID: "9f256f86-31a3-11ec-8d3d-0242ac130003",
			CreatedAt:      time.Time{},
			UpdatedAt:      time.Time{},
		},
	}

	testUsersRole1 = []user.User{
		{
			ID: "user-id-1",
		},
		{
			ID: "user-id-2",
		},
	}

	testUserIDRoleMapRole1 = map[string][]string{
		"user-id-1": {"group:role-1"},
		"user-id-2": {"group:role-1"},
	}

	testUsersRole2 = []user.User{
		{
			ID: "user-id-3",
		},
	}

	testUserIDRoleMapRole2 = map[string][]string{
		"user-id-3": {"group:role-2"},
	}

	testGroupsRole2 = []group.Group{
		{
			ID: "group-id-1",
		},
	}

	testGroupIDRoleMapRole2 = map[string][]string{
		"group-id-1": {"group:role-2"},
	}

	testUsersAnyRole = []user.User{
		{
			ID: "user-id-1",
		},
		{
			ID: "user-id-2",
		},
		{
			ID: "user-id-3",
		},
	}

	testUserIDRoleMapAnyRole = map[string][]string{
		"user-id-1": {"group:role-1"},
		"user-id-2": {"group:role-1"},
		"user-id-3": {"group:role-2"},
	}

	testGroupsAnyRole = []group.Group{
		{
			ID: "group-id-1",
		},
	}

	testGroupIDRoleMapAnyRole = map[string][]string{
		"group-id-1": {"group:role-2"},
	}
)

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
						Id:    testGroupID,
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
						Id:    testGroupID,
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
			name: "should return unauthenticated error if auth email in context is empty and group service return invalid user email",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",

					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, user.ErrInvalidEmail)
				return ctx
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
				OrgId:    someOrgID,
				Metadata: &structpb.Struct{},
			}},
			want:    nil,
			wantErr: grpcUnauthenticated,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(ctx context.Context, gs *mocks.GroupService) context.Context {
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some group",
					Slug: "some-group",

					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, errors.New("some error"))
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
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
					Name:           "some group",
					Slug:           "some-group",
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

					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{
					ID:   someGroupID,
					Name: "some group",
					Slug: "some-group",

					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}, nil)
				return user.SetContextWithEmail(ctx, email)
			},
			request: &shieldv1beta1.CreateGroupRequest{Body: &shieldv1beta1.GroupRequestBody{
				Name:     "some group",
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
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testGroupID).Return(testGroupMap[testGroupID], nil)
			},
			request: &shieldv1beta1.GetGroupRequest{Id: testGroupID},
			want: &shieldv1beta1.GetGroupResponse{
				Group: &shieldv1beta1.Group{
					Id:    testGroupID,
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
			name: "should return not found error if group id is empty",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					Name:           "new group",
					Slug:           "", // consider it by slug and make the slug empty
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
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
		{
			name: "should return success if updated by id and group service return nil error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}, nil)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group",
					OrgId: someOrgID,
				},
			},
			want: &shieldv1beta1.UpdateGroupResponse{
				Group: &shieldv1beta1.Group{
					Id:    someGroupID,
					Name:  "new group",
					Slug:  "new-group",
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
		{
			name: "should return success if updated by slug and group service return nil error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					Name:           "new group",
					Slug:           "some-slug",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{
					ID:             someGroupID,
					Name:           "new group",
					Slug:           "some-slug",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}, nil)
			},
			request: &shieldv1beta1.UpdateGroupRequest{
				Id: "some-slug",
				Body: &shieldv1beta1.GroupRequestBody{
					Name:  "new group",
					Slug:  "new-group", // will be ignored
					OrgId: someOrgID,
				},
			},
			want: &shieldv1beta1.UpdateGroupResponse{
				Group: &shieldv1beta1.Group{
					Id:    someGroupID,
					Name:  "new group",
					Slug:  "some-slug",
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

func TestHandler_ListGroupRelations(t *testing.T) {
	transformedUser1, _ := transformUserToPB(user.User{
		ID: testUsersRole1[0].ID,
	})

	transformedUser2, _ := transformUserToPB(user.User{
		ID: testUsersRole1[1].ID,
	})

	transformedUser3, _ := transformUserToPB(user.User{
		ID: testUsersRole2[0].ID,
	})

	transformedGroup1, _ := transformGroupToPB(group.Group{
		ID: testGroupsRole2[0].ID,
	})

	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *shieldv1beta1.ListGroupRelationsRequest
		want    *shieldv1beta1.ListGroupRelationsResponse
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListGroupRelations(mock.AnythingOfType("*context.emptyCtx"), "group-id", "", "").Return([]user.User{}, []group.Group{}, map[string][]string{}, map[string][]string{}, errors.New("some error"))
			},
			request: &shieldv1beta1.ListGroupRelationsRequest{
				Id:          "group-id",
				SubjectType: "",
				Role:        "",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return relations of subject_type-user and role-role1 if relation service return nil error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListGroupRelations(mock.AnythingOfType("*context.emptyCtx"), "group-id", schema.UserPrincipal, "role-1").Return(testUsersRole1, []group.Group{}, testUserIDRoleMapRole1, map[string][]string{}, nil)
			},
			request: &shieldv1beta1.ListGroupRelationsRequest{
				Id:          "group-id",
				SubjectType: schema.UserPrincipal,
				Role:        "role-1",
			},
			want: &shieldv1beta1.ListGroupRelationsResponse{
				Relations: []*shieldv1beta1.GroupRelation{
					{
						SubjectType: schema.UserPrincipal,
						Role:        "role-1",
						Subject: &shieldv1beta1.GroupRelation_User{
							User: &transformedUser1,
						},
					},
					{
						SubjectType: schema.UserPrincipal,
						Role:        "role-1",
						Subject: &shieldv1beta1.GroupRelation_User{
							User: &transformedUser2,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should return relations of subject_type-user and role-role2 if relation service return nil error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListGroupRelations(mock.AnythingOfType("*context.emptyCtx"), "group-id", schema.UserPrincipal, "role-2").Return(testUsersRole2, []group.Group{}, testUserIDRoleMapRole2, map[string][]string{}, nil)
			},
			request: &shieldv1beta1.ListGroupRelationsRequest{
				Id:          "group-id",
				SubjectType: schema.UserPrincipal,
				Role:        "role-2",
			},
			want: &shieldv1beta1.ListGroupRelationsResponse{
				Relations: []*shieldv1beta1.GroupRelation{
					{
						SubjectType: schema.UserPrincipal,
						Role:        "role-2",
						Subject: &shieldv1beta1.GroupRelation_User{
							User: &transformedUser3,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should return relations of subject_type-group and role-role2 if relation service return nil error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListGroupRelations(mock.AnythingOfType("*context.emptyCtx"), "group-id", schema.GroupPrincipal, "role-2").Return([]user.User{}, testGroupsRole2, map[string][]string{}, testGroupIDRoleMapRole2, nil)
			},
			request: &shieldv1beta1.ListGroupRelationsRequest{
				Id:          "group-id",
				SubjectType: schema.GroupPrincipal,
				Role:        "role-2",
			},
			want: &shieldv1beta1.ListGroupRelationsResponse{
				Relations: []*shieldv1beta1.GroupRelation{
					{
						SubjectType: schema.GroupPrincipal,
						Role:        "role-2",
						Subject: &shieldv1beta1.GroupRelation_Group{
							Group: &transformedGroup1,
						},
					},
				},
			},
			wantErr: nil,
		},

		{
			name: "should return relations both group and users for role-role2 if relation service return nil error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListGroupRelations(mock.AnythingOfType("*context.emptyCtx"), "group-id", "", "role-2").Return(testUsersRole2, testGroupsRole2, testUserIDRoleMapRole2, testGroupIDRoleMapRole2, nil)
			},
			request: &shieldv1beta1.ListGroupRelationsRequest{
				Id:          "group-id",
				SubjectType: "",
				Role:        "role-2",
			},
			want: &shieldv1beta1.ListGroupRelationsResponse{
				Relations: []*shieldv1beta1.GroupRelation{
					{
						SubjectType: schema.UserPrincipal,
						Role:        "role-2",
						Subject: &shieldv1beta1.GroupRelation_User{
							User: &transformedUser3,
						},
					},
					{
						SubjectType: schema.GroupPrincipal,
						Role:        "role-2",
						Subject: &shieldv1beta1.GroupRelation_Group{
							Group: &transformedGroup1,
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should return relations both group and users with all roles if relation service return nil error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListGroupRelations(mock.AnythingOfType("*context.emptyCtx"), "group-id", "", "").Return(testUsersAnyRole, testGroupsAnyRole, testUserIDRoleMapAnyRole, testGroupIDRoleMapAnyRole, nil)
			},
			request: &shieldv1beta1.ListGroupRelationsRequest{
				Id:          "group-id",
				SubjectType: "",
				Role:        "",
			},
			want: &shieldv1beta1.ListGroupRelationsResponse{
				Relations: []*shieldv1beta1.GroupRelation{
					{
						SubjectType: schema.UserPrincipal,
						Role:        "role-1",
						Subject: &shieldv1beta1.GroupRelation_User{
							User: &transformedUser1,
						},
					},
					{
						SubjectType: schema.UserPrincipal,
						Role:        "role-1",
						Subject: &shieldv1beta1.GroupRelation_User{
							User: &transformedUser2,
						},
					},
					{
						SubjectType: schema.UserPrincipal,
						Role:        "role-2",
						Subject: &shieldv1beta1.GroupRelation_User{
							User: &transformedUser3,
						},
					},
					{
						SubjectType: schema.GroupPrincipal,
						Role:        "role-2",
						Subject: &shieldv1beta1.GroupRelation_Group{
							Group: &transformedGroup1,
						},
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupSrv := new(mocks.GroupService)
			if tt.setup != nil {
				tt.setup(mockGroupSrv)
			}
			mockDep := Handler{groupService: mockGroupSrv}
			resp, err := mockDep.ListGroupRelations(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
