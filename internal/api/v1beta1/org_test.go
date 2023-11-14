// Package v1beta1 provides v1beta1  
package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/serviceuser"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestHandler_ListOrganization(t *testing.T) {
	table := []struct {
		title string
		setup func(os *mocks.OrganizationService)
		want  *frontierv1beta1.ListOrganizationsResponse
		err   error
	}{
		{
			title: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), organization.Filter{}).Return([]organization.Organization{}, errors.New("some error"))
			},
			want: nil,
			err:  status.Errorf(codes.Internal, ErrInternalServer.Error()),
		},
		{
			title: "should return success if org service return nil",
			setup: func(os *mocks.OrganizationService) {
				var testOrgList []organization.Organization
				for _, o := range testOrgMap {
					testOrgList = append(testOrgList, o)
				}
				os.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), organization.Filter{}).Return(testOrgList, nil)
			},
			want: &frontierv1beta1.ListOrganizationsResponse{Organizations: []*frontierv1beta1.Organization{
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
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			mockDep := Handler{orgService: mockOrgSrv}
			if tt.setup != nil {
				tt.setup(mockOrgSrv)
			}
			resp, err := mockDep.ListOrganizations(context.Background(), nil)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestHandler_CreateOrganization(t *testing.T) {
	email := "user@raystack.org"
	table := []struct {
		title string
		setup func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.CreateOrganizationRequest
		want  *frontierv1beta1.CreateOrganizationResponse
		err   error
	}{
		{
			title: "should return error if meta schema service gives error",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(errors.New("grpcBadBodyMetaSchemaError"))
				return ctx
			},
			req: &frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcBadBodyMetaSchemaError,
		},
		{
			title: "should return forbidden error if auth email in context is empty and org service return invalid user email",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), organization.Organization{
					Name:     "some-org",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, user.ErrInvalidEmail)
				return ctx
			},
			req: &frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "some-org",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcUnauthenticated,
		},
		{
			title: "should return internal error if org service return some error",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), organization.Organization{
					Name:     "abc",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, errors.New("some error"))
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "abc",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return bad request error if name is empty",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), organization.Organization{
					Name:     "abc",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, organization.ErrInvalidDetail)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "abc",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcBadBodyError,
		},
		{
			title: "should return already exist error if org service return error conflict",
			setup: func(ctx context.Context, os *mocks.OrganizationService, ms *mocks.MetaSchemaService) context.Context {
				ms.EXPECT().Validate(mock.AnythingOfType("metadata.Metadata"), orgMetaSchema).Return(nil)
				os.EXPECT().Create(mock.AnythingOfType("*context.valueCtx"), organization.Organization{
					Name:     "abc",
					Metadata: metadata.Metadata{},
				}).Return(organization.Organization{}, organization.ErrConflict)
				return authenticate.SetContextWithEmail(ctx, email)
			},
			req: &frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name:     "abc",
				Metadata: &structpb.Struct{},
			}},
			want: nil,
			err:  grpcConflictError,
		},
		{
			title: "should return success if org service return nil error",
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
			req: &frontierv1beta1.CreateOrganizationRequest{Body: &frontierv1beta1.OrganizationRequestBody{
				Name: "some-org",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"email": structpb.NewStringValue("a"),
					},
				},
			}},
			want: &frontierv1beta1.CreateOrganizationResponse{Organization: &frontierv1beta1.Organization{
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
			}},
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.Background()
			if tt.setup != nil {
				ctx = tt.setup(ctx, mockOrgSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{orgService: mockOrgSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.CreateOrganization(ctx, tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

func TestHandler_GetOrganization(t *testing.T) {
	someOrgID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService)
		request *frontierv1beta1.GetOrganizationRequest
		want    *frontierv1beta1.GetOrganizationResponse
		wantErr error
	}{

		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().GetRaw(mock.AnythingOfType("context.backgroundCtx"), someOrgID).Return(organization.Organization{}, errors.New("some error"))
			},
			request: &frontierv1beta1.GetOrganizationRequest{
				Id: someOrgID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if org id is not uuid (slug) and org not exist",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().GetRaw(mock.AnythingOfType("context.backgroundCtx"), someOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.GetOrganizationRequest{
				Id: someOrgID,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return not found error if org id is invalid",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().GetRaw(mock.AnythingOfType("context.backgroundCtx"), "").Return(organization.Organization{}, organization.ErrInvalidID)
			},
			request: &frontierv1beta1.GetOrganizationRequest{},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return success if org service return nil error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().GetRaw(mock.AnythingOfType("context.backgroundCtx"), "9f256f86-31a3-11ec-8d3d-0242ac130003").Return(testOrgMap["9f256f86-31a3-11ec-8d3d-0242ac130003"], nil)
			},
			request: &frontierv1beta1.GetOrganizationRequest{
				Id: "9f256f86-31a3-11ec-8d3d-0242ac130003",
			},
			want: &frontierv1beta1.GetOrganizationResponse{
				Organization: &frontierv1beta1.Organization{
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
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockOrgSrv)
			}
			mockDep := Handler{orgService: mockOrgSrv}
			got, err := mockDep.GetOrganization(ctx, tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateOrganization(t *testing.T) {
	someOrgID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ms *mocks.MetaSchemaService)
		request *frontierv1beta1.UpdateOrganizationRequest
		want    *frontierv1beta1.UpdateOrganizationResponse
		wantErr error
	}{
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
				}).Return(organization.Organization{}, errors.New("some error"))
			},
			request: &frontierv1beta1.UpdateOrganizationRequest{
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
			},
			want:    nil,
			wantErr: grpcInternalServerError,
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
			request: &frontierv1beta1.UpdateOrganizationRequest{
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
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
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
			request: &frontierv1beta1.UpdateOrganizationRequest{
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
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
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
			request: &frontierv1beta1.UpdateOrganizationRequest{
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
			},
			want:    nil,
			wantErr: grpcConflictError,
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
			request: &frontierv1beta1.UpdateOrganizationRequest{
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
			},
			want: &frontierv1beta1.UpdateOrganizationResponse{
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
					State:     "enabled",
					Avatar:    "",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
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
			request: &frontierv1beta1.UpdateOrganizationRequest{
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
			},
			want: &frontierv1beta1.UpdateOrganizationResponse{
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
					State:     "enabled",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgSrv := new(mocks.OrganizationService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockOrgSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{orgService: mockOrgSrv, metaSchemaService: mockMetaSchemaSvc}
			got, err := mockDep.UpdateOrganization(ctx, tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListOrganizationAdmins(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(us *mocks.UserService, os *mocks.OrganizationService)
		request *frontierv1beta1.ListOrganizationAdminsRequest
		want    *frontierv1beta1.ListOrganizationAdminsResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return([]user.User{}, errors.New("some error"))
			},
			request: &frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return error if org id is not exist",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
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
			request: &frontierv1beta1.ListOrganizationAdminsRequest{
				Id: testOrgID,
			},
			want: &frontierv1beta1.ListOrganizationAdminsResponse{
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
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			mockOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockUserService, mockOrgService)
			}
			mockDep := Handler{userService: mockUserService, orgService: mockOrgService}
			got, err := mockDep.ListOrganizationAdmins(ctx, tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListOrganizationUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(us *mocks.UserService, os *mocks.OrganizationService)
		request *frontierv1beta1.ListOrganizationUsersRequest
		want    *frontierv1beta1.ListOrganizationUsersResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(organization.Organization{}, errors.New("some error"))
			},
			request: &frontierv1beta1.ListOrganizationUsersRequest{
				Id: "some-org-id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return org not found error if org id is not exist",
			setup: func(us *mocks.UserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.ListOrganizationUsersRequest{
				Id: "some-org-id",
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
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
			request: &frontierv1beta1.ListOrganizationUsersRequest{
				Id: testOrgID,
			},
			want: &frontierv1beta1.ListOrganizationUsersResponse{
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
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			mockOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockUserService, mockOrgService)
			}
			mockDep := Handler{userService: mockUserService, orgService: mockOrgService}
			got, err := mockDep.ListOrganizationUsers(ctx, tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListOrganizationServiceUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(us *mocks.ServiceUserService, os *mocks.OrganizationService)
		req     *frontierv1beta1.ListOrganizationServiceUsersRequest
		want    *frontierv1beta1.ListOrganizationServiceUsersResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(us *mocks.ServiceUserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("some error"))
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return([]serviceuser.ServiceUser{}, errors.New("some error"))
			},
			req: &frontierv1beta1.ListOrganizationServiceUsersRequest{
				Id: testOrgID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return org not found error if org doesnt exist",
			setup: func(us *mocks.ServiceUserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			req: &frontierv1beta1.ListOrganizationServiceUsersRequest{
				Id: testOrgID,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return success if org service return nil error",
			setup: func(us *mocks.ServiceUserService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				var testUserList []user.User
				for _, u := range testUserMap {
					testUserList = append(testUserList, u)
				}
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return([]serviceuser.ServiceUser{
					{
						ID:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "Sample Service User",
						Metadata: map[string]interface{}{
							"foo": "bar",
						},
						CreatedAt: time.Time{},
						UpdatedAt: time.Time{},
					},
				}, nil)
			},
			req: &frontierv1beta1.ListOrganizationServiceUsersRequest{
				Id: testOrgID,
			},
			want: &frontierv1beta1.ListOrganizationServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{
					{
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "Sample Service User",
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
			mockSvcUserService := new(mocks.ServiceUserService)
			mockOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockSvcUserService, mockOrgService)
			}
			mockDep := Handler{serviceUserService: mockSvcUserService, orgService: mockOrgService}
			got, err := mockDep.ListOrganizationServiceUsers(ctx, tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListAllOrganizations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService)
		req     *frontierv1beta1.ListAllOrganizationsRequest
		want    *frontierv1beta1.ListAllOrganizationsResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"),
					organization.Filter{}).Return([]organization.Organization{}, errors.New("some error"))
			},
			req:     &frontierv1beta1.ListAllOrganizationsRequest{},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return empty list of orgs if org service return nil error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), organization.Filter{}).Return([]organization.Organization{}, nil)
			},
			req:     &frontierv1beta1.ListAllOrganizationsRequest{},
			want:    &frontierv1beta1.ListAllOrganizationsResponse{},
			wantErr: nil,
		},
		{
			name: "should return success if org service return nil error",
			setup: func(os *mocks.OrganizationService) {
				var testOrgList []organization.Organization
				for _, o := range testOrgMap {
					testOrgList = append(testOrgList, o)
				}
				os.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), organization.Filter{}).Return(testOrgList, nil)
			},
			req: &frontierv1beta1.ListAllOrganizationsRequest{},
			want: &frontierv1beta1.ListAllOrganizationsResponse{
				Organizations: []*frontierv1beta1.Organization{
					{
						Name:  "org-1",
						Id:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"age":    structpb.NewNumberValue(21),
								"intern": structpb.NewBoolValue(true),
								"email":  structpb.NewStringValue("org1@org1.com"),
							},
						},
						State:     "enabled",
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
			mockSvcOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockSvcOrgService)
			}
			mockDep := Handler{orgService: mockSvcOrgService}
			got, err := mockDep.ListAllOrganizations(ctx, tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_EnableOrganization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService)
		req     *frontierv1beta1.EnableOrganizationRequest
		want    *frontierv1beta1.EnableOrganizationResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Enable(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(errors.New("some error"))
			},
			req: &frontierv1beta1.EnableOrganizationRequest{
				Id: testOrgID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should enable org successfully",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Enable(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(nil)
			},
			req: &frontierv1beta1.EnableOrganizationRequest{
				Id: testOrgID,
			},
			want:    &frontierv1beta1.EnableOrganizationResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockOrgService)
			}
			mockDep := Handler{orgService: mockOrgService}
			got, err := mockDep.EnableOrganization(ctx, tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DisableOrganization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService)
		req     *frontierv1beta1.DisableOrganizationRequest
		want    *frontierv1beta1.DisableOrganizationResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Disable(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(errors.New("some error"))
			},
			req: &frontierv1beta1.DisableOrganizationRequest{
				Id: "some-org-id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should disable org successfully",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Disable(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(nil)
			},
			req: &frontierv1beta1.DisableOrganizationRequest{
				Id: "some-org-id",
			},
			want:    &frontierv1beta1.DisableOrganizationResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockOrgService)
			}
			mockDep := Handler{orgService: mockOrgService}
			got, err := mockDep.DisableOrganization(ctx, tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_AddOrganizationUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService)
		req     *frontierv1beta1.AddOrganizationUsersRequest
		want    *frontierv1beta1.AddOrganizationUsersResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("some error"))
			},
			req: &frontierv1beta1.AddOrganizationUsersRequest{
				Id:      testOrgID,
				UserIds: []string{"some-user-id"},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should add user to org successfully",
			setup: func(os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				os.EXPECT().AddUsers(mock.AnythingOfType("context.backgroundCtx"), testOrgID, []string{"some-user-id"}).Return(nil)
			},
			req: &frontierv1beta1.AddOrganizationUsersRequest{
				Id:      testOrgID,
				UserIds: []string{"some-user-id"},
			},
			want:    &frontierv1beta1.AddOrganizationUsersResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockOrgService)
			}
			mockDep := Handler{orgService: mockOrgService}
			got, err := mockDep.AddOrganizationUsers(ctx, tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_RemoveOrganizationUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter)
		req     *frontierv1beta1.RemoveOrganizationUserRequest
		want    *frontierv1beta1.RemoveOrganizationUserResponse
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("some error"))
			},
			req: &frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return the error and not remove user if it is the last admin user",
			setup: func(os *mocks.OrganizationService, us *mocks.UserService, ds *mocks.CascadeDeleter) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().ListByOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, organization.AdminRole).Return([]user.User{
					testUserMap[testUserID],
				}, nil)
				ds.EXPECT().RemoveUsersFromOrg(mock.AnythingOfType("context.backgroundCtx"), testOrgID, []string{testUserID}).Return(nil)
			},
			req: &frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: testUserID,
			},
			want:    nil,
			wantErr: grpcMinAdminCountErr,
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
			req: &frontierv1beta1.RemoveOrganizationUserRequest{
				Id:     testOrgID,
				UserId: "some-user-id",
			},
			want:    &frontierv1beta1.RemoveOrganizationUserResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockUserService := new(mocks.UserService)
			mockDeleterService := new(mocks.CascadeDeleter)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockOrgService, mockUserService, mockDeleterService)
			}
			mockDep := Handler{
				orgService:     mockOrgService,
				userService:    mockUserService,
				deleterService: mockDeleterService,
			}
			got, err := mockDep.RemoveOrganizationUser(ctx, tt.req)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_ListOrganizationProjects(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(ps *mocks.ProjectService, os *mocks.OrganizationService)
		req     *frontierv1beta1.ListOrganizationProjectsRequest
		want    *frontierv1beta1.ListOrganizationProjectsResponse
		wantErr error
	}{
		{
			name: "should return error if organization does not exist ",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-org-id").Return(organization.Organization{}, organization.ErrNotExist)
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{OrgID: "some-org-id"}).Return([]project.Project{}, organization.ErrNotExist)
			},
			req: &frontierv1beta1.ListOrganizationProjectsRequest{
				Id: "some-org-id",
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return internal error if org service return some error",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("some error"))
			},
			req: &frontierv1beta1.ListOrganizationProjectsRequest{
				Id: testOrgID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return list of projects successfully",
			setup: func(ps *mocks.ProjectService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgMap[testOrgID].Name).Return(testOrgMap[testOrgID], nil)
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{OrgID: testOrgID}).Return([]project.Project{
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
			req: &frontierv1beta1.ListOrganizationProjectsRequest{
				Id: testOrgMap[testOrgID].Name,
			},
			want: &frontierv1beta1.ListOrganizationProjectsResponse{
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
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectService := new(mocks.ProjectService)
			mockOrgService := new(mocks.OrganizationService)
			ctx := context.Background()
			if tt.setup != nil {
				tt.setup(mockProjectService, mockOrgService)
			}
			mockDep := Handler{projectService: mockProjectService, orgService: mockOrgService}
			got, err := mockDep.ListOrganizationProjects(ctx, tt.req)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
