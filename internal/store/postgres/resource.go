package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/raystack/frontier/core/resource"
)

type Resource struct {
	ID            string         `db:"id"`
	URN           string         `db:"urn"`
	Name          string         `db:"name"`
	Title         string         `db:"title"`
	ProjectID     string         `db:"project_id"`
	Project       Project        `db:"project"`
	Metadata      []byte         `db:"metadata"`
	NamespaceID   string         `db:"namespace_name"`
	Namespace     Namespace      `db:"namespace"`
	PrincipalID   sql.NullString `db:"principal_id"`
	PrincipalType sql.NullString `db:"principal_type"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	DeletedAt     sql.NullTime   `db:"deleted_at"`
}

func (from Resource) transformToResource() (resource.Resource, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return resource.Resource{}, err
		}
	}

	return resource.Resource{
		ID:            from.ID,
		URN:           from.URN,
		Name:          from.Name,
		Title:         from.Title,
		ProjectID:     from.ProjectID,
		NamespaceID:   from.NamespaceID,
		Metadata:      unmarshalledMetadata,
		PrincipalID:   from.PrincipalID.String,
		PrincipalType: from.PrincipalType.String,
		CreatedAt:     from.CreatedAt,
		UpdatedAt:     from.UpdatedAt,
	}, nil
}

type ResourceCols struct {
	ID            string         `db:"id"`
	URN           string         `db:"urn"`
	Title         string         `db:"title"`
	Name          string         `db:"name"`
	ProjectID     string         `db:"project_id"`
	NamespaceID   string         `db:"namespace_name"`
	PrincipalID   sql.NullString `db:"principal_id"`
	PrincipalType sql.NullString `db:"principal_type"`
	Metadata      []byte         `db:"metadata"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}
