package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/core/resource"

	"github.com/google/uuid"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/core/authenticate/token"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/pkg/errors"

	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/metadata"
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

func TestListUsers(t *testing.T) {
	table := []struct {
		title string
		setup func(us *mocks.UserService)
		req   *frontierv1beta1.ListUsersRequest
		want  *frontierv1beta1.ListUsersResponse
		err   error
	}{
		{
			title: "should return internal error in if user service return some error",
			setup: func(us *mocks.UserService) {
				us.EXPECT().List(mock.Anything, mock.Anything).Return([]user.User{}, errors.New("test error"))
			},
			req: &frontierv1beta1.ListUsersRequest{
				PageSize: 50,
				PageNum:  1,
				Keyword:  "",
			},
			want: nil,
			err:  errors.New("test error"),
		}, {
			title: "should return all users if user service return all users",
			setup: func(us *mocks.UserService) {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().List(mock.Anything, mock.Anything).Return(testUserList, nil)
			},
			req: &frontierv1beta1.ListUsersRequest{
				PageSize: 50,
				PageNum:  1,
				Keyword:  "",
			},
			want: &frontierv1beta1.ListUsersResponse{
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
			},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockUserSrv)
			}
			mockDep := Handler{userService: mockUserSrv}
			req := tt.req
			resp, err := mockDep.ListUsers(context.Background(), req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestCreateUser(t *testing.T) {
	email := "user@raystack.org"
	_ = email
	table := []struct {
		title string
		setup func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.CreateUserRequest
		want  *frontierv1beta1.CreateUserResponse
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
			req: &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "some user",
				Email: "",
			}},
			want: nil,
			err:  grpcBadBodyError,
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
			req: &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title:    "some user",
				Email:    "abc@test.com",
				Name:     "user-slug",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcConflictError,
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
			req: &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "some user",
				Email: "  abc@test.com  ",
				Name:  "user-slug",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			want: &frontierv1beta1.CreateUserResponse{User: &frontierv1beta1.User{
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
			}},
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
			req: &frontierv1beta1.CreateUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "some user",
				Email: "abc@test.com",
				Name:  "user-slug",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			want: &frontierv1beta1.CreateUserResponse{User: &frontierv1beta1.User{
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
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			var resp *frontierv1beta1.CreateUserResponse
			var err error

			ctx := context.Background()
			mockUserSrv := new(mocks.UserService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockUserSrv, mockMetaSrv)
			}
			mockDep := Handler{userService: mockUserSrv, metaSchemaService: mockMetaSrv}
			resp, err = mockDep.CreateUser(ctx, tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestGetUser(t *testing.T) {
	randomID := utils.NewString()
	table := []struct {
		title string
		req   *frontierv1beta1.GetUserRequest
		setup func(us *mocks.UserService)
		want  *frontierv1beta1.GetUserResponse
		err   error
	}{
		{
			title: "should return not found error if user does not exist",
			setup: func(us *mocks.UserService) {
				us.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), randomID).Return(user.User{}, user.ErrNotExist)
			},
			req: &frontierv1beta1.GetUserRequest{
				Id: randomID,
			},
			want: nil,
			err:  grpcUserNotFoundError,
		},
		{
			title: "should return not found error if user id is not uuid",
			setup: func(us *mocks.UserService) {
				us.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), "some-id").Return(user.User{}, user.ErrInvalidUUID)
			},
			req: &frontierv1beta1.GetUserRequest{
				Id: "some-id",
			},
			want: nil,
			err:  grpcUserNotFoundError,
		},
		{
			title: "should return not found error if user id is invalid",
			setup: func(us *mocks.UserService) {
				us.EXPECT().GetByID(mock.AnythingOfType("context.backgroundCtx"), "").Return(user.User{}, user.ErrInvalidID)
			},
			req:  &frontierv1beta1.GetUserRequest{},
			want: nil,
			err:  grpcUserNotFoundError,
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
			req: &frontierv1beta1.GetUserRequest{
				Id: randomID,
			},
			want: &frontierv1beta1.GetUserResponse{User: &frontierv1beta1.User{
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
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockUserSrv)
			}
			mockDep := Handler{userService: mockUserSrv}
			resp, err := mockDep.GetUser(context.Background(), tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	email := "user@raystack.org"
	table := []struct {
		title  string
		setup  func(ctx context.Context, us *mocks.AuthnService, ss *mocks.SessionService) context.Context
		header string
		want   *frontierv1beta1.GetCurrentUserResponse
		err    error
	}{
		{
			title: "should return unauthenticated error if no auth email header in context",
			want:  nil,
			err:   grpcUnauthenticated,
			setup: func(ctx context.Context, us *mocks.AuthnService, ss *mocks.SessionService) context.Context {
				us.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
				return ctx
			},
		},
		{
			title: "should return not found error if user does not exist",
			setup: func(ctx context.Context, us *mocks.AuthnService, ss *mocks.SessionService) context.Context {
				us.EXPECT().GetPrincipal(mock.AnythingOfType("*context.valueCtx")).Return(authenticate.Principal{}, user.ErrNotExist)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			want: nil,
			err:  grpcUserNotFoundError,
		},
		{
			title: "should return error if user service return some error",
			setup: func(ctx context.Context, us *mocks.AuthnService, ss *mocks.SessionService) context.Context {
				us.EXPECT().GetPrincipal(mock.AnythingOfType("*context.valueCtx")).Return(authenticate.Principal{}, errors.New("test error"))
				return authenticate.SetContextWithEmail(ctx, email)
			},
			want: nil,
			err:  errors.New("test error"),
		},
		{
			title: "should return user if user service return nil error",
			setup: func(ctx context.Context, us *mocks.AuthnService, ss *mocks.SessionService) context.Context {
				us.EXPECT().GetPrincipal(mock.AnythingOfType("*context.valueCtx")).Return(
					authenticate.Principal{
						ID: "user-id-1",
						User: &user.User{
							ID:    "user-id-1",
							Title: "some user",
							Email: "someuser@test.com",
							Metadata: metadata.Metadata{
								"foo": "bar",
							},
							CreatedAt: time.Time{},
							UpdatedAt: time.Time{},
						},
					}, nil)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			want: &frontierv1beta1.GetCurrentUserResponse{User: &frontierv1beta1.User{
				Id:    "user-id-1",
				Title: "some user",
				Email: "someuser@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			ctx := context.Background()
			mockAuthnSrv := new(mocks.AuthnService)

			mockOrgService := new(mocks.OrganizationService)
			mockOrgService.EXPECT().ListByUser(mock.Anything, "user-id-1", organization.Filter{}).Return([]organization.Organization{}, nil)
			mockAuthnSrv.EXPECT().BuildToken(mock.Anything,
				authenticate.Principal{
					ID:   "user-id-1",
					Type: schema.UserPrincipal,
				}, map[string]string{"orgs": ""}).Return(nil, token.ErrMissingRSADisableToken)

			if tt.setup != nil {
				ctx = tt.setup(ctx, mockAuthnSrv, nil)
			}
			mockDep := Handler{
				authnService: mockAuthnSrv,
				orgService:   mockOrgService,
			}
			resp, err := mockDep.GetCurrentUser(ctx, nil)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	someID := utils.NewString()
	table := []struct {
		title  string
		setup  func(us *mocks.UserService, ms *mocks.MetaSchemaService)
		req    *frontierv1beta1.UpdateUserRequest
		header string
		want   *frontierv1beta1.UpdateUserResponse
		err    error
	}{
		{
			title: "should return internal error if user service return some error",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:    someID,
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(user.User{}, errors.New("test error"))
			},
			req: &frontierv1beta1.UpdateUserRequest{
				Id: someID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				}},
			want: nil,
			err:  errors.New("test error"),
		},
		{
			title: "should return not found error if id is invalid",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(user.User{}, user.ErrInvalidID)
			},
			req: &frontierv1beta1.UpdateUserRequest{
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				}},
			want: nil,
			err:  grpcUserNotFoundError,
		},
		{
			title: "should return error if user meta schema service return error",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(grpcBadBodyMetaSchemaError)
			},
			req: &frontierv1beta1.UpdateUserRequest{
				Id: someID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				}},
			want: nil,
			err:  grpcBadBodyMetaSchemaError,
		},
		{
			title: "should return already exist error if user service return error conflict",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:    someID,
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(user.User{}, user.ErrConflict)
			},
			req: &frontierv1beta1.UpdateUserRequest{
				Id: someID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				}},
			want: nil,
			err:  grpcConflictError,
		},
		{
			title: "should return bad request error if email in request empty",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:    someID,
					Title: "abc user",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(user.User{}, user.ErrInvalidEmail)
			},
			req: &frontierv1beta1.UpdateUserRequest{
				Id: someID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return bad request error if empty request body",
			req:   &frontierv1beta1.UpdateUserRequest{Id: someID, Body: nil},
			want:  nil,
			err:   grpcBadBodyError,
		},
		{
			title: "should return success if user service return nil error",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:    someID,
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(
					user.User{
						ID:    someID,
						Title: "abc user",
						Email: "user@raystack.org",
						Metadata: metadata.Metadata{
							"foo": "bar",
						},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					}, nil)
			},
			req: &frontierv1beta1.UpdateUserRequest{
				Id: someID,
				Body: &frontierv1beta1.UserRequestBody{
					Title: "abc user",
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				}},
			want: &frontierv1beta1.UpdateUserResponse{User: &frontierv1beta1.User{
				Id:    someID,
				Title: "abc user",
				Email: "user@raystack.org",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
		{
			title: "should return success even though name is empty",
			setup: func(us *mocks.UserService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:    someID,
					Email: "user@raystack.org",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(
					user.User{
						ID:    someID,
						Email: "user@raystack.org",
						Metadata: metadata.Metadata{
							"foo": "bar",
						},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					}, nil)
			},
			req: &frontierv1beta1.UpdateUserRequest{
				Id: someID,
				Body: &frontierv1beta1.UserRequestBody{
					Email: "user@raystack.org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				}},
			want: &frontierv1beta1.UpdateUserResponse{User: &frontierv1beta1.User{
				Id:    someID,
				Email: "user@raystack.org",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockUserSrv, mockMetaSrv)
			}
			mockDep := Handler{userService: mockUserSrv, metaSchemaService: mockMetaSrv}
			resp, err := mockDep.UpdateUser(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestUpdateCurrentUser(t *testing.T) {
	userID := uuid.New().String()
	table := []struct {
		title  string
		setup  func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) context.Context
		req    *frontierv1beta1.UpdateCurrentUserRequest
		header string
		want   *frontierv1beta1.UpdateCurrentUserResponse
		err    error
	}{
		{
			title: "should return unauthenticated error if auth email header not exist",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) context.Context {
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
				return ctx
			},
			req: &frontierv1beta1.UpdateCurrentUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "abc user",
				Email: "abcuser123@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			want: nil,
			err:  grpcUnauthenticated,
		},
		{
			title: "should return internal error if user service return some error",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				us.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), user.User{
					ID:    userID,
					Title: "abc user",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
				}).Return(user.User{}, errors.New("test error"))
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: userID}, nil)
				return ctx
			},
			req: &frontierv1beta1.UpdateCurrentUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "abc user",
				Email: "user@raystack.org",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			want: nil,
			err:  errors.New("test error"),
		},
		{
			title: "should return bad request error if empty request body",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: userID}, nil)
				return ctx
			},
			req:  &frontierv1beta1.UpdateCurrentUserRequest{Body: nil},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return success if user service return nil error",
			setup: func(ctx context.Context, us *mocks.UserService, ms *mocks.MetaSchemaService, as *mocks.AuthnService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), userMetaSchema).Return(nil)
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: userID}, nil)
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
				return ctx
			},
			req: &frontierv1beta1.UpdateCurrentUserRequest{Body: &frontierv1beta1.UserRequestBody{
				Title: "abc user",
				Email: "user@raystack.org",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			want: &frontierv1beta1.UpdateCurrentUserResponse{User: &frontierv1beta1.User{
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
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			mockAuthSrv := new(mocks.AuthnService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockUserSrv, mockMetaSrv, mockAuthSrv)
			}
			mockDep := Handler{userService: mockUserSrv, metaSchemaService: mockMetaSrv, authnService: mockAuthSrv}
			resp, err := mockDep.UpdateCurrentUser(ctx, tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_ListUserGroups(t *testing.T) {
	someUserID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(gs *mocks.GroupService)
		request *frontierv1beta1.ListUserGroupsRequest
		want    *frontierv1beta1.ListUserGroupsResponse
		wantErr error
	}{
		{
			name: "should return internal error if group service return some error",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), someUserID, schema.UserPrincipal,
					group.Filter{}).Return([]group.Group{}, errors.New("test error"))
			},
			request: &frontierv1beta1.ListUserGroupsRequest{
				Id: someUserID,
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return empty list if user does not exist",
			setup: func(gs *mocks.GroupService) {
				gs.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), someUserID, schema.UserPrincipal,
					group.Filter{}).Return([]group.Group{}, nil)
			},
			request: &frontierv1beta1.ListUserGroupsRequest{
				Id: someUserID,
			},
			want:    &frontierv1beta1.ListUserGroupsResponse{},
			wantErr: nil,
		},
		// if user id empty, it would not go to this handler
		{
			name: "should return groups if group service return not nil",
			setup: func(gs *mocks.GroupService) {
				var testGroupList []group.Group
				for _, g := range testGroupMap {
					testGroupList = append(testGroupList, g)
				}
				gs.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), someUserID, schema.UserPrincipal,
					group.Filter{}).Return(testGroupList, nil)
			},
			request: &frontierv1beta1.ListUserGroupsRequest{
				Id: someUserID,
			},
			want: &frontierv1beta1.ListUserGroupsResponse{
				Groups: []*frontierv1beta1.Group{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
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
			got, err := h.ListUserGroups(context.Background(), tt.request)
			assert.EqualValues(t, got, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}
func TestHandler_ListAllUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(gs *mocks.UserService)
		request *frontierv1beta1.ListAllUsersRequest
		want    *frontierv1beta1.ListAllUsersResponse
		wantErr error
	}{
		{
			name: "should return internal error in if user service return some error",
			setup: func(us *mocks.UserService) {
				us.EXPECT().List(mock.Anything, mock.Anything).Return([]user.User{}, errors.New("test error"))
			},
			request: &frontierv1beta1.ListAllUsersRequest{
				PageSize: 50,
				PageNum:  1,
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return all users if user service return success",
			setup: func(us *mocks.UserService) {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().List(mock.Anything, mock.Anything).Return(testUserList, nil)
			},
			request: &frontierv1beta1.ListAllUsersRequest{
				PageSize: 50,
				PageNum:  1,
				Keyword:  "some_keyword",
				OrgId:    "some_id",
				GroupId:  "some_group_id",
				State:    "some_state",
			},
			want: &frontierv1beta1.ListAllUsersResponse{
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
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserSrv := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockUserSrv)
			}
			mockDep := Handler{userService: mockUserSrv}
			req := tt.request
			resp, err := mockDep.ListAllUsers(context.Background(), req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.wantErr)
		})
	}
}

func Test_ListCurrentUserGroups(t *testing.T) {
	md, _ := structpb.NewStruct(map[string]interface{}{})
	tests := []struct {
		name    string
		setup   func(g *mocks.GroupService, a *mocks.AuthnService, r *mocks.ResourceService)
		request *frontierv1beta1.ListCurrentUserGroupsRequest
		want    *frontierv1beta1.ListCurrentUserGroupsResponse
		wantErr error
	}{
		{
			name: "should list current user groups on success",
			setup: func(g *mocks.GroupService, a *mocks.AuthnService, r *mocks.ResourceService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{
					ID:   "some_id",
					Type: schema.UserPrincipal,
				}, nil)
				g.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), "some_id", schema.UserPrincipal,
					group.Filter{}).
					Return([]group.Group{
						{
							ID:             "some_id",
							Name:           "some_name",
							Title:          "some_title",
							OrganizationID: "some_org_id",
						},
					}, nil)
				r.EXPECT().BatchCheck(mock.AnythingOfType("context.backgroundCtx"), []resource.Check{}).Return(nil, nil)
			},
			request: &frontierv1beta1.ListCurrentUserGroupsRequest{},
			want: &frontierv1beta1.ListCurrentUserGroupsResponse{
				Groups: []*frontierv1beta1.Group{
					{
						Id:        "some_id",
						Name:      "some_name",
						Title:     "some_title",
						OrgId:     "some_org_id",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
						Metadata:  md,
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGrpSrv := new(mocks.GroupService)
			authServ := new(mocks.AuthnService)
			resourceServ := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockGrpSrv, authServ, resourceServ)
			}
			mockDep := Handler{
				groupService:    mockGrpSrv,
				authnService:    authServ,
				resourceService: resourceServ,
			}
			req := tt.request
			resp, err := mockDep.ListCurrentUserGroups(context.Background(), req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
