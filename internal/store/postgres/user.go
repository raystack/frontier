package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/user"
)

type User struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	Email     string       `db:"email"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (from User) transformToUser() (user.User, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return user.User{}, err
	}

	return user.User{
		ID:        from.ID,
		Name:      from.Name,
		Email:     from.Email,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}

func buildListUserGroupsQuery(dialect goqu.DialectWrapper) (string, error) {
	listUserGroupsQuery, _, err := dialect.Select(
		goqu.I("g.id").As("id"),
		goqu.I("g.metadata").As("metadata"),
		goqu.I("g.name").As("name"),
		goqu.I("g.slug").As("slug"),
		goqu.I("g.updated_at").As("updated_at"),
		goqu.I("g.created_at").As("created_at"),
		goqu.I("g.org_id").As("org_id"),
	).From(goqu.L("relations r")).
		Join(goqu.L("groups g"), goqu.On(
			goqu.I("g.id").Cast("VARCHAR").
				Eq(goqu.I("r.object_id")),
		)).Where(goqu.Ex{
		"r.object_namespace_id": namespace.DefinitionTeam.ID,
		"subject_namespace_id":  namespace.DefinitionUser.ID,
		"subject_id":            goqu.L("$1"),
		"role_id":               goqu.L("$2"),
	}).ToSQL()

	return listUserGroupsQuery, err
}
