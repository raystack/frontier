package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/raystack/shield/core/resource"
)

type Resource struct {
	ID          string         `db:"id"`
	URN         string         `db:"urn"`
	Name        string         `db:"name"`
	ProjectID   string         `db:"project_id"`
	Project     Project        `db:"project"`
	Metadata    []byte         `db:"metadata"`
	NamespaceID string         `db:"namespace_name"`
	Namespace   Namespace      `db:"namespace"`
	User        User           `db:"user"`
	UserID      sql.NullString `db:"user_id"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

func (from Resource) transformToResource() (resource.Resource, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return resource.Resource{}, err
		}
	}

	return resource.Resource{
		ID:          from.ID,
		URN:         from.URN,
		Name:        from.Name,
		ProjectID:   from.ProjectID,
		NamespaceID: from.NamespaceID,
		Metadata:    unmarshalledMetadata,
		UserID:      from.UserID.String,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}

type ResourceCols struct {
	ID          string         `db:"id"`
	URN         string         `db:"urn"`
	Name        string         `db:"name"`
	ProjectID   string         `db:"project_id"`
	NamespaceID string         `db:"namespace_name"`
	UserID      sql.NullString `db:"user_id"`
	Metadata    []byte         `db:"metadata"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}
