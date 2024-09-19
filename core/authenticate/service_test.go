package authenticate_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/mocks"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/authenticate/token"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
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

				return authenticate.NewService(nil, authenticate.Config{},
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
