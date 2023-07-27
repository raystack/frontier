package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
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
		request *frontierv1beta1.ListGroupsRequest
		want    *frontierv1beta1.ListGroupsResponse
		wantErr error
	}{
		{
			name: "should return empty groups if query param org_id is not uuid",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), group.Filter{
					OrganizationID: "some-id",
				}).Return([]group.Group{}, nil)
			},
			request: &frontierv1beta1.ListGroupsRequest{
				OrgId: "some-id",
			},
			want: &frontierv1beta1.ListGroupsResponse{
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
			request: &frontierv1beta1.ListGroupsRequest{
				OrgId: randomID,
			},
			want: &frontierv1beta1.ListGroupsResponse{
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
			request: &frontierv1beta1.ListGroupsRequest{},
			want: &frontierv1beta1.ListGroupsResponse{
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
			request: &frontierv1beta1.ListGroupsRequest{
				OrgId: "9f256f86-31a3-11ec-8d3d-0242ac130003",
			},
			want: &frontierv1beta1.ListGroupsResponse{
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
	email := "user@raystack.org"
	someOrgID := utils.NewString()
	someGroupID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context
		request *frontierv1beta1.CreateGroupRequest
		want    *frontierv1beta1.CreateGroupResponse
		wantErr error
	}{
		{
			name: "should return unauthenticated error if auth email in context is empty and group service return invalid user email",
			setup: func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, user.ErrInvalidEmail)

				return ctx
			},
			request: &frontierv1beta1.CreateGroupRequest{
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Metadata: &structpb.Struct{},
				}},
			want:    nil,
			wantErr: grpcUnauthenticated,
		},
		{
			name: "should return internal error if group service return some error",
			setup: func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name:           "some-group",
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, errors.New("some error"))
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateGroupRequest{
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				}},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return already exist error if group service return error conflict",
			setup: func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some-group",

					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrConflict)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateGroupRequest{
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				}},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return bad request error if name empty",
			setup: func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name:           "some-group",
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateGroupRequest{
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				}},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if org id is not uuid",
			setup: func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name:           "some-group",
					OrganizationID: "some-org-id",
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrInvalidUUID)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateGroupRequest{
				OrgId: "some-org-id",
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				}},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if org id not exist",
			setup: func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name: "some-group",

					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrNotExist)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateGroupRequest{
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				}},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name:    "should return bad request error if body is empty",
			request: &frontierv1beta1.CreateGroupRequest{Body: nil},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if group service return nil",
			setup: func(ctx context.Context, gs *mocks.GroupService, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), group.Group{
					Name:           "some-group",
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{
					ID:             someGroupID,
					Name:           "some-group",
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}, nil)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: &frontierv1beta1.CreateGroupRequest{
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name:     "some-group",
					Metadata: &structpb.Struct{},
				}},
			want: &frontierv1beta1.CreateGroupResponse{
				Group: &frontierv1beta1.Group{
					Id:    someGroupID,
					Name:  "some-group",
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
			mockUserSvc := new(mocks.UserService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockGroupSvc, mockUserSvc, mockMetaSchemaSvc)
			}
			h := Handler{
				userService:       mockUserSvc,
				groupService:      mockGroupSvc,
				metaSchemaService: mockMetaSchemaSvc,
			}
			got, err := h.CreateGroup(ctx, tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func TestHandler_GetGroup(t *testing.T) {
	someGroupID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *frontierv1beta1.GetGroupRequest
		want    *frontierv1beta1.GetGroupResponse
		wantErr error
	}{
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), someGroupID).Return(group.Group{}, errors.New("some error"))
			},
			request: &frontierv1beta1.GetGroupRequest{Id: someGroupID},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is invalid",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(group.Group{}, group.ErrInvalidID)
			},
			request: &frontierv1beta1.GetGroupRequest{},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found error if group not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(group.Group{}, group.ErrNotExist)
			},
			request: &frontierv1beta1.GetGroupRequest{},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return success if group service return nil",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testGroupID).Return(testGroupMap[testGroupID], nil)
			},
			request: &frontierv1beta1.GetGroupRequest{Id: testGroupID},
			want: &frontierv1beta1.GetGroupResponse{
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
	someGroupID := utils.NewString()
	someOrgID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService, ms *mocks.MetaSchemaService)
		request *frontierv1beta1.UpdateGroupRequest
		want    *frontierv1beta1.UpdateGroupResponse
		wantErr error
	}{

		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, errors.New("some error"))
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if body is empty",
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:   someGroupID,
				Body: nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return not found error if group id is not uuid (slug) and does not exist",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             "some-id",
					Name:           "some-id",
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrNotExist)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    "some-id",
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "some-id",
				},
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found error if group id is uuid and does not exist",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrNotExist)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return not found error if group id is empty",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					Name:           "new-group",
					OrganizationID: someOrgID,
					Metadata:       metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidID)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want:    nil,
			wantErr: grpcGroupNotFoundErr,
		},
		{
			name: "should return already exist error if group service return error conflict",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrConflict)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return bad request error if org id does not exist",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if org id is not uuid",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, organization.ErrInvalidUUID)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if name is empty",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if slug is empty",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:   someGroupID,
					Name: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{}, group.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id: someGroupID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: someOrgID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if updated by id and group service return nil error",
			setup: func(gs *mocks.GroupService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), groupMetaSchema).Return(nil)
				gs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}).Return(group.Group{
					ID:             someGroupID,
					Name:           "new-group",
					OrganizationID: someOrgID,

					Metadata: metadata.Metadata{},
				}, nil)
			},
			request: &frontierv1beta1.UpdateGroupRequest{
				Id:    someGroupID,
				OrgId: someOrgID,
				Body: &frontierv1beta1.GroupRequestBody{
					Name: "new-group",
				},
			},
			want: &frontierv1beta1.UpdateGroupResponse{
				Group: &frontierv1beta1.Group{
					Id:    someGroupID,
					Name:  "new-group",
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
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockGroupSvc, mockMetaSchemaSvc)
			}
			h := Handler{
				groupService:      mockGroupSvc,
				metaSchemaService: mockMetaSchemaSvc,
			}
			got, err := h.UpdateGroup(context.Background(), tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}
