package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
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
	testOrgID  = "9f256f86-31a3-11ec-8d3d-0242ac130003"
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

func TestHandler_ListOrganizationProjects(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ps *mocks.ProjectService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.ListOrganizationProjectsRequest]
		want    *connect.Response[frontierv1beta1.ListOrganizationProjectsResponse]
		wantErr error
	}{
		{
			name: "should return error if organization does not exist",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{
				Id: "some-org-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return error if organization is disabled",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return internal error if org service return some error",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return internal error if project service return some error",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{
					OrgID:           testOrgID,
					WithMemberCount: false,
				}).Return([]project.Project{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return list of projects successfully",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgMap[testOrgID].Name).Return(testOrgMap[testOrgID], nil)
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{
					OrgID:           testOrgID,
					WithMemberCount: false,
				}).Return([]project.Project{
					{
						ID:   "some-project-id",
						Name: "some-project-name",
						Metadata: metadata.Metadata{
							"foo": "bar",
						},
						Organization: organization.Organization{
							ID: testOrgID,
						},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{
				Id: testOrgMap[testOrgID].Name,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationProjectsResponse{
				Projects: []*frontierv1beta1.Project{
					{
						Id:    "some-project-id",
						Name:  "some-project-name",
						OrgId: testOrgID,
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
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return list of projects with member count successfully",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{
					OrgID:           testOrgID,
					WithMemberCount: true,
				}).Return([]project.Project{
					{
						ID:          "some-project-id",
						Name:        "some-project-name",
						MemberCount: 5,
						Metadata: metadata.Metadata{
							"foo": "bar",
						},
						Organization: organization.Organization{
							ID: testOrgID,
						},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{
				Id:              testOrgID,
				WithMemberCount: true,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationProjectsResponse{
				Projects: []*frontierv1beta1.Project{
					{
						Id:           "some-project-id",
						Name:         "some-project-name",
						OrgId:        testOrgID,
						MembersCount: 5,
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
					},
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectService := new(mocks.ProjectService)
			mockOrgService := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockProjectService, mockOrgService)
			}
			mockDep := &ConnectHandler{projectService: mockProjectService, orgService: mockOrgService}
			resp, err := mockDep.ListOrganizationProjects(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListOrganizationAdmins(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(us *mocks.UserService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.ListOrganizationAdminsRequest]
		want    *connect.Response[frontierv1beta1.ListOrganizationAdminsResponse]
		wantErr error
	}{
		{
			name: "should return internal error if user service return some error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return([]user.User{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return error if org id does not exist",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return error if org is disabled",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return internal error if org service return some error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success if org service return nil error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return(testUserList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationAdminsResponse{
				Users: []*frontierv1beta1.User{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "User 1",
						Name:  "user1",
						Email: "test@test.com",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"foo":    structpb.NewStringValue("bar"),
								"age":    structpb.NewNumberValue(21),
								"intern": structpb.NewBoolValue(true),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			mockOrgService := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockUserService, mockOrgService)
			}
			mockDep := &ConnectHandler{userService: mockUserService, orgService: mockOrgService}
			resp, err := mockDep.ListOrganizationAdmins(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListOrganizationUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(us *mocks.UserService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.ListOrganizationUsersRequest]
		want    *connect.Response[frontierv1beta1.ListOrganizationUsersResponse]
		wantErr error
	}{
		{
			name: "should return invalid argument error if role filters and with_roles are both provided",
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationUsersRequest{
				Id:          testOrgID,
				RoleFilters: []string{"admin"},
				WithRoles:   true,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return internal error if org service return some error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationUsersRequest{
				Id: "some-org-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org id does not exist",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationUsersRequest{
				Id: "some-org-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationUsersRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return internal error if user service return some error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, "").Return([]user.User{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationUsersRequest{
				Id: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return success if org service return nil error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, "").Return(testUserList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationUsersRequest{
				Id: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationUsersResponse{
				Users: []*frontierv1beta1.User{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "User 1",
						Name:  "user1",
						Email: "test@test.com",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"foo":    structpb.NewStringValue("bar"),
								"age":    structpb.NewNumberValue(21),
								"intern": structpb.NewBoolValue(true),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			mockOrgService := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockUserService, mockOrgService)
			}
			mockDep := &ConnectHandler{userService: mockUserService, orgService: mockOrgService}
			resp, err := mockDep.ListOrganizationUsers(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_AddOrganizationUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.AddOrganizationUsersRequest]
		want    *connect.Response[frontierv1beta1.AddOrganizationUsersResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.AddOrganizationUsersRequest{
				Id:      testOrgID,
				UserIds: []string{"some-user-id"},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.AddOrganizationUsersRequest{
				Id:      testOrgID,
				UserIds: []string{"some-user-id"},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.AddOrganizationUsersRequest{
				Id:      testOrgID,
				UserIds: []string{"some-user-id"},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return internal error if AddUsers fails",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				os.EXPECT().AddUsers(mock.AnythingOfType("context.backgroundCtx"), testOrgID, []string{"some-user-id"}).Return(errors.New("add users error"))
			},
			request: connect.NewRequest(&frontierv1beta1.AddOrganizationUsersRequest{
				Id:      testOrgID,
				UserIds: []string{"some-user-id"},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should add user to org successfully",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				os.EXPECT().AddUsers(mock.AnythingOfType("context.backgroundCtx"), testOrgID, []string{"some-user-id"}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.AddOrganizationUsersRequest{
				Id:      testOrgID,
				UserIds: []string{"some-user-id"},
			}),
			want:    connect.NewResponse(&frontierv1beta1.AddOrganizationUsersResponse{}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockOrgService)
			}
			mockDep := &ConnectHandler{orgService: mockOrgService}
			resp, err := mockDep.AddOrganizationUsers(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_RemoveOrganizationUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter)
		request *connect.Request[frontierv1beta1.RemoveOrganizationUserRequest]
		want    *connect.Response[frontierv1beta1.RemoveOrganizationUserResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return internal error if user service return some error",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return([]user.User{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return permission denied error and not remove user if it is the last admin user",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return([]user.User{
					testUserMap[testUserID],
				}, nil)
				// Note: deleterService should NOT be called when it's the last admin
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: testUserID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrMinAdminCount),
		},
		{
			name: "should return internal error if deleter service fails",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return([]user.User{
					testUserMap[testUserID],
					{
						ID:        "some-user-id",
						Title:     "User 1",
						Name:      "user1",
						Email:     "test@raystack.org",
						Metadata:  map[string]interface{}{},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
				ds.EXPECT().RemoveUsersFromOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, []string{"some-user-id"}).Return(errors.New("deleter error"))
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should remove user from org successfully",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return([]user.User{
					testUserMap[testUserID],
					{
						ID:        "some-user-id",
						Title:     "User 1",
						Name:      "user1",
						Email:     "test@raystack.org",
						Metadata:  map[string]interface{}{},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
				ds.EXPECT().RemoveUsersFromOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, []string{"some-user-id"}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			}),
			want:    connect.NewResponse(&frontierv1beta1.RemoveOrganizationUserResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockUserService := new(mocks.UserService)
			mockDeleterService := new(mocks.CascadeDeleter)
			if tt.setup != nil {
				tt.setup(mockOrgService, mockUserService, mockDeleterService)
			}
			mockDep := &ConnectHandler{
				orgService:     mockOrgService,
				userService:    mockUserService,
				deleterService: mockDeleterService,
			}
			resp, err := mockDep.RemoveOrganizationUser(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
