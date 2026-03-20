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
	UpdateLastUsedAt(ctx context.Context, id string, at time.Time) error
	Update(ctx context.Context, pat models.PAT) (models.PAT, error)
	Delete(ctx context.Context, id string) error
}
