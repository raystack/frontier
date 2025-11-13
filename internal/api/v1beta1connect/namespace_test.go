package v1beta1connect

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestHandler_ListNamespaces(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ns *mocks.NamespaceService)
		request *connect.Request[frontierv1beta1.ListNamespacesRequest]
		want    *connect.Response[frontierv1beta1.ListNamespacesResponse]
		wantErr error
	}{
		{
			name: "should return internal error if namespace service return some error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().List(mock.Anything).Return([]namespace.Namespace{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListNamespacesRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success if namespace service return nil error",
			setup: func(ns *mocks.NamespaceService) {
				var testNSList []namespace.Namespace
				for _, ns := range testNSMap {
					testNSList = append(testNSList, ns)
				}
				sort.Slice(testNSList, func(i, j int) bool {
					return strings.Compare(testNSList[i].ID, testNSList[j].ID) < 1
				})
				ns.EXPECT().List(mock.Anything).Return(testNSList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListNamespacesRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListNamespacesResponse{
				Namespaces: []*frontierv1beta1.Namespace{
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
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNamespaceSrv := new(mocks.NamespaceService)
			if tt.setup != nil {
				tt.setup(mockNamespaceSrv)
			}
			mockDep := &ConnectHandler{namespaceService: mockNamespaceSrv}
			resp, err := mockDep.ListNamespaces(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, resp)
		})
	}
}

func TestHandler_GetNamespace(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.NamespaceService)
		request *connect.Request[frontierv1beta1.GetNamespaceRequest]
		want    *connect.Response[frontierv1beta1.GetNamespaceResponse]
		wantErr error
	}{
		{
			name: "should return internal error if namespace service return some error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testNSID).Return(namespace.Namespace{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetNamespaceRequest{
				Id: testNSID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if namespace id is empty",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(namespace.Namespace{}, namespace.ErrInvalidID)
			},
			request: connect.NewRequest(&frontierv1beta1.GetNamespaceRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if namespace id not exist",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testNSID).Return(namespace.Namespace{}, namespace.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetNamespaceRequest{
				Id: testNSID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return success is namespace service return nil error",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testNSID).Return(namespace.Namespace{
					ID:   testNSMap[testNSID].ID,
					Name: testNSMap[testNSID].Name,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetNamespaceRequest{
				Id: testNSID,
			}),
			want: connect.NewResponse(&frontierv1beta1.GetNamespaceResponse{
				Namespace: &frontierv1beta1.Namespace{
					Id:        testNSMap[testNSID].ID,
					Name:      testNSMap[testNSID].Name,
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNamespaceSrv := new(mocks.NamespaceService)
			if tt.setup != nil {
				tt.setup(mockNamespaceSrv)
			}
			mockDep := &ConnectHandler{namespaceService: mockNamespaceSrv}
			resp, err := mockDep.GetNamespace(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, resp)
		})
	}
}
