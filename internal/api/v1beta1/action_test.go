package v1beta1

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testActionMap = map[string]action.Action{
	"read": {
		ID:   "read",
		Name: "Read",
		Namespace: namespace.Namespace{
			ID:        "resource-1",
			Name:      "Resource 1",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"write": {
		ID:   "write",
		Name: "Write",
		Namespace: namespace.Namespace{
			ID:        "resource-1",
			Name:      "Resource 1",
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"manage": {
		ID:   "manage",
		Name: "Manage",
		Namespace: namespace.Namespace{
			ID:        "resource-1",
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
		title string
		setup func(as *mocks.ActionService)
		req   *shieldv1beta1.ListActionsRequest
		want  *shieldv1beta1.ListActionsResponse
		err   error
	}{
		{
			title: "error in Action Service",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().List(mock.Anything).Return([]action.Action{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "success",
			setup: func(as *mocks.ActionService) {
				var testActionList []action.Action
				for _, act := range testActionMap {
					testActionList = append(testActionList, act)
				}

				sort.Slice(testActionList[:], func(i, j int) bool {
					return strings.Compare(testActionList[i].ID, testActionList[j].ID) < 1
				})
				as.EXPECT().List(mock.Anything).Return(testActionList, nil)
			},
			want: &shieldv1beta1.ListActionsResponse{Actions: []*shieldv1beta1.Action{
				{
					Id:   "manage",
					Name: "Manage",
					Namespace: &shieldv1beta1.Namespace{
						Id:        "resource-1",
						Name:      "Resource 1",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "read",
					Name: "Read",
					Namespace: &shieldv1beta1.Namespace{
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
					Namespace: &shieldv1beta1.Namespace{
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

			mockActionSrv := new(mocks.ActionService)
			if tt.setup != nil {
				tt.setup(mockActionSrv)
			}
			mockDep := Handler{actionService: mockActionSrv}
			resp, err := mockDep.ListActions(context.Background(), tt.req)

			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestCreateAction(t *testing.T) {
	t.Parallel()

	table := []struct {
		title string
		setup func(as *mocks.ActionService)
		req   *shieldv1beta1.CreateActionRequest
		want  *shieldv1beta1.CreateActionResponse
		err   error
	}{
		{
			title: "error in creating action",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Create(mock.Anything, action.Action{
					ID:          "read",
					Name:        "Read",
					NamespaceID: "team",
				}).Return(action.Action{
					ID:          "read",
					Name:        "Read",
					NamespaceID: "team",
				}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreateActionRequest{Body: &shieldv1beta1.ActionRequestBody{
				Id:          "read",
				Name:        "Read",
				NamespaceId: "team",
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "success",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Create(mock.Anything, action.Action{}).Return(
					action.Action{
						ID:   "read",
						Name: "Read",
						Namespace: namespace.Namespace{
							ID:   "team",
							Name: "Team",
						},
					}, nil)
			},
			req: &shieldv1beta1.CreateActionRequest{Body: &shieldv1beta1.ActionRequestBody{
				Id:          "read",
				Name:        "Read",
				NamespaceId: "team",
			}},
			want: &shieldv1beta1.CreateActionResponse{Action: &shieldv1beta1.Action{
				Id:   "read",
				Name: "Read",
				Namespace: &shieldv1beta1.Namespace{
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

			mockActionSrv := new(mocks.ActionService)
			if tt.setup != nil {
				tt.setup(mockActionSrv)
			}
			mockDep := Handler{actionService: mockActionSrv}
			resp, err := mockDep.CreateAction(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}
