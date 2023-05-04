package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/odpf/shield/core/permission"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	testPermissionIdx = 0
	testPermissions   = []permission.Permission{
		{
			ID:          uuid.New().String(),
			Name:        "Read",
			NamespaceID: "resource-1",
			Metadata:    map[string]any{},
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Write",
			NamespaceID: "resource-1",
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
		{
			ID:          uuid.New().String(),
			Name:        "Manage",
			NamespaceID: "resource-1",
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		},
	}
)

func TestListPermissions(t *testing.T) {
	table := []struct {
		title string
		setup func(as *mocks.PermissionService)
		req   *shieldv1beta1.ListPermissionsRequest
		want  *shieldv1beta1.ListPermissionsResponse
		err   error
	}{
		{
			title: "should return internal error if action service return some error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().List(mock.Anything).Return([]permission.Permission{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService) {
				var testPermissionList []permission.Permission
				for _, act := range testPermissions {
					testPermissionList = append(testPermissionList, act)
				}
				as.EXPECT().List(mock.Anything).Return(testPermissionList, nil)
			},
			want: &shieldv1beta1.ListPermissionsResponse{Permissions: []*shieldv1beta1.Permission{
				{
					Id:          testPermissions[0].ID,
					Name:        testPermissions[0].Name,
					NamespaceId: testPermissions[0].NamespaceID,
				},
				{
					Id:          testPermissions[1].ID,
					Name:        testPermissions[1].Name,
					NamespaceId: testPermissions[1].NamespaceID,
				},
				{
					Id:          testPermissions[2].ID,
					Name:        testPermissions[2].Name,
					NamespaceId: testPermissions[2].NamespaceID,
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
		setup func(as *mocks.PermissionService)
		req   *shieldv1beta1.CreatePermissionRequest
		want  *shieldv1beta1.CreatePermissionResponse
		err   error
	}{
		{
			title: "should return internal error if permission service return some error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(permission.Permission{}, errors.New("some error"))
			},
			req: &shieldv1beta1.CreatePermissionRequest{
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return bad request error if namespace id is wrong",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(permission.Permission{}, namespace.ErrNotExist)
			},
			req: &shieldv1beta1.CreatePermissionRequest{
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return bad request error if if id is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(permission.Permission{}, permission.ErrInvalidID)
			},
			req: &shieldv1beta1.CreatePermissionRequest{
				Body: &shieldv1beta1.PermissionRequestBody{
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return bad request error if if name is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(permission.Permission{}, permission.ErrInvalidDetail)
			},
			req: &shieldv1beta1.CreatePermissionRequest{
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}, nil)
			},
			req: &shieldv1beta1.CreatePermissionRequest{Body: &shieldv1beta1.PermissionRequestBody{
				Id:          testPermissions[testPermissionIdx].ID,
				Name:        testPermissions[testPermissionIdx].Name,
				NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
			}},
			want: &shieldv1beta1.CreatePermissionResponse{Permission: &shieldv1beta1.Permission{
				Id:          testPermissions[testPermissionIdx].ID,
				Name:        testPermissions[testPermissionIdx].Name,
				NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
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
		request *shieldv1beta1.GetPermissionRequest
		want    *shieldv1beta1.GetPermissionResponse
		wantErr error
	}{
		{
			name: "should return internal error if permission service return some error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testPermissions[testPermissionIdx].ID).Return(permission.Permission{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetPermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if permission id not exist",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testPermissions[testPermissionIdx].ID).Return(permission.Permission{}, permission.ErrNotExist)
			},
			request: &shieldv1beta1.GetPermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
			},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return not found error if permission id is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(permission.Permission{}, permission.ErrInvalidID)
			},
			request: &shieldv1beta1.GetPermissionRequest{},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testPermissions[testPermissionIdx].ID).Return(testPermissions[testPermissionIdx], nil)
			},
			request: &shieldv1beta1.GetPermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
			},
			want: &shieldv1beta1.GetPermissionResponse{
				Permission: &shieldv1beta1.Permission{
					Id:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
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
		request *shieldv1beta1.UpdatePermissionRequest
		want    *shieldv1beta1.UpdatePermissionResponse
		wantErr error
	}{
		{
			name: "should return internal error if permission service return some error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(permission.Permission{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &shieldv1beta1.PermissionRequestBody{
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if permission id not exist",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, permission.ErrNotExist)
			},
			request: &shieldv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return not found error if permission id is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, permission.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdatePermissionRequest{
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID, // id in body is being ignored
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcPermissionNotFoundErr,
		},
		{
			name: "should return bad request error if namespace id not exist",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, namespace.ErrNotExist)
			},
			request: &shieldv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if name is empty",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID}).Return(permission.Permission{}, permission.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if permission service return nil error",
			setup: func(as *mocks.PermissionService) {
				as.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), permission.Permission{
					ID:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceID: testPermissions[testPermissionIdx].NamespaceID,
				}).Return(testPermissions[testPermissionIdx], nil)
			},
			request: &shieldv1beta1.UpdatePermissionRequest{
				Id: testPermissions[testPermissionIdx].ID,
				Body: &shieldv1beta1.PermissionRequestBody{
					Id:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
				},
			},
			want: &shieldv1beta1.UpdatePermissionResponse{
				Permission: &shieldv1beta1.Permission{
					Id:          testPermissions[testPermissionIdx].ID,
					Name:        testPermissions[testPermissionIdx].Name,
					NamespaceId: testPermissions[testPermissionIdx].NamespaceID,
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
