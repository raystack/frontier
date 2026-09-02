package authenticate_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/raystack/frontier/core/consent"

	"github.com/go-webauthn/webauthn/webauthn"
	"golang.org/x/crypto/bcrypt"

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
				return authenticate.NewService(nil, authenticate.Config{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil, nil, nil)
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil, nil, nil)
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

				mockTokenService.EXPECT().Parse(mock.Anything, tokenBytes).Return(userID.String(), map[string]any{}, nil)
				mockUserService.EXPECT().GetByID(mock.Anything, userID.String()).Return(user.User{
					ID: userID.String(),
				}, nil)

				return authenticate.NewService(nil, authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil, nil, nil)
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

				mockTokenService.EXPECT().Parse(mock.Anything, tokenBytes).Return("", map[string]any{}, errors.New("invalid token"))

				return authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil, nil, nil)
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil, nil, nil)
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil, nil, nil)
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
					mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, nil, nil, nil)
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
	defaultHashCost := authenticate.OTPHashCost
	authenticate.OTPHashCost = bcrypt.MinCost
	t.Cleanup(func() { authenticate.OTPHashCost = defaultHashCost })

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

	// Set receives the flow with a bcrypt hash in Nonce; verify the hash
	// against the fixed OTP and compare the remaining fields to the fixture
	flowWithHashedNonce := mock.MatchedBy(func(got *authenticate.Flow) bool {
		if bcrypt.CompareHashAndPassword([]byte(got.Nonce), []byte(flow.Nonce)) != nil {
			return false
		}
		cp := *got
		cp.Nonce = flow.Nonce
		return reflect.DeepEqual(&cp, flow)
	})

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
					nil, nil, nil, nil, nil, nil, nil, nil)
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
				mockFlowRepo.EXPECT().Set(ctx, flowWithHashedNonce).Return(nil)
				srv := authenticate.NewService(
					nil,
					authenticate.Config{
						MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
						TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
					},
					mockFlowRepo, mockDialer, nil, nil,
					nil, nil, nil, nil, nil, nil)
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
				mockFlowRepo.EXPECT().Set(ctx, flowWithHashedNonce).Return(sampleErr)
				srv := authenticate.NewService(
					nil,
					authenticate.Config{
						MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
						TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
					},
					mockFlowRepo, mockDialer, nil, nil,
					nil, nil, nil, nil, nil, nil)
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
					nil, nil, nil, nil, nil, nil)
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
			if tt.want != nil && got != nil {
				// nonce is a bcrypt hash; the Set matcher already verified it
				got.Flow.Nonce = tt.want.Flow.Nonce
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func mailOTPFlow(id uuid.UUID, now time.Time, nonce string, md pkgMetadata.Metadata) *authenticate.Flow {
	return &authenticate.Flow{
		ID:        id,
		Method:    authenticate.MailOTPAuthMethod.String(),
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		Email:     "test@example.com",
		Nonce:     nonce,
		Metadata:  md,
	}
}

func TestService_FinishFlow(t *testing.T) {
	flowID := uuid.New()
	timeNow := time.Now()
	otpHash, err := bcrypt.GenerateFromPassword([]byte("111111"), bcrypt.MinCost)
	require.NoError(t, err)
	sampleUser := user.User{ID: "user-id", Email: "test@example.com"}

	type args struct {
		ctx     context.Context
		request authenticate.RegistrationFinishRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *authenticate.RegistrationFinishResponse
		wantErr error
		setup   func() *authenticate.Service
	}{
		{
			name: "return the user and consume the flow if the code is valid",
			args: args{
				ctx: context.Background(),
				request: authenticate.RegistrationFinishRequest{
					Method: authenticate.MailOTPAuthMethod.String(),
					State:  flowID.String(),
					Code:   "111111",
				},
			},
			want: &authenticate.RegistrationFinishResponse{
				User: sampleUser,
				Flow: mailOTPFlow(flowID, timeNow, string(otpHash), pkgMetadata.Metadata{"callback_url": ""}),
			},
			wantErr: nil,
			setup: func() *authenticate.Service {
				mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
				ctx := context.Background()
				mockFlowRepo.EXPECT().Get(ctx, flowID).
					Return(mailOTPFlow(flowID, timeNow, string(otpHash), pkgMetadata.Metadata{"callback_url": ""}), nil)
				mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
				mockUserService.EXPECT().GetByID(ctx, "test@example.com").Return(sampleUser, nil)
				srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
					nil, nil, mockUserService, nil, nil, nil, nil, nil)
				srv.Now = func() time.Time {
					return timeNow
				}
				return srv
			},
		},
		{
			name: "return ErrInvalidMailOTP and record the attempt if the code is wrong",
			args: args{
				ctx: context.Background(),
				request: authenticate.RegistrationFinishRequest{
					Method: authenticate.MailOTPAuthMethod.String(),
					State:  flowID.String(),
					Code:   "222222",
				},
			},
			want:    nil,
			wantErr: authenticate.ErrInvalidMailOTP,
			setup: func() *authenticate.Service {
				mockFlowRepo, _, _, _, _ := createMocks(t)
				ctx := context.Background()
				mockFlowRepo.EXPECT().Get(ctx, flowID).
					Return(mailOTPFlow(flowID, timeNow, string(otpHash), pkgMetadata.Metadata{"callback_url": ""}), nil)
				mockFlowRepo.EXPECT().Set(ctx, mock.MatchedBy(func(f *authenticate.Flow) bool {
					return f.Metadata["attempt"] == 1 && f.Nonce == string(otpHash)
				})).Return(nil)
				srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
					nil, nil, nil, nil, nil, nil, nil, nil)
				srv.Now = func() time.Time {
					return timeNow
				}
				return srv
			},
		},
		{
			name: "reject the correct code if the stored nonce is not hashed",
			args: args{
				ctx: context.Background(),
				request: authenticate.RegistrationFinishRequest{
					Method: authenticate.MailOTPAuthMethod.String(),
					State:  flowID.String(),
					Code:   "111111",
				},
			},
			want:    nil,
			wantErr: authenticate.ErrInvalidMailOTP,
			setup: func() *authenticate.Service {
				mockFlowRepo, _, _, _, _ := createMocks(t)
				ctx := context.Background()
				mockFlowRepo.EXPECT().Get(ctx, flowID).
					Return(mailOTPFlow(flowID, timeNow, "111111", pkgMetadata.Metadata{"callback_url": ""}), nil)
				mockFlowRepo.EXPECT().Set(ctx, mock.MatchedBy(func(f *authenticate.Flow) bool {
					return f.Metadata["attempt"] == 1 && f.Nonce == "111111"
				})).Return(nil)
				srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
					nil, nil, nil, nil, nil, nil, nil, nil)
				srv.Now = func() time.Time {
					return timeNow
				}
				return srv
			},
		},
		{
			name: "destroy the flow if the wrong code reaches the attempt cap",
			args: args{
				ctx: context.Background(),
				request: authenticate.RegistrationFinishRequest{
					Method: authenticate.MailOTPAuthMethod.String(),
					State:  flowID.String(),
					Code:   "222222",
				},
			},
			want:    nil,
			wantErr: authenticate.ErrInvalidMailOTP,
			setup: func() *authenticate.Service {
				mockFlowRepo, _, _, _, _ := createMocks(t)
				ctx := context.Background()
				mockFlowRepo.EXPECT().Get(ctx, flowID).
					Return(mailOTPFlow(flowID, timeNow, string(otpHash), pkgMetadata.Metadata{"callback_url": "", "attempt": 2}), nil)
				mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
				srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
					nil, nil, nil, nil, nil, nil, nil, nil)
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
			got, err := s.FinishFlow(tt.args.ctx, tt.args.request)
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

func TestService_FinishFlow_WrongThenRightOTP(t *testing.T) {
	flowID := uuid.New()
	timeNow := time.Now()
	otpHash, err := bcrypt.GenerateFromPassword([]byte("111111"), bcrypt.MinCost)
	require.NoError(t, err)
	sampleUser := user.User{ID: "user-id", Email: "test@example.com"}

	ctx := context.Background()
	mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
	flow := mailOTPFlow(flowID, timeNow, string(otpHash), pkgMetadata.Metadata{"callback_url": ""})

	mockFlowRepo.EXPECT().Get(ctx, flowID).Return(flow, nil).Twice()
	mockFlowRepo.EXPECT().Set(ctx, flow).Return(nil).Once()
	mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil).Once()
	mockUserService.EXPECT().GetByID(ctx, "test@example.com").Return(sampleUser, nil).Once()

	srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
		nil, nil, mockUserService, nil, nil, nil, nil, nil)
	srv.Now = func() time.Time {
		return timeNow
	}

	request := func(code string) authenticate.RegistrationFinishRequest {
		return authenticate.RegistrationFinishRequest{
			Method: authenticate.MailOTPAuthMethod.String(),
			State:  flowID.String(),
			Code:   code,
		}
	}

	got, err := srv.FinishFlow(ctx, request("222222"))
	assert.ErrorIs(t, err, authenticate.ErrInvalidMailOTP)
	assert.Nil(t, got)
	assert.Equal(t, 1, flow.Metadata["attempt"])

	got, err = srv.FinishFlow(ctx, request("111111"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, sampleUser, got.User)
}

// jsonFlowRepository keeps flow metadata as JSON bytes, the same way the
// Postgres repository stores it. Numbers in metadata come back from Get as
// float64, matching a real database round trip.
type jsonFlowRepository struct {
	flows map[uuid.UUID]jsonStoredFlow
}

type jsonStoredFlow struct {
	flow        authenticate.Flow
	rawMetadata []byte
}

func (r *jsonFlowRepository) Set(_ context.Context, flow *authenticate.Flow) error {
	raw, err := json.Marshal(flow.Metadata)
	if err != nil {
		return err
	}
	r.flows[flow.ID] = jsonStoredFlow{flow: *flow, rawMetadata: raw}
	return nil
}

func (r *jsonFlowRepository) Get(_ context.Context, id uuid.UUID) (*authenticate.Flow, error) {
	stored, ok := r.flows[id]
	if !ok {
		return nil, authenticate.ErrFlowInvalid
	}
	flow := stored.flow
	var md pkgMetadata.Metadata
	if err := json.Unmarshal(stored.rawMetadata, &md); err != nil {
		return nil, err
	}
	flow.Metadata = md
	return &flow, nil
}

func (r *jsonFlowRepository) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.flows, id)
	return nil
}

func (r *jsonFlowRepository) DeleteExpiredFlows(_ context.Context) error {
	return nil
}

func TestService_FinishFlow_OTPAttemptCapAfterJSONRoundTrip(t *testing.T) {
	flowID := uuid.New()
	timeNow := time.Now()
	otpHash, err := bcrypt.GenerateFromPassword([]byte("111111"), bcrypt.MinCost)
	require.NoError(t, err)

	ctx := context.Background()
	flowRepo := &jsonFlowRepository{flows: map[uuid.UUID]jsonStoredFlow{}}
	require.NoError(t, flowRepo.Set(ctx, mailOTPFlow(flowID, timeNow, string(otpHash), pkgMetadata.Metadata{"callback_url": ""})))

	srv := authenticate.NewService(nil, authenticate.Config{}, flowRepo, nil,
		nil, nil, nil, nil, nil, nil, nil, nil)
	srv.Now = func() time.Time {
		return timeNow
	}

	wrongCode := authenticate.RegistrationFinishRequest{
		Method: authenticate.MailOTPAuthMethod.String(),
		State:  flowID.String(),
		Code:   "222222",
	}

	for attempt := 1; attempt <= 2; attempt++ {
		got, err := srv.FinishFlow(ctx, wrongCode)
		assert.ErrorIs(t, err, authenticate.ErrInvalidMailOTP)
		assert.Nil(t, got)

		flow, err := flowRepo.Get(ctx, flowID)
		require.NoError(t, err)
		assert.Equal(t, float64(attempt), flow.Metadata["attempt"])
	}

	// the third wrong code reaches the cap and destroys the flow
	got, err := srv.FinishFlow(ctx, wrongCode)
	assert.ErrorIs(t, err, authenticate.ErrInvalidMailOTP)
	assert.Nil(t, got)

	_, err = flowRepo.Get(ctx, flowID)
	assert.ErrorIs(t, err, authenticate.ErrFlowInvalid)
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
		mockFlow, nil, mockTokenService, mockSessionService, mockUserService, mockServiceUserService, nil, mockPATService, nil, nil)

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

	svc := authenticate.NewService(nil, authenticate.Config{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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
		setup      func() *authenticate.Service
	}{
		{
			name: "reject PAT whose org is disabled or gone",
			ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
				consts.UserTokenGatewayKey: {"pat-token"},
			}),
			assertions: []authenticate.ClientAssertion{authenticate.PATClientAssertion},
			wantErr:    frontiererrors.ErrForbidden,
			setup: func() *authenticate.Service {
				pat := mocks.NewUserPATService(t)
				pat.EXPECT().Validate(mock.Anything, "pat-token").Return(patModels.PAT{ID: patID, UserID: userID, OrgID: orgID}, nil)
				usr := mocks.NewUserService(t)
				usr.EXPECT().GetByID(mock.Anything, userID).Return(user.User{ID: userID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(false, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, usr, nil, nil, pat, nil, nil)
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
			setup: func() *authenticate.Service {
				pat := mocks.NewUserPATService(t)
				pat.EXPECT().Validate(mock.Anything, "pat-token").Return(patModels.PAT{ID: patID, UserID: userID, OrgID: ""}, nil)
				usr := mocks.NewUserService(t)
				usr.EXPECT().GetByID(mock.Anything, userID).Return(user.User{ID: userID}, nil)
				org := mocks.NewOrgService(t)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, usr, nil, nil, pat, nil, nil)
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
			setup: func() *authenticate.Service {
				pat := mocks.NewUserPATService(t)
				pat.EXPECT().Validate(mock.Anything, "pat-token").Return(patModels.PAT{ID: patID, UserID: userID, OrgID: orgID}, nil)
				usr := mocks.NewUserService(t)
				usr.EXPECT().GetByID(mock.Anything, userID).Return(user.User{ID: userID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(true, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, usr, nil, nil, pat, nil, nil)
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
			setup: func() *authenticate.Service {
				su := mocks.NewServiceUserService(t)
				su.EXPECT().GetBySecret(mock.Anything, "user", "password").Return(serviceuser.ServiceUser{ID: userID, OrgID: orgID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(false, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, nil, su, nil, nil, nil, nil)
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
			setup: func() *authenticate.Service {
				su := mocks.NewServiceUserService(t)
				su.EXPECT().GetByJWT(mock.Anything, "grant-token").Return(serviceuser.ServiceUser{ID: userID, OrgID: orgID}, nil)
				org := mocks.NewOrgService(t)
				org.EXPECT().IsEnabled(mock.Anything, orgID).Return(false, nil)
				s := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
					nil, nil, nil, nil, nil, su, nil, nil, nil, nil)
				s.SetOrgService(org)
				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
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

func TestService_BuildToken(t *testing.T) {
	userID := uuid.NewString()

	tests := []struct {
		name       string
		principal  authenticate.Principal
		config     authenticate.Config
		wantClaims map[string]string
	}{
		{
			name: "user via session gets sub_type, auth_via and email claims",
			principal: authenticate.Principal{
				ID:      userID,
				Type:    schema.UserPrincipal,
				AuthVia: authenticate.SessionClientAssertion,
				User:    &user.User{ID: userID, Email: "jane@acme.org"},
			},
			config: authenticate.Config{Token: authenticate.TokenConfig{
				Claims: authenticate.TokenClaimConfig{AddUserEmailClaim: true},
			}},
			wantClaims: map[string]string{
				token.SubTypeClaimsKey:  schema.UserPrincipal,
				token.AuthViaClaimKey:   authenticate.SessionClientAssertion.String(),
				token.SubEmailClaimsKey: "jane@acme.org",
			},
		},
		{
			name: "service user via client credentials gets auth_via claim",
			principal: authenticate.Principal{
				ID:      userID,
				Type:    schema.ServiceUserPrincipal,
				AuthVia: authenticate.ClientCredentialsClientAssertion,
			},
			wantClaims: map[string]string{
				token.SubTypeClaimsKey: schema.ServiceUserPrincipal,
				token.AuthViaClaimKey:  authenticate.ClientCredentialsClientAssertion.String(),
			},
		},
		{
			name: "pat principal gets auth_via and user_id claims",
			principal: authenticate.Principal{
				ID:      "pat-token-id",
				Type:    schema.PATPrincipal,
				AuthVia: authenticate.PATClientAssertion,
				User:    &user.User{ID: userID},
			},
			wantClaims: map[string]string{
				token.SubTypeClaimsKey: schema.PATPrincipal,
				token.AuthViaClaimKey:  authenticate.PATClientAssertion.String(),
				token.UserIDClaimKey:   userID,
			},
		},
		{
			name: "principal without auth via gets empty auth_via claim",
			principal: authenticate.Principal{
				ID:   userID,
				Type: schema.UserPrincipal,
			},
			wantClaims: map[string]string{
				token.SubTypeClaimsKey: schema.UserPrincipal,
				token.AuthViaClaimKey:  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockToken := mocks.NewTokenService(t)
			mockToken.EXPECT().Build(tt.principal.ID, tt.wantClaims).Return([]byte("signed-token"), nil)
			s := authenticate.NewService(nil, tt.config, nil, nil, mockToken, nil, nil, nil, nil, nil, nil, nil)

			got, err := s.BuildToken(context.Background(), tt.principal, map[string]string{})
			assert.NoError(t, err)
			assert.Equal(t, []byte("signed-token"), got)
		})
	}
}

// passkeyUserMetadata returns metadata holding one stored passkey credential,
// base64 encoded the way startPassKeyLoginMethod expects it.
func passkeyUserMetadata(t *testing.T) pkgMetadata.Metadata {
	t.Helper()

	credBytes, err := json.Marshal([]webauthn.Credential{{
		ID:        []byte("credential-id"),
		PublicKey: []byte("public-key"),
	}})
	require.NoError(t, err)

	return pkgMetadata.Metadata{
		"passkey_credentials": base64.StdEncoding.EncodeToString(credBytes),
	}
}

func testWebAuthn(t *testing.T) *webauthn.WebAuthn {
	t.Helper()

	webAuth, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "frontier test",
		RPID:          "example.com",
		RPOrigins:     []string{"https://example.com"},
	})
	require.NoError(t, err)
	return webAuth
}

// TestService_StartFlow_Intent walks the intent-by-strategy table. For passkey
// the intent also picks the ceremony, which used to be guessed.
func TestService_StartFlow_Intent(t *testing.T) {
	defaultHashCost := authenticate.OTPHashCost
	authenticate.OTPHashCost = bcrypt.MinCost
	t.Cleanup(func() { authenticate.OTPHashCost = defaultHashCost })

	const email = "test@example.com"

	tests := []struct {
		name string

		method         string
		intent         authenticate.FlowIntent
		userExists     bool
		userHasPasskey bool

		wantErr             error
		wantPasskeyCeremony string
	}{
		{
			name:       "mail otp login starts the flow when the address has an account",
			method:     authenticate.MailOTPAuthMethod.String(),
			intent:     authenticate.FlowIntentLogin,
			userExists: true,
		},
		{
			name:    "mail otp login is rejected before the code is sent when it does not",
			method:  authenticate.MailOTPAuthMethod.String(),
			intent:  authenticate.FlowIntentLogin,
			wantErr: authenticate.ErrLoginUserNotFound,
		},
		{
			name:   "mail otp signup starts the flow when the address has no account",
			method: authenticate.MailOTPAuthMethod.String(),
			intent: authenticate.FlowIntentSignup,
		},
		{
			name:       "mail otp signup is rejected when the address already has one",
			method:     authenticate.MailOTPAuthMethod.String(),
			intent:     authenticate.FlowIntentSignup,
			userExists: true,
			wantErr:    authenticate.ErrSignupUserExists,
		},
		{
			name:       "mail otp without an intent starts the flow for a known address",
			method:     authenticate.MailOTPAuthMethod.String(),
			intent:     authenticate.FlowIntentUnspecified,
			userExists: true,
		},
		{
			name:   "mail otp without an intent starts the flow for an unknown address",
			method: authenticate.MailOTPAuthMethod.String(),
			intent: authenticate.FlowIntentUnspecified,
		},
		// mail link knows the address as early as mail otp and shares its finish
		// path, so the gate has to behave the same for both
		{
			name:       "mail link login starts the flow when the address has an account",
			method:     authenticate.MailLinkAuthMethod.String(),
			intent:     authenticate.FlowIntentLogin,
			userExists: true,
		},
		{
			name:    "mail link login is rejected before the link is sent when it does not",
			method:  authenticate.MailLinkAuthMethod.String(),
			intent:  authenticate.FlowIntentLogin,
			wantErr: authenticate.ErrLoginUserNotFound,
		},
		{
			name:   "mail link signup starts the flow when the address has no account",
			method: authenticate.MailLinkAuthMethod.String(),
			intent: authenticate.FlowIntentSignup,
		},
		{
			name:       "mail link signup is rejected when the address already has one",
			method:     authenticate.MailLinkAuthMethod.String(),
			intent:     authenticate.FlowIntentSignup,
			userExists: true,
			wantErr:    authenticate.ErrSignupUserExists,
		},
		{
			name:       "mail link without an intent starts the flow for a known address",
			method:     authenticate.MailLinkAuthMethod.String(),
			intent:     authenticate.FlowIntentUnspecified,
			userExists: true,
		},
		{
			name:   "mail link without an intent starts the flow for an unknown address",
			method: authenticate.MailLinkAuthMethod.String(),
			intent: authenticate.FlowIntentUnspecified,
		},
		{
			name:                "passkey login runs the login ceremony when the address has an account",
			method:              authenticate.PassKeyAuthMethod.String(),
			intent:              authenticate.FlowIntentLogin,
			userExists:          true,
			userHasPasskey:      true,
			wantPasskeyCeremony: strategy.PasskeyLoginType,
		},
		{
			name:    "passkey login is rejected when the address has no account",
			method:  authenticate.PassKeyAuthMethod.String(),
			intent:  authenticate.FlowIntentLogin,
			wantErr: authenticate.ErrLoginUserNotFound,
		},
		{
			name:       "passkey login fails when the account has no registered passkey",
			method:     authenticate.PassKeyAuthMethod.String(),
			intent:     authenticate.FlowIntentLogin,
			userExists: true,
			wantErr:    authenticate.ErrInvalidMethod,
		},
		{
			name:                "passkey signup runs the register ceremony when the address has no account",
			method:              authenticate.PassKeyAuthMethod.String(),
			intent:              authenticate.FlowIntentSignup,
			wantPasskeyCeremony: strategy.PasskeyRegisterType,
		},
		{
			name:           "passkey signup is rejected when the address already has an account",
			method:         authenticate.PassKeyAuthMethod.String(),
			intent:         authenticate.FlowIntentSignup,
			userExists:     true,
			userHasPasskey: true,
			wantErr:        authenticate.ErrSignupUserExists,
		},
		{
			name:                "passkey without an intent still guesses register for an unknown address",
			method:              authenticate.PassKeyAuthMethod.String(),
			intent:              authenticate.FlowIntentUnspecified,
			wantPasskeyCeremony: strategy.PasskeyRegisterType,
		},
		{
			name:                "passkey without an intent still guesses login for a registered passkey",
			method:              authenticate.PassKeyAuthMethod.String(),
			intent:              authenticate.FlowIntentUnspecified,
			userExists:          true,
			userHasPasskey:      true,
			wantPasskeyCeremony: strategy.PasskeyLoginType,
		},
		{
			name:                "passkey without an intent still guesses register when the account has no passkey",
			method:              authenticate.PassKeyAuthMethod.String(),
			intent:              authenticate.FlowIntentUnspecified,
			userExists:          true,
			wantPasskeyCeremony: strategy.PasskeyRegisterType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			isPasskey := tt.method == authenticate.PassKeyAuthMethod.String()

			mockFlowRepo, mockUserService, _, _, _ := createMocks(t)

			// passkey always looks the address up; the mail strategies only do
			// when there is an intent to check
			if isPasskey || tt.intent != authenticate.FlowIntentUnspecified {
				if tt.userExists {
					existing := user.User{ID: uuid.New().String(), Email: email}
					if tt.userHasPasskey {
						existing.Metadata = passkeyUserMetadata(t)
					}
					mockUserService.EXPECT().GetByID(ctx, email).Return(existing, nil)
				} else {
					mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))
				}
			}

			var storedFlow *authenticate.Flow
			if tt.wantErr == nil {
				mockFlowRepo.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, flow *authenticate.Flow) {
					storedFlow = flow
				}).Return(nil)
			}

			var webAuth *webauthn.WebAuthn
			var mockDialer mailer.Dialer
			if isPasskey {
				webAuth = testWebAuthn(t)
			} else {
				mockDialer = mailer.NewMockDialer()
			}

			srv := authenticate.NewService(nil, authenticate.Config{
				MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
				MailLink:  authenticate.MailLinkConfig{Validity: 10 * time.Minute},
				TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
			}, mockFlowRepo, mockDialer, nil, nil, mockUserService, nil, webAuth, nil, nil, nil)

			got, err := srv.StartFlow(ctx, authenticate.RegistrationStartRequest{
				Method: tt.method,
				Email:  email,
				Intent: tt.intent,
				// mail link embeds the callback host in the link it sends
				CallbackUrl: "http://localhost:7400/v1beta1/auth/callback",
			})

			switch {
			case tt.wantErr != nil:
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			default:
				require.NoError(t, err)
				require.NotNil(t, got)
				require.NotNil(t, storedFlow)
				if tt.wantPasskeyCeremony != "" {
					assert.Equal(t, tt.wantPasskeyCeremony, storedFlow.Metadata["passkey_type"])
				}
			}
		})
	}
}

// TestService_StartFlow_WritesIntentAndConsent covers what StartFlow puts on the
// flow, which is where both fields have to live to survive an OIDC redirect.
func TestService_StartFlow_WritesIntentAndConsent(t *testing.T) {
	defaultHashCost := authenticate.OTPHashCost
	authenticate.OTPHashCost = bcrypt.MinCost
	t.Cleanup(func() { authenticate.OTPHashCost = defaultHashCost })

	const email = "test@example.com"
	timeNow := time.Now().UTC()

	startFlow := func(t *testing.T, request authenticate.RegistrationStartRequest) *authenticate.Flow {
		t.Helper()

		ctx := context.Background()
		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		if request.Intent != authenticate.FlowIntentUnspecified {
			mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))
		}

		var storedFlow *authenticate.Flow
		mockFlowRepo.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, flow *authenticate.Flow) {
			storedFlow = flow
		}).Return(nil)

		srv := authenticate.NewService(nil, authenticate.Config{
			MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
			TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
		}, mockFlowRepo, mailer.NewMockDialer(), nil, nil, mockUserService, nil, nil, nil, nil, nil)
		srv.Now = func() time.Time { return timeNow }

		_, err := srv.StartFlow(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, storedFlow)
		return storedFlow
	}

	t.Run("writes the intent and the consent when the caller sends them", func(t *testing.T) {
		flow := startFlow(t, authenticate.RegistrationStartRequest{
			Method:              authenticate.MailOTPAuthMethod.String(),
			Email:               email,
			Intent:              authenticate.FlowIntentSignup,
			AcceptedDocumentIDs: []string{"terms_of_service", "privacy_policy"},
			IPAddress:           "10.0.0.1",
		})

		assert.Equal(t, authenticate.FlowIntentSignup, flow.Intent())

		consent, ok := flow.Consent()
		require.True(t, ok)
		assert.Equal(t, []string{"terms_of_service", "privacy_policy"}, consent.AcceptedDocumentIDs)
		assert.Equal(t, "10.0.0.1", consent.IPAddress)
		// the timestamp is when the user accepted, not when the flow finishes
		assert.Equal(t, timeNow, consent.At)
	})

	t.Run("leaves the flow untouched when the caller sends neither", func(t *testing.T) {
		flow := startFlow(t, authenticate.RegistrationStartRequest{
			Method: authenticate.MailOTPAuthMethod.String(),
			Email:  email,
		})

		// a client that sends no intent writes the same flow row it does today
		assert.Equal(t, pkgMetadata.Metadata{"callback_url": ""}, flow.Metadata)
		assert.Equal(t, authenticate.FlowIntentUnspecified, flow.Intent())
		_, ok := flow.Consent()
		assert.False(t, ok)
	})
}

// TestService_StartFlow_Consent covers the first of the two consent gates. It
// is the one that exists for the error: it runs before an OTP is sent and
// before the browser leaves for an identity provider, so a rejection costs the
// user a retry and nothing else.
//
// The intent decides which rule applies, because without one a signup and a
// login look identical.
func TestService_StartFlow_Consent(t *testing.T) {
	defaultHashCost := authenticate.OTPHashCost
	authenticate.OTPHashCost = bcrypt.MinCost
	t.Cleanup(func() { authenticate.OTPHashCost = defaultHashCost })

	const email = "test@example.com"
	documents := []consent.Document{
		{ID: "privacy_policy", Title: "Privacy Policy", Version: "2026-04-01", URL: "https://example.org/p"},
		{ID: "terms_of_service", Title: "Terms & Conditions", Version: "2026-04-01", URL: "https://example.org/t"},
	}
	acceptedIDs := []string{"privacy_policy", "terms_of_service"}

	// startFlow runs a mail otp flow start against whatever consent service it
	// is given. Mail otp is the strategy that shows the point of this gate,
	// because a rejection here is a code that never gets sent. The dialer is a
	// bare mock with no expectations, so a flow that reaches SendMail fails
	// rather than passing quietly.
	startFlow := func(t *testing.T, consentService authenticate.ConsentService,
		request authenticate.RegistrationStartRequest, wantFlow bool) (*authenticate.RegistrationStartResponse, error) {
		t.Helper()

		ctx := context.Background()
		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		if request.Intent != authenticate.FlowIntentUnspecified {
			mockUserService.EXPECT().GetByID(ctx, email).
				Return(user.User{}, errors.New("user not found")).Maybe()
		}

		var dialer mailer.Dialer = &mailerMock.Dialer{}
		if wantFlow {
			mockFlowRepo.EXPECT().Set(ctx, mock.Anything).Return(nil)
			dialer = mailer.NewMockDialer()
		}

		srv := authenticate.NewService(nil, authenticate.Config{
			MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
			TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
		}, mockFlowRepo, dialer, nil, nil, mockUserService, nil, nil, nil, consentService, nil)

		request.Method = authenticate.MailOTPAuthMethod.String()
		request.Email = email
		return srv.StartFlow(ctx, request)
	}

	t.Run("a signup carrying every document starts the flow", func(t *testing.T) {
		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll(acceptedIDs).Return(documents, nil)

		got, err := startFlow(t, mockConsent, authenticate.RegistrationStartRequest{
			Intent:              authenticate.FlowIntentSignup,
			AcceptedDocumentIDs: acceptedIDs,
		}, true)

		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("a signup missing a document is rejected before the code is sent", func(t *testing.T) {
		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll([]string{"privacy_policy"}).
			Return(nil, consent.ErrMissingDocuments)

		got, err := startFlow(t, mockConsent, authenticate.RegistrationStartRequest{
			Intent:              authenticate.FlowIntentSignup,
			AcceptedDocumentIDs: []string{"privacy_policy"},
		}, false)

		assert.ErrorIs(t, err, authenticate.ErrConsentRequired)
		// the wrapped error still names what was missing, for the log
		assert.ErrorIs(t, err, consent.ErrMissingDocuments)
		assert.Nil(t, got)
	})

	t.Run("a signup carrying no document at all is rejected too", func(t *testing.T) {
		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll([]string(nil)).Return(nil, consent.ErrMissingDocuments)

		_, err := startFlow(t, mockConsent, authenticate.RegistrationStartRequest{
			Intent: authenticate.FlowIntentSignup,
		}, false)

		assert.ErrorIs(t, err, authenticate.ErrConsentRequired)
	})

	t.Run("an unspecified intent checks only that the ids are known", func(t *testing.T) {
		// frontier cannot yet know this request will create a user, so
		// completeness waits for user creation, but a typo can still be caught
		// before the redirect
		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().Resolve([]string{"privacy_policy"}).Return(documents[:1], nil)

		_, err := startFlow(t, mockConsent, authenticate.RegistrationStartRequest{
			AcceptedDocumentIDs: []string{"privacy_policy"},
		}, true)

		require.NoError(t, err)
	})

	t.Run("an unspecified intent is rejected for an id config does not know", func(t *testing.T) {
		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().Resolve([]string{"not_a_document"}).
			Return(nil, consent.ErrUnknownDocuments)

		_, err := startFlow(t, mockConsent, authenticate.RegistrationStartRequest{
			AcceptedDocumentIDs: []string{"not_a_document"},
		}, false)

		assert.ErrorIs(t, err, authenticate.ErrConsentRequired)
		assert.ErrorIs(t, err, consent.ErrUnknownDocuments)
	})

	t.Run("a login checks nothing, because it writes no record", func(t *testing.T) {
		// an unexpected call fails the test rather than passing silently
		mockConsent := mocks.NewConsentService(t)

		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		ctx := context.Background()
		mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{ID: "user-id", Email: email}, nil)
		mockFlowRepo.EXPECT().Set(ctx, mock.Anything).Return(nil)

		srv := authenticate.NewService(nil, authenticate.Config{
			MailOTP:   authenticate.MailOTPConfig{Validity: 10 * time.Minute},
			TestUsers: testusers.Config{Enabled: true, OTP: "111111", Domain: "example.com"},
		}, mockFlowRepo, mailer.NewMockDialer(), nil, nil, mockUserService, nil, nil, nil, mockConsent, nil)

		_, err := srv.StartFlow(ctx, authenticate.RegistrationStartRequest{
			Method: authenticate.MailOTPAuthMethod.String(),
			Email:  email,
			Intent: authenticate.FlowIntentLogin,
		})
		require.NoError(t, err)
	})

	t.Run("a deployment with consent disabled ignores the ids rather than rejecting them", func(t *testing.T) {
		// one client build works against both kinds of deployment, so the same
		// signup that a configured deployment gates goes straight through here
		disabled := consent.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)),
			consent.Config{Enabled: false}, nil, nil)

		got, err := startFlow(t, disabled, authenticate.RegistrationStartRequest{
			Intent:              authenticate.FlowIntentSignup,
			AcceptedDocumentIDs: acceptedIDs,
		}, true)

		require.NoError(t, err)
		require.NotNil(t, got)
	})
}

// TestFlow_IntentAndConsent covers the accessors directly: the JSON round trip
// the database puts metadata through, and the nil receiver callers rely on.
func TestFlow_IntentAndConsent(t *testing.T) {
	acceptedAt := time.Now().UTC().Truncate(time.Second)

	t.Run("a nil flow reads as no intent and no consent", func(t *testing.T) {
		var flow *authenticate.Flow

		assert.Equal(t, authenticate.FlowIntentUnspecified, flow.Intent())
		_, ok := flow.Consent()
		assert.False(t, ok)
	})

	t.Run("survives the round trip the database puts metadata through", func(t *testing.T) {
		ctx := context.Background()
		flowRepo := &jsonFlowRepository{flows: map[uuid.UUID]jsonStoredFlow{}}
		flowID := uuid.New()

		require.NoError(t, flowRepo.Set(ctx, mailOTPFlow(flowID, acceptedAt, "nonce", pkgMetadata.Metadata{
			"callback_url": "",
			"intent":       authenticate.FlowIntentSignup.String(),
			"consent": map[string]any{
				"accepted_document_ids": []string{"terms_of_service"},
				"ip_address":            "10.0.0.1",
				"at":                    acceptedAt,
			},
		})))

		stored, err := flowRepo.Get(ctx, flowID)
		require.NoError(t, err)

		// JSON gives the ids back as []any and the timestamp as a string, so the
		// accessors parse rather than assert
		assert.Equal(t, authenticate.FlowIntentSignup, stored.Intent())
		consent, ok := stored.Consent()
		require.True(t, ok)
		assert.Equal(t, []string{"terms_of_service"}, consent.AcceptedDocumentIDs)
		assert.Equal(t, "10.0.0.1", consent.IPAddress)
		assert.True(t, acceptedAt.Equal(consent.At))
	})

	t.Run("an unparseable or empty consent is no consent", func(t *testing.T) {
		for name, md := range map[string]pkgMetadata.Metadata{
			"missing":       {"callback_url": ""},
			"wrong type":    {"consent": "yes"},
			"no documents":  {"consent": map[string]any{"ip_address": "10.0.0.1"}},
			"unknown types": {"consent": map[string]any{"accepted_document_ids": []any{1, 2}}},
		} {
			t.Run(name, func(t *testing.T) {
				flow := &authenticate.Flow{Metadata: md}
				_, ok := flow.Consent()
				assert.False(t, ok)
			})
		}
	})

	t.Run("an intent of the wrong type reads as unspecified", func(t *testing.T) {
		flow := &authenticate.Flow{Metadata: pkgMetadata.Metadata{"intent": 2}}
		assert.Equal(t, authenticate.FlowIntentUnspecified, flow.Intent())
	})
}

// TestService_FinishFlow_Intent covers the second gate, at user creation, which
// is the only point OIDC can be gated at since it has no email before then.
func TestService_FinishFlow_Intent(t *testing.T) {
	timeNow := time.Now()
	otpHash, err := bcrypt.GenerateFromPassword([]byte("111111"), bcrypt.MinCost)
	require.NoError(t, err)

	const email = "test@example.com"
	existingUser := user.User{ID: "user-id", Email: email}

	tests := []struct {
		name       string
		intent     authenticate.FlowIntent
		userExists bool

		wantErr        error
		wantUserCreate bool
	}{
		{
			name:       "login logs the existing user in",
			intent:     authenticate.FlowIntentLogin,
			userExists: true,
		},
		{
			name:    "login never creates the account",
			intent:  authenticate.FlowIntentLogin,
			wantErr: authenticate.ErrLoginUserNotFound,
		},
		{
			name:           "signup creates the account",
			intent:         authenticate.FlowIntentSignup,
			wantUserCreate: true,
		},
		{
			name:       "signup never logs the existing user in",
			intent:     authenticate.FlowIntentSignup,
			userExists: true,
			wantErr:    authenticate.ErrSignupUserExists,
		},
		{
			name:       "without an intent an existing user is logged in",
			intent:     authenticate.FlowIntentUnspecified,
			userExists: true,
		},
		{
			name:           "without an intent an unknown address is created, as before",
			intent:         authenticate.FlowIntentUnspecified,
			wantUserCreate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			flowID := uuid.New()

			md := pkgMetadata.Metadata{"callback_url": ""}
			if tt.intent != authenticate.FlowIntentUnspecified {
				md["intent"] = tt.intent.String()
			}

			mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
			mockFlowRepo.EXPECT().Get(ctx, flowID).Return(mailOTPFlow(flowID, timeNow, string(otpHash), md), nil)
			mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)

			if tt.userExists {
				mockUserService.EXPECT().GetByID(ctx, email).Return(existingUser, nil)
			} else {
				mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))
			}
			if tt.wantUserCreate {
				mockUserService.EXPECT().Create(ctx, mock.Anything).Return(existingUser, nil)
			}

			srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
				nil, nil, mockUserService, nil, nil, nil, nil, nil)
			srv.Now = func() time.Time { return timeNow }

			got, err := srv.FinishFlow(ctx, authenticate.RegistrationFinishRequest{
				Method: authenticate.MailOTPAuthMethod.String(),
				State:  flowID.String(),
				Code:   "111111",
			})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				// a rejection is the error and nothing else: the handler maps it
				// to a connect code and the caller decides what to do with it
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, existingUser, got.User)
		})
	}
}

// TestService_FinishFlow_Consent covers the invariant this feature rests on: a
// user row without a consent record is impossible. The rollback that backs it
// is exercised against a real database in
// internal/store/postgres/user_consent_repository_test.go; what is checked here
// is which of the three outcomes each request reaches, and what is written for
// it.
func TestService_FinishFlow_Consent(t *testing.T) {
	timeNow := time.Now()
	otpHash, err := bcrypt.GenerateFromPassword([]byte("111111"), bcrypt.MinCost)
	require.NoError(t, err)

	const email = "test@example.com"
	consentedAt := timeNow.Add(-time.Minute).UTC()
	newUser := user.User{ID: "user-id", Email: email}
	documents := []consent.Document{
		{ID: "privacy_policy", Title: "Privacy Policy", Version: "2026-04-01", URL: "https://example.org/p"},
		{ID: "terms_of_service", Title: "Terms & Conditions", Version: "2026-04-01", URL: "https://example.org/t"},
	}
	acceptedIDs := []string{"privacy_policy", "terms_of_service"}

	// consentMetadata is what StartFlow wrote before the redirect, after a JSON
	// round trip through the flows table.
	consentMetadata := func() pkgMetadata.Metadata {
		return pkgMetadata.Metadata{
			"callback_url": "",
			"intent":       authenticate.FlowIntentSignup.String(),
			"consent": map[string]any{
				"accepted_document_ids": []any{"privacy_policy", "terms_of_service"},
				"ip_address":            "203.0.113.9",
				"at":                    consentedAt.Format(time.RFC3339Nano),
			},
		}
	}

	finish := func(t *testing.T, srv *authenticate.Service, ctx context.Context, flowID uuid.UUID) (*authenticate.RegistrationFinishResponse, error) {
		t.Helper()
		return srv.FinishFlow(ctx, authenticate.RegistrationFinishRequest{
			Method: authenticate.MailOTPAuthMethod.String(),
			State:  flowID.String(),
			Code:   "111111",
		})
	}

	t.Run("a complete payload writes the user and the consent in one transaction", func(t *testing.T) {
		ctx := context.Background()
		flowID := uuid.New()

		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		mockFlowRepo.EXPECT().Get(ctx, flowID).Return(mailOTPFlow(flowID, timeNow, string(otpHash), consentMetadata()), nil)
		mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
		mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))
		mockUserService.EXPECT().CreateWithTx(ctx, (*sqlx.Tx)(nil), mock.Anything).Return(newUser, nil)

		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll(acceptedIDs).Return(documents, nil)

		granted := consent.Consent{ID: "consent-id", UserID: newUser.ID}
		var grantRequest consent.GrantRequest
		mockConsent.EXPECT().Grant(ctx, (*sqlx.Tx)(nil), mock.Anything).
			Run(func(_ context.Context, _ *sqlx.Tx, req consent.GrantRequest) { grantRequest = req }).
			Return(granted, nil)
		// after the commit, not inside it
		mockConsent.EXPECT().RecordGranted(ctx, granted).Return()

		srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
			nil, nil, mockUserService, nil, nil, nil, mockConsent, fakeTransactor{})
		srv.Now = func() time.Time { return timeNow }

		got, err := finish(t, srv, ctx, flowID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, newUser, got.User)

		assert.Equal(t, newUser.ID, grantRequest.UserID)
		assert.Equal(t, newUser.Email, grantRequest.UserEmail)
		assert.Equal(t, documents, grantRequest.Documents)
		assert.Equal(t, consent.SourceSignup, grantRequest.Source)
		// the flow's own word for how the consent came in
		assert.Equal(t, authenticate.MailOTPAuthMethod.String(), grantRequest.AuthStrategy)
		// the IP and the time are from when the user accepted, not from now
		assert.Equal(t, "203.0.113.9", grantRequest.IPAddress)
		assert.True(t, consentedAt.Equal(grantRequest.ConsentedAt))
	})

	t.Run("an incomplete payload writes neither row", func(t *testing.T) {
		ctx := context.Background()
		flowID := uuid.New()

		md := consentMetadata()
		md["consent"] = map[string]any{
			"accepted_document_ids": []any{"privacy_policy"},
			"ip_address":            "203.0.113.9",
			"at":                    consentedAt.Format(time.RFC3339Nano),
		}

		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		mockFlowRepo.EXPECT().Get(ctx, flowID).Return(mailOTPFlow(flowID, timeNow, string(otpHash), md), nil)
		mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
		mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))

		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll([]string{"privacy_policy"}).
			Return(nil, consent.ErrMissingDocuments)

		// no transactor at all: an incomplete payload must never open one, and
		// a nil one would panic if it did
		srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
			nil, nil, mockUserService, nil, nil, nil, mockConsent, nil)
		srv.Now = func() time.Time { return timeNow }

		got, err := finish(t, srv, ctx, flowID)
		assert.ErrorIs(t, err, authenticate.ErrConsentRequired)
		// the wrapped error still names what was missing
		assert.ErrorIs(t, err, consent.ErrMissingDocuments)
		// the rejection is the error and nothing else
		assert.Nil(t, got)
		mockUserService.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		mockUserService.AssertNotCalled(t, "CreateWithTx", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("a flow carrying no consent at all is rejected too", func(t *testing.T) {
		// the check runs under every intent, not for the error but as the
		// invariant guarding the write
		ctx := context.Background()
		flowID := uuid.New()

		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		mockFlowRepo.EXPECT().Get(ctx, flowID).Return(
			mailOTPFlow(flowID, timeNow, string(otpHash), pkgMetadata.Metadata{"callback_url": ""}), nil)
		mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
		mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))

		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll([]string(nil)).Return(nil, consent.ErrMissingDocuments)

		srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
			nil, nil, mockUserService, nil, nil, nil, mockConsent, nil)
		srv.Now = func() time.Time { return timeNow }

		got, err := finish(t, srv, ctx, flowID)
		assert.ErrorIs(t, err, authenticate.ErrConsentRequired)
		assert.Nil(t, got)
	})

	t.Run("an existing user gets no consent record", func(t *testing.T) {
		// absolute: a record written outside a user creation would carry this
		// moment's timestamp and IP for an agreement made elsewhere
		ctx := context.Background()
		flowID := uuid.New()

		md := consentMetadata()
		delete(md, "intent") // a signup intent would be rejected by the gate first

		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		mockFlowRepo.EXPECT().Get(ctx, flowID).Return(mailOTPFlow(flowID, timeNow, string(otpHash), md), nil)
		mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
		mockUserService.EXPECT().GetByID(ctx, email).Return(newUser, nil)

		// the consent service is never reached, so an unexpected call fails the
		// test rather than passing silently
		mockConsent := mocks.NewConsentService(t)

		srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
			nil, nil, mockUserService, nil, nil, nil, mockConsent, nil)
		srv.Now = func() time.Time { return timeNow }

		got, err := finish(t, srv, ctx, flowID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, newUser, got.User)
	})

	t.Run("a deployment that asks for no consent creates the user as before", func(t *testing.T) {
		// with app.consent disabled ResolveAll accepts anything and resolves
		// nothing, and an empty set means write no record
		ctx := context.Background()
		flowID := uuid.New()

		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		mockFlowRepo.EXPECT().Get(ctx, flowID).Return(mailOTPFlow(flowID, timeNow, string(otpHash), consentMetadata()), nil)
		mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
		mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))
		mockUserService.EXPECT().Create(ctx, mock.Anything).Return(newUser, nil)

		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll(acceptedIDs).Return(nil, nil)

		srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
			nil, nil, mockUserService, nil, nil, nil, mockConsent, nil)
		srv.Now = func() time.Time { return timeNow }

		got, err := finish(t, srv, ctx, flowID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, newUser, got.User)
	})

	t.Run("a failed consent write fails the signup", func(t *testing.T) {
		ctx := context.Background()
		flowID := uuid.New()

		mockFlowRepo, mockUserService, _, _, _ := createMocks(t)
		mockFlowRepo.EXPECT().Get(ctx, flowID).Return(mailOTPFlow(flowID, timeNow, string(otpHash), consentMetadata()), nil)
		mockFlowRepo.EXPECT().Delete(ctx, flowID).Return(nil)
		mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))
		mockUserService.EXPECT().CreateWithTx(ctx, (*sqlx.Tx)(nil), mock.Anything).Return(newUser, nil)

		mockConsent := mocks.NewConsentService(t)
		mockConsent.EXPECT().ResolveAll(acceptedIDs).Return(documents, nil)
		mockConsent.EXPECT().Grant(ctx, (*sqlx.Tx)(nil), mock.Anything).
			Return(consent.Consent{}, consent.ErrConsentExists)

		srv := authenticate.NewService(nil, authenticate.Config{}, mockFlowRepo, nil,
			nil, nil, mockUserService, nil, nil, nil, mockConsent, fakeTransactor{})
		srv.Now = func() time.Time { return timeNow }

		got, err := finish(t, srv, ctx, flowID)
		assert.ErrorIs(t, err, consent.ErrConsentExists)
		assert.Nil(t, got)
		// no breadcrumb for a consent that was rolled back
		mockConsent.AssertNotCalled(t, "RecordGranted", mock.Anything, mock.Anything)
	})
}

// fakeTransactor runs the function it is given without a database, which is all
// the service needs from it: the rollback that makes the two inserts atomic is
// Postgres's job, and is tested against a real one in
// internal/store/postgres/user_consent_repository_test.go.
type fakeTransactor struct{}

func (fakeTransactor) WithTxn(_ context.Context, _ sql.TxOptions, txFunc func(*sqlx.Tx) error) error {
	return txFunc(nil)
}

// TestService_PassthroughHeader_Consent pins the exemption. Three paths create
// a user with no flow behind them, and they stay exempt because no account
// holder is present to consent. This is the one of the three that runs through
// getOrCreateUser, so it is the one that could have been gated by accident.
func TestService_PassthroughHeader_Consent(t *testing.T) {
	const email = "passthrough@example.com"
	ctx := authenticate.SetContextWithEmail(context.Background(), email)
	newUser := user.User{ID: "user-id", Email: email}

	_, mockUserService, _, _, _ := createMocks(t)
	mockUserService.EXPECT().GetByID(ctx, email).Return(user.User{}, errors.New("user not found"))
	mockUserService.EXPECT().Create(ctx, mock.Anything).Return(newUser, nil)

	// the consent service is wired and enabled, and is still never reached:
	// there is no flow, so there is nothing that could carry a consent
	mockConsent := mocks.NewConsentService(t)

	srv := authenticate.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), authenticate.Config{},
		nil, nil, nil, nil, mockUserService, nil, nil, nil, mockConsent, nil)

	got, err := srv.GetPrincipal(ctx, authenticate.PassthroughHeaderClientAssertion)
	require.NoError(t, err)
	assert.Equal(t, newUser.ID, got.ID)
	assert.Equal(t, schema.UserPrincipal, got.Type)
}
