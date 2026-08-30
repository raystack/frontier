package v1beta1connect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/consent"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontiererrors "github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
					authenticate.JWTGrantClientAssertion,
					authenticate.PATClientAssertion).Return(authenticate.Principal{
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
					authenticate.JWTGrantClientAssertion,
					authenticate.PATClientAssertion).Return(authenticate.Principal{
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

			ctx := context.Background()
			resp, err := handler.AuthToken(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					connectErr := err.(*connect.Error)
					if tt.expectedErr == organization.ErrDisabled {
						assert.Equal(t, connect.CodeFailedPrecondition, connectErr.Code())
						assert.Contains(t, connectErr.Message(), "org is disabled")
					} else {
						assert.Equal(t, connect.CodeInternal, connectErr.Code())
						assert.Equal(t, "internal server error", connectErr.Message())
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

func TestConnectHandler_AuthToken_RejectsDisallowedCredential(t *testing.T) {
	mockAuthnSrv := new(mocks.AuthnService)
	mockOrgSrv := new(mocks.OrganizationService)

	// a credential type outside AuthToken's allowed list resolves to unauthenticated
	mockAuthnSrv.EXPECT().GetPrincipal(mock.Anything,
		authenticate.SessionClientAssertion,
		authenticate.ClientCredentialsClientAssertion,
		authenticate.JWTGrantClientAssertion,
		authenticate.PATClientAssertion).
		Return(authenticate.Principal{}, frontiererrors.ErrUnauthenticated)

	handler := &ConnectHandler{
		authnService: mockAuthnSrv,
		orgService:   mockOrgSrv,
		authConfig:   authenticate.Config{},
	}

	_, err := handler.AuthToken(context.Background(), connect.NewRequest(&frontierv1beta1.AuthTokenRequest{}))
	assert.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
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
				_ = testKey.Set(jwk.KeyIDKey, "test-key-id")
				_ = testKey.Set(jwk.KeyTypeKey, "oct")
				_ = testKeySet.AddKey(testKey)

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
				_ = testKey1.Set(jwk.KeyIDKey, "test-key-id-1")
				_ = testKey1.Set(jwk.KeyTypeKey, "oct")
				_ = testKeySet.AddKey(testKey1)

				// Second key
				testKey2, _ := jwk.FromRaw([]byte("test-key-data-2"))
				_ = testKey2.Set(jwk.KeyIDKey, "test-key-id-2")
				_ = testKey2.Set(jwk.KeyTypeKey, "oct")
				_ = testKeySet.AddKey(testKey2)

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
				_ = testKey.Set(jwk.KeyIDKey, "test-key-id")
				_ = testKey.Set(jwk.KeyTypeKey, "oct")
				_ = keySet.AddKey(testKey)
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
				require.NoError(t, json.Unmarshal(keySetJson, &expectedJWKS))
				assert.Equal(t, len(expectedJWKS.Keys), len(result.Keys))
			}
		})
	}
}

// testSessionHeaders is the default header set, so ExtractSessionMetadata reads
// the same headers the server reads.
func testSessionHeaders() authenticate.Config {
	return authenticate.Config{
		Session: authenticate.SessionConfig{
			Headers: authenticate.SessionMetadataHeaders{
				ClientIP:        "x-forwarded-for",
				ClientUserAgent: "User-Agent",
			},
		},
	}
}

// TestConnectHandler_Authenticate_PassesIntentConsentAndIP pins what the handler
// hands StartFlow. Authenticate is on the authentication skip list, so nothing
// puts session metadata on the context and the handler extracts it itself.
func TestConnectHandler_Authenticate_PassesIntentConsentAndIP(t *testing.T) {
	ctx := context.Background()

	mockAuthnSrv := mocks.NewAuthnService(t)
	mockSessionSrv := mocks.NewSessionService(t)

	mockAuthnSrv.EXPECT().SanitizeReturnToURL("https://example.org/done").Return("https://example.org/done")
	mockAuthnSrv.EXPECT().SanitizeCallbackURL("").Return("https://example.org/callback")
	mockSessionSrv.EXPECT().ExtractFromContext(ctx).Return(nil, frontiersession.ErrNoSession)

	var startRequest authenticate.RegistrationStartRequest
	mockAuthnSrv.EXPECT().StartFlow(ctx, mock.Anything).
		Run(func(_ context.Context, req authenticate.RegistrationStartRequest) { startRequest = req }).
		Return(&authenticate.RegistrationStartResponse{Flow: &authenticate.Flow{}}, nil)

	handler := &ConnectHandler{
		authnService:   mockAuthnSrv,
		sessionService: mockSessionSrv,
		authConfig:     testSessionHeaders(),
	}

	request := connect.NewRequest(&frontierv1beta1.AuthenticateRequest{
		StrategyName:        authenticate.MailOTPAuthMethod.String(),
		Email:               "test@example.com",
		ReturnTo:            "https://example.org/done",
		FlowIntent:          frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP,
		AcceptedDocumentIds: []string{"terms_of_service", "privacy_policy"},
	})
	request.Header().Set("x-forwarded-for", "203.0.113.9, 10.0.0.1")
	request.Header().Set("User-Agent", "Mozilla/5.0 (Macintosh) Chrome/124.0")

	_, err := handler.Authenticate(ctx, request)
	require.NoError(t, err)

	assert.Equal(t, authenticate.FlowIntentSignup, startRequest.Intent)
	assert.Equal(t, []string{"terms_of_service", "privacy_policy"}, startRequest.AcceptedDocumentIDs)
	// the first hop of the forwarded chain reaches the consent record, and the
	// user agent does not: the helper keeps an OS and a browser family from it
	// and drops the raw string, and neither is passed on
	assert.Equal(t, "203.0.113.9", startRequest.IPAddress)
}

// TestConnectHandler_Authenticate_RejectsIdsWithALoginIntent covers the one
// request shape the handler turns down on its own. A login writes no consent
// record, so accepting the ids would leave the client believing it recorded a
// consent that does not exist.
func TestConnectHandler_Authenticate_RejectsIdsWithALoginIntent(t *testing.T) {
	ctx := context.Background()

	mockAuthnSrv := mocks.NewAuthnService(t)
	mockSessionSrv := mocks.NewSessionService(t)
	mockConsentSrv := mocks.NewConsentService(t)
	mockAuthnSrv.EXPECT().SanitizeReturnToURL("").Return("")
	mockAuthnSrv.EXPECT().SanitizeCallbackURL("").Return("https://example.org/callback")
	mockSessionSrv.EXPECT().ExtractFromContext(ctx).Return(nil, frontiersession.ErrNoSession)
	mockConsentSrv.EXPECT().Documents().Return([]consent.Document{{ID: "terms_of_service"}})

	handler := &ConnectHandler{
		authnService:   mockAuthnSrv,
		sessionService: mockSessionSrv,
		consentService: mockConsentSrv,
		authConfig:     testSessionHeaders(),
	}

	resp, err := handler.Authenticate(ctx, connect.NewRequest(&frontierv1beta1.AuthenticateRequest{
		StrategyName:        authenticate.MailOTPAuthMethod.String(),
		Email:               "test@example.com",
		FlowIntent:          frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN,
		AcceptedDocumentIds: []string{"terms_of_service"},
	}))

	assert.Nil(t, resp)
	connectErr := err.(*connect.Error)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	assert.Equal(t, ErrConsentOnLoginIntent.Error(), connectErr.Message())
	// the flow is never started for a request the handler will not accept
	mockAuthnSrv.AssertNotCalled(t, "StartFlow", mock.Anything, mock.Anything)
}

// TestConnectHandler_Authenticate_IgnoresIdsWithALoginIntentWhenConsentIsDisabled
// is the other half of that rule. With consent off no record is written under
// any intent, so the ids mean nothing and a client built for a consent
// deployment still logs in against one that does not ask for consent.
func TestConnectHandler_Authenticate_IgnoresIdsWithALoginIntentWhenConsentIsDisabled(t *testing.T) {
	ctx := context.Background()

	mockAuthnSrv := mocks.NewAuthnService(t)
	mockSessionSrv := mocks.NewSessionService(t)
	mockConsentSrv := mocks.NewConsentService(t)
	mockAuthnSrv.EXPECT().SanitizeReturnToURL("").Return("")
	mockAuthnSrv.EXPECT().SanitizeCallbackURL("").Return("https://example.org/callback")
	mockSessionSrv.EXPECT().ExtractFromContext(ctx).Return(nil, frontiersession.ErrNoSession)
	// disabled: Documents is empty, which boot validation guarantees an enabled
	// deployment never is
	mockConsentSrv.EXPECT().Documents().Return(nil)

	var startRequest authenticate.RegistrationStartRequest
	mockAuthnSrv.EXPECT().StartFlow(ctx, mock.Anything).
		Run(func(_ context.Context, req authenticate.RegistrationStartRequest) { startRequest = req }).
		Return(&authenticate.RegistrationStartResponse{Flow: &authenticate.Flow{}}, nil)

	handler := &ConnectHandler{
		authnService:   mockAuthnSrv,
		sessionService: mockSessionSrv,
		consentService: mockConsentSrv,
		authConfig:     testSessionHeaders(),
	}

	_, err := handler.Authenticate(ctx, connect.NewRequest(&frontierv1beta1.AuthenticateRequest{
		StrategyName:        authenticate.MailOTPAuthMethod.String(),
		Email:               "test@example.com",
		FlowIntent:          frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN,
		AcceptedDocumentIds: []string{"terms_of_service"},
	}))

	require.NoError(t, err)
	assert.Equal(t, authenticate.FlowIntentLogin, startRequest.Intent)
	// passed through rather than stripped: the service resolves nothing while
	// disabled, so nothing downstream reads them
	assert.Equal(t, []string{"terms_of_service"}, startRequest.AcceptedDocumentIDs)
}

// TestConnectHandler_Authenticate_Rejections covers the three errors reaching
// the client with their own codes. Authenticate is an XHR, so it answers with
// the code and no redirect.
func TestConnectHandler_Authenticate_Rejections(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode connect.Code
		wantMsg  string
	}{
		{
			name:     "a login for an address with no account is a not found",
			err:      authenticate.ErrLoginUserNotFound,
			wantCode: connect.CodeNotFound,
			wantMsg:  authenticate.ErrLoginUserNotFound.Error(),
		},
		{
			name:     "a signup for an address that has one already exists",
			err:      authenticate.ErrSignupUserExists,
			wantCode: connect.CodeAlreadyExists,
			wantMsg:  authenticate.ErrSignupUserExists.Error(),
		},
		{
			// FailedPrecondition is what lets a client separate a consent
			// rejection from a bad code or an expired flow
			name:     "an incomplete consent is a failed precondition",
			err:      fmt.Errorf("%w: %w: terms_of_service", authenticate.ErrConsentRequired, consent.ErrMissingDocuments),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  authenticate.ErrConsentRequired.Error(),
		},
		{
			// shares the code with consent, and is told apart by the message.
			// Not InvalidArgument: the strategy name is valid, and would have
			// worked against an account that holds the credential
			name:     "a method the account cannot use is a failed precondition",
			err:      authenticate.ErrInvalidMethod,
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  authenticate.ErrInvalidMethod.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockAuthnSrv := mocks.NewAuthnService(t)
			mockSessionSrv := mocks.NewSessionService(t)
			mockAuthnSrv.EXPECT().SanitizeReturnToURL("").Return("")
			mockAuthnSrv.EXPECT().SanitizeCallbackURL("").Return("https://example.org/callback")
			mockSessionSrv.EXPECT().ExtractFromContext(ctx).Return(nil, frontiersession.ErrNoSession)
			mockAuthnSrv.EXPECT().StartFlow(ctx, mock.Anything).Return(nil, tt.err)

			handler := &ConnectHandler{
				authnService:   mockAuthnSrv,
				sessionService: mockSessionSrv,
				authConfig:     testSessionHeaders(),
			}

			resp, err := handler.Authenticate(ctx, connect.NewRequest(&frontierv1beta1.AuthenticateRequest{
				StrategyName: authenticate.MailOTPAuthMethod.String(),
				Email:        "test@example.com",
			}))

			assert.Nil(t, resp)
			connectErr := err.(*connect.Error)
			assert.Equal(t, tt.wantCode, connectErr.Code())
			// the bare sentinel, never the wrapped error: which documents were
			// missing belongs in a log, not in the response
			assert.Equal(t, tt.wantMsg, connectErr.Message())
			assert.NotContains(t, connectErr.Message(), "terms_of_service")
		})
	}
}

// TestConnectHandler_AuthCallback_Rejections covers the settled decision: the
// callback is a browser navigation, so a rejection sends the user back to the
// page they started from with a machine-readable code, not onto an error page.
func TestConnectHandler_AuthCallback_Rejections(t *testing.T) {
	tests := []struct {
		name string
		err  error

		wantCode connect.Code
		wantMsg  string
	}{
		{
			name:     "a login for an address with no account",
			err:      authenticate.ErrLoginUserNotFound,
			wantCode: connect.CodeNotFound,
			wantMsg:  authenticate.ErrLoginUserNotFound.Error(),
		},
		{
			name:     "a signup for an address that has one",
			err:      authenticate.ErrSignupUserExists,
			wantCode: connect.CodeAlreadyExists,
			wantMsg:  authenticate.ErrSignupUserExists.Error(),
		},
		{
			name:     "an incomplete consent",
			err:      fmt.Errorf("%w: %w: terms_of_service", authenticate.ErrConsentRequired, consent.ErrMissingDocuments),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  authenticate.ErrConsentRequired.Error(),
		},
		{
			name:     "a method the account cannot use",
			err:      authenticate.ErrInvalidMethod,
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  authenticate.ErrInvalidMethod.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockAuthnSrv := mocks.NewAuthnService(t)
			mockAuthnSrv.EXPECT().FinishFlow(ctx, mock.Anything).Return(nil, tt.err)

			handler := &ConnectHandler{authnService: mockAuthnSrv}

			resp, err := handler.AuthCallback(ctx, connect.NewRequest(&frontierv1beta1.AuthCallbackRequest{
				StrategyName: authenticate.MailOTPAuthMethod.String(),
				State:        "state",
				Code:         "111111",
			}))

			// the rejection is the answer. The callback URL points at a page the
			// application hosts, and that page is what called this RPC, so it
			// has the code in hand and decides where the user goes next.
			assert.Nil(t, resp)
			connectErr := err.(*connect.Error)
			assert.Equal(t, tt.wantCode, connectErr.Code())
			// the bare sentinel: the wrapped error names the documents that were
			// missing, and that belongs in the log rather than in the response
			assert.Equal(t, tt.wantMsg, connectErr.Message())
			assert.NotContains(t, connectErr.Message(), "terms_of_service")
		})
	}
}

// TestConnectHandler_AuthCallback_UnmappedErrorStaysInternal pins the boundary
// of the closed code set: an error with no code of its own is not turned into a
// redirect with some passthrough string.
func TestConnectHandler_AuthCallback_UnmappedErrorStaysInternal(t *testing.T) {
	ctx := context.Background()

	mockAuthnSrv := mocks.NewAuthnService(t)
	mockAuthnSrv.EXPECT().FinishFlow(ctx, mock.Anything).Return(nil, errors.New("database is down"))

	handler := &ConnectHandler{authnService: mockAuthnSrv}
	resp, err := handler.AuthCallback(ctx, connect.NewRequest(&frontierv1beta1.AuthCallbackRequest{
		StrategyName: authenticate.MailOTPAuthMethod.String(),
		State:        "state",
	}))

	assert.Nil(t, resp)
	assert.Equal(t, connect.CodeInternal, err.(*connect.Error).Code())
}

func TestToFlowIntent(t *testing.T) {
	assert.Equal(t, authenticate.FlowIntentLogin, toFlowIntent(frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN))
	assert.Equal(t, authenticate.FlowIntentSignup, toFlowIntent(frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP))
	assert.Equal(t, authenticate.FlowIntentUnspecified, toFlowIntent(frontierv1beta1.FlowIntent_FLOW_INTENT_UNSPECIFIED))
	// an unknown value reads as unspecified, which is the create-or-get
	// behaviour every client had before intents existed
	assert.Equal(t, authenticate.FlowIntentUnspecified, toFlowIntent(frontierv1beta1.FlowIntent(99)))
}
