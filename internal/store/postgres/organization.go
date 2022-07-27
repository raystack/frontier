package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/role"
)

type Organization struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	Slug      string       `db:"slug"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// *Get Organizations Query
func buildGetOrganizationsBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	getOrganizationsBySlugQuery, _, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).ToSQL()

	return getOrganizationsBySlugQuery, err
}

func buildGetOrganizationsByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	getOrganizationsByIDQuery, _, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).ToSQL()
	return getOrganizationsByIDQuery, err
}

// *Create Organization Query
func buildCreateOrganizationQuery(dialect goqu.DialectWrapper) (string, error) {
	createOrganizationQuery, _, err := dialect.Insert(TABLE_ORGANIZATIONS).Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"slug":     goqu.L("$2"),
			"metadata": goqu.L("$3"),
		}).Returning(&Organization{}).ToSQL()

	return createOrganizationQuery, err
}

// *List Organization Query
func buildListOrganizationsQuery(dialect goqu.DialectWrapper) (string, error) {
	listOrganizationsQuery, _, err := dialect.From(TABLE_ORGANIZATIONS).ToSQL()

	return listOrganizationsQuery, err
}

func buildListOrganizationAdmins(dialect goqu.DialectWrapper) (string, error) {
	listOrganizationAdmins, _, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).From(goqu.T(TABLE_RELATIONS).As("r")).
		Join(goqu.T(TABLE_USERS).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            goqu.L("$1"),
		"r.role_id":              role.DefinitionOrganizationAdmin.ID,
		"r.subject_namespace_id": namespace.DefinitionUser.ID,
		"r.object_namespace_id":  namespace.DefinitionOrg.ID,
	}).ToSQL()

	return listOrganizationAdmins, err
}

// *Update Organization Query
func buildUpdateOrganizationBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	updateOrganizationQuery, _, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"slug":       goqu.L("$3"),
			"metadata":   goqu.L("$4"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).Returning(&Organization{}).ToSQL()

	return updateOrganizationQuery, err
}

func buildUpdateOrganizationByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	updateOrganizationQuery, _, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"name":       goqu.L("$3"),
			"slug":       goqu.L("$4"),
			"metadata":   goqu.L("$5"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.ExOr{
		"slug": goqu.L("$1"),
		"id":   goqu.L("$2"),
	}).Returning(&Organization{}).ToSQL()

	return updateOrganizationQuery, err
}

func transformToOrg(from Organization) (organization.Organization, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return organization.Organization{}, err
	}

	return organization.Organization{
		ID:        from.ID,
		Name:      from.Name,
		Slug:      from.Slug,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
