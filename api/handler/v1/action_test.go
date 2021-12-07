package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/model"
	shieldv1 "github.com/odpf/shield/proto/odpf/shield/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testActionMap = map[string]model.Action{
	"read": {
		Id:   "read",
		Name: "Read",
		Namespace: model.Namespace{
			Id:        "resource-1",
			Name:      "Resource 1",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"write": {
		Id:   "write",
		Name: "Write",
		Namespace: model.Namespace{
			Id:        "resource-1",
			Name:      "Resource 1",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"manage": {
		Id:   "manage",
		Name: "Manage",
		Namespace: model.Namespace{
			Id:        "resource-1",
			Name:      "Resource 1",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestListActions(t *testing.T) {
	t.Parallel()
	table := []struct {
		title         string
		mockActionSrc mockActionSrv
		req           *shieldv1.ListActionsRequest
		want          *shieldv1.ListActionsResponse
		err           error
	}{
		{
			title: "error in Action Service",
			mockActionSrc: mockActionSrv{ListActionsFunc: func(ctx context.Context) (actions []model.Action, err error) {
				return []model.Action{}, errors.New("some error")
			}},
			want: nil,
			err:  status.Errorf(codes.Internal, internalServerError.Error()),
		},
		{
			title: "success",
			mockActionSrc: mockActionSrv{ListActionsFunc: func(ctx context.Context) (actions []model.Action, err error) {
				var testActionList []model.Action
				for _, act := range testActionMap {
					testActionList = append(testActionList, act)
				}
				return testActionList, nil
			}},
			want: &shieldv1.ListActionsResponse{Actions: []*shieldv1.Action{
				{
					Id:   "read",
					Name: "Read",
					Namespace: &shieldv1.Namespace{
						Id:        "resource-1",
						Name:      "Resource 1",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "write",
					Name: "Write",
					Namespace: &shieldv1.Namespace{
						Id:        "resource-1",
						Name:      "Resource 1",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "manage",
					Name: "Manage",
					Namespace: &shieldv1.Namespace{
						Id:        "resource-1",
						Name:      "Resource 1",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
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

			mockDep := Dep{ActionService: tt.mockActionSrc}
			resp, err := mockDep.ListActions(context.Background(), tt.req)

			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestCreateAction(t *testing.T) {
	t.Parallel()

	table := []struct {
		title         string
		mockActionSrv mockActionSrv
		req           *shieldv1.CreateActionRequest
		want          *shieldv1.CreateActionResponse
		err           error
	}{
		{
			title: "error in creating action",
			mockActionSrv: mockActionSrv{CreateActionFunc: func(ctx context.Context, act model.Action) (model.Action, error) {
				return model.Action{}, errors.New("some error")
			}},
			req: &shieldv1.CreateActionRequest{Body: &shieldv1.ActionRequestBody{
				Id:          "read",
				Name:        "Read",
				NamespaceId: "team",
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "success",
			mockActionSrv: mockActionSrv{CreateActionFunc: func(ctx context.Context, act model.Action) (model.Action, error) {
				return model.Action{
					Id:   "read",
					Name: "Read",
					Namespace: model.Namespace{
						Id:   "team",
						Name: "Team",
					},
				}, nil
			}},
			req: &shieldv1.CreateActionRequest{Body: &shieldv1.ActionRequestBody{
				Id:          "read",
				Name:        "Read",
				NamespaceId: "team",
			}},
			want: &shieldv1.CreateActionResponse{Action: &shieldv1.Action{
				Id:   "read",
				Name: "Read",
				Namespace: &shieldv1.Namespace{
					Id:        "team",
					Name:      "Team",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
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

			mockDep := Dep{ActionService: tt.mockActionSrv}
			resp, err := mockDep.CreateAction(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

type mockActionSrv struct {
	GetActionFunc    func(ctx context.Context, id string) (model.Action, error)
	CreateActionFunc func(ctx context.Context, act model.Action) (model.Action, error)
	ListActionsFunc  func(ctx context.Context) ([]model.Action, error)
	UpdateActionFunc func(ctx context.Context, id string, act model.Action) (model.Action, error)
}

func (m mockActionSrv) GetAction(ctx context.Context, id string) (model.Action, error) {
	return m.GetActionFunc(ctx, id)
}

func (m mockActionSrv) ListActions(ctx context.Context) ([]model.Action, error) {
	return m.ListActionsFunc(ctx)
}

func (m mockActionSrv) CreateAction(ctx context.Context, act model.Action) (model.Action, error) {
	return m.CreateActionFunc(ctx, act)
}

func (m mockActionSrv) UpdateAction(ctx context.Context, id string, act model.Action) (model.Action, error) {
	return m.UpdateActionFunc(ctx, id, act)
}
