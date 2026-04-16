package userpat

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/salt/rql"
)

type Repository interface {
	Create(ctx context.Context, pat models.PAT) (models.PAT, error)
	CountActive(ctx context.Context, userID, orgID string) (int64, error)
	GetByID(ctx context.Context, id string) (models.PAT, error)
	List(ctx context.Context, userID, orgID string, query *rql.Query) (models.PATList, error)
	GetBySecretHash(ctx context.Context, secretHash string) (models.PAT, error)
	IsTitleAvailable(ctx context.Context, userID, orgID, title string) (bool, error)
	UpdateUsedAt(ctx context.Context, id string, at time.Time) error
	Update(ctx context.Context, pat models.PAT) (models.PAT, error)
	Regenerate(ctx context.Context, id, secretHash string, expiresAt time.Time) (models.PAT, error)
	Delete(ctx context.Context, id string) error
	ListExpiryReminderPending(ctx context.Context, days int) ([]models.PAT, error)
	ListExpiredNoticePending(ctx context.Context) ([]models.PAT, error)
	SetAlertSentMetadata(ctx context.Context, id string, key string) error
}
