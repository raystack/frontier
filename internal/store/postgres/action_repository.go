package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/odpf/shield/pkg/db"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/pkg/str"
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
	var fetchedAction Action
	getActionQuery, err := buildGetActionQuery(dialect)
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedAction, getActionQuery, id)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return action.Action{}, action.ErrNotExist
		}
		return action.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedAction, err := transformToAction(fetchedAction)
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedAction, nil
}

// TODO this is actually an upsert
func (r ActionRepository) Create(ctx context.Context, actionToCreate action.Action) (action.Action, error) {
	if actionToCreate.ID == "" {
		return action.Action{}, action.ErrInvalidID
	}

	nsID := str.DefaultStringIfEmpty(actionToCreate.Namespace.ID, actionToCreate.NamespaceID)
	createActionQuery, err := buildCreateActionQuery(dialect)
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var actionModel Action
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, createActionQuery, actionToCreate.ID, actionToCreate.Name, nsID).StructScan(&actionModel)
	}); err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedAction, err := transformToAction(actionModel)
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedAction, nil
}

func (r ActionRepository) List(ctx context.Context) ([]action.Action, error) {
	var fetchedActions []Action
	listActionsQuery, err := buildListActionsQuery(dialect)
	if err != nil {
		return []action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedActions, listActionsQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []action.Action{}, nil
		}
		return []action.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedActions []action.Action
	for _, o := range fetchedActions {
		transformedAction, err := transformToAction(o)
		if err != nil {
			return []action.Action{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedActions = append(transformedActions, transformedAction)
	}

	return transformedActions, nil
}

func (r ActionRepository) Update(ctx context.Context, toUpdate action.Action) (action.Action, error) {
	if toUpdate.ID == "" {
		return action.Action{}, action.ErrInvalidID
	}

	updateActionQuery, err := buildUpdateActionQuery(dialect)
	if err != nil {
		return action.Action{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var actionModel Action
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, updateActionQuery, toUpdate.ID, toUpdate.Name, toUpdate.NamespaceID).StructScan(&actionModel)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return action.Action{}, action.ErrNotExist
		}
		return action.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedAction, err := transformToAction(actionModel)
	if err != nil {
		return action.Action{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedAction, nil
}
