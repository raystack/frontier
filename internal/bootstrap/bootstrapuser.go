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

// bootstrapBcryptCost matches serviceuser.CreateSecret's cost for client secrets.
// It is a var (not a const) so unit tests can lower it to bcrypt.MinCost — cost 14
// is deliberately expensive and, run under `-race -count 2`, blows the test timeout.
var bootstrapBcryptCost = 14

const defaultBootstrapTitle = "GitOps Bootstrap Superuser"

// SuperUserBootstrapConfig configures a superuser service account seeded from
// config. It gives a username/password-style credential (client_id + client_secret)
// so automation (like the GitOps reconcile flow) always has a superuser to log in as
// at boot. This means you don't need an existing superuser to create the first one.
// You provide the secret, and it is rotated when it changes. Leave client_id and
// client_secret empty to turn this off.
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

// ServiceUserCreator creates and looks up a bare service-user row. The bootstrap
// superuser deliberately skips org membership (a platform superuser needs none),
// so this is the repository's Create/GetByID, not serviceuser.Service.
type ServiceUserCreator interface {
	// GetByID looks up a service user by id and returns serviceuser.ErrNotExist
	// when it does not exist. Used to reuse the fixed-id bootstrap row instead of
	// creating a duplicate.
	GetByID(ctx context.Context, id string) (serviceuser.ServiceUser, error)
	Create(ctx context.Context, su serviceuser.ServiceUser) (serviceuser.ServiceUser, error)
	// Delete removes a service user by id. Used to roll back a bootstrap row that
	// this boot just created, when the credential write then fails.
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

// EnsureBootstrapSuperUser makes sure the configured superuser service account
// exists and is a platform superuser. It is safe to run on every boot. It does
// nothing when the account is not configured.
func (s Service) EnsureBootstrapSuperUser(ctx context.Context) error {
	return ensureBootstrapSuperUser(ctx, s.logger, s.adminConfig.Bootstrap, s.suCreator, s.suCredStore, s.suPromoter)
}

// ensureBootstrapSuperUser holds the testable core. The account lives in the
// platform/nil org (serviceusers.org_id is nullable, no FK) and is created
// without org membership. It decides what to do based on whether the credential
// id (client_id) already exists:
//   - not there: create the service user + client-secret credential + promote it;
//   - there: rotate the secret if it changed, then make sure the superuser relation is set.
func ensureBootstrapSuperUser(
	ctx context.Context,
	logger *slog.Logger,
	cfg SuperUserBootstrapConfig,
	users ServiceUserCreator,
	creds ServiceUserCredentialStore,
	promoter SuperUserPromoter,
) error {
	clientID := strings.TrimSpace(cfg.ClientID)
	// Only skip when the config is fully unset. A half-configured bootstrap (only
	// one of client_id/client_secret) is a mistake that would quietly skip seeding
	// the superuser — fail early instead.
	switch {
	case clientID == "" && cfg.ClientSecret == "":
		return nil
	case clientID == "" || cfg.ClientSecret == "":
		return errors.New("bootstrap superuser: client_id and client_secret must be set together")
	}
	// client_id is the service-user credential's id, which is a UUID column; a
	// non-UUID would otherwise fail with an opaque SQL error deep in the lookup.
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
	// creds.Create runs only because ensureBootstrapSuperUser already saw the
	// credential (id == clientID) is missing, so we never add a second credential
	// for the bootstrap SA.
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

// ensureBootstrapServiceUser returns the fixed-id bootstrap service-user row,
// creating it only when it does not already exist. Reusing an existing row keeps
// recovery idempotent: if an earlier boot created the row but failed before
// writing the credential (or a rotation deleted the credential), we must not try
// to create a second row with the same id. The bool reports whether this call
// created the row.
func ensureBootstrapServiceUser(ctx context.Context, cfg SuperUserBootstrapConfig, users ServiceUserCreator) (string, bool, error) {
	id := schema.BootstrapServiceUserID
	if _, err := users.GetByID(ctx, id); err == nil {
		return id, false, nil
	} else if !errors.Is(err, serviceuser.ErrNotExist) {
		return "", false, fmt.Errorf("bootstrap superuser: get service user: %w", err)
	}
	su, err := users.Create(ctx, serviceuser.ServiceUser{
		ID:    id, // fixed id (see schema.BootstrapServiceUserID)
		OrgID: schema.PlatformOrgID.String(),
		Title: cfg.title(),
	})
	if err != nil {
		return "", false, fmt.Errorf("bootstrap superuser: create service user: %w", err)
	}
	return su.ID, true, nil
}

// rollbackFreshBootstrapSU deletes the service-user row when a later step fails,
// but only if THIS boot created it. If we reused an existing row (an earlier boot
// created it, or a rotation left it in place), we leave it — the next boot finds
// it by its fixed id and retries the missing credential. Recovery is idempotent,
// so a failed rollback is not fatal. The original error is always returned.
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

// rotateBootstrapSecret replaces the stored secret. The credential repo has no
// update, so delete + recreate with the same id (a brief gap at boot is fine).
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
