package v1beta1connect

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testPermission = permission.Permission{
		Name:        schema.UpdatePermission,
		NamespaceID: testRelationV2.Object.Namespace,
	}
)

func TestHandler_CheckResourcePermission(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(res *mocks.ResourceService, perm *mocks.PermissionService)
		request *connect.Request[frontierv1beta1.CheckResourcePermissionRequest]
		want    *connect.Response[frontierv1beta1.CheckResourcePermissionResponse]
		wantErr error
	}{
		{
			name: "should return bad request error if resource is malformed",
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Resource: "not-namespace-uuid-format",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if resource is missing",
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Permission: schema.UpdatePermission,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if resource id part is empty",
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Resource:   "organization:",
				Permission: schema.UpdatePermission,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if resource namespace part is empty",
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Resource:   ":" + testRelationV2.Object.ID,
				Permission: schema.UpdatePermission,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if only the removed split fields are sent",
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				ObjectId:        testRelationV2.Object.ID,
				ObjectNamespace: testRelationV2.Object.Namespace,
				Permission:      schema.UpdatePermission,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return user unauthenticated error if CheckAuthz function returns ErrUnauthenticated",
			setup: func(res *mocks.ResourceService, perm *mocks.PermissionService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("context.backgroundCtx"), resource.Check{
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					}, Permission: schema.UpdatePermission,
				}).Return(false, errors.ErrUnauthenticated)
				perm.EXPECT().Get(mock.Anything, schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, schema.UpdatePermission)).
					Return(testPermission, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Permission: schema.UpdatePermission,
				Resource:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			name: "should return internal error if relation service's CheckAuthz function returns some error",
			setup: func(res *mocks.ResourceService, perm *mocks.PermissionService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("context.backgroundCtx"), resource.Check{
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					}, Permission: schema.UpdatePermission,
				}).Return(false, errors.New("test error"))
				perm.EXPECT().Get(mock.Anything, schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, schema.UpdatePermission)).
					Return(testPermission, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Permission: schema.UpdatePermission,
				Resource:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, fmt.Errorf("handleAuthErr: %w", errors.New("test error"))),
		},
		{
			name: "should return true when CheckAuthz function returns true bool",
			setup: func(res *mocks.ResourceService, perm *mocks.PermissionService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("context.backgroundCtx"), resource.Check{
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					}, Permission: schema.UpdatePermission,
				}).Return(true, nil)
				perm.EXPECT().Get(mock.Anything, schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, schema.UpdatePermission)).
					Return(testPermission, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Permission: schema.UpdatePermission,
				Resource:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
			}),
			want: connect.NewResponse(&frontierv1beta1.CheckResourcePermissionResponse{
				Status: true,
			}),
			wantErr: nil,
		},
		{
			name: "should return false when CheckAuthz function returns false bool",
			setup: func(res *mocks.ResourceService, perm *mocks.PermissionService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("context.backgroundCtx"), resource.Check{
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					}, Permission: schema.UpdatePermission,
				}).Return(false, nil)
				perm.EXPECT().Get(mock.Anything, schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, schema.UpdatePermission)).
					Return(testPermission, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckResourcePermissionRequest{
				Permission: schema.UpdatePermission,
				Resource:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
			}),
			want: connect.NewResponse(&frontierv1beta1.CheckResourcePermissionResponse{
				Status: false,
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			mockPermissionSrv := new(mocks.PermissionService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv, mockPermissionSrv)
			}

			mockDep := &ConnectHandler{resourceService: mockResourceSrv, permissionService: mockPermissionSrv}
			resp, err := mockDep.CheckResourcePermission(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, resp)
		})
	}
}

func TestHandler_BatchCheckPermission(t *testing.T) {
	tests := []struct {
		name    string
		request *connect.Request[frontierv1beta1.BatchCheckPermissionRequest]
		wantErr error
	}{
		{
			name: "should return bad request error if a body resource is malformed",
			request: connect.NewRequest(&frontierv1beta1.BatchCheckPermissionRequest{
				Bodies: []*frontierv1beta1.BatchCheckPermissionBody{
					{Resource: "not-namespace-uuid-format", Permission: schema.UpdatePermission},
				},
			}),
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if a body resource id part is empty",
			request: connect.NewRequest(&frontierv1beta1.BatchCheckPermissionRequest{
				Bodies: []*frontierv1beta1.BatchCheckPermissionBody{
					{Resource: "organization:", Permission: schema.UpdatePermission},
				},
			}),
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if a body resource namespace part is empty",
			request: connect.NewRequest(&frontierv1beta1.BatchCheckPermissionRequest{
				Bodies: []*frontierv1beta1.BatchCheckPermissionBody{
					{Resource: ":" + testRelationV2.Object.ID, Permission: schema.UpdatePermission},
				},
			}),
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDep := &ConnectHandler{}
			resp, err := mockDep.BatchCheckPermission(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Nil(t, resp)
		})
	}
}
