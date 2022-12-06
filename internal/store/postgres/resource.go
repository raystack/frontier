package postgres

import (
	"time"

	"database/sql"

	"github.com/odpf/shield/core/resource"
)

type Resource struct {
	ID             string         `db:"id"`
	URN            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectID      string         `db:"project_id"`
	Project        Project        `db:"project"`
	OrganizationID string         `db:"org_id"`
	Organization   Organization   `db:"organization"`
	NamespaceID    string         `db:"namespace_id"`
	Namespace      Namespace      `db:"namespace"`
	User           User           `db:"user"`
	UserID         sql.NullString `db:"user_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      sql.NullTime   `db:"deleted_at"`
}

func (from Resource) transformToResource() resource.Resource {
	// TODO: remove *ID
	return resource.Resource{
		Idxa:           from.ID,
		URN:            from.URN,
		Name:           from.Name,
		ProjectID:      from.ProjectID,
		NamespaceID:    from.NamespaceID,
		OrganizationID: from.OrganizationID,
		UserID:         from.UserID.String,
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}
}

type ResourceCols struct {
	ID             string         `db:"id"`
	URN            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectID      string         `db:"project_id"`
	OrganizationID string         `db:"org_id"`
	NamespaceID    string         `db:"namespace_id"`
	UserID         sql.NullString `db:"user_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}
