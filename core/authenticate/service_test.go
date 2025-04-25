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

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/mocks"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/authenticate/token"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	mailerMock "github.com/raystack/frontier/pkg/mailer/mocks"
	pkgMetadata "github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/server/consts"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/log"
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
				return authenticate.NewService(nil, authenticate.Config{}, nil, nil, nil, nil, nil, nil, nil)
			},
		},
		{
			name: "fetch principal from valid user session",
			args: args{
				ctx:        context.Background(),
				assertions: []authenticate.ClientAssertion{authenticate.SessionClientAssertion},
			},
			want: authenticate.Principal{
				ID:   userID.String(),
				Type: schema.UserPrincipal,
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
					Metadata:        nil,
				}
				mockSessionService.EXPECT().ExtractFromContext(mock.Anything).Return(mockSess, nil)

				mockUserService.EXPECT().GetByID(mock.Anything, mockSess.UserID).Return(user.User{
					ID: mockSess.UserID,
				}, nil)

				return authenticate.NewService(nil, authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
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
					Metadata:        nil,
				}
				mockSessionService.EXPECT().ExtractFromContext(mock.Anything).Return(mockSess, nil)

				return authenticate.NewService(log.NewLogrus(), authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
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
				ID:   userID.String(),
				Type: schema.UserPrincipal,
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
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

				return authenticate.NewService(log.NewLogrus(), authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
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
				ID:   userID.String(),
				Type: schema.ServiceUserPrincipal,
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
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

				return authenticate.NewService(log.NewLogrus(), authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
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
				ID:   userID.String(),
				Type: schema.ServiceUserPrincipal,
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
			},
		},
		{
			name: "fetch principal from opaque token",
			args: args{
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					consts.UserSecretGatewayKey: {userToken},
				}),
				assertions: []authenticate.ClientAssertion{authenticate.OpaqueTokenClientAssertion},
			},
			want: authenticate.Principal{
				ID:   userID.String(),
				Type: schema.ServiceUserPrincipal,
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil)
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
					nil, nil, nil, nil, nil)
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
					nil, nil, nil)
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
					nil, nil, nil)
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
					nil, nil, nil)
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
