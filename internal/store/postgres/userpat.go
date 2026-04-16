package postgres

import (
	"encoding/json"
	"time"

	"github.com/raystack/frontier/core/userpat/models"
)

type UserPAT struct {
	ID            string     `db:"id" goqu:"skipinsert"`
	UserID        string     `db:"user_id"`
	OrgID         string     `db:"org_id"`
	Title         string     `db:"title"`
	SecretHash    string     `db:"secret_hash"`
	Metadata      []byte     `db:"metadata"`
	LastUsedAt    *time.Time `db:"last_used_at"`
	RegeneratedAt *time.Time `db:"regenerated_at"`
	ExpiresAt     time.Time  `db:"expires_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (t UserPAT) transform() (models.PAT, error) {
	var unmarshalledMetadata map[string]any
	if len(t.Metadata) > 0 {
		if err := json.Unmarshal(t.Metadata, &unmarshalledMetadata); err != nil {
			return models.PAT{}, err
		}
	}
	return models.PAT{
		ID:            t.ID,
		UserID:        t.UserID,
		OrgID:         t.OrgID,
		Title:         t.Title,
		SecretHash:    t.SecretHash,
		Metadata:      unmarshalledMetadata,
		LastUsedAt:    t.LastUsedAt,
		RegeneratedAt: t.RegeneratedAt,
		ExpiresAt:     t.ExpiresAt,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}, nil
}
