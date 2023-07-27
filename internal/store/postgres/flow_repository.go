package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
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

func (s *FlowRepository) Set(ctx context.Context, flow *authenticate.Flow) error {
	if flow.ID == uuid.Nil {
		flow.ID = uuid.New()
	}
	if flow.Metadata == nil {
		flow.Metadata = make(map[string]any)
	}
	flow.Metadata["start_url"] = flow.StartURL
	flow.Metadata["finish_url"] = flow.FinishURL
	marshaledMetadata, err := json.Marshal(flow.Metadata)
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_FLOWS).Rows(
		goqu.Record{
			"id":         flow.ID,
			"method":     flow.Method,
			"email":      flow.Email,
			"nonce":      flow.Nonce,
			"metadata":   marshaledMetadata,
			"created_at": flow.CreatedAt,
			"expires_at": flow.ExpiresAt,
		}).OnConflict(goqu.DoUpdate("id", goqu.Record{
		"method":     flow.Method,
		"email":      flow.Email,
		"nonce":      flow.Nonce,
		"metadata":   marshaledMetadata,
		"expires_at": flow.ExpiresAt,
	})).Returning(&Flow{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var flowModel Flow
	if err = s.dbc.WithTimeout(ctx, TABLE_FLOWS, "Set", func(ctx context.Context) error {
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

	if err = s.dbc.WithTimeout(ctx, TABLE_FLOWS, "Get", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&flowModel)
	}); err != nil {
		err = checkPostgresError(err)
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	return flowModel.transformToFlow()
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

	return s.dbc.WithTimeout(ctx, TABLE_FLOWS,
		"Delete", func(ctx context.Context) error {
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
	query, params, err := dialect.Delete(TABLE_FLOWS).
		Where(
			goqu.Ex{
				"expires_at": goqu.Op{"lte": s.Now()},
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_FLOWS, "DeleteExpiredFlows", func(ctx context.Context) error {
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
