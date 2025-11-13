package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/metaschema"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectHandler_ListMetaSchemas(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *connect.Request[frontierv1beta1.ListMetaSchemasRequest]
		want    *connect.Response[frontierv1beta1.ListMetaSchemasResponse]
		wantErr error
	}{
		{
			name: "should list meta schemas on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().List(mock.Anything).Return([]metaschema.MetaSchema{
					{
						ID:        "some_id",
						Name:      "domain_name",
						Schema:    "schema_1",
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListMetaSchemasRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListMetaSchemasResponse{
				Metaschemas: []*frontierv1beta1.MetaSchema{
					{
						Id:        "some_id",
						Name:      "domain_name",
						Schema:    "schema_1",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return error when service fails",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().List(mock.Anything).Return(nil, errors.New("service error"))
			},
			req:     connect.NewRequest(&frontierv1beta1.ListMetaSchemasRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return empty list when no metaschemas exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().List(mock.Anything).Return([]metaschema.MetaSchema{}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.ListMetaSchemasRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListMetaSchemasResponse{
				Metaschemas: nil,
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaSchemaServ := new(mocks.MetaSchemaService)
			tt.setup(mockMetaSchemaServ)
			h := &ConnectHandler{
				metaSchemaService: mockMetaSchemaServ,
			}
			got, err := h.ListMetaSchemas(context.Background(), tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConnectHandler_CreateMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *connect.Request[frontierv1beta1.CreateMetaSchemaRequest]
		want    *connect.Response[frontierv1beta1.CreateMetaSchemaResponse]
		wantErr error
	}{
		{
			name: "should create metaschema on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Create(mock.Anything, metaschema.MetaSchema{
					Name:   "test_schema",
					Schema: "test_schema_body",
				}).Return(metaschema.MetaSchema{
					ID:        "created_id",
					Name:      "test_schema",
					Schema:    "test_schema_body",
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateMetaSchemaRequest{
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "test_schema",
					Schema: "test_schema_body",
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateMetaSchemaResponse{
				Metaschema: &frontierv1beta1.MetaSchema{
					Id:        "created_id",
					Name:      "test_schema",
					Schema:    "test_schema_body",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return bad request when body is nil",
			setup: func(m *mocks.MetaSchemaService) {
				// No expectation since validation happens before service call
			},
			req:     connect.NewRequest(&frontierv1beta1.CreateMetaSchemaRequest{Body: nil}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return conflict error when metaschema already exists",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(metaschema.MetaSchema{}, metaschema.ErrConflict)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateMetaSchemaRequest{
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "existing_schema",
					Schema: "test_schema_body",
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return bad request for invalid detail error",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(metaschema.MetaSchema{}, metaschema.ErrInvalidDetail)
			},
			req: connect.NewRequest(&frontierv1beta1.CreateMetaSchemaRequest{
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "invalid_schema",
					Schema: "invalid_schema_body",
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaSchemaServ := new(mocks.MetaSchemaService)
			tt.setup(mockMetaSchemaServ)
			h := &ConnectHandler{
				metaSchemaService: mockMetaSchemaServ,
			}
			got, err := h.CreateMetaSchema(context.Background(), tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConnectHandler_GetMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *connect.Request[frontierv1beta1.GetMetaSchemaRequest]
		want    *connect.Response[frontierv1beta1.GetMetaSchemaResponse]
		wantErr error
	}{
		{
			name: "should get metaschema on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Get(mock.Anything, "test_id").Return(metaschema.MetaSchema{
					ID:        "test_id",
					Name:      "test_schema",
					Schema:    "test_schema_body",
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.GetMetaSchemaRequest{Id: "test_id"}),
			want: connect.NewResponse(&frontierv1beta1.GetMetaSchemaResponse{
				Metaschema: &frontierv1beta1.MetaSchema{
					Id:        "test_id",
					Name:      "test_schema",
					Schema:    "test_schema_body",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return not found when id is empty",
			setup: func(m *mocks.MetaSchemaService) {
				// No expectation since validation happens before service call
			},
			req:     connect.NewRequest(&frontierv1beta1.GetMetaSchemaRequest{Id: ""}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound),
		},
		{
			name: "should return not found when metaschema doesn't exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Get(mock.Anything, "nonexistent_id").Return(metaschema.MetaSchema{}, metaschema.ErrNotExist)
			},
			req:     connect.NewRequest(&frontierv1beta1.GetMetaSchemaRequest{Id: "nonexistent_id"}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound),
		},
		{
			name: "should return internal error when service fails",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Get(mock.Anything, "test_id").Return(metaschema.MetaSchema{}, errors.New("service error"))
			},
			req:     connect.NewRequest(&frontierv1beta1.GetMetaSchemaRequest{Id: "test_id"}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaSchemaServ := new(mocks.MetaSchemaService)
			tt.setup(mockMetaSchemaServ)
			h := &ConnectHandler{
				metaSchemaService: mockMetaSchemaServ,
			}
			got, err := h.GetMetaSchema(context.Background(), tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConnectHandler_UpdateMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *connect.Request[frontierv1beta1.UpdateMetaSchemaRequest]
		want    *connect.Response[frontierv1beta1.UpdateMetaSchemaResponse]
		wantErr error
	}{
		{
			name: "should update metaschema on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.Anything, "test_id", metaschema.MetaSchema{
					Name:   "updated_schema",
					Schema: "updated_schema_body",
				}).Return(metaschema.MetaSchema{
					ID:        "test_id",
					Name:      "updated_schema",
					Schema:    "updated_schema_body",
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}, nil)
			},
			req: connect.NewRequest(&frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "test_id",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "updated_schema",
					Schema: "updated_schema_body",
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateMetaSchemaResponse{
				Metaschema: &frontierv1beta1.MetaSchema{
					Id:        "test_id",
					Name:      "updated_schema",
					Schema:    "updated_schema_body",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return not found when id is empty",
			setup: func(m *mocks.MetaSchemaService) {
				// No expectation since validation happens before service call
			},
			req: connect.NewRequest(&frontierv1beta1.UpdateMetaSchemaRequest{
				Id:   "",
				Body: &frontierv1beta1.MetaSchemaRequestBody{Name: "test", Schema: "test"},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound),
		},
		{
			name: "should return bad request when body is nil",
			setup: func(m *mocks.MetaSchemaService) {
				// No expectation since validation happens before service call
			},
			req: connect.NewRequest(&frontierv1beta1.UpdateMetaSchemaRequest{
				Id:   "test_id",
				Body: nil,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return not found when metaschema doesn't exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.Anything, "nonexistent_id", mock.Anything).Return(metaschema.MetaSchema{}, metaschema.ErrNotExist)
			},
			req: connect.NewRequest(&frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "nonexistent_id",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "test",
					Schema: "test",
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaSchemaServ := new(mocks.MetaSchemaService)
			tt.setup(mockMetaSchemaServ)
			h := &ConnectHandler{
				metaSchemaService: mockMetaSchemaServ,
			}
			got, err := h.UpdateMetaSchema(context.Background(), tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConnectHandler_DeleteMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *connect.Request[frontierv1beta1.DeleteMetaSchemaRequest]
		want    *connect.Response[frontierv1beta1.DeleteMetaSchemaResponse]
		wantErr error
	}{
		{
			name: "should delete metaschema on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Delete(mock.Anything, "test_id").Return(nil)
			},
			req:     connect.NewRequest(&frontierv1beta1.DeleteMetaSchemaRequest{Id: "test_id"}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteMetaSchemaResponse{}),
			wantErr: nil,
		},
		{
			name: "should return not found when id is empty",
			setup: func(m *mocks.MetaSchemaService) {
				// No expectation since validation happens before service call
			},
			req:     connect.NewRequest(&frontierv1beta1.DeleteMetaSchemaRequest{Id: ""}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound),
		},
		{
			name: "should return not found when metaschema doesn't exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Delete(mock.Anything, "nonexistent_id").Return(metaschema.ErrNotExist)
			},
			req:     connect.NewRequest(&frontierv1beta1.DeleteMetaSchemaRequest{Id: "nonexistent_id"}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrMetaschemaNotFound),
		},
		{
			name: "should return internal error when service fails",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Delete(mock.Anything, "test_id").Return(errors.New("service error"))
			},
			req:     connect.NewRequest(&frontierv1beta1.DeleteMetaSchemaRequest{Id: "test_id"}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaSchemaServ := new(mocks.MetaSchemaService)
			tt.setup(mockMetaSchemaServ)
			h := &ConnectHandler{
				metaSchemaService: mockMetaSchemaServ,
			}
			got, err := h.DeleteMetaSchema(context.Background(), tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
