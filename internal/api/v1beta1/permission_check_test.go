package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/core/relation"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
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
			name: "should return internal error if relation service's CheckAuthz function returns some error",
			setup: func(res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), relation.Object{
					ID:        testRelationV2.Object.ID,
					Namespace: testRelationV2.Object.Namespace,
				}, schema.UpdatePermission).Return(false, errors.New("some error"))
			},
			request: &frontierv1beta1.CheckResourcePermissionRequest{
				ObjectId:        testRelationV2.Object.ID,
				ObjectNamespace: testRelationV2.Object.Namespace,
				Permission:      schema.UpdatePermission,
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
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
