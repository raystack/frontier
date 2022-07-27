package postgres

import (
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
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

func buildListResourcesStatement(dialect goqu.DialectWrapper) *goqu.SelectDataset {
	listResourcesStatement := dialect.From(TABLE_RESOURCES)

	return listResourcesStatement
}

func buildGetResourcesByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	getResourcesByIDQuery, _, err := buildListResourcesStatement(dialect).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return getResourcesByIDQuery, err
}

func buildCreateResourceQuery(dialect goqu.DialectWrapper) (string, error) {
	createResourceQuery, _, err := dialect.Insert(TABLE_RESOURCES).Rows(
		goqu.Record{
			"urn":          goqu.L("$1"),
			"name":         goqu.L("$2"),
			"project_id":   goqu.L("$3"),
			"group_id":     goqu.L("$4"),
			"org_id":       goqu.L("$5"),
			"namespace_id": goqu.L("$6"),
			"user_id":      goqu.L("$7"),
		}).OnConflict(goqu.DoUpdate("ON CONSTRAINT resources_urn_unique", goqu.Record{
		"name":         goqu.L("$2"),
		"project_id":   goqu.L("$3"),
		"group_id":     goqu.L("$4"),
		"org_id":       goqu.L("$5"),
		"namespace_id": goqu.L("$6"),
		"user_id":      goqu.L("$7"),
	})).Returning(&ResourceCols{}).ToSQL()

	return createResourceQuery, err
}

func buildGetResourcesByURNQuery(dialect goqu.DialectWrapper) (string, error) {
	getResourcesByURNQuery, _, err := dialect.Select(&ResourceCols{}).From(TABLE_RESOURCES).Where(goqu.Ex{
		"urn": goqu.L("$1"),
	}).ToSQL()

	return getResourcesByURNQuery, err
}

func buildUpdateResourceQuery(dialect goqu.DialectWrapper) (string, error) {
	updateResourceQuery, _, err := dialect.Update(TABLE_RESOURCES).Set(
		goqu.Record{
			"name":         goqu.L("$2"),
			"project_id":   goqu.L("$3"),
			"group_id":     goqu.L("$4"),
			"org_id":       goqu.L("$5"),
			"namespace_id": goqu.L("$6"),
			"user_id":      goqu.L("$7"),
			"urn":          goqu.L("$8"),
		}).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return updateResourceQuery, err
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
