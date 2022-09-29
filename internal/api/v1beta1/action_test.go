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

var (
	testActionID  = "read"
	testActionMap = map[string]action.Action{
		"read": {
			ID:          "read",
			Name:        "Read",
			NamespaceID: "resource-1",
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
		"write": {
			ID:          "write",
			Name:        "Write",
			NamespaceID: "resource-1",
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
		"manage": {
			ID:          "manage",
			Name:        "Manage",
			NamespaceID: "resource-1",
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
	}
)

func TestListActions(t *testing.T) {
	table := []struct {
		title string
		setup func(as *mocks.ActionService)
		req   *shieldv1beta1.ListActionsRequest
		want  *shieldv1beta1.ListActionsResponse
		err   error
	}{
		{
			title: "should return internal error if action service return some error",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().List(mock.Anything).Return([]action.Action{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "should return success if action service return nil error",
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
					// @TODO(krtkvrm): issues/171
					//Namespace: &shieldv1beta1.Namespace{
					//	Id:        "resource-1",
					//	Name:      "Resource 1",
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "read",
					Name: "Read",
					// @TODO(krtkvrm): issues/171
					//Namespace: &shieldv1beta1.Namespace{
					//	Id:        "resource-1",
					//	Name:      "Resource 1",
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "write",
					Name: "Write",
					// @TODO(krtkvrm): issues/171
					//Namespace: &shieldv1beta1.Namespace{
					//	Id:        "resource-1",
					//	Name:      "Resource 1",
					//	CreatedAt: timestamppb.New(time.Time{}),
					//	UpdatedAt: timestamppb.New(time.Time{}),
					//},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
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
	table := []struct {
		title string
		setup func(as *mocks.ActionService)
		req   *shieldv1beta1.CreateActionRequest
		want  *shieldv1beta1.CreateActionResponse
		err   error
	}{
		{
			title: "should return internal error if action service return some error",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), action.Action{
					ID:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}).Return(action.Action{}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreateActionRequest{
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return bad request error if namespace id is wrong",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), action.Action{
					ID:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}).Return(action.Action{}, namespace.ErrNotExist)
			},
			req: &shieldv1beta1.CreateActionRequest{
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return bad request error if if id is empty",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), action.Action{
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}).Return(action.Action{}, action.ErrInvalidID)
			},
			req: &shieldv1beta1.CreateActionRequest{
				Body: &shieldv1beta1.ActionRequestBody{
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return bad request error if if name is empty",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), action.Action{
					ID:          testActionMap[testActionID].ID,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}).Return(action.Action{}, action.ErrInvalidDetail)
			},
			req: &shieldv1beta1.CreateActionRequest{
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return success if action service return nil error",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), action.Action{
					ID:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}).Return(action.Action{
					ID:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}, nil)
			},
			req: &shieldv1beta1.CreateActionRequest{Body: &shieldv1beta1.ActionRequestBody{
				Id:          testActionMap[testActionID].ID,
				Name:        testActionMap[testActionID].Name,
				NamespaceId: testActionMap[testActionID].NamespaceID,
			}},
			want: &shieldv1beta1.CreateActionResponse{Action: &shieldv1beta1.Action{
				Id:        testActionMap[testActionID].ID,
				Name:      testActionMap[testActionID].Name,
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
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

func TestHandler_GetAction(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.ActionService)
		request *shieldv1beta1.GetActionRequest
		want    *shieldv1beta1.GetActionResponse
		wantErr error
	}{
		{
			name: "should return internal error if action service return some error",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testActionID).Return(action.Action{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetActionRequest{
				Id: testActionID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if action id not exist",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testActionID).Return(action.Action{}, action.ErrNotExist)
			},
			request: &shieldv1beta1.GetActionRequest{
				Id: testActionID,
			},
			want:    nil,
			wantErr: grpcActionNotFoundErr,
		},
		{
			name: "should return not found error if action id is empty",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(action.Action{}, action.ErrInvalidID)
			},
			request: &shieldv1beta1.GetActionRequest{},
			want:    nil,
			wantErr: grpcActionNotFoundErr,
		},
		{
			name: "should return success if action service return nil error",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testActionID).Return(testActionMap[testActionID], nil)
			},
			request: &shieldv1beta1.GetActionRequest{
				Id: testActionID,
			},
			want: &shieldv1beta1.GetActionResponse{
				Action: &shieldv1beta1.Action{
					Id:        testActionMap[testActionID].ID,
					Name:      testActionMap[testActionID].Name,
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActionSrv := new(mocks.ActionService)
			if tt.setup != nil {
				tt.setup(mockActionSrv)
			}
			mockDep := Handler{actionService: mockActionSrv}
			resp, err := mockDep.GetAction(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateAction(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.ActionService)
		request *shieldv1beta1.UpdateActionRequest
		want    *shieldv1beta1.UpdateActionResponse
		wantErr error
	}{
		{
			name: "should return internal error if action service return some error",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testActionID, action.Action{
					ID:          testActionID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}).Return(action.Action{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateActionRequest{
				Id: testActionID,
				Body: &shieldv1beta1.ActionRequestBody{
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if action id not exist",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testActionID, action.Action{
					ID:          testActionID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID}).Return(action.Action{}, action.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateActionRequest{
				Id: testActionID,
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcActionNotFoundErr,
		},
		{
			name: "should return not found error if action id is empty",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), "", action.Action{
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID}).Return(action.Action{}, action.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdateActionRequest{
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID, // id in body is being ignored
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcActionNotFoundErr,
		},
		{
			name: "should return bad request error if namespace id not exist",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testActionMap[testActionID].ID, action.Action{
					ID:          testActionID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID}).Return(action.Action{}, namespace.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateActionRequest{
				Id: testActionID,
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if name is empty",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testActionMap[testActionID].ID, action.Action{
					ID:          testActionID,
					NamespaceID: testActionMap[testActionID].NamespaceID}).Return(action.Action{}, action.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateActionRequest{
				Id: testActionID,
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if action service return nil error",
			setup: func(as *mocks.ActionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), testActionMap[testActionID].ID, action.Action{
					ID:          testActionID,
					Name:        testActionMap[testActionID].Name,
					NamespaceID: testActionMap[testActionID].NamespaceID,
				}).Return(testActionMap[testActionID], nil)
			},
			request: &shieldv1beta1.UpdateActionRequest{
				Id: testActionID,
				Body: &shieldv1beta1.ActionRequestBody{
					Id:          testActionMap[testActionID].ID,
					Name:        testActionMap[testActionID].Name,
					NamespaceId: testActionMap[testActionID].NamespaceID,
				},
			},
			want: &shieldv1beta1.UpdateActionResponse{
				Action: &shieldv1beta1.Action{
					Id:        testActionMap[testActionID].ID,
					Name:      testActionMap[testActionID].Name,
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockActionSrv := new(mocks.ActionService)
			if tt.setup != nil {
				tt.setup(mockActionSrv)
			}
			mockDep := Handler{actionService: mockActionSrv}
			resp, err := mockDep.UpdateAction(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
