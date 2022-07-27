package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
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

// *Get Groups Query
func buildGetGroupsBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	getGroupsBySlugQuery, _, err := dialect.From(TABLE_GROUPS).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).ToSQL()

	return getGroupsBySlugQuery, err
}

func buildGetGroupsByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	getGroupsByIDQuery, _, err := dialect.From(TABLE_GROUPS).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).ToSQL()

	return getGroupsByIDQuery, err
}

// *Create Group Query
func buildCreateGroupQuery(dialect goqu.DialectWrapper) (string, error) {
	createGroupsQuery, _, err := dialect.Insert(TABLE_GROUPS).Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"slug":     goqu.L("$2"),
			"org_id":   goqu.L("$3"),
			"metadata": goqu.L("$4"),
		}).Returning(&Group{}).ToSQL()
	return createGroupsQuery, err
}

// *List Groups Query
func buildListGroupsQuery(dialect goqu.DialectWrapper) (string, error) {
	listGroupsQuery, _, err := dialect.From(TABLE_GROUPS).ToSQL()

	return listGroupsQuery, err
}

func buildListGroupUsersQuery(dialect goqu.DialectWrapper) (string, error) {
	listGroupUsersQuery, _, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).From(goqu.T(TABLE_RELATIONS).As("r")).
		Join(goqu.T(TABLE_USERS).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").
				Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            goqu.L("$1"),
		"r.role_id":              goqu.L("$2"),
		"r.subject_namespace_id": namespace.DefinitionUser.ID,
		"r.object_namespace_id":  namespace.DefinitionTeam.ID,
	}).ToSQL()

	return listGroupUsersQuery, err
}

func buildListUserGroupRelationsQuery(dialect goqu.DialectWrapper) (string, error) {
	listUserGroupRelationsQuery, _, err := dialect.From(TABLE_RELATIONS).Where(goqu.Ex{
		"subject_namespace_id": goqu.L(namespace.DefinitionUser.ID),
		"object_namespace_id":  goqu.L(namespace.DefinitionTeam.ID),
		"subject_id":           goqu.L("$1"),
		"object_id":            goqu.L("$2"),
	}).ToSQL()

	return listUserGroupRelationsQuery, err
}

// *Update Group Query
func buildUpdateGroupBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	updateGroupQuery, _, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"slug":       goqu.L("$3"),
			"org_id":     goqu.L("$4"),
			"metadata":   goqu.L("$5"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).Returning(&Group{}).ToSQL()

	return updateGroupQuery, err
}

func buildUpdateGroupByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	updateGroupQuery, _, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"name":       goqu.L("$3"),
			"slug":       goqu.L("$4"),
			"org_id":     goqu.L("$5"),
			"metadata":   goqu.L("$6"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).Returning(&Group{}).ToSQL()

	return updateGroupQuery, err
}

func transformToGroup(from Group) (group.Group, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return group.Group{}, err
	}

	return group.Group{
		ID:             from.ID,
		Name:           from.Name,
		Slug:           from.Slug,
		Organization:   organization.Organization{ID: from.OrgID},
		OrganizationID: from.OrgID,
		Metadata:       unmarshalledMetadata,
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}, nil
}
