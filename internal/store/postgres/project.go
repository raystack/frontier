package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
)

type Project struct {
	ID        string         `db:"id"`
	Name      string         `db:"name"`
	Title     sql.NullString `db:"title"`
	OrgID     string         `db:"org_id"`
	Metadata  []byte         `db:"metadata"`
	State     sql.NullString `db:"state"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at"`
}

func (from Project) transformToProject() (project.Project, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return project.Project{}, err
		}
	}

	return project.Project{
		ID:           from.ID,
		Name:         from.Name,
		Title:        from.Title.String,
		Organization: organization.Organization{ID: from.OrgID},
		Metadata:     unmarshalledMetadata,
		State:        project.State(from.State.String),
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}, nil
}
