package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/internal/user"
	"github.com/stretchr/testify/assert"
	shieldv1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testUserMap = map[string]user.User{
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
			mockUserSrv: mockUserSrv{ListUsersFunc: func(ctx context.Context) (users []user.User, err error) {
				return []user.User{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		}, {
			title: "success",
			mockUserSrv: mockUserSrv{ListUsersFunc: func(ctx context.Context) (users []user.User, err error) {
				var testUserList []user.User
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
			mockUserSrv: mockUserSrv{CreateUserFunc: func(ctx context.Context, u user.User) (user.User, error) {
				return user.User{}, errors.New("some error")
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
			mockUserSrv: mockUserSrv{CreateUserFunc: func(ctx context.Context, u user.User) (user.User, error) {
				return user.User{
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

type mockUserSrv struct {
	GetUserFunc    func(ctx context.Context, id string) (user.User, error)
	CreateUserFunc func(ctx context.Context, user user.User) (user.User, error)
	ListUsersFunc  func(ctx context.Context) ([]user.User, error)
	UpdateUserFunc func(ctx context.Context, toUpdate user.User) (user.User, error)
}

func (m mockUserSrv) GetUser(ctx context.Context, id string) (user.User, error) {
	return m.GetUserFunc(ctx, id)
}

func (m mockUserSrv) CreateUser(ctx context.Context, user user.User) (user.User, error) {
	return m.CreateUserFunc(ctx, user)
}

func (m mockUserSrv) ListUsers(ctx context.Context) ([]user.User, error) {
	return m.ListUsersFunc(ctx)
}

func (m mockUserSrv) UpdateUser(ctx context.Context, toUpdate user.User) (user.User, error) {
	return m.UpdateUserFunc(ctx, toUpdate)
}
