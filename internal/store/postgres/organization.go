package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/raystack/frontier/core/organization"
)

type Organization struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Title     sql.NullString `db:"title"`
	Metadata  []byte         `db:"metadata"`
	State     sql.NullString `db:"state"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at"`
}

func (from Organization) transformToOrg() (organization.Organization, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return organization.Organization{}, err
		}
	}

	return organization.Organization{
		ID:        from.ID,
		Name:      from.Name,
		Title:     from.Title.String,
		Metadata:  unmarshalledMetadata,
		State:     organization.State(from.State.String),
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
