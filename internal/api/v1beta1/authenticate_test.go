package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	shieldsession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_Authenticate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.AuthnService, ss *mocks.SessionService)
		request *frontierv1beta1.AuthenticateRequest
		want    *frontierv1beta1.AuthenticateResponse
		wantErr error
	}{
		{
			name: "should authenticate user if the user is already logged in",
			setup: func(as *mocks.AuthnService, ss *mocks.SessionService) {
				ss.EXPECT().ExtractFromContext(mock.AnythingOfType("*context.emptyCtx")).Return(&shieldsession.Session{
					ExpiresAt:       time.Now().UTC().Add(7 * 24 * time.Hour),
					AuthenticatedAt: time.Now().UTC(),
				}, nil)
			},
			request: &frontierv1beta1.AuthenticateRequest{},
			want:    &frontierv1beta1.AuthenticateResponse{},
		},
		{
			name: "should try to authenticate user if the user is not logged in",
			setup: func(as *mocks.AuthnService, ss *mocks.SessionService) {
				ss.EXPECT().ExtractFromContext(mock.AnythingOfType("*context.emptyCtx")).Return(&shieldsession.Session{}, nil)
				as.EXPECT().StartFlow(mock.AnythingOfType("*context.emptyCtx"), authenticate.RegistrationStartRequest{
					Email:    "test@raystack.org",
					ReturnTo: "http://localhost:8080",
					Method:   "",
				}).Return(&authenticate.RegistrationStartResponse{
					Flow: &authenticate.Flow{
						ID:        uuid.New(),
						Method:    "mailOtp",
						Email:     "test@raystack.org",
						FinishURL: "http://localhost:8080",
						StartURL:  "http://localhost:8080",
					},
					State: "test-state",
				}, nil)
			},
			request: &frontierv1beta1.AuthenticateRequest{
				Email:    "test@raystack.org",
				ReturnTo: "http://localhost:8080",
				Redirect: false,
			},
			want: &frontierv1beta1.AuthenticateResponse{
				Endpoint: "http://localhost:8080",
				State:    "test-state",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthSvc := new(mocks.AuthnService)
			mockSessionSvc := new(mocks.SessionService)
			if tt.setup != nil {
				tt.setup(mockAuthSvc, mockSessionSvc)
			}
			h := &Handler{
				authnService:   mockAuthSvc,
				sessionService: mockSessionSvc,
			}
			got, err := h.Authenticate(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_ListAuthStrategies(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.AuthnService)
		request *frontierv1beta1.ListAuthStrategiesRequest
		want    *frontierv1beta1.ListAuthStrategiesResponse
		wantErr error
	}{
		{
			name: "should return list of auth strategies",
			setup: func(as *mocks.AuthnService) {
				as.EXPECT().SupportedStrategies().Return([]string{"mailOtp", "smsOtp"})
			},
			request: &frontierv1beta1.ListAuthStrategiesRequest{},
			want: &frontierv1beta1.ListAuthStrategiesResponse{
				Strategies: []*frontierv1beta1.AuthStrategy{
					{
						Name: "mailOtp",
					},
					{
						Name: "smsOtp",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthSvc := new(mocks.AuthnService)
			if tt.setup != nil {
				tt.setup(mockAuthSvc)
			}
			h := &Handler{
				authnService: mockAuthSvc,
			}
			got, err := h.ListAuthStrategies(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}
