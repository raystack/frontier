package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	// converting the userID string to uuid
	userUUID, err := uuid.Parse(session.UserID)
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	query, params, err := dialect.Insert(TABLE_SESSIONS).Rows(
		goqu.Record{
			"id":               session.ID,
			"user_id":          userUUID,
			"authenticated_at": session.CreatedAt,
			"expires_at":       session.ExpiresAt,
			"created_at":       session.CreatedAt,
		}).Returning(&authenticate.Session{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var sessionModel authenticate.Session
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
	var session authenticate.Session
	query, params, err := dialect.From(TABLE_SESSIONS).Where(
		goqu.Ex{
			"id": id,
		}).ToSQL()
	if err != nil {
		return &authenticate.Session{}, fmt.Errorf("%w: %s", queryErr, err)
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
			return &authenticate.Session{}, fmt.Errorf("%w: %s", dbErr, authenticate.ErrNoSession)
		default:
			return &authenticate.Session{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	return &session, nil
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
