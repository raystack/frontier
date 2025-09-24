package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

var (
	testUserID  = "9f256f86-31a3-11ec-8d3d-0242ac130003"
	testUserMap = map[string]user.User{
		"9f256f86-31a3-11ec-8d3d-0242ac130003": {
			ID:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
			Title: "User 1",
			Name:  "user1",
			Email: "test@test.com",
			Metadata: metadata.Metadata{
				"foo":    "bar",
				"age":    21,
				"intern": true,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
)

func TestConnectHandler_ListUsers(t *testing.T) {
	table := []struct {
		title string
		setup func(us *mocks.UserService)
		req   *connect.Request[frontierv1beta1.ListUsersRequest]
		want  *connect.Response[frontierv1beta1.ListUsersResponse]
		err   error
	}{
		{
			title: "should return internal error if user service return some error",
			setup: func(us *mocks.UserService) {
				us.EXPECT().List(mock.Anything, mock.Anything).Return([]user.User{}, errors.New("test error"))
			},
			req: connect.NewRequest(&frontierv1beta1.ListUsersRequest{
				PageSize: 50,
				PageNum:  1,
				Keyword:  "",
			}),
			want: nil,
			err:  connect.NewError(connect.CodeInternal, ErrInternalServerError),
		}, {
			title: "should return all users if user service return all users",
			setup: func(us *mocks.UserService) {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().List(mock.Anything, mock.Anything).Return(testUserList, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListUsersRequest{
				PageSize: 50,
				PageNum:  1,
				Keyword:  "",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListUsersResponse{
				Count: 1,
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
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockUserSrv)
			}
			mockDep := &ConnectHandler{userService: mockUserSrv}
			req := tt.req

			resp, err := mockDep.ListUsers(context.Background(), req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestConnectHandler_CreateUser(t *testing.T) {
	email := "user@raystack.org"
	_ = email
	table := []struct {
		title string
		setup func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context
		req   *connect.Request[frontierv1beta1.CreateUserRequest]
		want  *connect.Response[frontierv1beta1.CreateUserResponse]
		err   error
	}{
		{
			title: "should return bad request error if email is empty",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), user.User{
					Title: "some user",
				}).Return(user.User{}, user.ErrInvalidEmail)
				return authenticate.SetContextWithEmail(ctx, "")
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "some user",
				Email: "",
			}}),
			want: nil,
			err:  connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			title: "should return already exist error if user service return error conflict",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), user.User{
					Title:    "some user",
					Email:    "abc@test.com",
					Name:     "user-slug",
					Metadata: metadata.Metadata{},
				}).Return(user.User{}, user.ErrConflict)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title:    "some user",
				Email:    "abc@test.com",
				Name:     "user-slug",
				Metadata: &structpb.Struct{},
			}}),
			want: nil,
			err:  connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			title: "should return success if user email contain whitespace but still valid service return nil error",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), user.User{
					Title:    "some user",
					Email:    "abc@test.com",
					Name:     "user-slug",
					Metadata: metadata.Metadata{"foo": "bar"},
				}).Return(
					user.User{
						ID:       "new-abc",
						Title:    "some user",
						Email:    "abc@test.com",
						Name:     "user-slug",
						Metadata: metadata.Metadata{"foo": "bar"},
					}, nil)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "some user",
				Email: "  abc@test.com  ",
				Name:  "user-slug",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}}),
			want: connect.NewResponse(&frontierv1beta1.CreateUserResponse{User: &frontierv1beta1.User{
				Id:    "new-abc",
				Title: "some user",
				Email: "abc@test.com",
				Name:  "user-slug",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}}),
			err: nil,
		},
		{
			title: "should return success if user service return nil error",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), user.User{
					Title:    "some user",
					Email:    "abc@test.com",
					Name:     "user-slug",
					Metadata: metadata.Metadata{"foo": "bar"},
				}).Return(
					user.User{
						ID:       "new-abc",
						Title:    "some user",
						Email:    "abc@test.com",
						Name:     "user-slug",
						Metadata: metadata.Metadata{"foo": "bar"},
					}, nil)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "some user",
				Email: "abc@test.com",
				Name:  "user-slug",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}}),
			want: connect.NewResponse(&frontierv1beta1.CreateUserResponse{User: &frontierv1beta1.User{
				Id:    "new-abc",
				Title: "some user",
				Email: "abc@test.com",
				Name:  "user-slug",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}}),
			err: nil,
		},
	}
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			var resp *connect.Response[frontierv1beta1.CreateUserResponse]
			var err error
			ctx := context.Background()
			mockUserSrv := new(mocks.UserService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockUserSrv, mockMetaSrv)
			}
			mockDep := &ConnectHandler{userService: mockUserSrv, metaSchemaService: mockMetaSrv}
			resp, err = mockDep.CreateUser(ctx, tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestConnectHandler_GetUser(t *testing.T) {
	randomID := utils.NewString()
	table := []struct {
		title string
		req   *connect.Request[frontierv1beta1.GetUserRequest]
		setup func(us *mocks.UserService)
		want  *connect.Response[frontierv1beta1.GetUserResponse]
		err   error
	}{
		{
			title: "should return not found error if user does not exist",
			setup: func(us *mocks.UserService) {
				us.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), randomID).Return(user.User{}, user.ErrNotExist)
			},
			req: connect.NewRequest(&frontierv1beta1.GetUserRequest{
				Id: randomID,
			}),
			want: nil,
			err:  connect.NewError(connect.CodeNotFound, ErrUserNotExist),
		},
		{
			title: "should return not found error if user id is not uuid",
			setup: func(us *mocks.UserService) {
				us.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), "some-id").Return(user.User{}, user.ErrInvalidUUID)
			},
			req: connect.NewRequest(&frontierv1beta1.GetUserRequest{
				Id: "some-id",
			}),
			want: nil,
			err:  connect.NewError(connect.CodeNotFound, ErrUserNotExist),
		},
		{
			title: "should return not found error if user id is invalid",
			setup: func(us *mocks.UserService) {
				us.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), "").Return(user.User{}, user.ErrInvalidID)
			},
			req:  connect.NewRequest(&frontierv1beta1.GetUserRequest{}),
			want: nil,
			err:  connect.NewError(connect.CodeNotFound, ErrUserNotExist),
		},
		{
			title: "should return user if user service return nil error",
			setup: func(us *mocks.UserService) {
				us.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), randomID).Return(
					user.User{
						ID:    randomID,
						Title: "some user",
						Email: "someuser@test.com",
						Metadata: metadata.Metadata{
							"foo": "bar",
						},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.GetUserRequest{
				Id: randomID,
			}),
			want: connect.NewResponse(&frontierv1beta1.GetUserResponse{User: &frontierv1beta1.User{
				Id:    randomID,
				Title: "some user",
				Email: "someuser@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}}),
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockUserSrv)
			}
			mockDep := &ConnectHandler{userService: mockUserSrv}
			resp, err := mockDep.GetUser(context.Background(), tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestConnectHandler_UpdateUser(t *testing.T) {
	randomID := utils.NewString()
	table := []struct {
		title string
		setup func(us *mocks.UserService, ms *mocks.MetaSchemaService)
		req   *connect.Request[frontierv1beta1.UpdateUserRequest]
		want  *connect.Response[frontierv1beta1.UpdateUserResponse]
		err   error
	}{
		{
			title: "should return bad request error if id is empty",
			req: connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id: "",
				Body: &frontierv1beta1.UserRequestBody{
					Title: "some user",
					Email: "test@test.com",
				},
			}),
			want: nil,
			err:  connect.NewError(connect.CodeNotFound, ErrUserNotExist),
		},
		{
			title: "should return bad request error if body is nil",
			req: connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id:   randomID,
				Body: nil,
			}),
			want: nil,
			err:  connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			title: "should return bad request error if email is empty",
			req: connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id: randomID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "some user",
					Email: "",
				},
			}),
			want: nil,
			err:  connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			title: "should return not found error if user does not exist",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:       randomID,
					Title:    "some user",
					Email:    "test@test.com",
					Name:     "",
					Avatar:   "",
					Metadata: metadata.Metadata{"foo": "bar"},
				}).Return(user.User{}, user.ErrNotExist)
			},
			req: connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id: randomID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "some user",
					Email: "test@test.com",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			}),
			want: nil,
			err:  connect.NewError(connect.CodeNotFound, ErrUserNotExist),
		},
		{
			title: "should return success if user service returns updated user",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:       randomID,
					Title:    "updated user",
					Email:    "test@test.com",
					Name:     "updated-name",
					Avatar:   "new-avatar.jpg",
					Metadata: metadata.Metadata{"foo": "updated"},
				}).Return(
					user.User{
						ID:        randomID,
						Title:     "updated user",
						Email:     "test@test.com",
						Name:      "updated-name",
						Avatar:    "new-avatar.jpg",
						Metadata:  metadata.Metadata{"foo": "updated"},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.UpdateUserRequest{
				Id: randomID,
				Body: &frontierv1beta1.UserRequestBody{
					Title:  "updated user",
					Email:  "test@test.com",
					Name:   "updated-name",
					Avatar: "new-avatar.jpg",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("updated"),
						},
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateUserResponse{User: &frontierv1beta1.User{
				Id:     randomID,
				Title:  "updated user",
				Email:  "test@test.com",
				Name:   "updated-name",
				Avatar: "new-avatar.jpg",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("updated"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}}),
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockUserSrv, mockMetaSrv)
			}
			mockDep := &ConnectHandler{userService: mockUserSrv, metaSchemaService: mockMetaSrv}
			resp, err := mockDep.UpdateUser(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestConnectHandler_UpdateCurrentUser(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		title string
		setup func(us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService)
		req   *frontierv1beta1.UpdateCurrentUserRequest
		want  *frontierv1beta1.UpdateCurrentUserResponse
		err   connect.Code
	}{
		{
			title: "should return unauthenticated error if GetPrincipal fails",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
			},
			req: &frontierv1beta1.UpdateCurrentUserRequest{
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "abcuser123@test.com",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: nil,
			err:  connect.CodeUnauthenticated,
		},
		{
			title: "should return internal error if user service returns error",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{ID: userID}, nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.Anything, user.User{
					ID:    userID,
					Title: "abc user",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(user.User{}, errors.New("test error"))
			},
			req: &frontierv1beta1.UpdateCurrentUserRequest{
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: nil,
			err:  connect.CodeInternal,
		},
		{
			title: "should return bad request error if empty request body",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) {
				// No mocks needed - the function returns early when body is nil
			},
			req: &frontierv1beta1.UpdateCurrentUserRequest{
				Body: nil,
			},
			want: nil,
			err:  connect.CodeInvalidArgument,
		},
		{
			title: "should return success if user service returns nil error",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{ID: userID}, nil)
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.Anything, mock.Anything).Return(
					user.User{
						ID:    "user-id-1",
						Title: "abc user",
						Email: "user@raystack.org",
						Metadata: metadata.Metadata{
							"foo": "bar",
						},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					}, nil)
			},
			req: &frontierv1beta1.UpdateCurrentUserRequest{
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: &frontierv1beta1.UpdateCurrentUserResponse{
				User: &frontierv1beta1.User{
					Id:    "user-id-1",
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			err: connect.Code(0), // Success case - no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			mockAuthSrv := new(mocks.AuthnService)

			if tt.setup != nil {
				tt.setup(mockUserSrv, mockMetaSrv, mockAuthSrv)
			}

			handler := &ConnectHandler{
				userService:       mockUserSrv,
				metaSchemaService: mockMetaSrv,
				authnService:      mockAuthSrv,
			}

			req := connect.NewRequest(tt.req)
			resp, err := handler.UpdateCurrentUser(context.Background(), req)

			if tt.err == connect.Code(0) {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, resp.Msg)
			} else {
				assert.Nil(t, resp)
				assert.Equal(t, tt.err, connect.CodeOf(err))
			}

			mockUserSrv.AssertExpectations(t)
			mockMetaSrv.AssertExpectations(t)
			mockAuthSrv.AssertExpectations(t)
		})
	}
}