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

var testNsMap = map[string]namespace.Namespace{
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
	t.Parallel()
	table := []struct {
		title string
		setup func(ns *mocks.NamespaceService)
		req   *shieldv1beta1.ListNamespacesRequest
		want  *shieldv1beta1.ListNamespacesResponse
		err   error
	}{
		{
			title: "error in Namespace Service",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().List(mock.Anything).Return([]namespace.Namespace{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "success",
			setup: func(ns *mocks.NamespaceService) {
				var testNSList []namespace.Namespace
				for _, ns := range testNsMap {
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
			t.Parallel()

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
	t.Parallel()

	table := []struct {
		title string
		setup func(ns *mocks.NamespaceService)
		req   *shieldv1beta1.CreateNamespaceRequest
		want  *shieldv1beta1.CreateNamespaceResponse
		err   error
	}{
		{
			title: "error in creating namespace",
			setup: func(ns *mocks.NamespaceService) {
				ns.EXPECT().Create(mock.Anything, mock.Anything).Return(namespace.Namespace{}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreateNamespaceRequest{Body: &shieldv1beta1.NamespaceRequestBody{
				Id:   "team",
				Name: "Team",
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "success",
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
			t.Parallel()

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
