package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
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
)

func TestHandler_CreateRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService, ms *mocks.MetaSchemaService)
		request *connect.Request[frontierv1beta1.CreateRoleRequest]
		want    *connect.Response[frontierv1beta1.CreateRoleResponse]
		wantErr error
	}{
		{
			name: "should return bad body error if request body is empty",
			request: connect.NewRequest(&frontierv1beta1.CreateRoleRequest{
				Body: nil,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should create role on success",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				expectedResp := testRoleMap[instanceLevelRoleID]
				expectedResp.ID = ""
				ms.EXPECT().Validate(testRoleMap[testRoleID].Metadata, roleMetaSchema).Return(nil)
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), expectedResp).Return(testRoleMap[instanceLevelRoleID], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateRoleRequest{
				Body: &frontierv1beta1.RoleRequestBody{
					Name:        testRoleMap[instanceLevelRoleID].Name,
					Permissions: testRoleMap[instanceLevelRoleID].Permissions,
					Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{
						"foo": {Kind: &structpb.Value_StringValue{StringValue: "bar"}},
					}},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateRoleResponse{
				Role: &testInstanceLevelRolePB,
			}),
			wantErr: nil,
		},
		{
			name: "should return bad request error if namespace or permission doesn't exist",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				expectedResp := testRoleMap[instanceLevelRoleID]
				expectedResp.ID = ""
				ms.EXPECT().Validate(testRoleMap[testRoleID].Metadata, roleMetaSchema).Return(nil)
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), expectedResp).Return(role.Role{}, namespace.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateRoleRequest{
				Body: &frontierv1beta1.RoleRequestBody{
					Name:        testRoleMap[instanceLevelRoleID].Name,
					Permissions: testRoleMap[instanceLevelRoleID].Permissions,
					Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{
						"foo": {Kind: &structpb.Value_StringValue{StringValue: "bar"}},
					}},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return conflict error if role already exists",
			setup: func(rs *mocks.RoleService, ms *mocks.MetaSchemaService) {
				expectedResp := testRoleMap[instanceLevelRoleID]
				expectedResp.ID = ""
				ms.EXPECT().Validate(testRoleMap[testRoleID].Metadata, roleMetaSchema).Return(nil)
				rs.EXPECT().Upsert(mock.AnythingOfType("context.backgroundCtx"), expectedResp).Return(role.Role{}, role.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateRoleRequest{
				Body: &frontierv1beta1.RoleRequestBody{
					Name:        testRoleMap[instanceLevelRoleID].Name,
					Permissions: testRoleMap[instanceLevelRoleID].Permissions,
					Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{
						"foo": {Kind: &structpb.Value_StringValue{StringValue: "bar"}},
					}},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, role.ErrConflict),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			mockMetaSrv := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv, mockMetaSrv)
			}
			mockDep := &ConnectHandler{roleService: mockRoleSrv, metaSchemaService: mockMetaSrv}
			resp, err := mockDep.CreateRole(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListRoles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *connect.Request[frontierv1beta1.ListRolesRequest]
		want    *connect.Response[frontierv1beta1.ListRolesResponse]
		wantErr error
	}{
		{
			name: "should return empty list if no roles exist",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), role.Filter{
					OrgID:  testRoleMap[instanceLevelRoleID].OrgID,
					Scopes: []string{},
				}).Return([]role.Role{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListRolesRequest{
				Scopes: []string{},
			}),
			want: connect.NewResponse(&frontierv1beta1.ListRolesResponse{
				Roles: nil,
			}),
			wantErr: nil,
		},
		{
			name: "should return list of roles on success",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), role.Filter{
					OrgID:  testRoleMap[instanceLevelRoleID].OrgID,
					Scopes: []string{"test"},
				}).Return([]role.Role{testRoleMap[instanceLevelRoleID]}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListRolesRequest{
				Scopes: []string{"test"},
			}),
			want: connect.NewResponse(&frontierv1beta1.ListRolesResponse{
				Roles: []*frontierv1beta1.Role{&testInstanceLevelRolePB},
			}),
			wantErr: nil,
		},
		{
			name: "should return internal error if role service fails",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), role.Filter{
					OrgID:  testRoleMap[instanceLevelRoleID].OrgID,
					Scopes: []string{},
				}).Return(nil, ErrInternalServerError)
			},
			request: connect.NewRequest(&frontierv1beta1.ListRolesRequest{
				Scopes: []string{},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv)
			}
			mockDep := &ConnectHandler{roleService: mockRoleSrv}
			resp, err := mockDep.ListRoles(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteRole(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RoleService)
		request *connect.Request[frontierv1beta1.DeleteRoleRequest]
		want    *connect.Response[frontierv1beta1.DeleteRoleResponse]
		wantErr error
	}{
		{
			name: "should return bad body error if role id is not valid uuid",
			request: connect.NewRequest(&frontierv1beta1.DeleteRoleRequest{
				Id: "invalid-role-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return not found error if role service return err not found",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleID).Return(role.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteRoleRequest{
				Id: testRoleID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return internal error if role service gives unknown error",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleID).Return(ErrInternalServerError)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteRoleRequest{
				Id: testRoleID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return nil if role service return nil",
			setup: func(rs *mocks.RoleService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testRoleID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteRoleRequest{
				Id: testRoleID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteRoleResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoleSrv := new(mocks.RoleService)
			if tt.setup != nil {
				tt.setup(mockRoleSrv)
			}
			mockDep := &ConnectHandler{roleService: mockRoleSrv}
			resp, err := mockDep.DeleteRole(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
