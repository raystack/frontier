package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/model"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1"
)

var testUserMap = map[string]model.User{
	"9f256f86-31a3-11ec-8d3d-0242ac130003": {
		Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
		Name:  "User 1",
		Email: "test@test.com",
		Metadata: map[string]string{
			"foo": "bar",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	table := []struct {
		title       string
		mockUserSrv mockUserSrv
		want        *shieldv1.ListUsersResponse
		err         error
	}{
		{
			title: "error in User Service",
			mockUserSrv: mockUserSrv{ListUsersFunc: func(ctx context.Context) (users []model.User, err error) {
				return []model.User{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		}, {
			title: "success",
			mockUserSrv: mockUserSrv{ListUsersFunc: func(ctx context.Context) (users []model.User, err error) {
				var testUserList []model.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				return testUserList, nil
			}},
			want: &shieldv1.ListUsersResponse{Users: []*shieldv1.User{
				{
					Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
					Name:  "User 1",
					Email: "test@test.com",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Dep{UserService: tt.mockUserSrv}
			resp, err := mockDep.ListUsers(context.Background(), nil)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	table := []struct {
		title       string
		mockUserSrv mockUserSrv
		req         *shieldv1.CreateUserRequest
		want        *shieldv1.CreateUserResponse
		err         error
	}{
		{
			title: "error in fetching user list",
			mockUserSrv: mockUserSrv{CreateUserFunc: func(ctx context.Context, u model.User) (model.User, error) {
				return model.User{}, grpcInternalServerError
			}},
			req: &shieldv1.CreateUserRequest{Body: &shieldv1.UserRequestBody{
				Name:     "some user",
				Email:    "abc@test.com",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "int values in metadata map",
			req: &shieldv1.CreateUserRequest{Body: &shieldv1.UserRequestBody{
				Name:  "some user",
				Email: "abc@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewNumberValue(10),
					},
				},
			}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "success",
			mockUserSrv: mockUserSrv{CreateUserFunc: func(ctx context.Context, u model.User) (model.User, error) {
				return model.User{
					Id:       "new-abc",
					Name:     "some user",
					Email:    "abc@test.com",
					Metadata: nil,
				}, nil
			}},
			req: &shieldv1.CreateUserRequest{Body: &shieldv1.UserRequestBody{
				Name:  "some user",
				Email: "abc@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			want: &shieldv1.CreateUserResponse{User: &shieldv1.User{
				Id:        "new-abc",
				Name:      "some user",
				Email:     "abc@test.com",
				Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Dep{UserService: tt.mockUserSrv}
			resp, err := mockDep.CreateUser(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	table := []struct {
		title       string
		mockUserSrv mockUserSrv
		header      string
		want        *shieldv1.GetCurrentUserResponse
		err         error
	}{
		{
			title: "error in User Service",
			mockUserSrv: mockUserSrv{GetCurrentUserFunc: func(ctx context.Context, email string) (user model.User, err error) {
				return model.User{}, errors.New("some error")
			}},
			header: "email-temp",
			want:   nil,
			err:    grpcInternalServerError,
		},
		{
			title: "success",
			mockUserSrv: mockUserSrv{GetCurrentUserFunc: func(ctx context.Context, email string) (user model.User, err error) {
				return model.User{
					Id:    "user-id-1",
					Name:  "some user",
					Email: "someuser@test.com",
					Metadata: map[string]string{
						"foo": "bar",
					},
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, nil
			}},
			header: "someuser@test.com",
			want: &shieldv1.GetCurrentUserResponse{User: &shieldv1.User{
				Id:    "user-id-1",
				Name:  "some user",
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
			t.Parallel()

			mockDep := Dep{UserService: tt.mockUserSrv, IdentityProxyHeader: "x-auth-email"}
			md := metadata.Pairs(mockDep.IdentityProxyHeader, tt.header)
			ctx := metadata.NewIncomingContext(context.Background(), md)

			resp, err := mockDep.GetCurrentUser(ctx, nil)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestUpdateCurrentUser(t *testing.T) {
	t.Parallel()

	table := []struct {
		title       string
		mockUserSrv mockUserSrv
		req         *shieldv1.UpdateCurrentUserRequest
		header      string
		want        *shieldv1.UpdateCurrentUserResponse
		err         error
	}{
		{
			title: "error in User Service",
			mockUserSrv: mockUserSrv{UpdateCurrentUserFunc: func(ctx context.Context, toUpdate model.User) (user model.User, err error) {
				return model.User{}, errors.New("some error")
			}},
			req: &shieldv1.UpdateCurrentUserRequest{Body: &shieldv1.UserRequestBody{
				Name:  "abc user",
				Email: "abcuser@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			header: "abcuser@test.com",
			want:   nil,
			err:    grpcInternalServerError,
		},
		{
			title: "diff emails in header and body",
			mockUserSrv: mockUserSrv{UpdateCurrentUserFunc: func(ctx context.Context, toUpdate model.User) (user model.User, err error) {
				return model.User{}, nil
			}},
			req: &shieldv1.UpdateCurrentUserRequest{Body: &shieldv1.UserRequestBody{
				Name:  "abc user",
				Email: "abcuser123@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			header: "abcuser@test.com",
			want:   nil,
			err:    grpcBadBodyError,
		},
		{
			title: "empty request body",
			mockUserSrv: mockUserSrv{UpdateCurrentUserFunc: func(ctx context.Context, toUpdate model.User) (user model.User, err error) {
				return model.User{}, nil
			}},
			req:    &shieldv1.UpdateCurrentUserRequest{Body: nil},
			header: "abcuser@test.com",
			want:   nil,
			err:    grpcBadBodyError,
		},
		{
			title: "success",
			mockUserSrv: mockUserSrv{UpdateCurrentUserFunc: func(ctx context.Context, toUpdate model.User) (user model.User, err error) {
				return model.User{
					Id:    "user-id-1",
					Name:  "abc user",
					Email: "abcuser@test.com",
					Metadata: map[string]string{
						"foo": "bar",
					},
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, nil
			}},
			req: &shieldv1.UpdateCurrentUserRequest{Body: &shieldv1.UserRequestBody{
				Name:  "abc user",
				Email: "abcuser@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			header: "abcuser@test.com",
			want: &shieldv1.UpdateCurrentUserResponse{User: &shieldv1.User{
				Id:    "user-id-1",
				Name:  "abc user",
				Email: "abcuser@test.com",
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
			t.Parallel()

			mockDep := Dep{UserService: tt.mockUserSrv, IdentityProxyHeader: "x-auth-email"}
			md := metadata.Pairs(mockDep.IdentityProxyHeader, tt.header)
			ctx := metadata.NewIncomingContext(context.Background(), md)

			resp, err := mockDep.UpdateCurrentUser(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

type mockUserSrv struct {
	GetUserFunc           func(ctx context.Context, id string) (model.User, error)
	GetCurrentUserFunc    func(ctx context.Context, email string) (model.User, error)
	CreateUserFunc        func(ctx context.Context, user model.User) (model.User, error)
	ListUsersFunc         func(ctx context.Context) ([]model.User, error)
	UpdateUserFunc        func(ctx context.Context, toUpdate model.User) (model.User, error)
	UpdateCurrentUserFunc func(ctx context.Context, toUpdate model.User) (model.User, error)
}

func (m mockUserSrv) GetUser(ctx context.Context, id string) (model.User, error) {
	return m.GetUserFunc(ctx, id)
}

func (m mockUserSrv) GetCurrentUser(ctx context.Context, email string) (model.User, error) {
	return m.GetCurrentUserFunc(ctx, email)
}

func (m mockUserSrv) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	return m.CreateUserFunc(ctx, user)
}

func (m mockUserSrv) ListUsers(ctx context.Context) ([]model.User, error) {
	return m.ListUsersFunc(ctx)
}

func (m mockUserSrv) UpdateUser(ctx context.Context, toUpdate model.User) (model.User, error) {
	return m.UpdateUserFunc(ctx, toUpdate)
}

func (m mockUserSrv) UpdateCurrentUser(ctx context.Context, toUpdate model.User) (model.User, error) {
	return m.UpdateCurrentUserFunc(ctx, toUpdate)
}
