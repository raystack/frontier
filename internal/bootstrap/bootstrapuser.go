package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// bootstrapBcryptCost is a var (not const) so tests can drop it to bcrypt.MinCost;
// cost 14 under `-race -count 2` blows the test timeout.
var bootstrapBcryptCost = 14

const defaultBootstrapTitle = "GitOps Bootstrap Superuser"

// SuperUserBootstrapConfig configures a superuser service account seeded from config,
// so automation (like GitOps reconcile) always has a superuser to log in as at boot —
// no existing superuser needed to create the first one. Leave both fields empty to disable.
type SuperUserBootstrapConfig struct {
	ClientID     string `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string `yaml:"client_secret" mapstructure:"client_secret"`
	Title        string `yaml:"title" mapstructure:"title"`
}

func (c SuperUserBootstrapConfig) title() string {
	if t := strings.TrimSpace(c.Title); t != "" {
		return t
	}
	return defaultBootstrapTitle
}

// ServiceUserCreator creates and looks up a bare service-user row (no org membership;
// a platform superuser needs none), backed by the repository, not serviceuser.Service.
type ServiceUserCreator interface {
	// GetByID returns serviceuser.ErrNotExist when absent, letting us reuse the
	// fixed-id bootstrap row instead of creating a duplicate.
	GetByID(ctx context.Context, id string) (serviceuser.ServiceUser, error)
	Create(ctx context.Context, su serviceuser.ServiceUser) (serviceuser.ServiceUser, error)
	// Delete rolls back a row this boot just created when the credential write fails.
	Delete(ctx context.Context, id string) error
}

// ServiceUserCredentialStore manages client-secret credentials keyed by id.
type ServiceUserCredentialStore interface {
	Get(ctx context.Context, id string) (serviceuser.Credential, error)
	Create(ctx context.Context, cred serviceuser.Credential) (serviceuser.Credential, error)
	Delete(ctx context.Context, id string) error
}

// SuperUserPromoter grants a platform relation (admin/member) to a principal.
type SuperUserPromoter interface {
	Sudo(ctx context.Context, id, relationName string) error
}

// EnsureBootstrapSuperUser makes sure the configured superuser service account exists
// and is a platform superuser. Safe to run on every boot; a no-op when unconfigured.
func (s Service) EnsureBootstrapSuperUser(ctx context.Context) error {
	return ensureBootstrapSuperUser(ctx, s.logger, s.adminConfig.Bootstrap, s.suCreator, s.suCredStore, s.suPromoter)
}

// ensureBootstrapSuperUser is the testable core, keyed on the credential id (client_id):
// if the credential is absent it creates the service user + credential + promotes;
// if present it rotates the secret when it changed, then re-asserts the superuser relation.
func ensureBootstrapSuperUser(
	ctx context.Context,
	logger *slog.Logger,
	cfg SuperUserBootstrapConfig,
	users ServiceUserCreator,
	creds ServiceUserCredentialStore,
	promoter SuperUserPromoter,
) error {
	clientID := strings.TrimSpace(cfg.ClientID)
	// Skip only when fully unset; a half-configured bootstrap is a mistake, so fail fast.
	switch {
	case clientID == "" && cfg.ClientSecret == "":
		return nil
	case clientID == "" || cfg.ClientSecret == "":
		return errors.New("bootstrap superuser: client_id and client_secret must be set together")
	}
	// client_id is a UUID credential-id column; validate up front for a clear error.
	if _, err := uuid.Parse(clientID); err != nil {
		return fmt.Errorf("bootstrap superuser: client_id must be a valid uuid: %w", err)
	}

	cred, err := creds.Get(ctx, clientID)
	switch {
	case errors.Is(err, serviceuser.ErrCredNotExist):
		return createBootstrapSuperUser(ctx, logger, cfg, clientID, users, creds, promoter)
	case err != nil:
		return fmt.Errorf("bootstrap superuser: get credential %q: %w", clientID, err)
	}

	if bcrypt.CompareHashAndPassword([]byte(cred.SecretHash), []byte(cfg.ClientSecret)) != nil {
		if err := rotateBootstrapSecret(ctx, cfg, clientID, cred, creds); err != nil {
			return err
		}
		logger.InfoContext(ctx, "rotated bootstrap superuser secret", "client_id", clientID)
	}
	return promoteBootstrapSuperUser(ctx, promoter, cred.ServiceUserID)
}

func createBootstrapSuperUser(
	ctx context.Context,
	logger *slog.Logger,
	cfg SuperUserBootstrapConfig,
	clientID string,
	users ServiceUserCreator,
	creds ServiceUserCredentialStore,
	promoter SuperUserPromoter,
) error {
	suID, created, err := ensureBootstrapServiceUser(ctx, cfg, users)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.ClientSecret), bootstrapBcryptCost)
	if err != nil {
		return rollbackFreshBootstrapSU(ctx, logger, users, suID, created,
			fmt.Errorf("bootstrap superuser: hash secret: %w", err))
	}
	// Reached only because the caller saw the credential is missing, so this never
	// adds a second credential for the bootstrap SA.
	if _, err := creds.Create(ctx, serviceuser.Credential{
		ID:            clientID,
		ServiceUserID: suID,
		Type:          serviceuser.ClientSecretCredentialType,
		SecretHash:    string(hash),
		Title:         cfg.title(),
	}); err != nil {
		return rollbackFreshBootstrapSU(ctx, logger, users, suID, created,
			fmt.Errorf("bootstrap superuser: create credential: %w", err))
	}
	logger.InfoContext(ctx, "created bootstrap superuser service account",
		"client_id", clientID, "serviceuser_id", suID)
	return promoteBootstrapSuperUser(ctx, promoter, suID)
}

// ensureBootstrapServiceUser returns the fixed-id bootstrap row, creating it only if
// absent so recovery is idempotent (never a duplicate row). The bool reports whether
// this call created it.
func ensureBootstrapServiceUser(ctx context.Context, cfg SuperUserBootstrapConfig, users ServiceUserCreator) (string, bool, error) {
	id := schema.BootstrapServiceUserID
	if _, err := users.GetByID(ctx, id); err == nil {
		return id, false, nil
	} else if !errors.Is(err, serviceuser.ErrNotExist) {
		return "", false, fmt.Errorf("bootstrap superuser: get service user: %w", err)
	}
	su, err := users.Create(ctx, serviceuser.ServiceUser{
		ID:    id,
		OrgID: schema.PlatformOrgID.String(),
		Title: cfg.title(),
	})
	if err != nil {
		return "", false, fmt.Errorf("bootstrap superuser: create service user: %w", err)
	}
	return su.ID, true, nil
}

// rollbackFreshBootstrapSU deletes the row on failure only if this boot created it;
// a reused (possibly promoted) row is left for the next boot to retry. Rollback is
// best-effort — the original error is always returned.
func rollbackFreshBootstrapSU(ctx context.Context, logger *slog.Logger, users ServiceUserCreator, suID string, created bool, cause error) error {
	if !created {
		return cause
	}
	if err := users.Delete(ctx, suID); err != nil {
		logger.WarnContext(ctx, "failed to roll back freshly created bootstrap service user",
			"serviceuser_id", suID, "err", err.Error())
	}
	return cause
}

// rotateBootstrapSecret replaces the stored secret. The credential repo has no update,
// so it deletes and recreates with the same id (a brief gap at boot is fine).
func rotateBootstrapSecret(
	ctx context.Context,
	cfg SuperUserBootstrapConfig,
	clientID string,
	cred serviceuser.Credential,
	creds ServiceUserCredentialStore,
) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.ClientSecret), bootstrapBcryptCost)
	if err != nil {
		return fmt.Errorf("bootstrap superuser: hash secret: %w", err)
	}
	if err := creds.Delete(ctx, clientID); err != nil {
		return fmt.Errorf("bootstrap superuser: rotate (delete): %w", err)
	}
	if _, err := creds.Create(ctx, serviceuser.Credential{
		ID:            clientID,
		ServiceUserID: cred.ServiceUserID,
		Type:          serviceuser.ClientSecretCredentialType,
		SecretHash:    string(hash),
		Title:         cred.Title,
	}); err != nil {
		return fmt.Errorf("bootstrap superuser: rotate (create): %w", err)
	}
	return nil
}

func promoteBootstrapSuperUser(ctx context.Context, promoter SuperUserPromoter, suID string) error {
	if err := promoter.Sudo(ctx, suID, schema.AdminRelationName); err != nil {
		return fmt.Errorf("bootstrap superuser: promote %q: %w", suID, err)
	}
	return nil
}
