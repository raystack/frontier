package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
)

type ActionRepository struct {
	dbc *db.Client
}

func NewActionRepository(dbc *db.Client) *ActionRepository {
	return &ActionRepository{
		dbc: dbc,
	}
}

func (r ActionRepository) Get(ctx context.Context, id string) (action.Action, error) {
	if strings.TrimSpace(id) == "" {
		return action.Action{}, action.ErrInvalidID
	}

	var fetchedAction Action
	query, params, err := dialect.Select(&returnedActionColumns{}).From(TABLE_ACTIONS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedAction, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return action.Action{}, action.ErrNotExist
		}
		return action.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return fetchedAction.transformToAction(), nil
}

// TODO this is actually an upsert
func (r ActionRepository) Create(ctx context.Context, act action.Action) (action.Action, error) {
	if strings.TrimSpace(act.ID) == "" {
		return action.Action{}, action.ErrInvalidID
	}

	nsID := str.DefaultStringIfEmpty(act.Namespace.ID, act.NamespaceID)
	query, params, err := dialect.Insert(TABLE_ACTIONS).Rows(
		goqu.Record{
			"id":           act.ID,
			"name":         act.Name,
			"namespace_id": nsID,
		}).OnConflict(
		goqu.DoUpdate("id", goqu.Record{
			"name": act.Name,
		})).Returning(&returnedActionColumns{}).ToSQL()
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var actionModel Action
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&actionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errForeignKeyViolation):
			return action.Action{}, namespace.ErrNotExist
		default:
			return action.Action{}, err
		}
	}

	return actionModel.transformToAction(), nil
}

func (r ActionRepository) List(ctx context.Context) ([]action.Action, error) {
	var fetchedActions []Action
	query, params, err := dialect.Select(&returnedActionColumns{}).From(TABLE_ACTIONS).ToSQL()
	if err != nil {
		return []action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedActions, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []action.Action{}, nil
		}
		return []action.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedActions []action.Action
	for _, o := range fetchedActions {
		transformedActions = append(transformedActions, o.transformToAction())
	}

	return transformedActions, nil
}

func (r ActionRepository) Update(ctx context.Context, act action.Action) (action.Action, error) {
	if strings.TrimSpace(act.ID) == "" {
		return action.Action{}, action.ErrInvalidID
	}

	if strings.TrimSpace(act.Name) == "" {
		return action.Action{}, action.ErrInvalidDetail
	}

	query, params, err := dialect.Update(TABLE_ACTIONS).Set(
		goqu.Record{
			"name":         act.Name,
			"namespace_id": act.NamespaceID,
			"updated_at":   goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": act.ID,
	}).Returning(&returnedActionColumns{}).ToSQL()
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var actionModel Action
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&actionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return action.Action{}, action.ErrNotExist
		case errors.Is(err, errForeignKeyViolation):
			return action.Action{}, namespace.ErrNotExist
		default:
			return action.Action{}, err
		}
	}

	return actionModel.transformToAction(), nil
}
