package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/odpf/salt/log"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	newrelic "github.com/newrelic/go-agent"
	"github.com/odpf/shield/core/authenticate"
	"github.com/odpf/shield/pkg/db"
)

type SessionRepository struct {
	dbc *db.Client
}

func NewSessionRepository(dbc *db.Client) *SessionRepository {
	return &SessionRepository{
		dbc: dbc,
	}
}

func (s *SessionRepository) Set(ctx context.Context, session *authenticate.Session) error {
	query, params, err := dialect.Insert(TABLE_SESSIONS).Rows(
		goqu.Record{
			"id":               session.ID,
			"user_id":          session.UserID,
			"authenticated_at": session.CreatedAt,
			"expires_at":       session.ExpiresAt,
			"created_at":       session.CreatedAt,
		}).Returning(&Session{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var sessionModel Session
	if err = s.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_SESSIONS,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&sessionModel)
	}); err != nil {
		err = checkPostgresError(err)
		return fmt.Errorf("%w: %s", dbErr, err)
	}

	return nil
}

func (s *SessionRepository) Get(ctx context.Context, id uuid.UUID) (*authenticate.Session, error) {
	var session Session
	query, params, err := dialect.From(TABLE_SESSIONS).Where(
		goqu.Ex{
			"id": id,
		}).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err := s.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_SESSIONS,
				Operation:  "Get",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&session)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("%w: %s", dbErr, authenticate.ErrNoSession)
		default:
			return nil, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return session.transformToSession(), nil
}

func (s *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, params, err := dialect.Delete(TABLE_SESSIONS).
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
				Collection: TABLE_SESSIONS,
				Operation:  "Delete",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return fmt.Errorf("%w: %s", dbErr, authenticate.ErrNoSession)
			default:
				return fmt.Errorf("%w: %s", dbErr, err)
			}
		}

		if count, _ := result.RowsAffected(); count > 0 {
			return nil
		}

		return authenticate.ErrDeletingSession
	})
}

func (s *SessionRepository) DeleteExpiredSessions(ctx context.Context, logger log.Logger) error {
	query, params, err := dialect.Delete(TABLE_SESSIONS).
		Where(
			goqu.Ex{
				"expires_at": goqu.Op{"lt": goqu.L("NOW()")},
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
				Collection: TABLE_SESSIONS,
				Operation:  "DeleteAllExpired",
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
		logger.Info("deleted expired sessions", "expired_session_count", count)

		return nil
	})
}
