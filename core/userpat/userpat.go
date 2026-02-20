package userpat

import (
	"context"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type PersonalAccessToken struct {
	ID         string `rql:"name=id,type=string"`
	UserID     string `rql:"name=user_id,type=string"`
	OrgID      string `rql:"name=org_id,type=string"`
	Title      string `rql:"name=title,type=string"`
	SecretHash string
	Metadata   metadata.Metadata
	LastUsedAt *time.Time `rql:"name=last_used_at,type=datetime"`
	ExpiresAt  time.Time  `rql:"name=expires_at,type=datetime"`
	CreatedAt  time.Time  `rql:"name=created_at,type=datetime"`
	UpdatedAt  time.Time  `rql:"name=updated_at,type=datetime"`
}

type Repository interface {
	Create(ctx context.Context, pat PersonalAccessToken) (PersonalAccessToken, error)
	CountActive(ctx context.Context, userID, orgID string) (int64, error)
}
