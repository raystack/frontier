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

// SuperUserBootstrapConfig configures a config-seeded superuser service account.
// It supplies a username/password-style credential (client_id + client_secret)
// so automation (e.g. the GitOps reconcile flow) has a guaranteed superuser
// identity at boot — removing the chicken-and-egg of needing a superuser to
// create the first superuser. The secret is operator-provided and rotated when
// it changes. Leave client_id/client_secret empty to disable.
type SuperUserBootstrapConfig struct {
	ClientID     string `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string `yaml:"client_secret" mapstructure:"client_secret"`
	Title        string `yaml:"title" mapstructure:"title"`
}

func (c SuperUserBootstrapConfig) enabled() bool {
	return strings.TrimSpace(c.ClientID) != "" && c.ClientSecret != ""
}

func (c SuperUserBootstrapConfig) title() string {
	if t := strings.TrimSpace(c.Title); t != "" {
		return t
	}
	return defaultBootstrapTitle
}

// ServiceUserCreator creates a bare service-user row. The bootstrap superuser
// deliberately skips org membership (a platform superuser needs none), so this
// is the repository's Create, not serviceuser.Service.Create.
type ServiceUserCreator interface {
	Create(ctx context.Context, su serviceuser.ServiceUser) (serviceuser.ServiceUser, error)
	// Delete removes a service user by id. Used to roll back a just-created
	// bootstrap service user when credential creation fails, so repeated boot
	// failures don't accumulate credential-less orphan rows.
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

// EnsureBootstrapSuperUser idempotently ensures the configured superuser service
// account exists and is a platform superuser. No-op when not configured.
func (s Service) EnsureBootstrapSuperUser(ctx context.Context) error {
	return ensureBootstrapSuperUser(ctx, s.logger, s.adminConfig.Bootstrap, s.suCreator, s.suCredStore, s.suPromoter)
}

// ensureBootstrapSuperUser holds the testable core. The account lives in the
// platform/nil org (serviceusers.org_id is nullable, no FK) and is created
// without org membership. Idempotency is keyed on the credential id (client_id):
//   - absent: create the service user + client-secret credential + promote;
//   - present: rotate the secret if it changed, then (re)ensure the superuser relation.
func ensureBootstrapSuperUser(
	ctx context.Context,
	logger *slog.Logger,
	cfg SuperUserBootstrapConfig,
	users ServiceUserCreator,
	creds ServiceUserCredentialStore,
	promoter SuperUserPromoter,
) error {
	if !cfg.enabled() {
		return nil
	}
	clientID := strings.TrimSpace(cfg.ClientID)
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
	su, err := users.Create(ctx, serviceuser.ServiceUser{
		ID:    schema.BootstrapServiceUserID, // fixed, well-known id (see schema.BootstrapServiceUserID)
		OrgID: schema.PlatformOrgID.String(),
		Title: cfg.title(),
	})
	if err != nil {
		return fmt.Errorf("bootstrap superuser: create service user: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.ClientSecret), bootstrapBcryptCost)
	if err != nil {
		return cleanupOrphanBootstrapSU(ctx, logger, users, su.ID,
			fmt.Errorf("bootstrap superuser: hash secret: %w", err))
	}
	if _, err := creds.Create(ctx, serviceuser.Credential{
		ID:            clientID,
		ServiceUserID: su.ID,
		Type:          serviceuser.ClientSecretCredentialType,
		SecretHash:    string(hash),
		Title:         cfg.title(),
	}); err != nil {
		return cleanupOrphanBootstrapSU(ctx, logger, users, su.ID,
			fmt.Errorf("bootstrap superuser: create credential: %w", err))
	}
	logger.InfoContext(ctx, "created bootstrap superuser service account",
		"client_id", clientID, "serviceuser_id", su.ID)
	return promoteBootstrapSuperUser(ctx, promoter, su.ID)
}

// cleanupOrphanBootstrapSU best-effort deletes a service user created moments
// earlier when a later step (hashing, credential creation) fails. Without it the
// row would be orphaned — credential-less and unpromoted — and the next boot,
// still finding no credential, would create another. The original cause is always
// returned; a failed cleanup is only logged (the next boot retries anyway).
func cleanupOrphanBootstrapSU(ctx context.Context, logger *slog.Logger, users ServiceUserCreator, suID string, cause error) error {
	if err := users.Delete(ctx, suID); err != nil {
		logger.WarnContext(ctx, "failed to roll back orphan bootstrap service user",
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
