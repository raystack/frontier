package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
)

type Action struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

const (
	getActionQuery             = `SELECT id, name, slug, created_at, updated_at from actions where id=$1;`
	createActionQuery          = `INSERT INTO actions(name, slug) values($1, $2) RETURNING id, name, slug, created_at, updated_at;`
	listActionsQuery           = `SELECT id, name, slug, created_at, updated_at from actions;`
	selectActionForUpdateQuery = `SELECT id, name, slug, version, updated_at from actions where id=$1;`
	updateActionQuery          = `UPDATE actions set name = $2, slug = $3 updated_at = now() where id = $1 RETURNING id, name, slug, created_at, updated_at;`
)

func (s Store) GetAction(ctx context.Context, id string) (model.Action, error) {
	fetchedAction, err := s.selectAction(ctx, id, false, nil)
	return fetchedAction, err
}

func (s Store) selectAction(ctx context.Context, id string, forUpdate bool, txn *sqlx.Tx) (model.Action, error) {
	var fetchedAction Action

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		if forUpdate {
			return txn.GetContext(ctx, &fetchedAction, selectActionForUpdateQuery, id)
		} else {
			return s.DB.GetContext(ctx, &fetchedAction, getActionQuery, id)
		}
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Action{}, schema.ActionDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Action{}, schema.InvalidUUID
	} else if err != nil {
		return model.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedAction, err := transformToAction(fetchedAction)
	if err != nil {
		return model.Action{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedAction, nil
}

func (s Store) CreateAction(ctx context.Context, actionToCreate model.Action) (model.Action, error) {

	var newAction Action
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newAction, createActionQuery, actionToCreate.Name, actionToCreate.Slug)
	})

	if err != nil {
		return model.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedAction, err := transformToAction(newAction)
	if err != nil {
		return model.Action{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedAction, nil
}

func (s Store) ListActions(ctx context.Context) ([]model.Action, error) {
	var fetchedActions []Action
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedActions, listActionsQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Action{}, schema.ActionDoesntExist
	}

	if err != nil {
		return []model.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedActions []model.Action

	for _, o := range fetchedActions {
		transformedAction, err := transformToAction(o)
		if err != nil {
			return []model.Action{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedActions = append(transformedActions, transformedAction)
	}

	return transformedActions, nil
}

func transformToAction(from Action) (model.Action, error) {

	return model.Action{
		Id:        from.Id,
		Name:      from.Name,
		Slug:      from.Slug,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
