package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnectHandler_RevokeSession(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*mocks.SessionService, *mocks.AuthnService)
		request     *connect.Request[frontierv1beta1.RevokeSessionRequest]
		wantErr     bool
		wantErrCode connect.Code
	}{
		{
			name: "should successfully revoke session when user owns the session",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				sessionID := uuid.New()
				
				// Mock authentication
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				
				// Mock session retrieval - use mock.Anything for sessionID to match any UUID
				sessionSvc.On("GetSession", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&frontiersession.Session{
					ID:     sessionID,
					UserID: userID,
				}, nil)
				
				// Mock session deletion
				sessionSvc.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeSessionRequest{
				SessionId: uuid.New().String(),
			}),
			wantErr: false,
		},
		{
			name: "should return unauthenticated error when user is not authenticated",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{}, errors.New("not authenticated"))
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeSessionRequest{
				SessionId: uuid.New().String(),
			}),
			wantErr:     true,
			wantErrCode: connect.CodeUnauthenticated,
		},
		{
			name: "should return invalid argument error when session ID is invalid",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeSessionRequest{
				SessionId: "invalid-uuid",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return not found error when session does not exist",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				sessionSvc.On("GetSession", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("session not found"))
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeSessionRequest{
				SessionId: uuid.New().String(),
			}),
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
		},
		{
			name: "should return not found error when user tries to revoke someone else's session",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				otherUserID := "other-user-456"
				
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				sessionSvc.On("GetSession", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&frontiersession.Session{
					ID:     uuid.New(),
					UserID: otherUserID, // Different user
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeSessionRequest{
				SessionId: uuid.New().String(),
			}),
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
		},
		{
			name: "should return internal error when session deletion fails",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				sessionSvc.On("GetSession", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&frontiersession.Session{
					ID:     uuid.New(),
					UserID: userID,
				}, nil)
				sessionSvc.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(errors.New("database error"))
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeSessionRequest{
				SessionId: uuid.New().String(),
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionSvc := new(mocks.SessionService)
			mockAuthnSvc := new(mocks.AuthnService)
			
			if tt.setup != nil {
				tt.setup(mockSessionSvc, mockAuthnSvc)
			}
			
			handler := &ConnectHandler{
				sessionService: mockSessionSvc,
				authnService:   mockAuthnSvc,
			}
			
			resp, err := handler.RevokeSession(context.Background(), tt.request)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrCode != 0 {
					var connectErr *connect.Error
					if assert.True(t, errors.As(err, &connectErr)) {
						assert.Equal(t, tt.wantErrCode, connectErr.Code())
					}
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.IsType(t, &frontierv1beta1.RevokeSessionResponse{}, resp.Msg)
			}
		})
	}
}

func TestConnectHandler_ListSessions(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*mocks.SessionService, *mocks.AuthnService)
		request     *connect.Request[frontierv1beta1.ListSessionsRequest]
		wantErr     bool
		wantErrCode connect.Code
		wantCount   int
	}{
		{
			name: "should successfully return user sessions",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				now := time.Now().UTC()
				
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				
				// Mock current session extraction
				currentSessionID := uuid.New()
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(&frontiersession.Session{
					ID: currentSessionID,
				}, nil)
				
				// Mock sessions list
				sessions := []*frontiersession.Session{
					{
						ID:              uuid.New(),
						UserID:          userID,
						AuthenticatedAt: now.Add(-2 * time.Hour),
						ExpiresAt:       now.Add(2 * time.Hour),
						CreatedAt:       now.Add(-2 * time.Hour),
						UpdatedAt:       now.Add(-1 * time.Hour),
						Metadata: frontiersession.SessionMetadata{
							OperatingSystem: "Mac OS X",
							Browser:         "Chrome",
							IpAddress:       "192.168.1.1",
							Location: struct {
								Country string
								City    string
							}{
								Country: "US",
								City:    "San Francisco",
							},
						},
					},
					{
						ID:              currentSessionID,
						UserID:          userID,
						AuthenticatedAt: now.Add(-1 * time.Hour),
						ExpiresAt:       now.Add(1 * time.Hour),
						CreatedAt:       now.Add(-1 * time.Hour),
						UpdatedAt:       now.Add(-30 * time.Minute),
						Metadata: frontiersession.SessionMetadata{
							OperatingSystem: "Windows",
							Browser:         "Firefox",
							IpAddress:       "192.168.1.2",
							Location: struct {
								Country string
								City    string
							}{
								Country: "US",
								City:    "New York",
							},
						},
					},
				}
				
				sessionSvc.On("ListSessions", mock.Anything, userID).Return(sessions, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr: false,
			wantCount: 2,
		},
		{
			name: "should return unauthenticated error when user is not authenticated",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{}, errors.New("not authenticated"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr:     true,
			wantErrCode: connect.CodeUnauthenticated,
		},
		{
			name: "should return internal error when session service fails",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(nil, errors.New("no session"))
				sessionSvc.On("ListSessions", mock.Anything, userID).Return(nil, errors.New("database error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
		{
			name: "should return empty list when user has no sessions",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(nil, errors.New("no session"))
				sessionSvc.On("ListSessions", mock.Anything, userID).Return([]*frontiersession.Session{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr: false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionSvc := new(mocks.SessionService)
			mockAuthnSvc := new(mocks.AuthnService)
			
			if tt.setup != nil {
				tt.setup(mockSessionSvc, mockAuthnSvc)
			}
			
			handler := &ConnectHandler{
				sessionService: mockSessionSvc,
				authnService:   mockAuthnSvc,
			}
			
			resp, err := handler.ListSessions(context.Background(), tt.request)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrCode != 0 {
					var connectErr *connect.Error
					if assert.True(t, errors.As(err, &connectErr)) {
						assert.Equal(t, tt.wantErrCode, connectErr.Code())
					}
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.IsType(t, &frontierv1beta1.ListSessionsResponse{}, resp.Msg)
				assert.Len(t, resp.Msg.Sessions, tt.wantCount)
			}
		})
	}
}
