package userpat_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/userpat"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/mocks"
	"github.com/raystack/frontier/core/userpat/models"
	"io"
	"log/slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

func validPATValue(t *testing.T, prefix string) (value string, secretHash string) {
	t.Helper()
	secretBytes := make([]byte, 32)
	_, err := rand.Read(secretBytes)
	require.NoError(t, err)
	value = prefix + "_" + base64.RawURLEncoding.EncodeToString(secretBytes)
	hash := sha3.Sum256(secretBytes)
	secretHash = hex.EncodeToString(hash[:])
	return value, secretHash
}

func TestValidator_Validate(t *testing.T) {
	const prefix = "fpt"
	cfg := userpat.Config{
		Enabled: true,
		Prefix:  prefix,
	}

	t.Run("disabled feature returns ErrDisabled", func(t *testing.T) {
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, userpat.Config{Enabled: false})
		_, err := v.Validate(context.Background(), "fpt_anything")
		assert.ErrorIs(t, err, paterrors.ErrDisabled)
	})

	t.Run("wrong prefix returns ErrInvalidPAT", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		_, err := v.Validate(context.Background(), "ghp_sometoken")
		assert.ErrorIs(t, err, paterrors.ErrInvalidPAT)
	})

	t.Run("no prefix separator returns ErrInvalidPAT", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		_, err := v.Validate(context.Background(), "randomstring")
		assert.ErrorIs(t, err, paterrors.ErrInvalidPAT)
	})

	t.Run("malformed base64 returns ErrMalformedPAT", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		_, err := v.Validate(context.Background(), "fpt_!!!not-base64!!!")
		assert.ErrorIs(t, err, paterrors.ErrMalformedPAT)
		assert.NotErrorIs(t, err, paterrors.ErrInvalidPAT)
	})

	t.Run("unknown hash returns ErrNotFound", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		value, secretHash := validPATValue(t, prefix)
		repo.EXPECT().GetBySecretHash(mock.Anything, secretHash).Return(models.PAT{}, paterrors.ErrNotFound)

		_, err := v.Validate(context.Background(), value)
		assert.ErrorIs(t, err, paterrors.ErrNotFound)
	})

	t.Run("expired PAT returns ErrExpired", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		value, secretHash := validPATValue(t, prefix)
		repo.EXPECT().GetBySecretHash(mock.Anything, secretHash).Return(models.PAT{
			ID:        "pat-1",
			ExpiresAt: time.Now().Add(-time.Hour),
		}, nil)

		_, err := v.Validate(context.Background(), value)
		assert.ErrorIs(t, err, paterrors.ErrExpired)
	})

	t.Run("db error propagates as-is", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		value, secretHash := validPATValue(t, prefix)
		dbErr := errors.New("connection refused")
		repo.EXPECT().GetBySecretHash(mock.Anything, secretHash).Return(models.PAT{}, dbErr)

		_, err := v.Validate(context.Background(), value)
		assert.ErrorIs(t, err, dbErr)
		assert.NotErrorIs(t, err, paterrors.ErrInvalidPAT)
	})

	t.Run("UpdateUsedAt failure returns error", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		value, secretHash := validPATValue(t, prefix)
		repo.EXPECT().GetBySecretHash(mock.Anything, secretHash).Return(models.PAT{
			ID:        "pat-1",
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)
		dbErr := errors.New("connection refused")
		repo.EXPECT().UpdateUsedAt(mock.Anything, "pat-1", mock.AnythingOfType("time.Time")).Return(dbErr)

		_, err := v.Validate(context.Background(), value)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("valid PAT returns PAT and updates used_at", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		v := userpat.NewValidator(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg)

		value, secretHash := validPATValue(t, prefix)
		expectedPAT := models.PAT{
			ID:        "pat-1",
			UserID:    "user-1",
			OrgID:     "org-1",
			Title:     "my-pat",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		repo.EXPECT().GetBySecretHash(mock.Anything, secretHash).Return(expectedPAT, nil)
		repo.EXPECT().UpdateUsedAt(mock.Anything, "pat-1", mock.AnythingOfType("time.Time")).Return(nil)

		pat, err := v.Validate(context.Background(), value)
		require.NoError(t, err)
		assert.Equal(t, expectedPAT.ID, pat.ID)
		assert.Equal(t, expectedPAT.UserID, pat.UserID)
		assert.Equal(t, expectedPAT.OrgID, pat.OrgID)
		assert.Equal(t, expectedPAT.Title, pat.Title)
	})
}
