package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testRoleID          = uuid.New().String()
	instanceLevelRoleID = uuid.New().String()
	testRoleMap         = map[string]role.Role{
		testRoleID: {
			ID:    testRoleID,
			Name:  "a new role",
			Title: "Test Title",
			Permissions: []string{
				"member",
				"user",
			},
			OrgID: uuid.New().String(),
			Metadata: metadata.Metadata{
				"foo": "bar",
			},
		},
		instanceLevelRoleID: {
			ID:   uuid.NewString(),
			Name: "a new role",
			Permissions: []string{
				"member",
				"user",
			},
			OrgID: uuid.Nil.String(),
			Metadata: metadata.Metadata{
				"foo": "bar",
			},
		},
	}
	testInstanceLevelRolePB = frontierv1beta1.Role{
		Id:          testRoleMap[instanceLevelRoleID].ID,
		Name:        testRoleMap[instanceLevelRoleID].Name,
		Permissions: testRoleMap[instanceLevelRoleID].Permissions,
		OrgId:       testRoleMap[instanceLevelRoleID].OrgID,
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"foo": {
					Kind: &structpb.Value_StringValue{
						StringValue: "bar",
					},
				},
			},
		},
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
	testRolePB = frontierv1beta1.Role{
		Id:          testRoleMap[testRoleID].ID,
		Name:        testRoleMap[testRoleID].Name,
		Title:       testRoleMap[testRoleID].Title,
		Permissions: testRoleMap[testRoleID].Permissions,
		OrgId:       testRoleMap[testRoleID].OrgID,
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"foo": {
					Kind: &structpb.Value_StringValue{
						StringValue: "bar",
					},
				},
			},
		},
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
)

func TestHandler_ListOrganizationRoles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *frontierv1beta1.ListOrganizationRolesRequest
		want    *frontierv1beta1.ListOrganizationRolesResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), role.Filter{}).Return([]role.Role{}, errors.New("test error"))
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return success if role service return nil error",
			setup: func(rs *mocks.RoleService) {
				var testRolesList []role.Role
				for _, rl := range testRoleMap {
					if rl.OrgID == uuid.Nil.String() {
						continue
					}
					testRolesList = append(testRolesList, rl)
				}
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), role.Filter{}).Return(testRolesList, nil)
			},
			want: &frontierv1beta1.ListOrganizationRolesResponse{
				Roles: []*frontierv1beta1.Role{
					{
						Id:          testRoleMap[testRoleID].ID,
						Name:        testRoleMap[testRoleID].Name,
						Title:       testRoleMap[testRoleID].Title,
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
		request *frontierv1beta1.CreateOrganizationRoleRequest
		want    *frontierv1beta1.CreateOrganizationRoleResponse
		wantErr error
	}{
		{
			name: "should return error if org id is not uuid",
			request: &frontierv1beta1.CreateOrganizationRoleRequest{
				OrgId: "not-uuid",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad body error if metaschema validation fails",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(errors.New("test error"))
			},
			request: &frontierv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyMetaSchemaError,
		},
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, errors.New("test error"))
			},
			request: &frontierv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
			wantErr: errors.New("test error"),
		},
		{
			name: "should return bad request error if namespace id not exist",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, namespace.ErrNotExist)
			},
			request: &frontierv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &frontierv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidID)
			}, request: &frontierv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(testRoleMap[testRoleID], nil)
			},
			request: &frontierv1beta1.CreateOrganizationRoleRequest{
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: &frontierv1beta1.CreateOrganizationRoleResponse{
				Role: &frontierv1beta1.Role{
					Id:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Title:       testRoleMap[testRoleID].Title,
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
		request *frontierv1beta1.GetOrganizationRoleRequest
		want    *frontierv1beta1.GetOrganizationRoleResponse
		wantErr error
	}{
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testRoleID).Return(role.Role{}, errors.New("test error"))
			},
			request: &frontierv1beta1.GetOrganizationRoleRequest{
				Id: testRoleID,
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testRoleID).Return(role.Role{}, role.ErrNotExist)
			},
			request: &frontierv1beta1.GetOrganizationRoleRequest{
				Id: testRoleID,
			},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return not found error if id empty",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(role.Role{}, role.ErrInvalidID)
			},
			request: &frontierv1beta1.GetOrganizationRoleRequest{},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return success if role service return nil error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testRoleID).Return(testRoleMap[testRoleID], nil)
			},
			request: &frontierv1beta1.GetOrganizationRoleRequest{
				Id: testRoleID,
			},
			want: &frontierv1beta1.GetOrganizationRoleResponse{
				Role: &frontierv1beta1.Role{
					Id:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Title:       testRoleMap[testRoleID].Title,
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
		request *frontierv1beta1.UpdateOrganizationRoleRequest
		want    *frontierv1beta1.UpdateOrganizationRoleResponse
		wantErr error
	}{
		{
			name: "should return bad body error if org id is not valid uuid",
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id: "not-valid-uuid",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad body error if permissions field is empty",
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
					Permissions: []string{},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return internal error if role service return some error",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, errors.New("test error"))
			},
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
			wantErr: errors.New("test error"),
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrNotExist)
			},
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidID)
			},
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrInvalidDetail)
			},
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(role.Role{}, role.ErrConflict)
			},
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
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
		{
			name: "should update role successfully",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), roleMetaSchema).Return(nil)
				rs.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), role.Role{
					ID:          testRoleMap[testRoleID].ID,
					Title:       testRoleMap[testRoleID].Title,
					Name:        testRoleMap[testRoleID].Name,
					Permissions: testRoleMap[testRoleID].Permissions,
					OrgID:       testRoleMap[testRoleID].OrgID,
					Metadata:    testRoleMap[testRoleID].Metadata,
				}).Return(testRoleMap[testRoleID], nil)
			},
			request: &frontierv1beta1.UpdateOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
				Body: &frontierv1beta1.RoleRequestBody{
					Name:        testRoleMap[testRoleID].Name,
					Title:       testRoleMap[testRoleID].Title,
					Permissions: testRoleMap[testRoleID].Permissions,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": structpb.NewStringValue("bar"),
						},
					},
				},
			},
			want: &frontierv1beta1.UpdateOrganizationRoleResponse{
				Role: &testRolePB,
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
			resp, err := mockDep.UpdateOrganizationRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteOrganizationRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *frontierv1beta1.DeleteOrganizationRoleRequest
		want    *frontierv1beta1.DeleteOrganizationRoleResponse
		wantErr error
	}{
		{
			name: "should return bad body error if org id or role id is not valid uuid",
			request: &frontierv1beta1.DeleteOrganizationRoleRequest{
				Id: "invalid-role-id",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return not found error if role service return err not found",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleMap[testRoleID].ID).Return(role.ErrNotExist)
			},
			request: &frontierv1beta1.DeleteOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
			},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return internal error if role service gives unknown error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleMap[testRoleID].ID).Return(errors.New("test error"))
			},
			request: &frontierv1beta1.DeleteOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return nil if role service return nil",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleMap[testRoleID].ID).Return(nil)
			},
			request: &frontierv1beta1.DeleteOrganizationRoleRequest{
				Id:    testRoleMap[testRoleID].ID,
				OrgId: testRoleMap[testRoleID].OrgID,
			},
			want:    &frontierv1beta1.DeleteOrganizationRoleResponse{},
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
			resp, err := mockDep.DeleteOrganizationRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *frontierv1beta1.DeleteRoleRequest
		want    *frontierv1beta1.DeleteRoleResponse
		wantErr error
	}{
		{
			name: "should return bad body error if role id is not valid uuid",
			request: &frontierv1beta1.DeleteRoleRequest{
				Id: "invalid-role-id",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return not found error if role service return err not found",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleMap[testRoleID].ID).Return(role.ErrNotExist)
			},
			request: &frontierv1beta1.DeleteRoleRequest{
				Id: testRoleMap[testRoleID].ID,
			},
			want:    nil,
			wantErr: grpcRoleNotFoundErr,
		},
		{
			name: "should return internal error if role service gives unknown error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleMap[testRoleID].ID).Return(errors.New("test error"))
			},
			request: &frontierv1beta1.DeleteRoleRequest{
				Id: testRoleMap[testRoleID].ID,
			},
			want:    nil,
			wantErr: errors.New("test error"),
		},
		{
			name: "should return nil if role service return nil",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleMap[testRoleID].ID).Return(nil)
			},
			request: &frontierv1beta1.DeleteRoleRequest{
				Id: testRoleMap[testRoleID].ID,
			},
			want:    &frontierv1beta1.DeleteRoleResponse{},
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
			resp, err := mockDep.DeleteRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListRoles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *frontierv1beta1.ListRolesRequest
		want    *frontierv1beta1.ListRolesResponse
		wantErr error
	}{
		{
			name: "should return instance level roles on success",
			setup: func(rs *mocks.RoleService) {
				var testRolesList []role.Role
				for _, rl := range testRoleMap {
					if rl.OrgID == uuid.Nil.String() {
						testRolesList = append(testRolesList, rl)
					}
				}
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), role.Filter{OrgID: testRoleMap[instanceLevelRoleID].OrgID}).Return(testRolesList, nil)
			},
			request: &frontierv1beta1.ListRolesRequest{},
			want: &frontierv1beta1.ListRolesResponse{
				Roles: []*frontierv1beta1.Role{
					{
						Id:          testRoleMap[instanceLevelRoleID].ID,
						OrgId:       testRoleMap[instanceLevelRoleID].OrgID,
						Name:        testRoleMap[instanceLevelRoleID].Name,
						Permissions: testRoleMap[instanceLevelRoleID].Permissions,
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{
							"foo": {Kind: &structpb.Value_StringValue{StringValue: "bar"}},
						}},
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
		request *frontierv1beta1.CreateRoleRequest
		want    *frontierv1beta1.CreateRoleResponse
		wantErr error
	}{
		{
			name: "should return bad body error if request body is empty",
			request: &frontierv1beta1.CreateRoleRequest{
				Body: nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should create role on success",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				expectedResp := testRoleMap[instanceLevelRoleID]
				expectedResp.ID = ""
				ms.EXPECT().Validate(testRoleMap[testRoleID].Metadata, roleMetaSchema).Return(nil)
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), expectedResp).Return(testRoleMap[instanceLevelRoleID], nil)
			},
			request: &frontierv1beta1.CreateRoleRequest{
				Body: &frontierv1beta1.RoleRequestBody{
					Name:        testRoleMap[instanceLevelRoleID].Name,
					Permissions: testRoleMap[instanceLevelRoleID].Permissions,
					Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{
						"foo": {Kind: &structpb.Value_StringValue{StringValue: "bar"}},
					}},
				},
			},
			want: &frontierv1beta1.CreateRoleResponse{
				Role: &testInstanceLevelRolePB,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv, mockMetaSrv)
			}
			mockDep := Handler{roleService: mockRoleSrv, metaSchemaService: mockMetaSrv}
			resp, err := mockDep.CreateRole(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
