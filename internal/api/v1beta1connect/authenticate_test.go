package v1beta1connect

import (
	"context"
	"encoding/json"
	"testing"

	"connectrpc.com/connect"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestConnectHandler_AuthToken_ServiceUser(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(authn *mocks.AuthnService, org *mocks.OrganizationService)
		request     *connect.Request[frontierv1beta1.AuthTokenRequest]
		want        *connect.Response[frontierv1beta1.AuthTokenResponse]
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should return error when service user org is disabled",
			setup: func(authn *mocks.AuthnService, org *mocks.OrganizationService) {
				orgID := "test-org-id"
				serviceUserID := "test-service-user-id"

				authn.EXPECT().GetPrincipal(mock.Anything,
					authenticate.SessionClientAssertion,
					authenticate.ClientCredentialsClientAssertion,
					authenticate.JWTGrantClientAssertion).Return(authenticate.Principal{
					ID:   serviceUserID,
					Type: schema.ServiceUserPrincipal,
					ServiceUser: &serviceuser.ServiceUser{
						ID:    serviceUserID,
						OrgID: orgID,
					},
				}, nil)

				org.EXPECT().Get(mock.Anything, orgID).Return(
					organization.Organization{}, organization.ErrDisabled)
			},
			request:     connect.NewRequest(&frontierv1beta1.AuthTokenRequest{}),
			wantErr:     true,
			expectedErr: organization.ErrDisabled,
			want:        nil,
		},
		{
			name: "should return token when service user org is enabled",
			setup: func(authn *mocks.AuthnService, org *mocks.OrganizationService) {
				orgID := "test-org-id"
				serviceUserID := "test-service-user-id"
				expectedToken := []byte("test-access-token")

				authn.EXPECT().GetPrincipal(mock.Anything,
					authenticate.SessionClientAssertion,
					authenticate.ClientCredentialsClientAssertion,
					authenticate.JWTGrantClientAssertion).Return(authenticate.Principal{
					ID:   serviceUserID,
					Type: schema.ServiceUserPrincipal,
					ServiceUser: &serviceuser.ServiceUser{
						ID:    serviceUserID,
						OrgID: orgID,
					},
				}, nil)

				org.EXPECT().Get(mock.Anything, orgID).Return(
					organization.Organization{
						ID:    orgID,
						State: organization.Enabled,
					}, nil)

				authn.EXPECT().BuildToken(mock.Anything, mock.AnythingOfType("authenticate.Principal"), mock.AnythingOfType("map[string]string")).Return(expectedToken, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.AuthTokenRequest{}),
			want: connect.NewResponse(&frontierv1beta1.AuthTokenResponse{
				AccessToken: "test-access-token",
				TokenType:   "Bearer",
			}),
			wantErr:     false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthnSrv := new(mocks.AuthnService)
			mockOrgSrv := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockAuthnSrv, mockOrgSrv)
			}

			handler := &ConnectHandler{
				authnService: mockAuthnSrv,
				orgService:   mockOrgSrv,
				authConfig: authenticate.Config{
					Token: authenticate.TokenConfig{
						Claims: authenticate.TokenClaimConfig{
							AddOrgIDsClaim: false,
						},
					},
				},
			}

			logger := zap.NewNop()
			ctx := context.WithValue(context.Background(), "logger", logger)
			resp, err := handler.AuthToken(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					if tt.expectedErr == organization.ErrDisabled {
						connectErr := err.(*connect.Error)
						assert.Equal(t, connect.CodeInternal, connectErr.Code())
						assert.Contains(t, connectErr.Message(), "org is disabled")
					} else {
						connectErr := err.(*connect.Error)
						assert.Equal(t, connect.CodeInternal, connectErr.Code())
					}
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.want.Msg.GetAccessToken(), resp.Msg.GetAccessToken())
				assert.Equal(t, tt.want.Msg.GetTokenType(), resp.Msg.GetTokenType())
			}
		})
	}
}

func TestConnectHandler_GetJWKs(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(authn *mocks.AuthnService)
		request     *connect.Request[frontierv1beta1.GetJWKsRequest]
		want        *connect.Response[frontierv1beta1.GetJWKsResponse]
		wantErr     bool
		expectedErr error
	}{
		{
			name: "should return jwks successfully",
			setup: func(authn *mocks.AuthnService) {
				// Create a test key set
				testKeySet := jwk.NewSet()
				testKey, _ := jwk.FromRaw([]byte("test-key-data"))
				testKey.Set(jwk.KeyIDKey, "test-key-id")
				testKey.Set(jwk.KeyTypeKey, "oct")
				testKeySet.AddKey(testKey)

				authn.EXPECT().JWKs(mock.Anything).Return(testKeySet)
			},
			request: connect.NewRequest(&frontierv1beta1.GetJWKsRequest{}),
			want: connect.NewResponse(&frontierv1beta1.GetJWKsResponse{
				Keys: []*frontierv1beta1.JSONWebKey{
					{
						Kid: "test-key-id",
						Kty: "oct",
					},
				},
			}),
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "should return empty keys when keySet is empty",
			setup: func(authn *mocks.AuthnService) {
				emptyKeySet := jwk.NewSet()
				authn.EXPECT().JWKs(mock.Anything).Return(emptyKeySet)
			},
			request: connect.NewRequest(&frontierv1beta1.GetJWKsRequest{}),
			want: connect.NewResponse(&frontierv1beta1.GetJWKsResponse{
				Keys: []*frontierv1beta1.JSONWebKey{},
			}),
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "should handle multiple keys in keySet",
			setup: func(authn *mocks.AuthnService) {
				testKeySet := jwk.NewSet()

				// First key
				testKey1, _ := jwk.FromRaw([]byte("test-key-data-1"))
				testKey1.Set(jwk.KeyIDKey, "test-key-id-1")
				testKey1.Set(jwk.KeyTypeKey, "oct")
				testKeySet.AddKey(testKey1)

				// Second key
				testKey2, _ := jwk.FromRaw([]byte("test-key-data-2"))
				testKey2.Set(jwk.KeyIDKey, "test-key-id-2")
				testKey2.Set(jwk.KeyTypeKey, "oct")
				testKeySet.AddKey(testKey2)

				authn.EXPECT().JWKs(mock.Anything).Return(testKeySet)
			},
			request: connect.NewRequest(&frontierv1beta1.GetJWKsRequest{}),
			want: connect.NewResponse(&frontierv1beta1.GetJWKsResponse{
				Keys: []*frontierv1beta1.JSONWebKey{
					{
						Kid: "test-key-id-1",
						Kty: "oct",
					},
					{
						Kid: "test-key-id-2",
						Kty: "oct",
					},
				},
			}),
			wantErr:     false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthnSrv := new(mocks.AuthnService)
			if tt.setup != nil {
				tt.setup(mockAuthnSrv)
			}

			handler := &ConnectHandler{
				authnService: mockAuthnSrv,
			}

			resp, err := handler.GetJWKs(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					connectErr := err.(*connect.Error)
					assert.Equal(t, connect.CodeInternal, connectErr.Code())
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)

				// Verify the response has the expected structure
				assert.NotNil(t, resp.Msg)
				assert.NotNil(t, resp.Msg.GetKeys())
				assert.Equal(t, len(tt.want.Msg.GetKeys()), len(resp.Msg.GetKeys()))

				// Verify each key matches expected properties
				for i, expectedKey := range tt.want.Msg.GetKeys() {
					if i < len(resp.Msg.GetKeys()) {
						actualKey := resp.Msg.GetKeys()[i]
						assert.Equal(t, expectedKey.GetKid(), actualKey.GetKid())
						assert.Equal(t, expectedKey.GetKty(), actualKey.GetKty())
					}
				}
			}
		})
	}
}

func TestToJSONWebKey(t *testing.T) {
	tests := []struct {
		name        string
		keySet      jwk.Set
		expectError bool
	}{
		{
			name: "should convert valid key set to JSON web key",
			keySet: func() jwk.Set {
				keySet := jwk.NewSet()
				testKey, _ := jwk.FromRaw([]byte("test-key-data"))
				testKey.Set(jwk.KeyIDKey, "test-key-id")
				testKey.Set(jwk.KeyTypeKey, "oct")
				keySet.AddKey(testKey)
				return keySet
			}(),
			expectError: false,
		},
		{
			name:        "should handle empty key set",
			keySet:      jwk.NewSet(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toJSONWebKey(tt.keySet)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Keys)

				// Verify the structure is correct
				keySetJson, _ := json.Marshal(tt.keySet)
				var expectedJWKS JsonWebKeySet
				json.Unmarshal(keySetJson, &expectedJWKS)
				assert.Equal(t, len(expectedJWKS.Keys), len(result.Keys))
			}
		})
	}
}
