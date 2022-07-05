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
	"github.com/odpf/shield/pkg/utils"

	newrelic "github.com/newrelic/go-agent"
)

type Action struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	Namespace   Namespace `db:"namespace"`
	NamespaceID string    `db:"namespace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

const (
	getActionQuery    = `SELECT id, name, namespace_id, created_at, updated_at from actions where id=$1;`
	createActionQuery = `INSERT INTO actions(id, name, namespace_id)
		values($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET name=$2
		RETURNING id, name, namespace_id, created_at, updated_at;`
	listActionsQuery  = `SELECT id, name, namespace_id, created_at, updated_at from actions;`
	updateActionQuery = `UPDATE actions set name = $2, namespace_id = $3, updated_at = now() where id = $1 RETURNING id, name, namespace_id, created_at, updated_at;`
)

func (s Store) GetAction(ctx context.Context, id string) (model.Action, error) {
	fetchedAction, err := s.selectAction(ctx, id, nil)
	return fetchedAction, err
}

func (s Store) selectAction(ctx context.Context, id string, txn *sqlx.Tx) (model.Action, error) {
	var fetchedAction Action

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("actions"),
			Operation:  "Get Action",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

		return s.DB.GetContext(ctx, &fetchedAction, getActionQuery, id)
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

	nsId := utils.DefaultStringIfEmpty(actionToCreate.Namespace.Id, actionToCreate.NamespaceId)
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("actions"),
			Operation:  "Create Action",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

		return s.DB.GetContext(ctx, &newAction, createActionQuery, actionToCreate.Id, actionToCreate.Name, nsId)
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
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("actions"),
			Operation:  "List Actions",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

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

func (s Store) UpdateAction(ctx context.Context, toUpdate model.Action) (model.Action, error) {
	var updatedAction Action

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("actions"),
			Operation:  "Update Action",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

		return s.DB.GetContext(ctx, &updatedAction, updateActionQuery, toUpdate.Id, toUpdate.Name, toUpdate.NamespaceId)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Action{}, schema.ActionDoesntExist
	} else if err != nil {
		return model.Action{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedAction, err := transformToAction(updatedAction)
	if err != nil {
		return model.Action{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedAction, nil
}

func transformToAction(from Action) (model.Action, error) {
	from.Namespace.Id = from.NamespaceID
	namespace, err := transformToNamespace(from.Namespace)
	if err != nil {
		return model.Action{}, err
	}

	return model.Action{
		Id:          from.Id,
		Name:        from.Name,
		NamespaceId: from.NamespaceID,
		Namespace:   namespace,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
