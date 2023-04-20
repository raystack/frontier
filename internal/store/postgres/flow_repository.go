package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	newrelic "github.com/newrelic/go-agent"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/authenticate"
	"github.com/odpf/shield/pkg/db"
)

type FlowRepository struct {
	log log.Logger
	dbc *db.Client
	Now func() time.Time
}

func NewFlowRepository(logger log.Logger, dbc *db.Client) *FlowRepository {
	return &FlowRepository{
		dbc: dbc,
		log: logger,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *FlowRepository) Set(ctx context.Context, session *authenticate.Flow) error {
	query, params, err := dialect.Insert(TABLE_FLOWS).Rows(
		goqu.Record{
			"id":         session.ID,
			"method":     session.Method,
			"start_url":  session.StartURL,
			"finish_url": session.FinishURL,
			"nonce":      session.Nonce,
			"created_at": session.CreatedAt,
		}).Returning(&Flow{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var flowModel Flow
	if err = s.dbc.WithTimeout(context.TODO(), func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_FLOWS,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&flowModel)
	}); err != nil {
		err = checkPostgresError(err)
		return fmt.Errorf("%w: %s", dbErr, err)
	}

	return nil
}

func (s *FlowRepository) Get(ctx context.Context, id uuid.UUID) (*authenticate.Flow, error) {
	var flowModel Flow
	query, params, err := dialect.From(TABLE_FLOWS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()

	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_FLOWS,
				Operation:  "Get",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&flowModel)
	}); err != nil {
		err = checkPostgresError(err)
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	return flowModel.transformToFlow(), nil
}

func (s *FlowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, params, err := dialect.Delete(TABLE_FLOWS).
		Where(
			goqu.Ex{
				"id": id,
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_FLOWS,
				Operation:  "Delete",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		if count, _ := result.RowsAffected(); count > 0 {
			return nil
		}

		return fmt.Errorf("no entry to delete")
	})
}

func (s *FlowRepository) DeleteExpiredFlows(ctx context.Context) error {
	expiryTime := s.Now().AddDate(0, 0, -7)
	query, params, err := dialect.Delete(TABLE_FLOWS).
		Where(
			goqu.Ex{
				"created_at": goqu.Op{"lte": expiryTime},
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_FLOWS,
				Operation:  "DeleteAll",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		count, _ := result.RowsAffected()
		s.log.Debug("deleted expired flows", "expired_flows_count", count)

		return nil
	})
}
