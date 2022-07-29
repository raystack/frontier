package postgres

import (
	"time"

	"database/sql"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/user"
)

type Resource struct {
	ID             string         `db:"id"`
	URN            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectID      string         `db:"project_id"`
	Project        Project        `db:"project"`
	GroupID        sql.NullString `db:"group_id"`
	Group          Group          `db:"group"`
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
		Project:        project.Project{ID: from.ProjectID},
		ProjectID:      from.ProjectID,
		Namespace:      namespace.Namespace{ID: from.NamespaceID},
		NamespaceID:    from.NamespaceID,
		Organization:   organization.Organization{ID: from.OrganizationID},
		OrganizationID: from.OrganizationID,
		GroupID:        from.GroupID.String,
		Group:          group.Group{ID: from.GroupID.String},
		User:           user.User{ID: from.UserID.String},
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}
}

type ResourceCols struct {
	ID             string         `db:"id"`
	URN            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectID      string         `db:"project_id"`
	GroupID        sql.NullString `db:"group_id"`
	OrganizationID string         `db:"org_id"`
	NamespaceID    string         `db:"namespace_id"`
	UserID         sql.NullString `db:"user_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}
