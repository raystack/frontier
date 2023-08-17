package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	su1 = serviceuser.ServiceUser{
		ID:        "1",
		Title:     "1",
		OrgID:     "1",
		State:     "1",
		Metadata:  metadata.Metadata{},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	su2 = serviceuser.ServiceUser{
		ID:    "2",
		OrgID: "2",
		Title: "2",
		State: "2",
		Metadata: metadata.Metadata{
			"key": "value",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	su1PB = &frontierv1beta1.ServiceUser{
		Id:    "1",
		Title: "1",
		OrgId: "1",
		State: "1",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
	su2PB = &frontierv1beta1.ServiceUser{
		Id:    "2",
		Title: "2",
		OrgId: "2",
		State: "2",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"key": {
					Kind: &structpb.Value_StringValue{
						StringValue: "value",
					},
				},
			},
		},
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
)

func TestHandler_ListServiveUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.ListServiceUsersRequest
		want    *frontierv1beta1.ListServiceUsersResponse
		wantErr error
	}{
		{
			name:    "should return internal server error when list service user service returns error",
			request: &frontierv1beta1.ListServiceUsersRequest{},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), serviceuser.Filter{
					OrgID: "",
					State: "",
				}).Return(nil, errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "Test List Service Users",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), serviceuser.Filter{
					OrgID: "",
					State: "",
				}).Return([]serviceuser.ServiceUser{su1, su2}, nil)
			},
			request: &frontierv1beta1.ListServiceUsersRequest{},
			want: &frontierv1beta1.ListServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{
					su1PB,
					su2PB,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.ListServiceUsers(context.Background(), &frontierv1beta1.ListServiceUsersRequest{})
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetServiceUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.GetServiceUserRequest
		want    *frontierv1beta1.GetServiceUserResponse
		wantErr error
	}{
		{
			name: "should return internal server error when get service user service returns error",
			request: &frontierv1beta1.GetServiceUserRequest{
				Id: "1",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "1").Return(serviceuser.ServiceUser{}, errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error when service user is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "1").Return(serviceuser.ServiceUser{}, serviceuser.ErrNotExist)
			},
			request: &frontierv1beta1.GetServiceUserRequest{
				Id: "1",
			},
			want:    nil,
			wantErr: grpcServiceUserNotFound,
		},
		{
			name: "should return service user",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "1").Return(su1, nil)
			},
			request: &frontierv1beta1.GetServiceUserRequest{
				Id: "1",
			},
			want: &frontierv1beta1.GetServiceUserResponse{
				Serviceuser: su1PB,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.GetServiceUser(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateServiceUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.CreateServiceUserRequest
		want    *frontierv1beta1.CreateServiceUserResponse
		wantErr error
	}{
		{
			name: "should return internal server error when create service user service returns error",
			request: &frontierv1beta1.CreateServiceUserRequest{
				Body: &frontierv1beta1.ServiceUserRequestBody{
					Title:    su1PB.Title,
					Metadata: su1PB.Metadata,
				},
				OrgId: su1PB.OrgId,
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), serviceuser.ServiceUser{
					Title:    su1.Title,
					Metadata: su1.Metadata,
					OrgID:    su1.OrgID,
				}).Return(serviceuser.ServiceUser{}, errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return service user",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), serviceuser.ServiceUser{
					Title:    su1.Title,
					Metadata: su1.Metadata,
					OrgID:    su1.OrgID,
				}).Return(su1, nil)
			},
			request: &frontierv1beta1.CreateServiceUserRequest{
				Body: &frontierv1beta1.ServiceUserRequestBody{
					Title:    su1PB.Title,
					Metadata: su1PB.Metadata,
				},
				OrgId: su1PB.OrgId,
			},
			want: &frontierv1beta1.CreateServiceUserResponse{
				Serviceuser: su1PB,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.CreateServiceUser(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteServiceUser(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.DeleteServiceUserRequest
		want    *frontierv1beta1.DeleteServiceUserResponse
		wantErr error
	}{
		{
			name: "should return internal server error when delete service user service returns error",
			request: &frontierv1beta1.DeleteServiceUserRequest{
				Id: "1",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "1").Return(errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error when service user is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "1").Return(serviceuser.ErrNotExist)
			},
			request: &frontierv1beta1.DeleteServiceUserRequest{
				Id: "1",
			},
			want:    nil,
			wantErr: grpcServiceUserNotFound,
		},
		{
			name: "should return service user",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "1").Return(nil)
			},
			request: &frontierv1beta1.DeleteServiceUserRequest{
				Id: "1",
			},
			want:    &frontierv1beta1.DeleteServiceUserResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.DeleteServiceUser(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateServiceUserKey(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.CreateServiceUserKeyRequest
		want    *frontierv1beta1.CreateServiceUserKeyResponse
		wantErr error
	}{
		{
			name: "should return internal server error when create service user key service returns error",
			request: &frontierv1beta1.CreateServiceUserKeyRequest{
				Id:    "1",
				Title: "title",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateKey(mock.AnythingOfType("*context.emptyCtx"), serviceuser.Credential{
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Credential{}, errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error when service user is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateKey(mock.AnythingOfType("*context.emptyCtx"), serviceuser.Credential{
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Credential{}, serviceuser.ErrNotExist)
			},
			request: &frontierv1beta1.CreateServiceUserKeyRequest{
				Id:    "1",
				Title: "title",
			},
			want:    nil,
			wantErr: grpcServiceUserNotFound,
		},
		{
			name: "should return service user key",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateKey(mock.AnythingOfType("*context.emptyCtx"), serviceuser.Credential{
					ServiceUserID: "1",
					Title:         "title",
				}).Return(suKey1PB, nil)
			},
			request: &frontierv1beta1.CreateServiceUserKeyRequest{
				Id:    "1",
				Title: "title",
			},
			want: &frontierv1beta1.CreateServiceUserKeyResponse{
				Key: &Key1PB,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.CreateServiceUserKey(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

// Todo
var suKey1PB = serviceuser.Credential{
	ID:            "1",
	ServiceUserID: "1",
	Title:         "title",
	SecretHash:    []byte("hash"),
	PublicKey:     jwk.NewSet(),
	PrivateKey:    []byte("private"),
}
var Key1PB = frontierv1beta1.KeyCredential{
	Type:        "sv_rsa",
	PrincipalId: "1",
	PrivateKey:  "private",
	Kid:         "1",
}

func TestHandler_ListServiceUserKeys(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.ListServiceUserKeysRequest
		want    *frontierv1beta1.ListServiceUserKeysResponse
		wantErr error
	}{
		{
			name: "should return internal server error when list service user keys service returns error",
			request: &frontierv1beta1.ListServiceUserKeysRequest{
				Id: "1",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().ListKeys(mock.AnythingOfType("*context.emptyCtx"), "1").Return(nil, errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error when service user is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().ListKeys(mock.AnythingOfType("*context.emptyCtx"), "1").Return(nil, serviceuser.ErrNotExist)
			},
			request: &frontierv1beta1.ListServiceUserKeysRequest{
				Id: "1",
			},
			want:    nil,
			wantErr: grpcServiceUserNotFound,
		},
		{
			name: "should return service user keys",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().ListKeys(mock.AnythingOfType("*context.emptyCtx"), "1").Return([]serviceuser.Credential{suKey1PB}, nil)
			},
			request: &frontierv1beta1.ListServiceUserKeysRequest{
				Id: "1",
			},
			want: &frontierv1beta1.ListServiceUserKeysResponse{
				Keys: []*frontierv1beta1.ServiceUserKey{
					{
						Id:          "1",
						Title:       "title",
						PrincipalId: "1",
						PublicKey:   "{\"keys\":[]}",
						CreatedAt:   timestamppb.New(time.Time{}),
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.ListServiceUserKeys(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetServiceUserKey(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.GetServiceUserKeyRequest
		want    *frontierv1beta1.GetServiceUserKeyResponse
		wantErr error
	}{
		{
			name: "should return internal server error when get service user key service returns error",
			request: &frontierv1beta1.GetServiceUserKeyRequest{
				Id:    "1",
				KeyId: "1",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().GetKey(mock.AnythingOfType("*context.emptyCtx"), "1").Return(serviceuser.Credential{}, errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error when service user is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().GetKey(mock.AnythingOfType("*context.emptyCtx"), "1").Return(serviceuser.Credential{}, serviceuser.ErrCredNotExist)
			},
			request: &frontierv1beta1.GetServiceUserKeyRequest{
				Id:    "1",
				KeyId: "1",
			},
			want:    nil,
			wantErr: grpcSvcUserCredNotFound,
		},
		{
			name: "should return service user key",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().GetKey(mock.AnythingOfType("*context.emptyCtx"), "1").Return(suKey1PB, nil)
			},
			request: &frontierv1beta1.GetServiceUserKeyRequest{
				Id:    "1",
				KeyId: "1",
			},
			want: &frontierv1beta1.GetServiceUserKeyResponse{
				Keys: []*frontierv1beta1.JSONWebKey{
					// {
					// 	// Kid: "1",
					// 	// Kty: "RSA",
					// 	// N:   "null",
					// 	// E:   "null",
					// 	// Alg: "RS256",
					// },
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.GetServiceUserKey(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteServiceUserKey(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.DeleteServiceUserKeyRequest
		want    *frontierv1beta1.DeleteServiceUserKeyResponse
		wantErr error
	}{
		{
			name: "should return internal server error when delete service user key service returns error",
			request: &frontierv1beta1.DeleteServiceUserKeyRequest{
				Id:    "1",
				KeyId: "1",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().DeleteKey(mock.AnythingOfType("*context.emptyCtx"), "1").Return(errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error when service user is not found",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().DeleteKey(mock.AnythingOfType("*context.emptyCtx"), "1").Return(serviceuser.ErrCredNotExist)
			},
			request: &frontierv1beta1.DeleteServiceUserKeyRequest{
				Id:    "1",
				KeyId: "1",
			},
			want:    nil,
			wantErr: grpcSvcUserCredNotFound,
		},
		{
			name: "should return service user key",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().DeleteKey(mock.AnythingOfType("*context.emptyCtx"), "1").Return(nil)
			},
			request: &frontierv1beta1.DeleteServiceUserKeyRequest{
				Id:    "1",
				KeyId: "1",
			},
			want:    &frontierv1beta1.DeleteServiceUserKeyResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.DeleteServiceUserKey(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteServiceUserSecret(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.DeleteServiceUserSecretRequest
		want    *frontierv1beta1.DeleteServiceUserSecretResponse
		wantErr error
	}{
		{
			name: "should return internal server error when delete service user secret service returns error",
			request: &frontierv1beta1.DeleteServiceUserSecretRequest{
				Id:       "1",
				SecretId: "1",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().DeleteSecret(mock.AnythingOfType("*context.emptyCtx"), "1").Return(errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return service user secret",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().DeleteSecret(mock.AnythingOfType("*context.emptyCtx"), "1").Return(nil)
			},
			request: &frontierv1beta1.DeleteServiceUserSecretRequest{
				Id:       "1",
				SecretId: "1",
			},
			want:    &frontierv1beta1.DeleteServiceUserSecretResponse{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.DeleteServiceUserSecret(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateServiceUserSecret(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(su *mocks.ServiceUserService)
		request *frontierv1beta1.CreateServiceUserSecretRequest
		want    *frontierv1beta1.CreateServiceUserSecretResponse
		wantErr error
	}{
		{
			name: "should return internal server error when create service user secret service returns error",
			request: &frontierv1beta1.CreateServiceUserSecretRequest{
				Id:    "1",
				Title: "title",
			},
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateSecret(mock.AnythingOfType("*context.emptyCtx"), serviceuser.Credential{
					// ID:            "1",
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Secret{
					ID:        "1",
					Value:     []byte("value"),
					CreatedAt: time.Now(),
				}, errors.New("error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return service user secret",
			setup: func(su *mocks.ServiceUserService) {
				su.EXPECT().CreateSecret(mock.AnythingOfType("*context.emptyCtx"), serviceuser.Credential{
					Title:         "title",
					ServiceUserID: "1",
				}).Return(serviceuser.Secret{
					ID:        "1",
					Value:     []byte("value"),
					CreatedAt: time.Time{},
				}, nil)
			},
			request: &frontierv1beta1.CreateServiceUserSecretRequest{
				Id:    "1",
				Title: "title",
			},
			want: &frontierv1beta1.CreateServiceUserSecretResponse{
				Secret: &frontierv1beta1.SecretCredential{
					Id:        "1",
					Secret:    "value",
					CreatedAt: timestamppb.New(time.Time{}),
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiveUserSvc := new(mocks.ServiceUserService)
			if tt.setup != nil {
				tt.setup(mockServiveUserSvc)
			}
			h := Handler{
				serviceUserService: mockServiveUserSvc,
			}
			got, err := h.CreateServiceUserSecret(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, got)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
