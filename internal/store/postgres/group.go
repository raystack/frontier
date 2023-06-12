package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/raystack/shield/core/group"
)

type Group struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	Slug      string       `db:"slug"`
	OrgID     string       `db:"org_id"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (from Group) transformToGroup() (group.Group, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return group.Group{}, err
	}

	return group.Group{
		ID:             from.ID,
		Name:           from.Name,
		Slug:           from.Slug,
		OrganizationID: from.OrgID,
		Metadata:       unmarshalledMetadata,
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}, nil
}
