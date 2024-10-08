package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/google/uuid"

	"github.com/raystack/frontier/core/permission"

	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestListPermissions(t *testing.T) {
	table := []struct {
		title string
		setup func(as *mocks.PermissionService)
		req   *frontierv1beta1.ListPermissionsRequest
		want  *frontierv1beta1.ListPermissionsResponse
		err   error
	}{
		{
			title: "should return internal error if action service return some error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().List(mock.Anything, permission.Filter{}).Return([]permission.Permission{}, errors.New("test error"))
			},
			want: nil,
			err:  errors.New("test error"),
		},
		{
			title: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService) {
				var testPermissionList []permission.Permission
				for _, act := range testPermissions {
					testPermissionList = append(testPermissionList, act)
				}
				as.EXPECT().List(mock.Anything, permission.Filter{}).Return(testPermissionList, nil)
			},
			want: &frontierv1beta1.ListPermissionsResponse{Permissions: []*frontierv1beta1.Permission{
				{
					Id:        testPermissions[0].ID,
					Name:      testPermissions[0].Name,
					Namespace: testPermissions[0].NamespaceID,
					Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[0].NamespaceID, testPermissions[0].Name),
					CreatedAt: timestamppb.New(testPermissions[0].CreatedAt),
					UpdatedAt: timestamppb.New(testPermissions[0].UpdatedAt),
				},
				{
					Id:        testPermissions[1].ID,
					Name:      testPermissions[1].Name,
					Namespace: testPermissions[1].NamespaceID,
					Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[1].NamespaceID, testPermissions[1].Name),
					CreatedAt: timestamppb.New(testPermissions[1].CreatedAt),
					UpdatedAt: timestamppb.New(testPermissions[1].UpdatedAt),
				},
				{
					Id:        testPermissions[2].ID,
					Name:      testPermissions[2].Name,
					Namespace: testPermissions[2].NamespaceID,
					Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[2].NamespaceID, testPermissions[2].Name),
					CreatedAt: timestamppb.New(testPermissions[2].CreatedAt),
					UpdatedAt: timestamppb.New(testPermissions[2].UpdatedAt),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockPermissionSrv := new(mocks.PermissionService)
			if tt.setup != nil {
				tt.setup(mockPermissionSrv)
			}
			mockDep := Handler{permissionService: mockPermissionSrv}
			resp, err := mockDep.ListPermissions(context.Background(), tt.req)

			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestCreatePermission(t *testing.T) {
	table := []struct {
		title string
		setup func(as *mocks.PermissionService, bs *mocks.BootstrapService)
		req   *frontierv1beta1.CreatePermissionRequest
		want  *frontierv1beta1.CreatePermissionResponse
		err   error
	}{
		{
			title: "should return internal error if permission service return some error",
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
			req: &frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Name:      testPermissions[testPermissionIdx].Name,
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
					},
				},
			},
			want: nil,
			err:  errors.New("test error"),
		},
		{
			title: "should return bad request error if namespace id is empty",
			setup: func(as *mocks.PermissionService, bs *mocks.BootstrapService) {},
			req: &frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Name: testPermissions[testPermissionIdx].Name,
					},
				},
			},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return bad request error if if name is empty",
			setup: func(as *mocks.PermissionService, bs *mocks.BootstrapService) {
			},
			req: &frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Namespace: testPermissions[testPermissionIdx].NamespaceID,
					},
				},
			},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return success if permission service return nil error",
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
			req: &frontierv1beta1.CreatePermissionRequest{
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
			},
			want: &frontierv1beta1.CreatePermissionResponse{Permissions: []*frontierv1beta1.Permission{
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
			}},
			err: nil,
		},
		{
			title: "should return success if permission service return nil error with permission key",
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
			req: &frontierv1beta1.CreatePermissionRequest{
				Bodies: []*frontierv1beta1.PermissionRequestBody{
					{
						Key: schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"0"),
					},
				},
			},
			want: &frontierv1beta1.CreatePermissionResponse{Permissions: []*frontierv1beta1.Permission{
				{
					Id:        testPermissions[testPermissionIdx].ID,
					Name:      testPermissions[testPermissionIdx].Name + "0",
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
					Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name+"0"),
					CreatedAt: timestamppb.New(testPermissions[testPermissionIdx].CreatedAt),
					UpdatedAt: timestamppb.New(testPermissions[testPermissionIdx].UpdatedAt),
				},
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockPermissionSrv := new(mocks.PermissionService)
			mockBootstrapSrv := new(mocks.BootstrapService)
			if tt.setup != nil {
				tt.setup(mockPermissionSrv, mockBootstrapSrv)
			}
			mockDep := Handler{permissionService: mockPermissionSrv, bootstrapService: mockBootstrapSrv}
			resp, err := mockDep.CreatePermission(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_GetPermission(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.PermissionService)
		request *frontierv1beta1.GetPermissionRequest
		want    *frontierv1beta1.GetPermissionResponse
		wantErr error
	}{
		{
			name: "should return internal error if permission service return some error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testPermissions[testPermissionIdx].ID).Return(permission.Permission{}, errors.New("test error"))
			},
			request: &frontierv1beta1.GetPermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return not found error if permission id not exist",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testPermissions[testPermissionIdx].ID).Return(permission.Permission{}, permission.ErrNotExist)
			},
			request: &frontierv1beta1.GetPermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
			},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return not found error if permission id is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(permission.Permission{}, permission.ErrInvalidID)
			},
			request: &frontierv1beta1.GetPermissionRequest{},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"),
					testPermissions[testPermissionIdx].ID).Return(testPermissions[testPermissionIdx], nil)
			},
			request: &frontierv1beta1.GetPermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
			},
			want: &frontierv1beta1.GetPermissionResponse{
				Permission: &frontierv1beta1.Permission{
					Id:        testPermissions[testPermissionIdx].ID,
					Name:      testPermissions[testPermissionIdx].Name,
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
					Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name),
					CreatedAt: timestamppb.New(testPermissions[testPermissionIdx].CreatedAt),
					UpdatedAt: timestamppb.New(testPermissions[testPermissionIdx].UpdatedAt),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPermissionSrv := new(mocks.PermissionService)
			if tt.setup != nil {
				tt.setup(mockPermissionSrv)
			}
			mockDep := Handler{permissionService: mockPermissionSrv}
			resp, err := mockDep.GetPermission(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdatePermission(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.PermissionService)
		request *frontierv1beta1.UpdatePermissionRequest
		want    *frontierv1beta1.UpdatePermissionResponse
		wantErr error
	}{
		{
			name: "should return internal error if permission service return some error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(permission.Permission{}, errors.New("test error"))
			},
			request: &frontierv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &frontierv1beta1.PermissionRequestBody{
					Name:      testPermissions[testPermissionIdx].Name,
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return not found error if permission id not exist",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, permission.ErrNotExist)
			},
			request: &frontierv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &frontierv1beta1.PermissionRequestBody{
					Name:      testPermissions[testPermissionIdx].Name,
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return not found error if permission id is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), permission.Permission{
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, permission.ErrInvalidID)
			},
			request: &frontierv1beta1.UpdatePermissionRequest{
				Body: &frontierv1beta1.PermissionRequestBody{
					Name:      testPermissions[testPermissionIdx].Name,
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return bad request error if namespace id not exist",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, namespace.ErrNotExist)
			},
			request: &frontierv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &frontierv1beta1.PermissionRequestBody{
					Name:      testPermissions[testPermissionIdx].Name,
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if name is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, permission.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &frontierv1beta1.PermissionRequestBody{
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(testPermissions[testPermissionIdx], nil)
			},
			request: &frontierv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &frontierv1beta1.PermissionRequestBody{
					Name:      testPermissions[testPermissionIdx].Name,
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want: &frontierv1beta1.UpdatePermissionResponse{
				Permission: &frontierv1beta1.Permission{
					Id:        testPermissions[testPermissionIdx].ID,
					Name:      testPermissions[testPermissionIdx].Name,
					Namespace: testPermissions[testPermissionIdx].NamespaceID,
					Key:       schema.PermissionKeyFromNamespaceAndName(testPermissions[testPermissionIdx].NamespaceID, testPermissions[testPermissionIdx].Name),
					CreatedAt: timestamppb.New(testPermissions[testPermissionIdx].CreatedAt),
					UpdatedAt: timestamppb.New(testPermissions[testPermissionIdx].UpdatedAt),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPermissionSrv := new(mocks.PermissionService)
			if tt.setup != nil {
				tt.setup(mockPermissionSrv)
			}
			mockDep := Handler{permissionService: mockPermissionSrv}
			resp, err := mockDep.UpdatePermission(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
