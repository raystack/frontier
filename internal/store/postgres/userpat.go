package postgres

import (
	"encoding/json"
	"time"

	"github.com/raystack/frontier/core/userpat"
)

type UserToken struct {
	ID         string     `db:"id" goqu:"skipinsert"`
	UserID     string     `db:"user_id"`
	OrgID      string     `db:"org_id"`
	Title      string     `db:"title"`
	SecretHash string     `db:"secret_hash"`
	Metadata   []byte     `db:"metadata"`
	LastUsedAt *time.Time `db:"last_used_at"`
	ExpiresAt  time.Time  `db:"expires_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

func (t UserToken) transform() (userpat.PersonalAccessToken, error) {
	var unmarshalledMetadata map[string]any
	if len(t.Metadata) > 0 {
		if err := json.Unmarshal(t.Metadata, &unmarshalledMetadata); err != nil {
			return userpat.PersonalAccessToken{}, err
		}
	}
	return userpat.PersonalAccessToken{
		ID:         t.ID,
		UserID:     t.UserID,
		OrgID:      t.OrgID,
		Title:      t.Title,
		SecretHash: t.SecretHash,
		Metadata:   unmarshalledMetadata,
		LastUsedAt: t.LastUsedAt,
		ExpiresAt:  t.ExpiresAt,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
	}, nil
}
