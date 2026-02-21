package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_CreateCurrentUserPersonalToken(t *testing.T) {
	testTime := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	testCreatedAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	testOrgID := "9f256f86-31a3-11ec-8d3d-0242ac130003"
	testUserID := "8e256f86-31a3-11ec-8d3d-0242ac130003"
	defaultPATConfig := userpat.Config{
		Enabled:                true,
		TokenPrefix:            "fpt",
		MaxTokensPerUserPerOrg: 50,
		MaxTokenLifetime:       "8760h",
	}

	tests := []struct {
		name      string
		setup     func(ps *mocks.UserPATService, as *mocks.AuthnService)
		patConfig userpat.Config
		request   *connect.Request[frontierv1beta1.CreateCurrentUserPersonalTokenRequest]
		want      *frontierv1beta1.CreateCurrentUserPersonalTokenResponse
		wantErr   error
	}{
		{
			name: "should return unauthenticated error when GetLoggedInPrincipal fails",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			name: "should return permission denied when principal is not a user",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   "sv-1",
					Type: schema.ServiceUserPrincipal,
				}, nil)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated),
		},
		{
			name: "should return invalid argument when expiry is in the past",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, userpat.ErrExpiryInPast),
		},
		{
			name: "should return invalid argument when expiry exceeds max lifetime",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
			},
			patConfig: userpat.Config{
				Enabled:                true,
				TokenPrefix:            "fpt",
				MaxTokensPerUserPerOrg: 50,
				MaxTokenLifetime:       "24h",
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, userpat.ErrExpiryExceeded),
		},
		{
			name: "should return failed precondition when PAT is disabled",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(userpat.PersonalAccessToken{}, "", userpat.ErrDisabled)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeFailedPrecondition, userpat.ErrDisabled),
		},
		{
			name: "should return already exists when title conflicts",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(userpat.PersonalAccessToken{}, "", userpat.ErrConflict)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, userpat.ErrConflict),
		},
		{
			name: "should return resource exhausted when limit exceeded",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(userpat.PersonalAccessToken{}, "", userpat.ErrLimitExceeded)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeResourceExhausted, userpat.ErrLimitExceeded),
		},
		{
			name: "should return internal error for unknown service failure",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(userpat.PersonalAccessToken{}, "", errors.New("unexpected error"))
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should create token successfully and return response",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Create(mock.Anything, mock.MatchedBy(func(req userpat.CreateRequest) bool {
					return req.UserID == testUserID &&
						req.OrgID == testOrgID &&
						req.Title == "my-token" &&
						len(req.Roles) == 1 && req.Roles[0] == "role-1"
				})).Return(userpat.PersonalAccessToken{
					ID:        "pat-1",
					UserID:    testUserID,
					OrgID:     testOrgID,
					Title:     "my-token",
					ExpiresAt: testTime,
					CreatedAt: testCreatedAt,
					UpdatedAt: testCreatedAt,
				}, "fpt_abc123", nil)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want: &frontierv1beta1.CreateCurrentUserPersonalTokenResponse{
				Token: &frontierv1beta1.PersonalAccessToken{
					Id:        "pat-1",
					UserId:    testUserID,
					OrgId:     testOrgID,
					Title:     "my-token",
					Token:     "fpt_abc123",
					ExpiresAt: timestamppb.New(testTime),
					CreatedAt: timestamppb.New(testCreatedAt),
					UpdatedAt: timestamppb.New(testCreatedAt),
				},
			},
			wantErr: nil,
		},
		{
			name: "should create token with metadata",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(userpat.PersonalAccessToken{
						ID:        "pat-1",
						UserID:    testUserID,
						OrgID:     testOrgID,
						Title:     "my-token",
						ExpiresAt: testTime,
						CreatedAt: testCreatedAt,
						UpdatedAt: testCreatedAt,
						Metadata:  metadata.Metadata{"env": "staging"},
					}, "fpt_xyz789", nil)
			},
			patConfig: defaultPATConfig,
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPersonalTokenRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				Roles:     []string{"role-1"},
				ExpiresAt: timestamppb.New(testTime),
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"env": structpb.NewStringValue("staging"),
					},
				},
			}),
			want: &frontierv1beta1.CreateCurrentUserPersonalTokenResponse{
				Token: &frontierv1beta1.PersonalAccessToken{
					Id:        "pat-1",
					UserId:    testUserID,
					OrgId:     testOrgID,
					Title:     "my-token",
					Token:     "fpt_xyz789",
					ExpiresAt: timestamppb.New(testTime),
					CreatedAt: timestamppb.New(testCreatedAt),
					UpdatedAt: timestamppb.New(testCreatedAt),
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"env": structpb.NewStringValue("staging"),
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPATSrv := new(mocks.UserPATService)
			mockAuthnSrv := new(mocks.AuthnService)

			if tt.setup != nil {
				tt.setup(mockPATSrv, mockAuthnSrv)
			}

			handler := &ConnectHandler{
				userPATService: mockPATSrv,
				authnService:   mockAuthnSrv,
				patConfig:      tt.patConfig,
			}

			resp, err := handler.CreateCurrentUserPersonalToken(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, resp.Msg)
			} else {
				assert.Nil(t, resp)
			}

			mockPATSrv.AssertExpectations(t)
			mockAuthnSrv.AssertExpectations(t)
		})
	}
}

func TestTransformPATToPB(t *testing.T) {
	testTime := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	testCreatedAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	testLastUsed := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		pat        userpat.PersonalAccessToken
		tokenValue string
		want       *frontierv1beta1.PersonalAccessToken
	}{
		{
			name: "should transform minimal PAT",
			pat: userpat.PersonalAccessToken{
				ID:        "pat-1",
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				ExpiresAt: testTime,
				CreatedAt: testCreatedAt,
				UpdatedAt: testCreatedAt,
			},
			tokenValue: "",
			want: &frontierv1beta1.PersonalAccessToken{
				Id:        "pat-1",
				UserId:    "user-1",
				OrgId:     "org-1",
				Title:     "my-token",
				ExpiresAt: timestamppb.New(testTime),
				CreatedAt: timestamppb.New(testCreatedAt),
				UpdatedAt: timestamppb.New(testCreatedAt),
			},
		},
		{
			name: "should include token value when provided",
			pat: userpat.PersonalAccessToken{
				ID:        "pat-1",
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				ExpiresAt: testTime,
				CreatedAt: testCreatedAt,
				UpdatedAt: testCreatedAt,
			},
			tokenValue: "fpt_abc123",
			want: &frontierv1beta1.PersonalAccessToken{
				Id:        "pat-1",
				UserId:    "user-1",
				OrgId:     "org-1",
				Title:     "my-token",
				Token:     "fpt_abc123",
				ExpiresAt: timestamppb.New(testTime),
				CreatedAt: timestamppb.New(testCreatedAt),
				UpdatedAt: timestamppb.New(testCreatedAt),
			},
		},
		{
			name: "should include last_used_at when set",
			pat: userpat.PersonalAccessToken{
				ID:         "pat-1",
				UserID:     "user-1",
				OrgID:      "org-1",
				Title:      "my-token",
				ExpiresAt:  testTime,
				CreatedAt:  testCreatedAt,
				UpdatedAt:  testCreatedAt,
				LastUsedAt: &testLastUsed,
			},
			tokenValue: "",
			want: &frontierv1beta1.PersonalAccessToken{
				Id:         "pat-1",
				UserId:     "user-1",
				OrgId:      "org-1",
				Title:      "my-token",
				ExpiresAt:  timestamppb.New(testTime),
				CreatedAt:  timestamppb.New(testCreatedAt),
				UpdatedAt:  timestamppb.New(testCreatedAt),
				LastUsedAt: timestamppb.New(testLastUsed),
			},
		},
		{
			name: "should include metadata when set",
			pat: userpat.PersonalAccessToken{
				ID:        "pat-1",
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				ExpiresAt: testTime,
				CreatedAt: testCreatedAt,
				UpdatedAt: testCreatedAt,
				Metadata:  metadata.Metadata{"env": "prod"},
			},
			tokenValue: "fpt_xyz",
			want: &frontierv1beta1.PersonalAccessToken{
				Id:        "pat-1",
				UserId:    "user-1",
				OrgId:     "org-1",
				Title:     "my-token",
				Token:     "fpt_xyz",
				ExpiresAt: timestamppb.New(testTime),
				CreatedAt: timestamppb.New(testCreatedAt),
				UpdatedAt: timestamppb.New(testCreatedAt),
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"env": structpb.NewStringValue("prod"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformPATToPB(tt.pat, tt.tokenValue)
			assert.Equal(t, tt.want, got)
		})
	}
}
