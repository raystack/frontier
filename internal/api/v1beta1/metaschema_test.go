package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/metaschema"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_ListMetaSchemas(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *frontierv1beta1.ListMetaSchemasRequest
		want    *frontierv1beta1.ListMetaSchemasResponse
		wantErr error
	}{
		{
			name: "Should list meta schemas on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx")).Return([]metaschema.MetaSchema{
					{
						ID:     "some_id",
						Name:   "domain_name",
						Schema: "schema_1",
					},
				}, nil)
			},
			req: &frontierv1beta1.ListMetaSchemasRequest{},
			want: &frontierv1beta1.ListMetaSchemasResponse{Metaschemas: []*frontierv1beta1.MetaSchema{
				{
					Id:        "some_id",
					Name:      "domain_name",
					Schema:    "schema_1",
					UpdatedAt: timestamppb.New(time.Time{}),
					CreatedAt: timestamppb.New(time.Time{}),
				},
			}},
			wantErr: nil,
		},
		{
			name: "should return an error if Meta schema service return some error",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().List(mock.AnythingOfType("context.backgroundCtx")).Return([]metaschema.MetaSchema{}, errors.New("some_err"))
			},
			req:     &frontierv1beta1.ListMetaSchemasRequest{},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaServ := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockMetaServ)
			}
			mockMeta := Handler{metaSchemaService: mockMetaServ}
			resp, err := mockMeta.ListMetaSchemas(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_CreateMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *frontierv1beta1.CreateMetaSchemaRequest
		want    *frontierv1beta1.CreateMetaSchemaResponse
		wantErr error
	}{
		{
			name: "should successfully create  Meta Schema",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{
					ID:     "some_id",
					Name:   "some_name",
					Schema: "some_schema",
				}, nil)
			},
			req: &frontierv1beta1.CreateMetaSchemaRequest{
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want: &frontierv1beta1.CreateMetaSchemaResponse{
				Metaschema: &frontierv1beta1.MetaSchema{
					Id:        "some_id",
					Name:      "some_name",
					Schema:    "some_schema",
					UpdatedAt: timestamppb.New(time.Time{}),
					CreatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
		{
			name: "should return error if the request body is nil",
			req: &frontierv1beta1.CreateMetaSchemaRequest{
				Body: nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return bad body error if meta scheme does not exist ",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, metaschema.ErrNotExist)
			},
			req: &frontierv1beta1.CreateMetaSchemaRequest{
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return conflict error if metaschema already exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, metaschema.ErrConflict)
			},
			req: &frontierv1beta1.CreateMetaSchemaRequest{
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return an error if meta scheme service return some error",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, errors.New("some_err"))
			},
			req: &frontierv1beta1.CreateMetaSchemaRequest{
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaServ := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockMetaServ)
			}
			mockMeta := Handler{metaSchemaService: mockMetaServ}
			resp, err := mockMeta.CreateMetaSchema(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_GetMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *frontierv1beta1.GetMetaSchemaRequest
		want    *frontierv1beta1.GetMetaSchemaResponse
		wantErr error
	}{
		{
			name: "should get meta schema on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some_id").Return(metaschema.MetaSchema{
					ID:     "some_id",
					Name:   "some_name",
					Schema: "some_schema",
				}, nil)
			},
			req: &frontierv1beta1.GetMetaSchemaRequest{
				Id: "some_id",
			},
			want: &frontierv1beta1.GetMetaSchemaResponse{
				Metaschema: &frontierv1beta1.MetaSchema{
					Id:        "some_id",
					Name:      "some_name",
					Schema:    "some_schema",
					UpdatedAt: timestamppb.New(time.Time{}),
					CreatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
		{
			name: "should return err if id is empty",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(metaschema.MetaSchema{}, grpcMetaSchemaNotFoundErr)
			},
			req: &frontierv1beta1.GetMetaSchemaRequest{
				Id: "",
			},
			want:    nil,
			wantErr: grpcMetaSchemaNotFoundErr,
		},
		{
			name: "should return error if ID is Invalid",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some_id").Return(metaschema.MetaSchema{}, metaschema.ErrInvalidID)
			},
			req: &frontierv1beta1.GetMetaSchemaRequest{
				Id: "some_id",
			},
			want:    nil,
			wantErr: grpcMetaSchemaNotFoundErr,
		},
		{
			name: "should return an error if meta schema service return some error",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some_id").Return(metaschema.MetaSchema{}, errors.New("some_error"))
			},
			req: &frontierv1beta1.GetMetaSchemaRequest{
				Id: "some_id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaServ := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockMetaServ)
			}
			mockMeta := Handler{metaSchemaService: mockMetaServ}
			resp, err := mockMeta.GetMetaSchema(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_UpdateMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *frontierv1beta1.UpdateMetaSchemaRequest
		want    *frontierv1beta1.UpdateMetaSchemaResponse
		wantErr error
	}{
		{
			name: "should update meta schema on  success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), "some_id", metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{
					ID:     "some_id",
					Name:   "some_name",
					Schema: "some_schema",
				}, nil)
			},
			req: &frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "some_id",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want: &frontierv1beta1.UpdateMetaSchemaResponse{
				Metaschema: &frontierv1beta1.MetaSchema{
					Id:        "some_id",
					Name:      "some_name",
					Schema:    "some_schema",
					UpdatedAt: timestamppb.New(time.Time{}),
					CreatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
		{
			name: "should return errorif meta schema doesn't exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), "", metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, nil)
			},
			req: &frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcMetaSchemaNotFoundErr,
		},
		{
			name: "should return error if request body is nil",
			req: &frontierv1beta1.UpdateMetaSchemaRequest{
				Id:   "some_id",
				Body: nil,
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return error if invalid metadata detail",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), "some_id", metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, metaschema.ErrInvalidDetail)
			},
			req: &frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "some_id",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return error if metaschema doesn't exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), "some_id", metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, metaschema.ErrNotExist)
			},
			req: &frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "some_id",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcMetaSchemaNotFoundErr,
		},
		{
			name: "should return error if metaschema already exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), "some_id", metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, metaschema.ErrConflict)
			},
			req: &frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "some_id",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return an error if Meta schema service return some error",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), "some_id", metaschema.MetaSchema{
					Name:   "some_name",
					Schema: "some_schema",
				}).Return(metaschema.MetaSchema{}, errors.New("some_err"))
			},
			req: &frontierv1beta1.UpdateMetaSchemaRequest{
				Id: "some_id",
				Body: &frontierv1beta1.MetaSchemaRequestBody{
					Name:   "some_name",
					Schema: "some_schema",
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaServ := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockMetaServ)
			}
			mockMeta := Handler{metaSchemaService: mockMetaServ}
			resp, err := mockMeta.UpdateMetaSchema(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_DeleteMetaSchema(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.MetaSchemaService)
		req     *frontierv1beta1.DeleteMetaSchemaRequest
		want    *frontierv1beta1.DeleteMetaSchemaResponse
		wantErr error
	}{
		{
			name: "should delete meta schema on success",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), "some_id").Return(nil)
			},
			req: &frontierv1beta1.DeleteMetaSchemaRequest{
				Id: "some_id",
			},
			want:    &frontierv1beta1.DeleteMetaSchemaResponse{},
			wantErr: nil,
		},
		{
			name: "should return error if Id is empty",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), "").Return(nil)
			},
			req: &frontierv1beta1.DeleteMetaSchemaRequest{
				Id: "",
			},
			want:    nil,
			wantErr: grpcMetaSchemaNotFoundErr,
		},
		{
			name: "should return error if metaschema doesn't exist",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), "some_id").Return(metaschema.ErrNotExist)
			},
			req: &frontierv1beta1.DeleteMetaSchemaRequest{
				Id: "some_id",
			},
			want:    nil,
			wantErr: grpcMetaSchemaNotFoundErr,
		},
		{
			name: "should return an error if Meta schema service return some error ",
			setup: func(m *mocks.MetaSchemaService) {
				m.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), "some_id").Return(errors.New("some_error"))
			},
			req: &frontierv1beta1.DeleteMetaSchemaRequest{
				Id: "some_id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaServ := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockMetaServ)
			}
			mockMeta := Handler{metaSchemaService: mockMetaServ}
			resp, err := mockMeta.DeleteMetaSchema(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
