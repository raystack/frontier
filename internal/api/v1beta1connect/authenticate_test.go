package v1beta1connect

import (
	"context"
	"testing"

	"connectrpc.com/connect"
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
				assert.Equal(t, tt.want.Msg.AccessToken, resp.Msg.AccessToken)
				assert.Equal(t, tt.want.Msg.TokenType, resp.Msg.TokenType)
			}
		})
	}
}
