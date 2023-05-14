package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/odpf/shield/pkg/metadata"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testRoleID  = "role-id"
	testRoleMap = map[string]role.Role{
		testRoleID: {
			ID:   testRoleID,
			Name: "a new role",
			Types: []string{
				"member",
				"user",
			},
			NamespaceID: "ns-1",
			Metadata: metadata.Metadata{
				"foo": "bar",
			},
		},
	}
)

func TestHandler_ListRoles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *shieldv1beta1.ListRolesRequest
		want    *shieldv1beta1.ListRolesResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return([]role.Role{}, errors.New("some error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return success if role service return nil error",
			setup: func(rs *mocks.RoleService) {
				var testRolesList []role.Role
				for _, rl := range testRoleMap {
					testRolesList = append(testRolesList, rl)
				}
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return(testRolesList, nil)
			},
			want: &shieldv1beta1.ListRolesResponse{
				Roles: []*shieldv1beta1.Role{
					{
						Id:    testRoleMap[testRoleID].ID,
						Name:  testRoleMap[testRoleID].Name,
						Types: testRoleMap[testRoleID].Types,
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"foo": structpb.NewStringValue("bar"),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv)
			}
			mockDep := Handler{roleService: mockRoleSrv}
			resp, err := mockDep.ListRoles(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService, ms *mocks.MetaSchemaService)
		request *shieldv1beta1.CreateRoleRequest
		want    *shieldv1beta1.CreateRoleResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, errors.New("some error"))
			},
			request: &shieldv1beta1.CreateRoleRequest{
				Body: &shieldv1beta1.RoleRequestBody{
					Id:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if namespace id not exist",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, namespace.ErrNotExist)
			},
			request: &shieldv1beta1.CreateRoleRequest{
				Body: &shieldv1beta1.RoleRequestBody{
					Id:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if name empty",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &shieldv1beta1.CreateRoleRequest{
				Body: &shieldv1beta1.RoleRequestBody{
					Id:          testRoleMap[testRoleID].ID,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if id empty",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidID)
			}, request: &shieldv1beta1.CreateRoleRequest{
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if role service return nil error",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(testRoleMap[testRoleID], nil)
			},
			request: &shieldv1beta1.CreateRoleRequest{
				Body: &shieldv1beta1.RoleRequestBody{
					Id:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: &shieldv1beta1.CreateRoleResponse{
				Role: &shieldv1beta1.Role{
					Id:    testRoleMap[testRoleID].ID,
					Name:  testRoleMap[testRoleID].Name,
					Types: testRoleMap[testRoleID].Types,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{roleService: mockRoleSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.CreateRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *shieldv1beta1.GetRoleRequest
		want    *shieldv1beta1.GetRoleResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRoleID).Return(role.Role{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetRoleRequest{
				Id: testRoleID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRoleID).Return(role.Role{}, role.ErrNotExist)
			},
			request: &shieldv1beta1.GetRoleRequest{
				Id: testRoleID,
			},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return not found error if id empty",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(role.Role{}, role.ErrInvalidID)
			},
			request: &shieldv1beta1.GetRoleRequest{},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return success if role service return nil error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRoleID).Return(testRoleMap[testRoleID], nil)
			},
			request: &shieldv1beta1.GetRoleRequest{
				Id: testRoleID,
			},
			want: &shieldv1beta1.GetRoleResponse{
				Role: &shieldv1beta1.Role{
					Id:    testRoleMap[testRoleID].ID,
					Name:  testRoleMap[testRoleID].Name,
					Types: testRoleMap[testRoleID].Types,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv)
			}
			mockDep := Handler{roleService: mockRoleSrv}
			resp, err := mockDep.GetRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService, ms *mocks.MetaSchemaService)
		request *shieldv1beta1.UpdateRoleRequest
		want    *shieldv1beta1.UpdateRoleResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateRoleRequest{
				Id: testRoleMap[testRoleID].ID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateRoleRequest{
				Id: testRoleMap[testRoleID].ID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdateRoleRequest{
				Id: testRoleMap[testRoleID].ID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return bad request error if name is empty",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateRoleRequest{
				Id: testRoleMap[testRoleID].ID,
				Body: &shieldv1beta1.RoleRequestBody{
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad request error if namespace id not exist",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateRoleRequest{
				Id: testRoleMap[testRoleID].ID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return already exist error if role service return err conflict",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceID: testRoleMap[testRoleID].NamespaceID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrConflict)
			},
			request: &shieldv1beta1.UpdateRoleRequest{
				Id: testRoleMap[testRoleID].ID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Types:       testRoleMap[testRoleID].Types,
					NamespaceId: testRoleMap[testRoleID].NamespaceID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{roleService: mockRoleSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.UpdateRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
