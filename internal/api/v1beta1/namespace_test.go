package v1beta1

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testNSID = "team"
var testNSMap = map[string]namespace.Namespace{
	"team": {
		ID:        "team",
		Name:      "Team",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"org": {
		ID:        "org",
		Name:      "Org",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"project": {
		ID:        "project",
		Name:      "Project",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestListNamespaces(t *testing.T) {
	table := []struct {
		title string
		setup func(ns *mocks.NamespaceService)
		req   *shieldv1beta1.ListNamespacesRequest
		want  *shieldv1beta1.ListNamespacesResponse
		err   error
	}{
		{
			title: "should return internal error if namespace service return some error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().List(mock.Anything).Return([]namespace.Namespace{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "should return success if namespace service return nil error",
			setup: func(ns *mocks.NamespaceService) {
				var testNSList []namespace.Namespace
				for _, ns := range testNSMap {
					testNSList = append(testNSList, ns)
				}
				sort.Slice(testNSList[:], func(i, j int) bool {
					return strings.Compare(testNSList[i].ID, testNSList[j].ID) < 1
				})
				ns.EXPECT().List(mock.Anything).Return(testNSList, nil)
			},
			want: &shieldv1beta1.ListNamespacesResponse{Namespaces: []*shieldv1beta1.Namespace{
				{
					Id:        "org",
					Name:      "Org",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:        "project",
					Name:      "Project",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:        "team",
					Name:      "Team",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockNamespaceSrv := new(mocks.NamespaceService)
			if tt.setup != nil {
				tt.setup(mockNamespaceSrv)
			}
			mockDep := Handler{namespaceService: mockNamespaceSrv}
			resp, err := mockDep.ListNamespaces(context.Background(), tt.req)

			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestCreateNamespace(t *testing.T) {
	table := []struct {
		title string
		setup func(ns *mocks.NamespaceService)
		req   *shieldv1beta1.CreateNamespaceRequest
		want  *shieldv1beta1.CreateNamespaceResponse
		err   error
	}{
		{
			title: "should return internal error if namespace service return some error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID:   "team",
					Name: "Team",
				}).Return(namespace.Namespace{}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreateNamespaceRequest{Body: &shieldv1beta1.NamespaceRequestBody{
				Id:   "team",
				Name: "Team",
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return already exist error if namespace service return err conflict",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID:   "team",
					Name: "Team",
				}).Return(namespace.Namespace{}, namespace.ErrConflict)
			},
			req: &shieldv1beta1.CreateNamespaceRequest{Body: &shieldv1beta1.NamespaceRequestBody{
				Id:   "team",
				Name: "Team",
			}},
			want: nil,
			err:  grpcConflictError,
		},
		{
			title: "should return bad request error if id is empty",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					Name: "Team",
				}).Return(namespace.Namespace{}, namespace.ErrInvalidID)
			},
			req: &shieldv1beta1.CreateNamespaceRequest{Body: &shieldv1beta1.NamespaceRequestBody{
				Name: "Team",
			}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return bad request error if name is empty",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID: "team",
				}).Return(namespace.Namespace{}, namespace.ErrInvalidDetail)
			},
			req: &shieldv1beta1.CreateNamespaceRequest{Body: &shieldv1beta1.NamespaceRequestBody{
				Id: "team",
			}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return success if namespace service return nil error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Create(mock.Anything, mock.Anything).Return(
					namespace.Namespace{
						ID:   "team",
						Name: "Team",
					}, nil)
			},
			req: &shieldv1beta1.CreateNamespaceRequest{Body: &shieldv1beta1.NamespaceRequestBody{
				Id:   "team",
				Name: "Team",
			}},
			want: &shieldv1beta1.CreateNamespaceResponse{Namespace: &shieldv1beta1.Namespace{
				Id:        "team",
				Name:      "Team",
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockNamespaceSrv := new(mocks.NamespaceService)
			if tt.setup != nil {
				tt.setup(mockNamespaceSrv)
			}
			mockDep := Handler{namespaceService: mockNamespaceSrv}
			resp, err := mockDep.CreateNamespace(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_GetNamespace(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.NamespaceService)
		request *shieldv1beta1.GetNamespaceRequest
		want    *shieldv1beta1.GetNamespaceResponse
		wantErr error
	}{
		{
			name: "should return internal error if namespace service return some error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testNSID).Return(namespace.Namespace{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetNamespaceRequest{
				Id: testNSID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if namespace id is empty",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(namespace.Namespace{}, namespace.ErrInvalidID)
			},
			request: &shieldv1beta1.GetNamespaceRequest{},
			want:    nil,
			wantErr: grpcNamespaceNotFoundErr,
		},
		{
			name: "should return not found error if namespace id not exist",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testNSID).Return(namespace.Namespace{}, namespace.ErrNotExist)
			},
			request: &shieldv1beta1.GetNamespaceRequest{
				Id: testNSID,
			},
			want:    nil,
			wantErr: grpcNamespaceNotFoundErr,
		},
		{
			name: "should return success is namespace service return nil error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testNSID).Return(namespace.Namespace{
					ID:   testNSMap[testNSID].ID,
					Name: testNSMap[testNSID].Name,
				}, nil)
			},
			request: &shieldv1beta1.GetNamespaceRequest{
				Id: testNSID,
			},
			want: &shieldv1beta1.GetNamespaceResponse{
				Namespace: &shieldv1beta1.Namespace{
					Id:        testNSMap[testNSID].ID,
					Name:      testNSMap[testNSID].Name,
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNamespaceSrv := new(mocks.NamespaceService)
			if tt.setup != nil {
				tt.setup(mockNamespaceSrv)
			}
			mockDep := Handler{namespaceService: mockNamespaceSrv}
			resp, err := mockDep.GetNamespace(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateNamespace(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.NamespaceService)
		request *shieldv1beta1.UpdateNamespaceRequest
		want    *shieldv1beta1.UpdateNamespaceResponse
		wantErr error
	}{
		{
			name: "should return internal error if namespace service return some error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID:   testNSID,
					Name: testNSMap[testNSID].Name,
				}).Return(namespace.Namespace{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateNamespaceRequest{
				Id: testNSID,
				Body: &shieldv1beta1.NamespaceRequestBody{
					Id:   testNSID, // id in body is ignored
					Name: testNSMap[testNSID].Name,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if namespace id not exist",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID:   testNSID,
					Name: testNSMap[testNSID].Name,
				}).Return(namespace.Namespace{}, namespace.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateNamespaceRequest{
				Id: testNSID,
				Body: &shieldv1beta1.NamespaceRequestBody{
					Id:   testNSID, // id in body is ignored
					Name: testNSMap[testNSID].Name,
				},
			},
			want:    nil,
			wantErr: grpcNamespaceNotFoundErr,
		},
		{
			name: "should return already exist error if namespace service return err conflict",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID:   testNSID,
					Name: testNSMap[testNSID].Name,
				}).Return(namespace.Namespace{}, namespace.ErrConflict)
			},
			request: &shieldv1beta1.UpdateNamespaceRequest{
				Id: testNSID,
				Body: &shieldv1beta1.NamespaceRequestBody{
					Id:   testNSID, // id in body is ignored
					Name: testNSMap[testNSID].Name,
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return bad request error if namespace name is empty",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID: testNSID,
				}).Return(namespace.Namespace{}, namespace.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateNamespaceRequest{
				Id: testNSID,
				Body: &shieldv1beta1.NamespaceRequestBody{
					Id: testNSID, // id in body is ignored
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if namespace service return nil error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), namespace.Namespace{
					ID:   testNSID,
					Name: testNSMap[testNSID].Name,
				}).Return(namespace.Namespace{
					ID:   testNSMap[testNSID].ID,
					Name: testNSMap[testNSID].Name,
				}, nil)
			},
			request: &shieldv1beta1.UpdateNamespaceRequest{
				Id: testNSID,
				Body: &shieldv1beta1.NamespaceRequestBody{
					Id:   testNSID, // id in body is ignored
					Name: testNSMap[testNSID].Name,
				},
			},
			want: &shieldv1beta1.UpdateNamespaceResponse{
				Namespace: &shieldv1beta1.Namespace{
					Id:        testNSMap[testNSID].ID,
					Name:      testNSMap[testNSID].Name,
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNamespaceSrv := new(mocks.NamespaceService)
			if tt.setup != nil {
				tt.setup(mockNamespaceSrv)
			}
			mockDep := Handler{namespaceService: mockNamespaceSrv}
			resp, err := mockDep.UpdateNamespace(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
