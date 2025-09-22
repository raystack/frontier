package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testOrgMap = map[string]organization.Organization{
		"9f256f86-31a3-11ec-8d3d-0242ac130003": {
			ID:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
			Name:  "org-1",
			State: organization.Enabled,
			Metadata: metadata.Metadata{
				"email":  "org1@org1.com",
				"age":    21,
				"intern": true,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
)

const DefaultPageSize = 1000

func TestHandler_ListOrganizations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.ListOrganizationsRequest]
		want    *connect.Response[frontierv1beta1.ListOrganizationsResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), organization.Filter{
					Pagination: pagination.NewPagination(1, DefaultPageSize)},
				).Return([]organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationsRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success if org service return nil",
			setup: func(os *mocks.OrganizationService) {
				var testOrgList []organization.Organization
				for _, o := range testOrgMap {
					testOrgList = append(testOrgList, o)
				}
				os.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), organization.Filter{
					Pagination: pagination.NewPagination(1, DefaultPageSize),
				}).Return(testOrgList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationsRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationsResponse{Organizations: []*frontierv1beta1.Organization{
				{
					Id:   "9f256f86-31a3-11ec-8d3d-0242ac130003",
					Name: "org-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email":  structpb.NewStringValue("org1@org1.com"),
							"age":    structpb.NewNumberValue(21),
							"intern": structpb.NewBoolValue(true),
						},
					},
					State:     "enabled",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockOrgSrv)
			}
			mockDep := &ConnectHandler{orgService: mockOrgSrv}
			resp, err := mockDep.ListOrganizations(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateOrganization(t *testing.T) {
	email := "user@raystack.org"
	tests := []struct {
		name    string
		setup   func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context
		request *connect.Request[frontierv1beta1.CreateOrganizationRequest]
		want    *connect.Response[frontierv1beta1.CreateOrganizationResponse]
		wantErr error
	}{
		{
			name: "should return error if meta schema service gives error",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(errors.New("meta schema error"))
				return ctx
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Metadata: &structpb.Struct{},
			}}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError),
		},
		{
			name: "should return unauthenticated error if auth email in context is empty and org service return invalid user email",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					Name:     "some-org",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, user.ErrInvalidEmail)
				return ctx
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "some-org",
				Metadata: &structpb.Struct{},
			}}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			name: "should return internal error if org service return some error",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), organization.Organization{
					Name:     "abc",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, errors.New("test error"))
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "abc",
				Metadata: &structpb.Struct{},
			}}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return invalid argument error if name is empty",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), organization.Organization{
					Name:     "abc",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, organization.ErrInvalidDetail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "abc",
				Metadata: &structpb.Struct{},
			}}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return already exists error if org service return error conflict",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), organization.Organization{
					Name:     "abc",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, organization.ErrConflict)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "abc",
				Metadata: &structpb.Struct{},
			}}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return success if org service return nil error",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), organization.Organization{
					Name: "some-org",
					Metadata: metadata.Metadata{
						"email": "a",
					},
				}).Return(organization.Organization{
					ID:   "new-abc",
					Name: "some-org",
					Metadata: metadata.Metadata{
						"email": "a",
					},
					State: organization.Enabled,
				}, nil)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "some-org",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("a"),
					},
				},
			}}),
			want: connect.NewResponse(&frontierv1beta1.CreateOrganizationResponse{Organization: &frontierv1beta1.Organization{
				Id:   "new-abc",
				Name: "some-org",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("a"),
					}},
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
				State:     "enabled",
				Avatar:    "",
			}}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockOrgSrv, mockMetaSchemaSvc)
			}
			mockDep := &ConnectHandler{orgService: mockOrgSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.CreateOrganization(ctx, tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateOrganization(t *testing.T) {
	someOrgID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService)
		request *connect.Request[frontierv1beta1.UpdateOrganizationRequest]
		want    *connect.Response[frontierv1beta1.UpdateOrganizationResponse]
		wantErr error
	}{
		{
			name: "should return bad request error if request body is nil",
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Id:   someOrgID,
				Body: nil,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return error if meta schema service gives error",
			setup: func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(errors.New("meta schema error"))
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Id: someOrgID,
				Body: &frontierv1beta1.OrganizationRequestBody{
					Name:     "new-org",
					Metadata: &structpb.Struct{},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError),
		},
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					ID: someOrgID,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
					Name: "new-org",
				}).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Id: someOrgID,
				Body: &frontierv1beta1.OrganizationRequestBody{
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org id is not uuid (slug) and not exist",
			setup: func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					ID: someOrgID,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
					Name: "new-org",
				}).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Id: someOrgID,
				Body: &frontierv1beta1.OrganizationRequestBody{
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if org id is empty",
			setup: func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					Name: "new-org",
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
				}).Return(organization.Organization{}, organization.ErrInvalidID)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Body: &frontierv1beta1.OrganizationRequestBody{
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return already exist error if org service return err conflict",
			setup: func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					ID: someOrgID,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
					Name: "new-org",
				}).Return(organization.Organization{}, organization.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Id: someOrgID,
				Body: &frontierv1beta1.OrganizationRequestBody{
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest),
		},
		{
			name: "should return success if org service is updated by id and return nil error",
			setup: func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					ID: someOrgID,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
					Name: "new-org",
				}).Return(organization.Organization{
					ID: someOrgID,
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
					Name:  "new-org",
					State: organization.Enabled,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Id: someOrgID,
				Body: &frontierv1beta1.OrganizationRequestBody{
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateOrganizationResponse{
				Organization: &frontierv1beta1.Organization{
					Id:   someOrgID,
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
					State:     "enabled",
					Avatar:    "",
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return success if org service is updated by name and return nil error",
			setup: func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService) {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Update(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					Name: "new-org",
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
				}).Return(organization.Organization{
					ID:   someOrgID,
					Name: "new-org",
					Metadata: metadata.Metadata{
						"email": "org1@org1.com",
						"age":   float64(21),
						"valid": true,
					},
					State: organization.Enabled,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
				Body: &frontierv1beta1.OrganizationRequestBody{
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.UpdateOrganizationResponse{
				Organization: &frontierv1beta1.Organization{
					Id:   someOrgID,
					Name: "new-org",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
							"age":   structpb.NewNumberValue(21),
							"valid": structpb.NewBoolValue(true),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
					State:     "enabled",
					Avatar:    "",
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			if tt.setup != nil {
				tt.setup(mockOrgSrv, mockMetaSchemaSvc)
			}
			mockDep := &ConnectHandler{orgService: mockOrgSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.UpdateOrganization(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
