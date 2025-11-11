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
				sessionSvc.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&frontiersession.Session{
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
				sessionSvc.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("session not found"))
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
				sessionSvc.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&frontiersession.Session{
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
				sessionSvc.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(&frontiersession.Session{
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
								Country   string
								City      string
								Latitude  string
								Longitude string
							}{
								Country:   "US",
								City:      "San Francisco",
								Latitude:  "",
								Longitude: "",
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
								Country   string
								City      string
								Latitude  string
								Longitude string
							}{
								Country:   "US",
								City:      "New York",
								Latitude:  "",
								Longitude: "",
							},
						},
					},
				}

				sessionSvc.On("List", mock.Anything, userID).Return(sessions, nil)
			},
			request:   connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "should return unauthenticated error when user is not authenticated",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{}, errors.New("not authenticated"))
			},
			request:     connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr:     true,
			wantErrCode: connect.CodeUnauthenticated,
		},
		{
			name: "should return internal error when session service fails",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(nil, errors.New("no session"))
				sessionSvc.On("List", mock.Anything, userID).Return(nil, errors.New("database error"))
			},
			request:     connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
		{
			name: "should return empty list when user has no sessions",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				authnSvc.On("GetPrincipal", mock.Anything, authenticate.SessionClientAssertion).Return(authenticate.Principal{ID: userID}, nil)
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(nil, errors.New("no session"))
				sessionSvc.On("List", mock.Anything, userID).Return([]*frontiersession.Session{}, nil)
			},
			request:   connect.NewRequest(&frontierv1beta1.ListSessionsRequest{}),
			wantErr:   false,
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
				assert.Len(t, resp.Msg.GetSessions(), tt.wantCount)

				// Verify location fields are present in all sessions
				for _, session := range resp.Msg.GetSessions() {
					if session.GetMetadata() != nil {
						location := session.GetMetadata().GetLocation()
						assert.NotNil(t, location)
						// Verify latitude and longitude fields exist (can be empty)
						_ = location.GetLatitude()
						_ = location.GetLongitude()
					}
				}
			}
		})
	}
}

func TestConnectHandler_PingUserSession(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*mocks.SessionService, *mocks.AuthnService)
		request     *connect.Request[frontierv1beta1.PingUserSessionRequest]
		wantErr     bool
		wantErrCode connect.Code
	}{
		{
			name: "should successfully ping user session and update metadata",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				sessionID := uuid.New()
				now := time.Now().UTC()

				// Mock session extraction from context
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(&frontiersession.Session{
					ID:              sessionID,
					UserID:          userID,
					AuthenticatedAt: now.Add(-1 * time.Hour),
					ExpiresAt:       now.Add(1 * time.Hour), // Valid session
					CreatedAt:       now.Add(-1 * time.Hour),
					UpdatedAt:       now.Add(-30 * time.Minute),
					Metadata: frontiersession.SessionMetadata{
						OperatingSystem: "Mac OS X",
						Browser:         "Chrome",
						IpAddress:       "192.168.1.1",
						Location: struct {
							Country   string
							City      string
							Latitude  string
							Longitude string
						}{
							Country:   "",
							City:      "",
							Latitude:  "",
							Longitude: "",
						},
					},
				}, nil)

				// Mock session metadata update
				sessionSvc.On("Ping", mock.Anything, sessionID, mock.AnythingOfType("session.SessionMetadata")).Return(nil)

				// Mock GetByID to return updated session
				sessionSvc.On("GetByID", mock.Anything, sessionID).Return(&frontiersession.Session{
					ID:              sessionID,
					UserID:          userID,
					AuthenticatedAt: now.Add(-1 * time.Hour),
					ExpiresAt:       now.Add(1 * time.Hour),
					CreatedAt:       now.Add(-1 * time.Hour),
					UpdatedAt:       now.Add(-30 * time.Minute),
					Metadata: frontiersession.SessionMetadata{
						OperatingSystem: "Mac OS X",
						Browser:         "Chrome",
						IpAddress:       "192.168.1.1",
						Location: struct {
							Country   string
							City      string
							Latitude  string
							Longitude string
						}{
							Country:   "",
							City:      "",
							Latitude:  "",
							Longitude: "",
						},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.PingUserSessionRequest{}),
			wantErr: false,
		},
		{
			name: "should return unauthenticated error when no session in context",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(nil, errors.New("no session"))
			},
			request:     connect.NewRequest(&frontierv1beta1.PingUserSessionRequest{}),
			wantErr:     true,
			wantErrCode: connect.CodeUnauthenticated,
		},
		{
			name: "should return unauthenticated error when session is expired",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				sessionID := uuid.New()
				now := time.Now().UTC()

				// Mock expired session
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(&frontiersession.Session{
					ID:              sessionID,
					UserID:          userID,
					AuthenticatedAt: now.Add(-3 * time.Hour),
					ExpiresAt:       now.Add(-1 * time.Hour), // Expired session
					CreatedAt:       now.Add(-3 * time.Hour),
					UpdatedAt:       now.Add(-2 * time.Hour),
					Metadata: frontiersession.SessionMetadata{
						Location: struct {
							Country   string
							City      string
							Latitude  string
							Longitude string
						}{
							Country:   "",
							City:      "",
							Latitude:  "",
							Longitude: "",
						},
					},
				}, nil)
			},
			request:     connect.NewRequest(&frontierv1beta1.PingUserSessionRequest{}),
			wantErr:     true,
			wantErrCode: connect.CodeUnauthenticated,
		},
		{
			name: "should return internal error when ping session fails",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				sessionID := uuid.New()
				now := time.Now().UTC()

				// Mock valid session
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(&frontiersession.Session{
					ID:              sessionID,
					UserID:          userID,
					AuthenticatedAt: now.Add(-1 * time.Hour),
					ExpiresAt:       now.Add(1 * time.Hour), // Valid session
					CreatedAt:       now.Add(-1 * time.Hour),
					UpdatedAt:       now.Add(-30 * time.Minute),
					Metadata: frontiersession.SessionMetadata{
						Location: struct {
							Country   string
							City      string
							Latitude  string
							Longitude string
						}{
							Country:   "",
							City:      "",
							Latitude:  "",
							Longitude: "",
						},
					},
				}, nil)

				// Mock ping session failure
				sessionSvc.On("Ping", mock.Anything, sessionID, mock.AnythingOfType("session.SessionMetadata")).Return(errors.New("database error"))
			},
			request:     connect.NewRequest(&frontierv1beta1.PingUserSessionRequest{}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
		{
			name: "should handle session with valid metadata extraction",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := "user-123"
				sessionID := uuid.New()
				now := time.Now().UTC()

				// Mock session with existing metadata
				sessionSvc.On("ExtractFromContext", mock.Anything).Return(&frontiersession.Session{
					ID:              sessionID,
					UserID:          userID,
					AuthenticatedAt: now.Add(-1 * time.Hour),
					ExpiresAt:       now.Add(1 * time.Hour),
					CreatedAt:       now.Add(-1 * time.Hour),
					UpdatedAt:       now.Add(-30 * time.Minute),
					Metadata: frontiersession.SessionMetadata{
						OperatingSystem: "Windows",
						Browser:         "Firefox",
						IpAddress:       "10.0.0.1",
						Location: struct {
							Country   string
							City      string
							Latitude  string
							Longitude string
						}{
							Country:   "US",
							City:      "New York",
							Latitude:  "",
							Longitude: "",
						},
					},
				}, nil)

				// Mock successful ping with metadata update
				sessionSvc.On("Ping", mock.Anything, sessionID, mock.AnythingOfType("session.SessionMetadata")).Return(nil)

				// Mock GetByID to return updated session with location data
				sessionSvc.On("GetByID", mock.Anything, sessionID).Return(&frontiersession.Session{
					ID:              sessionID,
					UserID:          userID,
					AuthenticatedAt: now.Add(-1 * time.Hour),
					ExpiresAt:       now.Add(1 * time.Hour),
					CreatedAt:       now.Add(-1 * time.Hour),
					UpdatedAt:       now.Add(-30 * time.Minute),
					Metadata: frontiersession.SessionMetadata{
						OperatingSystem: "Windows",
						Browser:         "Firefox",
						IpAddress:       "10.0.0.1",
						Location: struct {
							Country   string
							City      string
							Latitude  string
							Longitude string
						}{
							Country:   "US",
							City:      "New York",
							Latitude:  "40.7128",
							Longitude: "-74.0060",
						},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.PingUserSessionRequest{}),
			wantErr: false,
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
				authConfig: authenticate.Config{
					Session: authenticate.SessionConfig{
						Headers: authenticate.SessionMetadataHeaders{
							ClientIP:        "X-Forwarded-For",
							ClientCountry:   "X-Country",
							ClientCity:      "X-City",
							ClientLatitude:  "CloudFront-Viewer-Latitude",
							ClientLongitude: "CloudFront-Viewer-Longitude",
							ClientUserAgent: "User-Agent",
						},
					},
				},
			}

			resp, err := handler.PingUserSession(context.Background(), tt.request)

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
				assert.IsType(t, &frontierv1beta1.PingUserSessionResponse{}, resp.Msg)

				// Verify metadata is present
				if resp.Msg.GetMetadata() != nil {
					location := resp.Msg.GetMetadata().GetLocation()
					if location != nil {
						// Verify location fields are present (can be empty strings)
						assert.NotNil(t, location)
						// If this is the test case with location data, verify values
						if location.GetCity() == "New York" {
							assert.Equal(t, "US", location.GetCountry())
							assert.Equal(t, "New York", location.GetCity())
							assert.Equal(t, "40.7128", location.GetLatitude())
							assert.Equal(t, "-74.0060", location.GetLongitude())
						}
					}
				}
			}
		})
	}
}

func TestConnectHandler_ListUserSessions(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*mocks.SessionService, *mocks.AuthnService)
		request     *connect.Request[frontierv1beta1.ListUserSessionsRequest]
		wantErr     bool
		wantErrCode connect.Code
		wantCount   int
	}{
		{
			name: "should successfully return user sessions for admin",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				userID := uuid.New().String()
				now := time.Now().UTC()

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
								Country   string
								City      string
								Latitude  string
								Longitude string
							}{
								Country:   "US",
								City:      "San Francisco",
								Latitude:  "",
								Longitude: "",
							},
						},
					},
					{
						ID:              uuid.New(),
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
								Country   string
								City      string
								Latitude  string
								Longitude string
							}{
								Country:   "US",
								City:      "New York",
								Latitude:  "",
								Longitude: "",
							},
						},
					},
				}

				sessionSvc.On("List", mock.Anything, mock.AnythingOfType("string")).Return(sessions, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListUserSessionsRequest{
				UserId: uuid.New().String(),
			}),
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "should return invalid argument error when user_id is empty",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				// No setup needed for this test
			},
			request: connect.NewRequest(&frontierv1beta1.ListUserSessionsRequest{
				UserId: "",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return invalid argument error when user_id is not a valid UUID",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				// No setup needed for this test
			},
			request: connect.NewRequest(&frontierv1beta1.ListUserSessionsRequest{
				UserId: "invalid-uuid",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return internal error when session service fails",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				sessionSvc.On("List", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("database error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListUserSessionsRequest{
				UserId: uuid.New().String(),
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
		{
			name: "should return empty list when user has no sessions",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				sessionSvc.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*frontiersession.Session{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListUserSessionsRequest{
				UserId: uuid.New().String(),
			}),
			wantErr:   false,
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

			resp, err := handler.ListUserSessions(context.Background(), tt.request)

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
				assert.IsType(t, &frontierv1beta1.ListUserSessionsResponse{}, resp.Msg)
				assert.Len(t, resp.Msg.GetSessions(), tt.wantCount)

				// Verify location fields are present in all sessions
				for _, session := range resp.Msg.GetSessions() {
					if session.GetMetadata() != nil {
						location := session.GetMetadata().GetLocation()
						assert.NotNil(t, location)
						// Verify latitude and longitude fields exist (can be empty)
						_ = location.GetLatitude()
						_ = location.GetLongitude()
					}
				}
			}
		})
	}
}

func TestConnectHandler_RevokeUserSession(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*mocks.SessionService, *mocks.AuthnService)
		request     *connect.Request[frontierv1beta1.RevokeUserSessionRequest]
		wantErr     bool
		wantErrCode connect.Code
	}{
		{
			name: "should successfully revoke user session for admin",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				// Mock session deletion
				sessionSvc.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeUserSessionRequest{
				SessionId: uuid.New().String(),
			}),
			wantErr: false,
		},
		{
			name: "should return invalid argument error when session_id is invalid",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				// No setup needed for this test
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeUserSessionRequest{
				SessionId: "invalid-uuid",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return internal error when session deletion fails",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				// Mock session deletion failure
				sessionSvc.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(errors.New("database error"))
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeUserSessionRequest{
				SessionId: uuid.New().String(),
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
		{
			name: "should handle session not found gracefully",
			setup: func(sessionSvc *mocks.SessionService, authnSvc *mocks.AuthnService) {
				// Mock session not found error
				sessionSvc.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(frontiersession.ErrNoSession)
			},
			request: connect.NewRequest(&frontierv1beta1.RevokeUserSessionRequest{
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

			resp, err := handler.RevokeUserSession(context.Background(), tt.request)

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
				assert.IsType(t, &frontierv1beta1.RevokeUserSessionResponse{}, resp.Msg)
			}
		})
	}
}
