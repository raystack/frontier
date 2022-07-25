package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/role"
)

type Project struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	Slug      string       `db:"slug"`
	OrgID     string       `db:"org_id"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// *Get Projects Query
func buildGetProjectsBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	getProjectsBySlugQuery, _, err := dialect.From(TABLE_PROJECTS).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).ToSQL()

	return getProjectsBySlugQuery, err
}

func buildGetProjectsByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	getProjectsByIDQuery, _, err := dialect.From(TABLE_PROJECTS).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).ToSQL()

	return getProjectsByIDQuery, err
}

// *Create Project Query
func buildCreateProjectQuery(dialect goqu.DialectWrapper) (string, error) {
	createProjectQuery, _, err := dialect.Insert(TABLE_PROJECTS).Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"slug":     goqu.L("$2"),
			"org_id":   goqu.L("$3"),
			"metadata": goqu.L("$4"),
		}).Returning(&Project{}).ToSQL()

	return createProjectQuery, err
}

// *List Project Query
func buildListProjectAdminsQuery(dialect goqu.DialectWrapper) (string, error) {
	listProjectAdminsQuery, _, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).From(goqu.T(TABLE_RELATION).As("r")).Join(
		goqu.T(TABLE_USER).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            goqu.L("$1"),
		"r.role_id":              role.DefinitionProjectAdmin.ID,
		"r.subject_namespace_id": namespace.DefinitionUser.ID,
		"r.object_namespace_id":  namespace.DefinitionProject.ID,
	}).ToSQL()

	return listProjectAdminsQuery, err
}

func buildListProjectQuery(dialect goqu.DialectWrapper) (string, error) {
	listProjectQuery, _, err := dialect.From(TABLE_PROJECTS).ToSQL()

	return listProjectQuery, err
}

// *Update Project Query
func buildUpdateProjectBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	updateProjectQuery, _, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"slug":       goqu.L("$3"),
			"org_id":     goqu.L("$4"),
			"metadata":   goqu.L("$5"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).Returning(&Project{}).ToSQL()

	return updateProjectQuery, err
}

func buildUpdateProjectByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	updateProjectQuery, _, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"name":       goqu.L("$3"),
			"slug":       goqu.L("$4"),
			"org_id":     goqu.L("$5"),
			"metadata":   goqu.L("$6"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).Returning(&Project{}).ToSQL()

	return updateProjectQuery, err
}

func transformToProject(from Project) (project.Project, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return project.Project{}, err
	}

	return project.Project{
		ID:           from.ID,
		Name:         from.Name,
		Slug:         from.Slug,
		Organization: organization.Organization{ID: from.OrgID},
		Metadata:     unmarshalledMetadata,
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}, nil
}
