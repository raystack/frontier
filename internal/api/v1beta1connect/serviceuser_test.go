package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testServiceUserID = "su-9f256f86-31a3-11ec-8d3d-0242ac130003"

	testServiceUserMap = map[string]serviceuser.ServiceUser{
		"su-9f256f86-31a3-11ec-8d3d-0242ac130003": {
			ID:    "su-9f256f86-31a3-11ec-8d3d-0242ac130003",
			Title: "Test Service User",
			OrgID: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
			State: "enabled",
			Metadata: metadata.Metadata{
				"purpose": "testing",
				"team":    "backend",
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
)

func TestHandler_ListServiceUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(sus *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.ListServiceUsersRequest]
		want    *connect.Response[frontierv1beta1.ListServiceUsersResponse]
		wantErr bool
	}{
		{
			name: "should list service users successfully",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, serviceuser.Filter{
					OrgID: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
					State: serviceuser.State("enabled"),
				}).Return([]serviceuser.ServiceUser{
					testServiceUserMap["su-9f256f86-31a3-11ec-8d3d-0242ac130003"],
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
				State: "enabled",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{
					{
						Id:    "su-9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "Test Service User",
						OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
						State: "enabled",
						Metadata: func() *structpb.Struct {
							md, _ := metadata.Metadata{
								"purpose": "testing",
								"team":    "backend",
							}.ToStructPB()
							return md
						}(),
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: false,
		},
		{
			name: "should list service users with only org filter",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, serviceuser.Filter{
					OrgID: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
					State: serviceuser.State(""),
				}).Return([]serviceuser.ServiceUser{
					testServiceUserMap["su-9f256f86-31a3-11ec-8d3d-0242ac130003"],
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{
					{
						Id:    "su-9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "Test Service User",
						OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
						State: "enabled",
						Metadata: func() *structpb.Struct {
							md, _ := metadata.Metadata{
								"purpose": "testing",
								"team":    "backend",
							}.ToStructPB()
							return md
						}(),
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: false,
		},
		{
			name: "should return empty list when no service users found",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, serviceuser.Filter{
					OrgID: "org-nonexistent",
					State: serviceuser.State(""),
				}).Return([]serviceuser.ServiceUser{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-nonexistent",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{},
			}),
			wantErr: false,
		},
		{
			name: "should return internal error when service fails",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
			}),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceUserService := &mocks.ServiceUserService{}
			if tt.setup != nil {
				tt.setup(serviceUserService)
			}

			h := ConnectHandler{
				serviceUserService: serviceUserService,
			}

			got, err := h.ListServiceUsers(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.want != nil {
					assert.Equal(t, len(tt.want.Msg.GetServiceusers()), len(got.Msg.GetServiceusers()))
					for i, expectedSU := range tt.want.Msg.GetServiceusers() {
						actualSU := got.Msg.GetServiceusers()[i]
						assert.Equal(t, expectedSU.GetId(), actualSU.GetId())
						assert.Equal(t, expectedSU.GetTitle(), actualSU.GetTitle())
						assert.Equal(t, expectedSU.GetOrgId(), actualSU.GetOrgId())
						assert.Equal(t, expectedSU.GetState(), actualSU.GetState())
					}
				}
			}

			serviceUserService.AssertExpectations(t)
		})
	}
}

func TestHandler_ListAllServiceUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(sus *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.ListAllServiceUsersRequest]
		want    *connect.Response[frontierv1beta1.ListAllServiceUsersResponse]
		wantErr bool
	}{
		{
			name: "should list all service users successfully",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("ListAll", mock.Anything).Return([]serviceuser.ServiceUser{
					testServiceUserMap["su-9f256f86-31a3-11ec-8d3d-0242ac130003"],
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAllServiceUsersRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListAllServiceUsersResponse{
				ServiceUsers: []*frontierv1beta1.ServiceUser{
					{
						Id:    "su-9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "Test Service User",
						OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
						State: "enabled",
						Metadata: func() *structpb.Struct {
							md, _ := metadata.Metadata{
								"purpose": "testing",
								"team":    "backend",
							}.ToStructPB()
							return md
						}(),
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: false,
		},
		{
			name: "should return empty list when no service users found",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("ListAll", mock.Anything).Return([]serviceuser.ServiceUser{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAllServiceUsersRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListAllServiceUsersResponse{
				ServiceUsers: []*frontierv1beta1.ServiceUser{},
			}),
			wantErr: false,
		},
		{
			name: "should return internal error when service fails",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("ListAll", mock.Anything).Return(nil, errors.New("service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListAllServiceUsersRequest{}),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceUserService := &mocks.ServiceUserService{}
			if tt.setup != nil {
				tt.setup(serviceUserService)
			}

			h := ConnectHandler{
				serviceUserService: serviceUserService,
			}

			got, err := h.ListAllServiceUsers(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.want != nil {
					assert.Equal(t, len(tt.want.Msg.GetServiceUsers()), len(got.Msg.GetServiceUsers()))
					for i, expectedSU := range tt.want.Msg.GetServiceUsers() {
						actualSU := got.Msg.GetServiceUsers()[i]
						assert.Equal(t, expectedSU.GetId(), actualSU.GetId())
						assert.Equal(t, expectedSU.GetTitle(), actualSU.GetTitle())
						assert.Equal(t, expectedSU.GetOrgId(), actualSU.GetOrgId())
						assert.Equal(t, expectedSU.GetState(), actualSU.GetState())
					}
				}
			}

			serviceUserService.AssertExpectations(t)
		})
	}
}

func TestConnectHandler_CreateServiceUser(t *testing.T) {
	type args struct {
		request *connect.Request[frontierv1beta1.CreateServiceUserRequest]
	}
	tests := []struct {
		name               string
		args               args
		want               *connect.Response[frontierv1beta1.CreateServiceUserResponse]
		wantErr            bool
		serviceUserService func() *mocks.ServiceUserService
	}{
		{
			name: "should create service user successfully",
			args: args{
				request: &connect.Request[frontierv1beta1.CreateServiceUserRequest]{
					Msg: &frontierv1beta1.CreateServiceUserRequest{
						OrgId: testOrgID,
						Body: &frontierv1beta1.ServiceUserRequestBody{
							Title: "test-service-user",
							Metadata: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"key": structpb.NewStringValue("value"),
								},
							},
						},
					},
				},
			},
			want: &connect.Response[frontierv1beta1.CreateServiceUserResponse]{
				Msg: &frontierv1beta1.CreateServiceUserResponse{
					Serviceuser: &frontierv1beta1.ServiceUser{
						Id:       testServiceUserID,
						OrgId:    testOrgID,
						Title:    "test-service-user",
						State:    "enabled",
						Metadata: &structpb.Struct{},
					},
				},
			},
			wantErr: false,
			serviceUserService: func() *mocks.ServiceUserService {
				mockSvc := &mocks.ServiceUserService{}
				mockSvc.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(su serviceuser.ServiceUser) bool {
					return su.Title == "test-service-user" && su.OrgID == testOrgID
				})).Return(serviceuser.ServiceUser{
					ID:       testServiceUserID,
					OrgID:    testOrgID,
					Title:    "test-service-user",
					State:    "enabled",
					Metadata: metadata.Metadata{},
				}, nil)
				return mockSvc
			},
		},
		{
			name: "should return error when service user creation fails",
			args: args{
				request: &connect.Request[frontierv1beta1.CreateServiceUserRequest]{
					Msg: &frontierv1beta1.CreateServiceUserRequest{
						OrgId: testOrgID,
						Body: &frontierv1beta1.ServiceUserRequestBody{
							Title: "test-service-user",
						},
					},
				},
			},
			wantErr: true,
			serviceUserService: func() *mocks.ServiceUserService {
				mockSvc := &mocks.ServiceUserService{}
				mockSvc.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("serviceuser.ServiceUser")).Return(serviceuser.ServiceUser{}, errors.New("creation failed"))
				return mockSvc
			},
		},
		{
			name: "should create service user without metadata",
			args: args{
				request: &connect.Request[frontierv1beta1.CreateServiceUserRequest]{
					Msg: &frontierv1beta1.CreateServiceUserRequest{
						OrgId: testOrgID,
						Body: &frontierv1beta1.ServiceUserRequestBody{
							Title: "simple-service-user",
						},
					},
				},
			},
			want: &connect.Response[frontierv1beta1.CreateServiceUserResponse]{
				Msg: &frontierv1beta1.CreateServiceUserResponse{
					Serviceuser: &frontierv1beta1.ServiceUser{
						Id:       testServiceUserID,
						OrgId:    testOrgID,
						Title:    "simple-service-user",
						State:    "enabled",
						Metadata: &structpb.Struct{},
					},
				},
			},
			wantErr: false,
			serviceUserService: func() *mocks.ServiceUserService {
				mockSvc := &mocks.ServiceUserService{}
				mockSvc.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(su serviceuser.ServiceUser) bool {
					return su.Title == "simple-service-user" && su.OrgID == testOrgID
				})).Return(serviceuser.ServiceUser{
					ID:       testServiceUserID,
					OrgID:    testOrgID,
					Title:    "simple-service-user",
					State:    "enabled",
					Metadata: metadata.Metadata{},
				}, nil)
				return mockSvc
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceUserService := tt.serviceUserService()

			h := ConnectHandler{
				serviceUserService: serviceUserService,
			}

			got, err := h.CreateServiceUser(context.Background(), tt.args.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.want != nil {
					assert.Equal(t, tt.want.Msg.GetServiceuser().GetId(), got.Msg.GetServiceuser().GetId())
					assert.Equal(t, tt.want.Msg.GetServiceuser().GetTitle(), got.Msg.GetServiceuser().GetTitle())
					assert.Equal(t, tt.want.Msg.GetServiceuser().GetOrgId(), got.Msg.GetServiceuser().GetOrgId())
					assert.Equal(t, tt.want.Msg.GetServiceuser().GetState(), got.Msg.GetServiceuser().GetState())
				}
			}

			serviceUserService.AssertExpectations(t)
		})
	}
}

func TestConnectHandler_DeleteServiceUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(sus *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.DeleteServiceUserRequest]
		want    *connect.Response[frontierv1beta1.DeleteServiceUserResponse]
		wantErr bool
	}{
		{
			name: "should delete service user successfully",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("Delete", mock.Anything, testServiceUserID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserRequest{
				Id:    testServiceUserID,
				OrgId: testOrgID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteServiceUserResponse{}),
			wantErr: false,
		},
		{
			name: "should return not found error when service user does not exist",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("Delete", mock.Anything, "non-existent-id").Return(serviceuser.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserRequest{
				Id:    "non-existent-id",
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: true,
		},
		{
			name: "should return internal error when delete service fails",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("Delete", mock.Anything, testServiceUserID).Return(errors.New("service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserRequest{
				Id:    testServiceUserID,
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceUserService := &mocks.ServiceUserService{}
			if tt.setup != nil {
				tt.setup(serviceUserService)
			}

			h := ConnectHandler{
				serviceUserService: serviceUserService,
			}

			got, err := h.DeleteServiceUser(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if err != nil {
					connectErr := err.(*connect.Error)
					switch {
					case tt.request.Msg.GetId() == "non-existent-id":
						assert.Equal(t, connect.CodeNotFound, connectErr.Code())
					default:
						assert.Equal(t, connect.CodeInternal, connectErr.Code())
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.Msg, got.Msg)
			}

			serviceUserService.AssertExpectations(t)
		})
	}
}

func TestConnectHandler_CreateServiceUserJWK(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.CreateServiceUserJWKRequest]
		want    *connect.Response[frontierv1beta1.CreateServiceUserJWKResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when create service user key service returns error",
			request: connect.NewRequest(&frontierv1beta1.CreateServiceUserJWKRequest{
				Id:    "1",
				Title: "title",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateKey(mock.Anything, serviceuser.Credential{
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Credential{}, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return service user not found error when service user does not exist",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateKey(mock.Anything, serviceuser.Credential{
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Credential{}, serviceuser.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateServiceUserJWKRequest{
				Id:    "1",
				Title: "title",
			}),
			want:    nil,
			wantErr: serviceuser.ErrNotExist,
			errCode: connect.CodeNotFound,
		},
		{
			name: "should return service user key",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateKey(mock.Anything, serviceuser.Credential{
					ServiceUserID: "1",
					Title:         "title",
				}).Return(serviceuser.Credential{
					ID:            "1",
					ServiceUserID: "1",
					Title:         "title",
					SecretHash:    "hash",
					PrivateKey:    []byte("private"),
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateServiceUserJWKRequest{
				Id:    "1",
				Title: "title",
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateServiceUserJWKResponse{
				Key: &frontierv1beta1.KeyCredential{
					Type:        serviceuser.DefaultKeyType,
					Kid:         "1",
					PrincipalId: "1",
					PrivateKey:  "private",
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.CreateServiceUserJWK(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnectHandler_ListServiceUserJWKs(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.ListServiceUserJWKsRequest]
		want    *connect.Response[frontierv1beta1.ListServiceUserJWKsResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when list service user keys service returns error",
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserJWKsRequest{
				Id: "1",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().ListKeys(mock.Anything, "1").Return(nil, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return not found error when service user is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().ListKeys(mock.Anything, "1").Return(nil, serviceuser.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserJWKsRequest{
				Id: "1",
			}),
			want:    nil,
			wantErr: serviceuser.ErrNotExist,
			errCode: connect.CodeNotFound,
		},
		{
			name: "should return service user keys",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().ListKeys(mock.Anything, "1").Return([]serviceuser.Credential{
					{
						ID:            "1",
						ServiceUserID: "1",
						Title:         "title",
						SecretHash:    "hash",
						PublicKey:     jwk.NewSet(),
						PrivateKey:    []byte("private"),
						CreatedAt:     time.Time{},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserJWKsRequest{
				Id: "1",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUserJWKsResponse{
				Keys: []*frontierv1beta1.ServiceUserJWK{
					{
						Id:          "1",
						Title:       "title",
						PrincipalId: "1",
						PublicKey:   "{\"keys\":[]}",
						CreatedAt:   timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.ListServiceUserJWKs(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_GetServiceUserJWK(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.GetServiceUserJWKRequest]
		want    *connect.Response[frontierv1beta1.GetServiceUserJWKResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when get service user key service returns error",
			request: connect.NewRequest(&frontierv1beta1.GetServiceUserJWKRequest{
				Id:    "1",
				KeyId: "1",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("GetKey", mock.Anything, "1").Return(serviceuser.Credential{}, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return not found error when service user credential is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.On("GetKey", mock.Anything, "1").Return(serviceuser.Credential{}, serviceuser.ErrCredNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetServiceUserJWKRequest{
				Id:    "1",
				KeyId: "1",
			}),
			want:    nil,
			wantErr: serviceuser.ErrCredNotExist,
			errCode: connect.CodeNotFound,
		},
		{
			name: "should return service user key",
			setup: func(su *mocks.ServiceUserService) {
				su.On("GetKey", mock.Anything, "1").Return(serviceuser.Credential{
					ID:            "1",
					ServiceUserID: "1",
					Title:         "title",
					SecretHash:    "hash",
					PublicKey:     jwk.NewSet(),
					PrivateKey:    []byte("private"),
					CreatedAt:     time.Time{},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetServiceUserJWKRequest{
				Id:    "1",
				KeyId: "1",
			}),
			want: connect.NewResponse(&frontierv1beta1.GetServiceUserJWKResponse{
				Keys: []*frontierv1beta1.JSONWebKey{},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.GetServiceUserJWK(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_DeleteServiceUserJWK(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.DeleteServiceUserJWKRequest]
		want    *connect.Response[frontierv1beta1.DeleteServiceUserJWKResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when delete service user key service returns error",
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserJWKRequest{
				Id:    "1",
				KeyId: "1",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("DeleteKey", mock.Anything, "1").Return(errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return not found error when service user credential is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.On("DeleteKey", mock.Anything, "1").Return(serviceuser.ErrCredNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserJWKRequest{
				Id:    "1",
				KeyId: "1",
			}),
			want:    nil,
			wantErr: serviceuser.ErrCredNotExist,
			errCode: connect.CodeNotFound,
		},
		{
			name: "should delete service user key successfully",
			setup: func(su *mocks.ServiceUserService) {
				su.On("DeleteKey", mock.Anything, "1").Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserJWKRequest{
				Id:    "1",
				KeyId: "1",
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteServiceUserJWKResponse{}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.DeleteServiceUserJWK(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_CreateServiceUserCredential(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.CreateServiceUserCredentialRequest]
		want    *connect.Response[frontierv1beta1.CreateServiceUserCredentialResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when create service user secret service returns error",
			request: connect.NewRequest(&frontierv1beta1.CreateServiceUserCredentialRequest{
				Id:    "1",
				Title: "title",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("CreateSecret", mock.Anything, serviceuser.Credential{
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Secret{
					ID:        "1",
					Value:     "value",
					CreatedAt: time.Now(),
				}, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return service user secret successfully",
			setup: func(su *mocks.ServiceUserService) {
				su.On("CreateSecret", mock.Anything, serviceuser.Credential{
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Secret{
					ID:        "1",
					Title:     "title",
					Value:     "value",
					CreatedAt: time.Time{},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateServiceUserCredentialRequest{
				Id:    "1",
				Title: "title",
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateServiceUserCredentialResponse{
				Secret: &frontierv1beta1.SecretCredential{
					Id:        "1",
					Title:     "title",
					Secret:    "value",
					CreatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.CreateServiceUserCredential(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_ListServiceUserCredentials(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.ListServiceUserCredentialsRequest]
		want    *connect.Response[frontierv1beta1.ListServiceUserCredentialsResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when list service user credentials service returns error",
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserCredentialsRequest{
				Id: "service-user-id",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("ListSecret", mock.Anything, "service-user-id").Return(nil, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return empty list when service user has no credentials",
			setup: func(su *mocks.ServiceUserService) {
				su.On("ListSecret", mock.Anything, "service-user-id").Return([]serviceuser.Credential{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserCredentialsRequest{
				Id: "service-user-id",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUserCredentialsResponse{
				Secrets: []*frontierv1beta1.SecretCredential{},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should return service user credentials successfully",
			setup: func(su *mocks.ServiceUserService) {
				su.On("ListSecret", mock.Anything, "service-user-id").Return([]serviceuser.Credential{
					{
						ID:        "cred-1",
						Title:     "Test Credential 1",
						CreatedAt: time.Time{},
					},
					{
						ID:        "cred-2",
						Title:     "Test Credential 2",
						CreatedAt: time.Time{},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserCredentialsRequest{
				Id: "service-user-id",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUserCredentialsResponse{
				Secrets: []*frontierv1beta1.SecretCredential{
					{
						Id:        "cred-1",
						Title:     "Test Credential 1",
						CreatedAt: timestamppb.New(time.Time{}),
					},
					{
						Id:        "cred-2",
						Title:     "Test Credential 2",
						CreatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.ListServiceUserCredentials(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_DeleteServiceUserCredential(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.DeleteServiceUserCredentialRequest]
		want    *connect.Response[frontierv1beta1.DeleteServiceUserCredentialResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when delete service user credential service returns error",
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserCredentialRequest{
				Id:       "service-user-id",
				SecretId: "credential-id",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("DeleteSecret", mock.Anything, "credential-id").Return(errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should delete service user credential successfully",
			setup: func(su *mocks.ServiceUserService) {
				su.On("DeleteSecret", mock.Anything, "credential-id").Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserCredentialRequest{
				Id:       "service-user-id",
				SecretId: "credential-id",
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteServiceUserCredentialResponse{}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.DeleteServiceUserCredential(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_CreateServiceUserToken(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.CreateServiceUserTokenRequest]
		want    *connect.Response[frontierv1beta1.CreateServiceUserTokenResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when create service user token service returns error",
			request: connect.NewRequest(&frontierv1beta1.CreateServiceUserTokenRequest{
				Id:    "service-user-id",
				Title: "Test Token",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("CreateToken", mock.Anything, serviceuser.Credential{
					Title:         "Test Token",
					ServiceUserID: "service-user-id",
				}).Return(serviceuser.Token{}, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should create service user token successfully",
			setup: func(su *mocks.ServiceUserService) {
				su.On("CreateToken", mock.Anything, serviceuser.Credential{
					Title:         "Test Token",
					ServiceUserID: "service-user-id",
				}).Return(serviceuser.Token{
					ID:        "token-id",
					Title:     "Test Token",
					Value:     "token-value-123",
					CreatedAt: time.Time{},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateServiceUserTokenRequest{
				Id:    "service-user-id",
				Title: "Test Token",
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateServiceUserTokenResponse{
				Token: &frontierv1beta1.ServiceUserToken{
					Id:        "token-id",
					Title:     "Test Token",
					Token:     "token-value-123",
					CreatedAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.CreateServiceUserToken(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_ListServiceUserTokens(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.ListServiceUserTokensRequest]
		want    *connect.Response[frontierv1beta1.ListServiceUserTokensResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when list service user tokens service returns error",
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserTokensRequest{
				Id: "service-user-id",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("ListToken", mock.Anything, "service-user-id").Return(nil, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should list service user tokens successfully with empty list",
			setup: func(su *mocks.ServiceUserService) {
				su.On("ListToken", mock.Anything, "service-user-id").Return([]serviceuser.Credential{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserTokensRequest{
				Id: "service-user-id",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUserTokensResponse{
				Tokens: []*frontierv1beta1.ServiceUserToken{},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should list service user tokens successfully with multiple tokens",
			setup: func(su *mocks.ServiceUserService) {
				su.On("ListToken", mock.Anything, "service-user-id").Return([]serviceuser.Credential{
					{
						ID:        "token-1",
						Title:     "Token 1",
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        "token-2",
						Title:     "Token 2",
						CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserTokensRequest{
				Id: "service-user-id",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUserTokensResponse{
				Tokens: []*frontierv1beta1.ServiceUserToken{
					{
						Id:        "token-1",
						Title:     "Token 1",
						CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
					{
						Id:        "token-2",
						Title:     "Token 2",
						CreatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.ListServiceUserTokens(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_DeleteServiceUserToken(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.DeleteServiceUserTokenRequest]
		want    *connect.Response[frontierv1beta1.DeleteServiceUserTokenResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when delete service user token service returns error",
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserTokenRequest{
				TokenId: "token-id",
			}),
			setup: func(su *mocks.ServiceUserService) {
				su.On("DeleteToken", mock.Anything, "token-id").Return(errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should delete service user token successfully",
			setup: func(su *mocks.ServiceUserService) {
				su.On("DeleteToken", mock.Anything, "token-id").Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteServiceUserTokenRequest{
				TokenId: "token-id",
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteServiceUserTokenResponse{}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiceUserSvc)
			}
			h := &ConnectHandler{
				serviceUserService: mockServiceUserSvc,
			}
			got, err := h.DeleteServiceUserToken(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_ListServiceUserProjects(t *testing.T) {
	testProjectMap := map[string]project.Project{
		"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
			ID:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
			Name: "prj-1",
			Metadata: metadata.Metadata{
				"email": "org1@org1.com",
			},
			Organization: organization.Organization{
				ID: testOrgID,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
			ID:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
			Name: "prj-2",
			Metadata: metadata.Metadata{
				"email": "org1@org2.com",
			},
			Organization: organization.Organization{
				ID: testOrgID,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}

	tests := []struct {
		name    string
		setup   func(projSvc *mocks.ProjectService, permSvc *mocks.PermissionService, resourceSvc *mocks.ResourceService)
		request *connect.Request[frontierv1beta1.ListServiceUserProjectsRequest]
		want    *connect.Response[frontierv1beta1.ListServiceUserProjectsResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return internal server error when list service user project returns error",
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserProjectsRequest{
				Id: "1",
			}),
			setup: func(projSvc *mocks.ProjectService, permSvc *mocks.PermissionService, resourceSvc *mocks.ResourceService) {
				projSvc.EXPECT().ListByUser(mock.Anything, "1", schema.ServiceUserPrincipal, project.Filter{}).Return(nil, errors.New("test error"))
			},
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should return project list when there is no error",
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserProjectsRequest{
				Id: "1",
			}),
			setup: func(projSvc *mocks.ProjectService, permSvc *mocks.PermissionService, resourceSvc *mocks.ResourceService) {
				var projects []project.Project
				for _, projectID := range testProjectIDList {
					projects = append(projects, testProjectMap[projectID])
				}
				projSvc.EXPECT().ListByUser(mock.Anything, "1", schema.ServiceUserPrincipal, project.Filter{}).Return(projects, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListServiceUserProjectsResponse{
				Projects: []*frontierv1beta1.Project{{
					Id:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
					Name: "prj-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
						},
					},
					OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
					{
						Id:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
						Name: "prj-2",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"email": structpb.NewStringValue("org1@org2.com"),
							},
						},
						OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
		{
			name: "should return project list with access pairs if withPermission is passed",
			request: connect.NewRequest(&frontierv1beta1.ListServiceUserProjectsRequest{
				Id:              "1",
				WithPermissions: []string{"get"},
			}),
			setup: func(projSvc *mocks.ProjectService, permSvc *mocks.PermissionService, resourceSvc *mocks.ResourceService) {
				var projects []project.Project
				for _, projectID := range testProjectIDList {
					projects = append(projects, testProjectMap[projectID])
				}

				ctx := mock.Anything
				projSvc.EXPECT().ListByUser(ctx, "1", schema.ServiceUserPrincipal, project.Filter{}).Return(projects, nil)

				permSvc.EXPECT().Get(ctx, "app/project:get").Return(
					permission.Permission{
						ID:          uuid.New().String(),
						Name:        "get",
						NamespaceID: "app/project",
						Metadata:    map[string]any{},
						CreatedAt:   time.Time{},
						UpdatedAt:   time.Time{},
					}, nil)

				resourceSvc.EXPECT().BatchCheck(ctx, []resource.Check{
					{
						Object: relation.Object{
							ID:        "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
							Namespace: "app/project",
						},
						Permission: "get",
					},
					{
						Object: relation.Object{
							ID:        "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
							Namespace: "app/project",
						},
						Permission: "get",
					},
				}).Return([]relation.CheckPair{
					{
						Relation: relation.Relation{
							Object: relation.Object{
								ID:        "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
								Namespace: "app/project",
							},
							RelationName: "get",
						},
						Status: true,
					},
					{
						Relation: relation.Relation{
							Object: relation.Object{
								ID:        "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
								Namespace: "app/project",
							},
							RelationName: "get",
						},
						Status: true,
					},
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListServiceUserProjectsResponse{
				Projects: []*frontierv1beta1.Project{{
					Id:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
					Name: "prj-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
						},
					},
					OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
					{
						Id:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
						Name: "prj-2",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"email": structpb.NewStringValue("org1@org2.com"),
							},
						},
						OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
				AccessPairs: []*frontierv1beta1.ListServiceUserProjectsResponse_AccessPair{
					{
						ProjectId:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
						Permissions: []string{"get"},
					},
					{
						ProjectId:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
						Permissions: []string{"get"},
					},
				},
			}),
			wantErr: nil,
			errCode: connect.Code(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectSvc := new(mocks.ProjectService)
			mockPermissionSvc := new(mocks.PermissionService)
			mockResourceSvc := new(mocks.ResourceService)

			if tt.setup != nil {
				tt.setup(mockProjectSvc, mockPermissionSvc, mockResourceSvc)
			}
			h := &ConnectHandler{
				projectService:    mockProjectSvc,
				permissionService: mockPermissionSvc,
				resourceService:   mockResourceSvc,
			}
			got, err := h.ListServiceUserProjects(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.EqualValues(t, tt.errCode, connect.CodeOf(err))
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
