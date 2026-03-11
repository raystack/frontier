package userpat

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/userpat/models"
)

type Repository interface {
	Create(ctx context.Context, pat models.PAT) (models.PAT, error)
	CountActive(ctx context.Context, userID, orgID string) (int64, error)
	GetBySecretHash(ctx context.Context, secretHash string) (models.PAT, error)
	UpdateLastUsedAt(ctx context.Context, id string, at time.Time) error
}
