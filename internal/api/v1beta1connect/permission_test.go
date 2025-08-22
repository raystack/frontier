package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testPermissionIdx = 0
	testPermissions   = []permission.Permission{
		{
			ID:          uuid.New().String(),
			Name:        "Read",
			NamespaceID: "app/resource",
			Metadata:    map[string]any{},
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Write",
			NamespaceID: "app/resource",
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Manage",
			NamespaceID: "app/resource",
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
	}
)

func TestHandler_CreatePermission(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.PermissionService, bs *mocks.BootstrapService)
		request *connect.Request[frontierv1beta1.CreatePermissionRequest]
		want    *connect.Response[frontierv1beta1.CreatePermissionResponse]
		wantErr error
	}{
		{
			name: "should return internal error if permission service return some error",
			setup: func(as *mocks.PermissionService, bs *mocks.BootstrapService) {
				bs.EXPECT().AppendSchema(mock.AnythingOfType("context.backgroundCtx"), schema.ServiceDefinition{
					Permissions: []schema.ResourcePermission{
						{
							Name:        testPermissions[testPermissionIdx].Name,
							Namespace:   testPermissions[testPermissionIdx].NamespaceID,
							Description: "",
						},
					},
				}).Return(errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Name:      testPermissions[testPermissionIdx].Name,
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name:  "should return bad request error if namespace id is empty",
			setup: func(as *mocks.PermissionService, bs *mocks.BootstrapService) {},
			request: connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Name: testPermissions[testPermissionIdx].Name,
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name:  "should return bad request error if name is empty",
			setup: func(as *mocks.PermissionService, bs *mocks.BootstrapService) {},
			request: connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService, bs *mocks.BootstrapService) {
				bs.EXPECT().AppendSchema(mock.AnythingOfType("context.backgroundCtx"), schema.ServiceDefinition{
					Permissions: []schema.ResourcePermission{
						{
							Name:      testPermissions[testPermissionIdx].Name + "0",
							Namespace: testPermissions[testPermissionIdx].NamespaceID,
						},
						{
							Name:      testPermissions[testPermissionIdx].Name + "1",
							Namespace: testPermissions[testPermissionIdx].NamespaceID,
						},
					},
				}).Return(nil)
				as.EXPECT().List(mock.Anything, permission.Filter{
					Slugs: []string{
						schema.FQPermissionNameFromNamespace(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"0"),
						schema.FQPermissionNameFromNamespace(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"1"),
					},
				}).Return([]permission.Permission{
					{
						ID:          testPermissions[testPermissionIdx].ID,
						Name:        testPermissions[testPermissionIdx].Name + "0",
						NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
					},
					{
						ID:          testPermissions[testPermissionIdx].ID,
						Name:        testPermissions[testPermissionIdx].Name + "1",
						NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Name:      testPermissions[testPermissionIdx].Name + "0",
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
					},
					{
						Name:      testPermissions[testPermissionIdx].Name + "1",
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreatePermissionResponse{
				Permissions: []*frontierv1beta1.Permission{
					{
						Id:        testPermissions[testPermissionIdx].ID,
						Name:      testPermissions[testPermissionIdx].Name + "0",
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
						Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"0"),
						CreatedAt: timestamppb.New(testPermissions[testPermissionIdx].CreatedAt),
						UpdatedAt: timestamppb.New(testPermissions[testPermissionIdx].UpdatedAt),
					},
					{
						Id:        testPermissions[testPermissionIdx].ID,
						Name:      testPermissions[testPermissionIdx].Name + "1",
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
						Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"1"),
						CreatedAt: timestamppb.New(testPermissions[testPermissionIdx].CreatedAt),
						UpdatedAt: timestamppb.New(testPermissions[testPermissionIdx].UpdatedAt),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return success if permission service return nil error with permission key",
			setup: func(as *mocks.PermissionService, bs *mocks.BootstrapService) {
				bs.EXPECT().AppendSchema(mock.AnythingOfType("context.backgroundCtx"), schema.ServiceDefinition{
					Permissions: []schema.ResourcePermission{
						{
							Name:      testPermissions[testPermissionIdx].Name + "0",
							Namespace: testPermissions[testPermissionIdx].NamespaceID,
						},
					},
				}).Return(nil)
				as.EXPECT().List(mock.Anything, permission.Filter{
					Slugs: []string{
						schema.FQPermissionNameFromNamespace(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"0"),
					},
				}).Return([]permission.Permission{
					{
						ID:          testPermissions[testPermissionIdx].ID,
						Name:        testPermissions[testPermissionIdx].Name + "0",
						NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Key: schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"0"),
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreatePermissionResponse{
				Permissions: []*frontierv1beta1.Permission{
					{
						Id:        testPermissions[testPermissionIdx].ID,
						Name:      testPermissions[testPermissionIdx].Name + "0",
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
						Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"0"),
						CreatedAt: timestamppb.New(testPermissions[testPermissionIdx].CreatedAt),
						UpdatedAt: timestamppb.New(testPermissions[testPermissionIdx].UpdatedAt),
					},
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPermissionSrv := new(mocks.PermissionService)
			mockBootstrapSrv := new(mocks.BootstrapService)
			if tt.setup != nil {
				tt.setup(mockPermissionSrv, mockBootstrapSrv)
			}
			mockDep := &ConnectHandler{permissionService: mockPermissionSrv, bootstrapService: mockBootstrapSrv}
			resp, err := mockDep.CreatePermission(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
