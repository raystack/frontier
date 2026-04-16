package userpat

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/salt/log"
	"golang.org/x/crypto/sha3"
)

// Validator validates PAT values during authentication.
type Validator struct {
	repo   Repository
	config Config
	logger log.Logger
}

func NewValidator(logger log.Logger, repo Repository, config Config) *Validator {
	return &Validator{
		repo:   repo,
		config: config,
		logger: logger,
	}
}

// Validate checks a PAT value and returns the corresponding PAT.
// Returns ErrInvalidPAT if the value doesn't match the configured prefix (allowing
// the auth chain to fall through to the next authenticator).
// Returns ErrMalformedPAT, ErrExpired, ErrNotFound, or ErrDisabled for terminal auth failures.
func (v *Validator) Validate(ctx context.Context, value string) (models.PAT, error) {
	if !v.config.Enabled {
		return models.PAT{}, paterrors.ErrDisabled
	}

	prefix := v.config.Prefix + "_"
	if !strings.HasPrefix(value, prefix) {
		return models.PAT{}, paterrors.ErrInvalidPAT
	}

	encoded := value[len(prefix):]
	secretBytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: invalid encoding", paterrors.ErrMalformedPAT)
	}

	hash := sha3.Sum256(secretBytes)
	secretHash := hex.EncodeToString(hash[:])

	pat, err := v.repo.GetBySecretHash(ctx, secretHash)
	if err != nil {
		return models.PAT{}, err
	}

	if pat.ExpiresAt.Before(time.Now()) {
		return models.PAT{}, paterrors.ErrExpired
	}

	if err := v.repo.UpdateUsedAt(ctx, pat.ID, time.Now()); err != nil {
		return models.PAT{}, fmt.Errorf("updating used_at: %w", err)
	}

	return pat, nil
}

// GetByID retrieves a PAT by ID. Used by the access token authenticator
// to reconstruct the PAT principal from JWT claims.
func (v *Validator) GetByID(ctx context.Context, id string) (models.PAT, error) {
	return v.repo.GetByID(ctx, id)
}
