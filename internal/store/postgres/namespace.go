package postgres

import (
	"time"

	"database/sql"

	"github.com/odpf/shield/core/namespace"
)

type Namespace struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// func buildGetNamespaceQuery(dialect goqu.DialectWrapper) (string, error) {
// 	getNamespaceQuery, _, err := dialect.Select(&Namespace{}).From(TABLE_NAMESPACES).Where(goqu.Ex{
// 		"id": goqu.L("$1"),
// 	}).ToSQL()

// 	return getNamespaceQuery, err
// }
// func buildCreateNamespaceQuery(dialect goqu.DialectWrapper) (string, error) {
// 	createNamespaceQuery, _, err := dialect.Insert(TABLE_NAMESPACES).Rows(
// 		goqu.Record{
// 			"id":   goqu.L("$1"),
// 			"name": goqu.L("$2"),
// 		}).OnConflict(goqu.DoUpdate("id", goqu.Record{
// 		"name": goqu.L("$2"),
// 	})).Returning(&Namespace{}).ToSQL()

// 	return createNamespaceQuery, err
// }
// func buildListNamespacesQuery(dialect goqu.DialectWrapper) (string, error) {
// 	listNamespacesQuery, _, err := dialect.Select(&Namespace{}).From(TABLE_NAMESPACES).ToSQL()

// 	return listNamespacesQuery, err
// }
// func buildUpdateNamespaceQuery(dialect goqu.DialectWrapper) (string, error) {
// 	updateNamespaceQuery, _, err := dialect.Update(TABLE_NAMESPACES).Set(
// 		goqu.Record{
// 			"id":         goqu.L("$2"),
// 			"name":       goqu.L("$3"),
// 			"updated_at": goqu.L("now()"),
// 		}).Where(goqu.Ex{
// 		"id": goqu.L("$1"),
// 	}).Returning(&Namespace{}).ToSQL()

// 	return updateNamespaceQuery, err
// }

func transformToNamespace(from Namespace) (namespace.Namespace, error) {
	return namespace.Namespace{
		ID:        from.ID,
		Name:      from.Name,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
