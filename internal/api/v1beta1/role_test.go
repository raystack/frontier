package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/core/role"
	"github.com/raystack/shield/internal/api/v1beta1/mocks"
	"github.com/raystack/shield/pkg/metadata"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
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
			Permissions: []string{
				"member",
				"user",
			},
			OrgID: uuid.New().String(),
			Metadata: metadata.Metadata{
				"foo": "bar",
			},
		},
	}
)

func TestHandler_ListOrganizationRoles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *shieldv1beta1.ListOrganizationRolesRequest
		want    *shieldv1beta1.ListOrganizationRolesResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), role.Filter{}).Return([]role.Role{}, errors.New("some error"))
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
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), role.Filter{}).Return(testRolesList, nil)
			},
			want: &shieldv1beta1.ListOrganizationRolesResponse{
				Roles: []*shieldv1beta1.Role{
					{
						Id:          testRoleMap[testRoleID].ID,
						Name:        testRoleMap[testRoleID].Name,
						Permissions: testRoleMap[testRoleID].Permissions,
						OrgId:       testRoleMap[testRoleID].OrgID,
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
			resp, err := mockDep.ListOrganizationRoles(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateOrganizationRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService, ms *mocks.MetaSchemaService)
		request *shieldv1beta1.CreateOrganizationRoleRequest
		want    *shieldv1beta1.CreateOrganizationRoleResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, errors.New("some error"))
			},
			request: &shieldv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
				rs.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, namespace.ErrNotExist)
			},
			request: &shieldv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
				rs.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &shieldv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Permissions: testRoleMap[testRoleID].Permissions,
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
				rs.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidID)
			}, request: &shieldv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
				rs.EXPECT().Upsert(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(testRoleMap[testRoleID], nil)
			},
			request: &shieldv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: &shieldv1beta1.CreateOrganizationRoleResponse{
				Role: &shieldv1beta1.Role{
					Id:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgId:       testRoleMap[testRoleID].OrgID,
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
			resp, err := mockDep.CreateOrganizationRole(context.Background(), tt.request)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			}
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetOrganizationRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *shieldv1beta1.GetOrganizationRoleRequest
		want    *shieldv1beta1.GetOrganizationRoleResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRoleID).Return(role.Role{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetOrganizationRoleRequest{
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
			request: &shieldv1beta1.GetOrganizationRoleRequest{
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
			request: &shieldv1beta1.GetOrganizationRoleRequest{},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return success if role service return nil error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRoleID).Return(testRoleMap[testRoleID], nil)
			},
			request: &shieldv1beta1.GetOrganizationRoleRequest{
				Id: testRoleID,
			},
			want: &shieldv1beta1.GetOrganizationRoleResponse{
				Role: &shieldv1beta1.Role{
					Id:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgId:       testRoleMap[testRoleID].OrgID,
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
			resp, err := mockDep.GetOrganizationRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateOrganizationRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService, ms *mocks.MetaSchemaService)
		request *shieldv1beta1.UpdateOrganizationRoleRequest
		want    *shieldv1beta1.UpdateOrganizationRoleResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Permissions: testRoleMap[testRoleID].Permissions,
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
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrConflict)
			},
			request: &shieldv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &shieldv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
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
			resp, err := mockDep.UpdateOrganizationRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
