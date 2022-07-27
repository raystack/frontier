package postgres

import (
	"time"

	"github.com/odpf/shield/core/action"
)

type Action struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Namespace   Namespace `db:"namespace"`
	NamespaceID string    `db:"namespace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type returnedActionColumns struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	NamespaceID string    `db:"namespace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// func buildGetActionQuery(dialect goqu.DialectWrapper) (string, error) {
// 	getActionQuery, _, err := dialect.Select(&returnedActionColumns{}).From(TABLE_ACTIONS).Where(goqu.Ex{
// 		"id": goqu.L("$1"),
// 	}).ToSQL()

// 	return getActionQuery, err
// }

// func buildCreateActionQuery(dialect goqu.DialectWrapper) (string, error) {
// 	createActionQuery, _, err := dialect.Insert(TABLE_ACTIONS).Rows(
// 		goqu.Record{
// 			"id":           goqu.L("$1"),
// 			"name":         goqu.L("$2"),
// 			"namespace_id": goqu.L("$3"),
// 		}).OnConflict(goqu.DoUpdate("id", goqu.Record{
// 		"name": goqu.L("$2"),
// 	})).Returning(&returnedActionColumns{}).ToSQL()

// 	return createActionQuery, err
// }

// func buildListActionsQuery(dialect goqu.DialectWrapper) (string, error) {
// 	listActionsQuery, _, err := dialect.Select(&returnedActionColumns{}).From(TABLE_ACTIONS).ToSQL()

// 	return listActionsQuery, err
// }

// func buildUpdateActionQuery(dialect goqu.DialectWrapper) (string, error) {
// 	updateActionQuery, _, err := dialect.Update(TABLE_ACTIONS).Set(
// 		goqu.Record{
// 			"name":         goqu.L("$2"),
// 			"namespace_id": goqu.L("$3"),
// 			"updated_at":   goqu.L("now()"),
// 		}).Where(goqu.Ex{
// 		"id": goqu.L("$1"),
// 	}).Returning(&returnedActionColumns{}).ToSQL()

// 	return updateActionQuery, err
// }

func transformToAction(from Action) (action.Action, error) {
	from.Namespace.ID = from.NamespaceID
	namespace, err := transformToNamespace(from.Namespace)
	if err != nil {
		return action.Action{}, err
	}

	return action.Action{
		ID:          from.ID,
		Name:        from.Name,
		NamespaceID: from.NamespaceID,
		Namespace:   namespace,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
