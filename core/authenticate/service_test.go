package authenticate_test

import (
	"context"
	"encoding/base64"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate/strategy"
	testusers "github.com/raystack/frontier/core/authenticate/test_users"
	"github.com/raystack/frontier/pkg/mailer"
	"github.com/stretchr/testify/assert"

	"io"
	"log/slog"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/mocks"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/authenticate/token"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	patModels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontiererrors "github.com/raystack/frontier/pkg/errors"
	mailerMock "github.com/raystack/frontier/pkg/mailer/mocks"
	pkgMetadata "github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/server/consts"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func createMocks(t *testing.T) (*mocks.FlowRepository, *mocks.UserService, *mocks.TokenService,
	*mocks.SessionService, *mocks.ServiceUserService) {
	t.Helper()

	return mocks.NewFlowRepository(t), mocks.NewUserService(t), mocks.NewTokenService(t),
		mocks.NewSessionService(t), mocks.NewServiceUserService(t)
}

func TestService_GetPrincipal(t *testing.T) {
	userID := uuid.New()
	testKey, err := utils.CreateJWKWithKID("test-id")
	require.NoError(t, err)
	tokenBytes, err := utils.BuildToken(testKey, "test", userID.String(), time.Hour, map[string]string{
		token.GeneratedClaimKey: token.GeneratedClaimValue,
	})
	require.NoError(t, err)
	userToken := base64.StdEncoding.EncodeToString([]byte("user:password"))

	type args struct {
		ctx        context.Context
		assertions []authenticate.ClientAssertion
	}
	tests := []struct {
		name    string
		args    args
		want    authenticate.Principal
		wantErr bool
		setup   func() *authenticate.Service
	}{
		{
			name: "fetch principal from context if available",
			args: args{
				ctx: authenticate.SetContextWithPrincipal(context.Background(), &authenticate.Principal{
					ID:   userID.String(),
					Type: schema.UserPrincipal,
				}),
				assertions: []authenticate.ClientAssertion{},
			},
			want: authenticate.Principal{
				ID:   userID.String(),
				Type: schema.UserPrincipal,
			},
			wantErr: false,
			setup: func() *authenticate.Service {
				return authenticate.NewService(nil, authenticate.Config{}, nil, nil, nil, nil, nil, nil, nil, nil)
			},
		},
		{
			name: "fetch principal from valid user session",
			args: args{
				ctx:        context.Background(),
				assertions: []authenticate.ClientAssertion{authenticate.SessionClientAssertion},
			},
			want: authenticate.Principal{
				ID:      userID.String(),
				Type:    schema.UserPrincipal,
				AuthVia: authenticate.SessionClientAssertion,
				User: &user.User{
					ID: userID.String(),
				},
			},
			wantErr: false,
			setup: func() *authenticate.Service {
				mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)

				mockSess := &frontiersession.Session{
					ID:              uuid.New(),
					UserID:          userID.String(),
					AuthenticatedAt: time.Now().Add(-time.Hour),
					ExpiresAt:       time.Now().Add(time.Hour),
					CreatedAt:       time.Time{},
					Metadata:        frontiersession.SessionMetadata{},
				}
				mockSessionService.EXPECT().ExtractFromContext(mock.Anything).Return(mockSess, nil)

				mockUserService.EXPECT().GetByID(mock.Anything, mockSess.UserID).Return(user.User{
					ID: mockSess.UserID,
				}, nil)

				return authenticate.NewService(nil, authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil)
			},
		},
		{
			name: "reject principal from expired user session",
			args: args{
				ctx:        context.Background(),
				assertions: []authenticate.ClientAssertion{authenticate.SessionClientAssertion},
			},
			wantErr: true,
			setup: func() *authenticate.Service {
				mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)

				mockSess := &frontiersession.Session{
					ID:              uuid.New(),
					UserID:          userID.String(),
					AuthenticatedAt: time.Now().Add(-time.Hour),
					ExpiresAt:       time.Now().Add(-time.Hour),
					CreatedAt:       time.Time{},
					Metadata:        frontiersession.SessionMetadata{},
				}
				mockSessionService.EXPECT().ExtractFromContext(mock.Anything).Return(mockSess, nil)

				return authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil)
			},
		},
		{
			name: "fetch principal from access token",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					consts.UserTokenGatewayKey: {string(tokenBytes)},
				}),
				assertions: []authenticate.ClientAssertion{authenticate.AccessTokenClientAssertion},
			},
			want: authenticate.Principal{
				ID:      userID.String(),
				Type:    schema.UserPrincipal,
				AuthVia: authenticate.AccessTokenClientAssertion,
				User: &user.User{
					ID: userID.String(),
				},
			},
			wantErr: false,
			setup: func() *authenticate.Service {
				mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)

				mockTokenService.EXPECT().Parse(mock.Anything, tokenBytes).Return(userID.String(), map[string]interface{}{}, nil)
				mockUserService.EXPECT().GetByID(mock.Anything, userID.String()).Return(user.User{
					ID: userID.String(),
				}, nil)

				return authenticate.NewService(nil, authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil)
			},
		},
		{
			name: "reject principal from invalid access token",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					consts.UserTokenGatewayKey: {string(tokenBytes)},
				}),
				assertions: []authenticate.ClientAssertion{authenticate.AccessTokenClientAssertion},
			},
			wantErr: true,
			setup: func() *authenticate.Service {
				mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)

				mockTokenService.EXPECT().Parse(mock.Anything, tokenBytes).Return("", map[string]interface{}{}, errors.New("invalid token"))

				return authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil)
			},
		},
		{
			name: "fetch principal from jwt grant",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					consts.UserTokenGatewayKey: {string(tokenBytes)},
				}),
				assertions: []authenticate.ClientAssertion{authenticate.JWTGrantClientAssertion},
			},
			want: authenticate.Principal{
				ID:      userID.String(),
				Type:    schema.ServiceUserPrincipal,
				AuthVia: authenticate.JWTGrantClientAssertion,
				ServiceUser: &serviceuser.ServiceUser{
					ID: userID.String(),
				},
			},
			wantErr: false,
			setup: func() *authenticate.Service {
				mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)

				mockServiceUserService.EXPECT().GetByJWT(mock.Anything, string(tokenBytes)).Return(serviceuser.ServiceUser{
					ID: userID.String(),
				}, nil)

				return authenticate.NewService(nil, authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil)
			},
		},
		{
			name: "failed to fetch principal from jwt grant",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					consts.UserTokenGatewayKey: {string(tokenBytes)},
				}),
				assertions: []authenticate.ClientAssertion{authenticate.JWTGrantClientAssertion},
			},
			wantErr: true,
			setup: func() *authenticate.Service {
				mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)

				mockServiceUserService.EXPECT().GetByJWT(mock.Anything, string(tokenBytes)).Return(serviceuser.ServiceUser{}, errors.New("invalid"))

				return authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil)
			},
		},
		{
			name: "fetch principal from client credential",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					consts.UserSecretGatewayKey: {userToken},
				}),
				assertions: []authenticate.ClientAssertion{authenticate.ClientCredentialsClientAssertion},
			},
			want: authenticate.Principal{
				ID:      userID.String(),
				Type:    schema.ServiceUserPrincipal,
				AuthVia: authenticate.ClientCredentialsClientAssertion,
				ServiceUser: &serviceuser.ServiceUser{
					ID: userID.String(),
				},
			},
			wantErr: false,
			setup: func() *authenticate.Service {
				mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)

				mockServiceUserService.EXPECT().GetBySecret(mock.Anything, "user", "password").Return(serviceuser.ServiceUser{
					ID: userID.String(),
				}, nil)

				return authenticate.NewService(nil, authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.GetPrincipal(tt.args.ctx, tt.args.assertions...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrincipal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetPrincipal() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_StartFlow(t *testing.T) {
	// Since, 'Flow' contains a call to UUID.New(), it will return a new UUID on each call.
	// We manipulate the seed so that fixed UUID is returned. This is done in setup.
	id := uuid.MustParse("52fdfc07-2182-454f-963f-5f0f9a621d72") // fixed UUID returned for first call of UUID.New()
	timeNow := time.Now()
	sampleErr := errors.New("sample error")

	flow := &authenticate.Flow{
		ID:        id,
		Method:    authenticate.MailOTPAuthMethod.String(),
		CreatedAt: timeNow,
		ExpiresAt: timeNow.Add(10 * time.Minute),
		Email:     "test@example.com",
		Nonce:     "111111", // fixed OTP
		Metadata: pkgMetadata.Metadata{
			"callback_url": "",
		},
	}

	type args struct {
		ctx     context.Context
		request authenticate.RegistrationStartRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *authenticate.RegistrationStartResponse
		wantErr error
		setup   func() *authenticate.Service
	}{
		{
			name: "return ErrUnsupportedMethod if request method is not supported",
			args: args{
				ctx:     context.Background(),
				request: authenticate.RegistrationStartRequest{},
			},
			want:    nil,
			wantErr: authenticate.ErrUnsupportedMethod,
			setup: func() *authenticate.Service {
				return authenticate.NewService(nil, authenticate.Config{}, nil, nil,
					nil, nil, nil, nil, nil, nil)
			},
		},
		{
			name: "simulate a successful StartFlow call",
			args: args{
				ctx: context.Background(),
				request: authenticate.RegistrationStartRequest{
					Method: authenticate.MailOTPAuthMethod.String(),
					Email:  "test@example.com",
				},
			},
			want: &authenticate.RegistrationStartResponse{
				Flow:  flow,
				State: flow.ID.String(),
			},
			wantErr: nil,
			setup: func() *authenticate.Service {
				uuid.SetRand(rand.New(rand.NewSource(1)))
				mockDialer := mailer.NewMockDialer()
				mockFlowRepo, _, _, _, _ := createMocks(t)
				ctx := context.Background()
				_ = strategy.NewMailOTP(mockDialer, "test-subject", "test-body")
				mockFlowRepo.EXPECT().Set(ctx, flow).Return(nil)
				srv := authenticate.NewService(
					nil,
					authenticate.Config{
						MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
						TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
					},
					mockFlowRepo, mockDialer, nil, nil,
					nil, nil, nil, nil)
				srv.Now = func() time.Time {
					return timeNow
				}
				return srv
			},
		},
		{
			name: "return sampleErr if flowRepo Set returns error",
			args: args{
				ctx: context.Background(),
				request: authenticate.RegistrationStartRequest{
					Method: authenticate.MailOTPAuthMethod.String(),
					Email:  "test@example.com",
				},
			},
			want:    nil,
			wantErr: sampleErr,
			setup: func() *authenticate.Service {
				uuid.SetRand(rand.New(rand.NewSource(1)))
				mockDialer := mailer.NewMockDialer()
				mockFlowRepo, _, _, _, _ := createMocks(t)
				ctx := context.Background()
				_ = strategy.NewMailOTP(mockDialer, "test-subject", "test-body")
				mockFlowRepo.EXPECT().Set(ctx, flow).Return(sampleErr)
				srv := authenticate.NewService(
					nil,
					authenticate.Config{
						MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
						TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
					},
					mockFlowRepo, mockDialer, nil, nil,
					nil, nil, nil, nil)
				srv.Now = func() time.Time {
					return timeNow
				}
				return srv
			},
		},
		{
			name: "return sampleErr if SendMail returns error",
			args: args{
				ctx: context.Background(),
				request: authenticate.RegistrationStartRequest{
					Method: authenticate.MailOTPAuthMethod.String(),
					Email:  "test@example.com",
				},
			},
			want:    nil,
			wantErr: sampleErr,
			setup: func() *authenticate.Service {
				mockDialer := &mailerMock.Dialer{}
				mockDialer.EXPECT().DialAndSend(mock.Anything).Return(sampleErr) // SendMail internally calls DialAndSend
				mockDialer.EXPECT().FromHeader().Return("")

				mockFlowRepo, _, _, _, _ := createMocks(t)
				_ = strategy.NewMailOTP(mockDialer, "test-subject", "test-body")
				srv := authenticate.NewService(
					nil,
					authenticate.Config{
						MailOTP: authenticate.MailOTPConfig{},
					},
					mockFlowRepo, mockDialer, nil, nil,
					nil, nil, nil, nil)
				srv.Now = func() time.Time {
					return timeNow
				}
				return srv
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.StartFlow(tt.args.ctx, tt.args.request)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetPrincipal_JWTGrantSkipsNonGrantToken(t *testing.T) {
	userID := uuid.New()
	patValue := "fpt_opaque-not-a-jwt"

	mockFlow, mockUserService, mockTokenService, mockSessionService, mockServiceUserService := createMocks(t)
	mockPATService := mocks.NewUserPATService(t)

	mockServiceUserService.EXPECT().GetByJWT(mock.Anything, patValue).
		Return(serviceuser.ServiceUser{}, serviceuser.ErrTokenNotJWT)
	pat := patModels.PAT{ID: "pat-1", UserID: userID.String(), ExpiresAt: time.Now().Add(time.Hour)}
	mockPATService.EXPECT().Validate(mock.Anything, patValue).Return(pat, nil)
	mockUserService.EXPECT().GetByID(mock.Anything, userID.String()).
		Return(user.User{ID: userID.String()}, nil)

	svc := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
		mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, mockPATService)

	ctx := metadata.NewIncomingContext(context.Background(), map[string][]string{
		consts.UserTokenGatewayKey: {patValue},
	})

	got, err := svc.GetPrincipal(ctx,
		authenticate.JWTGrantClientAssertion, authenticate.PATClientAssertion)
	require.NoError(t, err)
	assert.Equal(t, schema.PATPrincipal, got.Type)
	require.NotNil(t, got.PAT)
	assert.Equal(t, "pat-1", got.ID)
}

func TestService_GetPrincipal_RestrictsByAuthVia(t *testing.T) {
	// lists mirror what the handlers pass: session.go uses {Session}; AuthToken uses the token-exchange set.
	sessionOnly := []authenticate.ClientAssertion{authenticate.SessionClientAssertion}
	authTokenSet := []authenticate.ClientAssertion{
		authenticate.SessionClientAssertion,
		authenticate.ClientCredentialsClientAssertion,
		authenticate.JWTGrantClientAssertion,
		authenticate.PATClientAssertion,
	}

	tests := []struct {
		name    string
		authVia authenticate.ClientAssertion
		allowed []authenticate.ClientAssertion
		wantErr bool
	}{
		{"session endpoints accept a session", authenticate.SessionClientAssertion, sessionOnly, false},
		{"session endpoints reject a PAT", authenticate.PATClientAssertion, sessionOnly, true},
		{"session endpoints reject an access token", authenticate.AccessTokenClientAssertion, sessionOnly, true},
		{"session endpoints reject client credentials", authenticate.ClientCredentialsClientAssertion, sessionOnly, true},
		{"session endpoints reject a jwt grant", authenticate.JWTGrantClientAssertion, sessionOnly, true},

		{"authtoken accepts a session", authenticate.SessionClientAssertion, authTokenSet, false},
		{"authtoken accepts client credentials", authenticate.ClientCredentialsClientAssertion, authTokenSet, false},
		{"authtoken accepts a jwt grant", authenticate.JWTGrantClientAssertion, authTokenSet, false},
		{"authtoken accepts a PAT", authenticate.PATClientAssertion, authTokenSet, false},
		{"authtoken rejects an access token", authenticate.AccessTokenClientAssertion, authTokenSet, true},
		{"authtoken rejects passthrough", authenticate.PassthroughHeaderClientAssertion, authTokenSet, true},
	}

	svc := authenticate.NewService(nil, authenticate.Config{}, nil, nil, nil, nil, nil, nil, nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := authenticate.SetContextWithPrincipal(context.Background(), &authenticate.Principal{
				ID:      "principal-1",
				Type:    schema.UserPrincipal,
				AuthVia: tt.authVia,
			})
			if _, err := svc.GetPrincipal(ctx, tt.allowed...); (err != nil) != tt.wantErr {
				t.Errorf("GetPrincipal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetPrincipal_OrgStateGate(t *testing.T) {
	orgID := uuid.New().String()
	userID := uuid.New().String()
	patID := uuid.New().String()
	secret := base64.StdEncoding.EncodeToString([]byte("user:password"))

	tests := []struct {
		name       string
		ctx        context.Context
		assertions []authenticate.ClientAssertion
		want       authenticate.Principal
		wantErr    error
		setup      func(t *testing.T) *authenticate.Service
	}{
		{
			name: "reject PAT whose org is disabled or gone",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				consts.UserTokenGatewayKey: {"pat-token"},
			}),
			assertions: []authenticate.ClientAssertion{authenticate.PATClientAssertion},
			wantErr:    frontiererrors.ErrForbidden,
			setup: func(t *testing.T) *authenticate.Service {
				pat := mocks.NewUserPATService(t)
				pat.EXPECT().Validate(mock.Anything, "pat-token").Return(patModels.PAT{ID: patID, UserID: userID, OrgID: orgID}, nil)
				usr := mocks.NewUserService(t)
				usr.EXPECT().GetByID(mock.Anything, userID).Return(user.User{ID: userID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(false, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, usr, nil, nil, pat)
				s.SetOrgService(org)
				return s
			},
		},
		{
			name: "reject PAT with empty org id",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				consts.UserTokenGatewayKey: {"pat-token"},
			}),
			assertions: []authenticate.ClientAssertion{authenticate.PATClientAssertion},
			wantErr:    frontiererrors.ErrForbidden,
			setup: func(t *testing.T) *authenticate.Service {
				pat := mocks.NewUserPATService(t)
				pat.EXPECT().Validate(mock.Anything, "pat-token").Return(patModels.PAT{ID: patID, UserID: userID, OrgID: ""}, nil)
				usr := mocks.NewUserService(t)
				usr.EXPECT().GetByID(mock.Anything, userID).Return(user.User{ID: userID}, nil)
				org := mocks.NewOrgService(t)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, usr, nil, nil, pat)
				s.SetOrgService(org)
				return s
			},
		},
		{
			name: "allow PAT whose org is enabled",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				consts.UserTokenGatewayKey: {"pat-token"},
			}),
			assertions: []authenticate.ClientAssertion{authenticate.PATClientAssertion},
			want: authenticate.Principal{
				ID:      patID,
				Type:    schema.PATPrincipal,
				AuthVia: authenticate.PATClientAssertion,
				PAT:     &patModels.PAT{ID: patID, UserID: userID, OrgID: orgID},
				User:    &user.User{ID: userID},
			},
			setup: func(t *testing.T) *authenticate.Service {
				pat := mocks.NewUserPATService(t)
				pat.EXPECT().Validate(mock.Anything, "pat-token").Return(patModels.PAT{ID: patID, UserID: userID, OrgID: orgID}, nil)
				usr := mocks.NewUserService(t)
				usr.EXPECT().GetByID(mock.Anything, userID).Return(user.User{ID: userID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(true, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, usr, nil, nil, pat)
				s.SetOrgService(org)
				return s
			},
		},
		{
			name: "reject service user (client credentials) whose org is disabled or gone",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				consts.UserSecretGatewayKey: {secret},
			}),
			assertions: []authenticate.ClientAssertion{authenticate.ClientCredentialsClientAssertion},
			wantErr:    frontiererrors.ErrForbidden,
			setup: func(t *testing.T) *authenticate.Service {
				su := mocks.NewServiceUserService(t)
				su.EXPECT().GetBySecret(mock.Anything, "user", "password").Return(serviceuser.ServiceUser{ID: userID, OrgID: orgID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(false, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, nil, su, nil, nil)
				s.SetOrgService(org)
				return s
			},
		},
		{
			name: "reject service user (jwt grant) whose org is disabled or gone",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				consts.UserTokenGatewayKey: {"grant-token"},
			}),
			assertions: []authenticate.ClientAssertion{authenticate.JWTGrantClientAssertion},
			wantErr:    frontiererrors.ErrForbidden,
			setup: func(t *testing.T) *authenticate.Service {
				su := mocks.NewServiceUserService(t)
				su.EXPECT().GetByJWT(mock.Anything, "grant-token").Return(serviceuser.ServiceUser{ID: userID, OrgID: orgID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(false, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, nil, su, nil, nil)
				s.SetOrgService(org)
				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup(t)
			got, err := s.GetPrincipal(tt.ctx, tt.assertions...)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetPrincipal() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
