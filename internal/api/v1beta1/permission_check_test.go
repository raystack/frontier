package v1beta1

import (
	"context"
	"testing"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_CheckResourcePermission(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(res *mocks.ResourceService)
		request *frontierv1beta1.CheckResourcePermissionRequest
		want    *frontierv1beta1.CheckResourcePermissionResponse
		wantErr error
	}{
		{
			name: "should return bad request error if object id is empty or namespace is empty",
			request: &frontierv1beta1.CheckResourcePermissionRequest{
				Resource: "not-namespace-uuid-format",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return user unauthenticated error if CheckAuthz function returns ErrUnauthenticated",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        testRelationV2.Object.ID,
					Namespace: testRelationV2.Object.Namespace,
				}, schema.UpdatePermission).Return(false, errors.ErrUnauthenticated)
			},
			request: &frontierv1beta1.CheckResourcePermissionRequest{
				Permission: schema.UpdatePermission,
				Resource:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
			},
			want:    nil,
			wantErr: grpcUnauthenticated,
		},
		{
			name: "should return internal error if relation service's CheckAuthz function returns some error",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        testRelationV2.Object.ID,
					Namespace: testRelationV2.Object.Namespace,
				}, schema.UpdatePermission).Return(false, errors.New("some error"))
			},
			request: &frontierv1beta1.CheckResourcePermissionRequest{
				Permission: schema.UpdatePermission,
				Resource:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return true when CheckAuthz function returns true bool",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        testRelationV2.Object.ID,
					Namespace: testRelationV2.Object.Namespace,
				}, schema.UpdatePermission).Return(true, nil)
			},
			request: &frontierv1beta1.CheckResourcePermissionRequest{
				ObjectId:        testRelationV2.Object.ID,
				ObjectNamespace: testRelationV2.Object.Namespace,
				Permission:      schema.UpdatePermission,
			},
			want: &frontierv1beta1.CheckResourcePermissionResponse{
				Status: true,
			},
			wantErr: nil,
		},
		{
			name: "should return false when CheckAuthz function returns false bool",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        testRelationV2.Object.ID,
					Namespace: testRelationV2.Object.Namespace,
				}, schema.UpdatePermission).Return(false, nil)
			},
			request: &frontierv1beta1.CheckResourcePermissionRequest{
				ObjectId:        testRelationV2.Object.ID,
				ObjectNamespace: testRelationV2.Object.Namespace,
				Permission:      schema.UpdatePermission,
			},
			want: &frontierv1beta1.CheckResourcePermissionResponse{
				Status: false,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}

			mockDep := Handler{resourceService: mockResourceSrv}
			resp, err := mockDep.CheckResourcePermission(context.Background(), tt.request)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, resp)
		})
	}
}

func TestHandler_IsAuthorized(t *testing.T) {
	type autA struct {
		objectNamespace string
		objectID        string
		permission      string
	}
	tests := []struct {
		name    string
		setup   func(res *mocks.ResourceService)
		args    autA
		wantErr error
	}{
		{
			name: "Should return Unauthenticated error if user is not authorize",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        "objectID",
					Namespace: "objectNamespace",
				}, "permis").Return(true, user.ErrInvalidEmail)
			},
			args: autA{
				objectNamespace: "objectNamespace",
				objectID:        "objectID",
				permission:      "permis",
			},
			wantErr: grpcUnauthenticated,
		},
		{
			name: "Should return Internal Server Error if user is not authorize",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        "objectID",
					Namespace: "objectNamespace",
				}, "permis").Return(true, errors.New("some error"))
			},
			args: autA{
				objectNamespace: "objectNamespace",
				objectID:        "objectID",
				permission:      "permis",
			},
			wantErr: status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			name: "should return bad request error if object id is empty or namespace is empty",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        "objectID",
					Namespace: "objectNamespace",
				}, "permis").Return(false, nil)
			},
			args: autA{
				objectNamespace: "objectNamespace",
				objectID:        "objectID",
				permission:      "permis",
			},

			wantErr: grpcPermissionDenied,
		},
		{
			name: "should show success if Id is authorized ",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        "objectID",
					Namespace: "objectNamespace",
				}, "permis").Return(true, nil)
			},
			args: autA{
				objectNamespace: "objectNamespace",
				objectID:        "objectID",
				permission:      "permis",
			},

			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			mockDep := Handler{resourceService: mockResourceSrv}
			err := mockDep.IsAuthorized(ctx, tt.args.objectNamespace, tt.args.objectID, tt.args.permission)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
