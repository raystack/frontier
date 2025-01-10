package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_Authenticate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(authn *mocks.AuthnService, session *mocks.SessionService)
		request *frontierv1beta1.AuthenticateRequest
		want    *frontierv1beta1.AuthenticateResponse
		wantErr error
	}{
		{
			name: "should return an error if session service returns an error",
			setup: func(authn *mocks.AuthnService, session *mocks.SessionService) {
				authn.EXPECT().SanitizeReturnToURL("").Return("")
				authn.EXPECT().SanitizeCallbackURL("").Return("")
				session.EXPECT().ExtractFromContext(mock.AnythingOfType("context.backgroundCtx")).Return(nil, errors.New("new-error"))
			},
			request: &frontierv1beta1.AuthenticateRequest{},
			wantErr: status.Error(codes.Internal, "new-error"),
			want:    nil,
		},
		{
			name: "should return empty response if session is alreay there",
			setup: func(authn *mocks.AuthnService, session *mocks.SessionService) {
				authn.EXPECT().SanitizeReturnToURL("").Return("")
				authn.EXPECT().SanitizeCallbackURL("").Return("")
				session.EXPECT().ExtractFromContext(mock.AnythingOfType("context.backgroundCtx")).Return(&frontiersession.Session{
					ExpiresAt:       time.Now().Add(1 * time.Hour),
					AuthenticatedAt: time.Now(),
				}, nil)
			},
			request: &frontierv1beta1.AuthenticateRequest{},
			wantErr: nil,
			want:    &frontierv1beta1.AuthenticateResponse{},
		},
		{
			name: "should return an error if auth service returns an error",
			setup: func(authn *mocks.AuthnService, session *mocks.SessionService) {
				authn.EXPECT().SanitizeReturnToURL("").Return("")
				authn.EXPECT().SanitizeCallbackURL("").Return("")
				session.EXPECT().ExtractFromContext(mock.AnythingOfType("context.backgroundCtx")).Return(nil, frontiersession.ErrNoSession)
				authn.EXPECT().StartFlow(mock.AnythingOfType("context.backgroundCtx"), authenticate.RegistrationStartRequest{}).Return(nil, errors.New("new-error"))
			},
			request: &frontierv1beta1.AuthenticateRequest{},
			wantErr: status.Error(codes.Internal, "new-error"),
			want:    nil,
		},
		{
			name: "should create state and endpoint for callback",
			setup: func(authn *mocks.AuthnService, session *mocks.SessionService) {
				authn.EXPECT().SanitizeReturnToURL("").Return("")
				authn.EXPECT().SanitizeCallbackURL("").Return("")
				session.EXPECT().ExtractFromContext(mock.AnythingOfType("context.backgroundCtx")).Return(nil, frontiersession.ErrNoSession)
				authn.EXPECT().StartFlow(mock.AnythingOfType("context.backgroundCtx"), authenticate.RegistrationStartRequest{
					Email:       "",
					Method:      authenticate.MailOTPAuthMethod.String(),
					ReturnToURL: "",
					CallbackUrl: "",
				}).Return(&authenticate.RegistrationStartResponse{
					Flow: &authenticate.Flow{
						StartURL: "",
					},
					State: "",
				}, nil)
			},
			request: &frontierv1beta1.AuthenticateRequest{
				StrategyName: authenticate.MailOTPAuthMethod.String(),
				Email:        "",
			},
			wantErr: nil,
			want: &frontierv1beta1.AuthenticateResponse{
				Endpoint: "",
				State:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthnSrv := new(mocks.AuthnService)
			mockSessionSrv := new(mocks.SessionService)
			if tt.setup != nil {
				tt.setup(mockAuthnSrv, mockSessionSrv)
			}
			mockDep := Handler{authnService: mockAuthnSrv, sessionService: mockSessionSrv}
			resp, err := mockDep.Authenticate(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
