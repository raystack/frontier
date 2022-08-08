package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/metadata"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var testUserMap = map[string]user.User{
	"9f256f86-31a3-11ec-8d3d-0242ac130003": {
		ID:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
		Name:  "User 1",
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

func TestListUsers(t *testing.T) {
	t.Parallel()

	table := []struct {
		title       string
		mockUserSrv mockUserSrv
		req         *shieldv1beta1.ListUsersRequest
		want        *shieldv1beta1.ListUsersResponse
		err         error
	}{
		{
			title: "error in User Service",
			mockUserSrv: mockUserSrv{ListFunc: func(ctx context.Context, flt user.Filter) (users user.PagedUsers, err error) {
				return user.PagedUsers{}, errors.New("some error")
			}},
			req: &shieldv1beta1.ListUsersRequest{
				PageSize: 50,
				PageNum:  1,
				Keyword:  "",
			},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		}, {
			title: "success",
			mockUserSrv: mockUserSrv{ListFunc: func(ctx context.Context, flt user.Filter) (users user.PagedUsers, err error) {
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				return user.PagedUsers{
					Users: testUserList,
					Count: int32(len(testUserList)),
				}, nil
			}},
			req: &shieldv1beta1.ListUsersRequest{
				PageSize: 50,
				PageNum:  1,
				Keyword:  "",
			},
			want: &shieldv1beta1.ListUsersResponse{
				Count: 1,
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
				}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()

			mockDep := Handler{userService: tt.mockUserSrv}
			req := tt.req
			resp, err := mockDep.ListUsers(context.Background(), req)
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
		header      string
		req         *shieldv1beta1.CreateUserRequest
		want        *shieldv1beta1.CreateUserResponse
		err         error
	}{
		{
			title: "error in fetching user list",
			mockUserSrv: mockUserSrv{CreateFunc: func(ctx context.Context, u user.User) (user.User, error) {
				return user.User{}, emptyEmailId
			}},
			req: &shieldv1beta1.CreateUserRequest{Body: &shieldv1beta1.UserRequestBody{
				Name:     "some user",
				Email:    "abc@test.com",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  emptyEmailId,
		},
		{
			title: "int values in metadata map",
			req: &shieldv1beta1.CreateUserRequest{Body: &shieldv1beta1.UserRequestBody{
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
			mockUserSrv: mockUserSrv{CreateFunc: func(ctx context.Context, u user.User) (user.User, error) {
				return user.User{
					ID:       "new-abc",
					Name:     "some user",
					Email:    "abc@test.com",
					Metadata: nil,
				}, nil
			}},
			header: "abc@test.com",
			req: &shieldv1beta1.CreateUserRequest{Body: &shieldv1beta1.UserRequestBody{
				Name:  "some user",
				Email: "abc@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			want: &shieldv1beta1.CreateUserResponse{User: &shieldv1beta1.User{
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

			var resp *shieldv1beta1.CreateUserResponse
			var err error
			if tt.title == "success" {
				mockDep := Handler{userService: tt.mockUserSrv}
				ctx := user.SetContextWithEmail(context.Background(), tt.header)
				resp, err = mockDep.CreateUser(ctx, tt.req)
			} else {
				mockDep := Handler{userService: tt.mockUserSrv}
				resp, err = mockDep.CreateUser(context.Background(), tt.req)
			}

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
		want        *shieldv1beta1.GetCurrentUserResponse
		err         error
	}{
		{
			title: "error in User Service",
			mockUserSrv: mockUserSrv{GetByEmailFunc: func(ctx context.Context, email string) (usr user.User, err error) {
				return user.User{}, errors.New("some error")
			}},
			header: "email-temp",
			want:   nil,
			err:    grpcInternalServerError,
		},
		{
			title: "success",
			mockUserSrv: mockUserSrv{GetByEmailFunc: func(ctx context.Context, email string) (usr user.User, err error) {
				return user.User{
					ID:    "user-id-1",
					Name:  "some user",
					Email: "someuser@test.com",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, nil
			}},
			header: "someuser@test.com",
			want: &shieldv1beta1.GetCurrentUserResponse{User: &shieldv1beta1.User{
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

			mockDep := Handler{userService: tt.mockUserSrv}
			ctx := user.SetContextWithEmail(context.Background(), tt.header)

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
		req         *shieldv1beta1.UpdateCurrentUserRequest
		header      string
		want        *shieldv1beta1.UpdateCurrentUserResponse
		err         error
	}{
		{
			title: "error in User Service",
			mockUserSrv: mockUserSrv{UpdateByEmailFunc: func(ctx context.Context, toUpdate user.User) (usr user.User, err error) {
				return user.User{}, errors.New("some error")
			}},
			req: &shieldv1beta1.UpdateCurrentUserRequest{Body: &shieldv1beta1.UserRequestBody{
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
			mockUserSrv: mockUserSrv{UpdateByEmailFunc: func(ctx context.Context, toUpdate user.User) (usr user.User, err error) {
				return user.User{}, nil
			}},
			req: &shieldv1beta1.UpdateCurrentUserRequest{Body: &shieldv1beta1.UserRequestBody{
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
			mockUserSrv: mockUserSrv{UpdateByEmailFunc: func(ctx context.Context, toUpdate user.User) (usr user.User, err error) {
				return user.User{}, nil
			}},
			req:    &shieldv1beta1.UpdateCurrentUserRequest{Body: nil},
			header: "abcuser@test.com",
			want:   nil,
			err:    grpcBadBodyError,
		},
		{
			title: "success",
			mockUserSrv: mockUserSrv{UpdateByEmailFunc: func(ctx context.Context, toUpdate user.User) (usr user.User, err error) {
				return user.User{
					ID:    "user-id-1",
					Name:  "abc user",
					Email: "abcuser@test.com",
					Metadata: metadata.Metadata{
						"foo": "bar",
					},
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, nil
			}},
			req: &shieldv1beta1.UpdateCurrentUserRequest{Body: &shieldv1beta1.UserRequestBody{
				Name:  "abc user",
				Email: "abcuser@test.com",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": structpb.NewStringValue("bar"),
					},
				},
			}},
			header: "abcuser@test.com",
			want: &shieldv1beta1.UpdateCurrentUserResponse{User: &shieldv1beta1.User{
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

			mockDep := Handler{userService: tt.mockUserSrv}
			ctx := user.SetContextWithEmail(context.Background(), tt.header)

			resp, err := mockDep.UpdateCurrentUser(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

type mockUserSrv struct {
	GetByIDFunc       func(ctx context.Context, id string) (user.User, error)
	GetByEmailFunc    func(ctx context.Context, email string) (user.User, error)
	CreateFunc        func(ctx context.Context, usr user.User) (user.User, error)
	ListFunc          func(ctx context.Context, flt user.Filter) (user.PagedUsers, error)
	UpdateByIDFunc    func(ctx context.Context, toUpdate user.User) (user.User, error)
	UpdateByEmailFunc func(ctx context.Context, toUpdate user.User) (user.User, error)
	FetchCurrentFunc  func(ctx context.Context) (user.User, error)
}

func (m mockUserSrv) GetByID(ctx context.Context, id string) (user.User, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m mockUserSrv) GetByEmail(ctx context.Context, email string) (user.User, error) {
	return m.GetByEmailFunc(ctx, email)
}

func (m mockUserSrv) Create(ctx context.Context, usr user.User) (user.User, error) {
	return m.CreateFunc(ctx, usr)
}

func (m mockUserSrv) List(ctx context.Context, flt user.Filter) (user.PagedUsers, error) {
	return m.ListFunc(ctx, flt)
}

func (m mockUserSrv) UpdateByID(ctx context.Context, toUpdate user.User) (user.User, error) {
	return m.UpdateByIDFunc(ctx, toUpdate)
}

func (m mockUserSrv) UpdateByEmail(ctx context.Context, toUpdate user.User) (user.User, error) {
	return m.UpdateByEmailFunc(ctx, toUpdate)
}

func (m mockUserSrv) FetchCurrentUser(ctx context.Context) (user.User, error) {
	return m.FetchCurrentFunc(ctx)
}
