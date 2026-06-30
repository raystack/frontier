package bootstrap

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSUCreator struct{ mock.Mock }

func (m *mockSUCreator) Create(ctx context.Context, su serviceuser.ServiceUser) (serviceuser.ServiceUser, error) {
	args := m.Called(ctx, su)
	return args.Get(0).(serviceuser.ServiceUser), args.Error(1)
}

func (m *mockSUCreator) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

type mockCredStore struct{ mock.Mock }

func (m *mockCredStore) Get(ctx context.Context, id string) (serviceuser.Credential, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(serviceuser.Credential), args.Error(1)
}

func (m *mockCredStore) Create(ctx context.Context, cred serviceuser.Credential) (serviceuser.Credential, error) {
	args := m.Called(ctx, cred)
	return args.Get(0).(serviceuser.Credential), args.Error(1)
}

func (m *mockCredStore) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

type mockSUPromoter struct{ mock.Mock }

func (m *mockSUPromoter) Sudo(ctx context.Context, id, relationName string) error {
	return m.Called(ctx, id, relationName).Error(0)
}

func bcryptHash(t *testing.T, secret string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(secret), bootstrapBcryptCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return string(h)
}

func TestEnsureBootstrapSuperUser(t *testing.T) {
	// Cost 14 is intentionally slow; under `-race -count 2` even a handful of
	// hash/compare calls blow the 150s test timeout (measured ~151s). Behaviour is
	// identical at any cost, so run the unit test at the minimum.
	orig := bootstrapBcryptCost
	bootstrapBcryptCost = bcrypt.MinCost
	t.Cleanup(func() { bootstrapBcryptCost = orig })

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	const clientID = "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	t.Run("no-op when not configured", func(t *testing.T) {
		users, creds, prom := new(mockSUCreator), new(mockCredStore), new(mockSUPromoter)

		assert.NoError(t, ensureBootstrapSuperUser(ctx, logger, SuperUserBootstrapConfig{}, users, creds, prom))
		creds.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
		users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		prom.AssertNotCalled(t, "Sudo", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("rejects a non-uuid client_id", func(t *testing.T) {
		users, creds, prom := new(mockSUCreator), new(mockCredStore), new(mockSUPromoter)
		cfg := SuperUserBootstrapConfig{ClientID: "not-a-uuid", ClientSecret: "s3cret"}

		assert.Error(t, ensureBootstrapSuperUser(ctx, logger, cfg, users, creds, prom))
		creds.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	})

	t.Run("creates service user + credential + promotes when absent", func(t *testing.T) {
		users, creds, prom := new(mockSUCreator), new(mockCredStore), new(mockSUPromoter)
		cfg := SuperUserBootstrapConfig{ClientID: clientID, ClientSecret: "s3cret"}

		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{}, serviceuser.ErrCredNotExist)
		users.On("Create", mock.Anything, mock.MatchedBy(func(su serviceuser.ServiceUser) bool {
			return su.OrgID == schema.PlatformOrgID.String() && su.Title == defaultBootstrapTitle
		})).Return(serviceuser.ServiceUser{ID: "su-id"}, nil)
		// Capture the credential and verify the hash once after the call rather
		// than inside the matcher: testify re-invokes argument matchers, and bcrypt
		// is costly enough there to stall CI.
		var created serviceuser.Credential
		creds.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { created = args.Get(1).(serviceuser.Credential) }).
			Return(serviceuser.Credential{ID: clientID}, nil)
		prom.On("Sudo", mock.Anything, "su-id", schema.AdminRelationName).Return(nil)

		assert.NoError(t, ensureBootstrapSuperUser(ctx, logger, cfg, users, creds, prom))
		users.AssertExpectations(t)
		creds.AssertExpectations(t)
		prom.AssertExpectations(t)

		assert.Equal(t, clientID, created.ID)
		assert.Equal(t, "su-id", created.ServiceUserID)
		assert.Equal(t, serviceuser.ClientSecretCredentialType, created.Type)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(created.SecretHash), []byte("s3cret")))
	})

	t.Run("ensures superuser without rotating when the secret matches", func(t *testing.T) {
		users, creds, prom := new(mockSUCreator), new(mockCredStore), new(mockSUPromoter)
		cfg := SuperUserBootstrapConfig{ClientID: clientID, ClientSecret: "s3cret"}

		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{
			ID: clientID, ServiceUserID: "su-id", SecretHash: bcryptHash(t, "s3cret"),
		}, nil)
		prom.On("Sudo", mock.Anything, "su-id", schema.AdminRelationName).Return(nil)

		assert.NoError(t, ensureBootstrapSuperUser(ctx, logger, cfg, users, creds, prom))
		prom.AssertExpectations(t)
		creds.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
		creds.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("rotates the secret when it changed, then ensures superuser", func(t *testing.T) {
		users, creds, prom := new(mockSUCreator), new(mockCredStore), new(mockSUPromoter)
		cfg := SuperUserBootstrapConfig{ClientID: clientID, ClientSecret: "new-secret"}

		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{
			ID: clientID, ServiceUserID: "su-id", SecretHash: bcryptHash(t, "old-secret"), Title: "t",
		}, nil)
		creds.On("Delete", mock.Anything, clientID).Return(nil)
		var rotated serviceuser.Credential
		creds.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { rotated = args.Get(1).(serviceuser.Credential) }).
			Return(serviceuser.Credential{ID: clientID}, nil)
		prom.On("Sudo", mock.Anything, "su-id", schema.AdminRelationName).Return(nil)

		assert.NoError(t, ensureBootstrapSuperUser(ctx, logger, cfg, users, creds, prom))
		creds.AssertExpectations(t)
		prom.AssertExpectations(t)
		users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

		assert.Equal(t, clientID, rotated.ID)
		assert.Equal(t, "su-id", rotated.ServiceUserID)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(rotated.SecretHash), []byte("new-secret")))
	})

	t.Run("rolls back the service user when credential creation fails", func(t *testing.T) {
		users, creds, prom := new(mockSUCreator), new(mockCredStore), new(mockSUPromoter)
		cfg := SuperUserBootstrapConfig{ClientID: clientID, ClientSecret: "s3cret"}

		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{}, serviceuser.ErrCredNotExist)
		users.On("Create", mock.Anything, mock.Anything).Return(serviceuser.ServiceUser{ID: "su-id"}, nil)
		creds.On("Create", mock.Anything, mock.Anything).Return(serviceuser.Credential{}, errors.New("db down"))
		users.On("Delete", mock.Anything, "su-id").Return(nil)

		assert.Error(t, ensureBootstrapSuperUser(ctx, logger, cfg, users, creds, prom))
		users.AssertExpectations(t) // Create + compensating Delete both invoked
		creds.AssertExpectations(t)
		prom.AssertNotCalled(t, "Sudo", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("aborts on a non-not-found credential lookup error", func(t *testing.T) {
		users, creds, prom := new(mockSUCreator), new(mockCredStore), new(mockSUPromoter)
		cfg := SuperUserBootstrapConfig{ClientID: clientID, ClientSecret: "s3cret"}

		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{}, errors.New("db down"))

		assert.Error(t, ensureBootstrapSuperUser(ctx, logger, cfg, users, creds, prom))
		users.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		prom.AssertNotCalled(t, "Sudo", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestBootstrapServiceUserID(t *testing.T) {
	ctx := context.Background()
	const clientID = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	cfg := SuperUserBootstrapConfig{ClientID: clientID, ClientSecret: "s3cret"}

	t.Run("empty when not configured", func(t *testing.T) {
		creds := new(mockCredStore)
		s := Service{suCredStore: creds}

		id, err := s.BootstrapServiceUserID(ctx)
		assert.NoError(t, err)
		assert.Empty(t, id)
		creds.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	})

	t.Run("returns the service-user id of the configured credential", func(t *testing.T) {
		creds := new(mockCredStore)
		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{ServiceUserID: "su-id"}, nil)
		s := Service{adminConfig: AdminConfig{Bootstrap: cfg}, suCredStore: creds}

		id, err := s.BootstrapServiceUserID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "su-id", id)
	})

	t.Run("empty when the credential does not exist yet", func(t *testing.T) {
		creds := new(mockCredStore)
		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{}, serviceuser.ErrCredNotExist)
		s := Service{adminConfig: AdminConfig{Bootstrap: cfg}, suCredStore: creds}

		id, err := s.BootstrapServiceUserID(ctx)
		assert.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("errors on a backend failure", func(t *testing.T) {
		creds := new(mockCredStore)
		creds.On("Get", mock.Anything, clientID).Return(serviceuser.Credential{}, errors.New("db down"))
		s := Service{adminConfig: AdminConfig{Bootstrap: cfg}, suCredStore: creds}

		_, err := s.BootstrapServiceUserID(ctx)
		assert.Error(t, err)
	})
}
