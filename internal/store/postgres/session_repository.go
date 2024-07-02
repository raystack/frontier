package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/salt/log"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/raystack/frontier/pkg/db"
)

type SessionRepository struct {
	log log.Logger
	dbc *db.Client
	Now func() time.Time
}

func NewSessionRepository(logger log.Logger, dbc *db.Client) *SessionRepository {
	return &SessionRepository{
		log: logger,
		dbc: dbc,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *SessionRepository) Set(ctx context.Context, session *frontiersession.Session) error {
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		return fmt.Errorf("error parsing user id: %w", err)
	}

	marshaledMetadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_SESSIONS).Rows(
		goqu.Record{
			"id":               session.ID,
			"user_id":          userID,
			"authenticated_at": session.CreatedAt,
			"expires_at":       session.ExpiresAt,
			"created_at":       session.CreatedAt,
			"metadata":         marshaledMetadata,
		}).Returning(&Session{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var sessionModel Session
	if err = s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "Upsert", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&sessionModel)
	}); err != nil {
		err = checkPostgresError(err)
		return fmt.Errorf("%w: %s", dbErr, err)
	}

	return nil
}

func (s *SessionRepository) Get(ctx context.Context, id uuid.UUID) (*frontiersession.Session, error) {
	var session Session
	query, params, err := dialect.From(TABLE_SESSIONS).Where(
		goqu.Ex{
			"id":         id,
			"deleted_at": nil,
		}).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err := s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "Get", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&session)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, fmt.Errorf("%s: %w", dbErr.Error(), frontiersession.ErrNoSession)
		default:
			return nil, fmt.Errorf("%s: %w", dbErr.Error(), err)
		}
	}

	return session.transformToSession()
}

func (s *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, params, err := dialect.Update(TABLE_SESSIONS).
		Set(goqu.Record{
			"deleted_at": time.Now().UTC(),
		}).
		Where(
			goqu.Ex{
				"id": id,
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "Update", func(ctx context.Context) error {
		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return fmt.Errorf("%w: %s", dbErr, frontiersession.ErrNoSession)
			default:
				return fmt.Errorf("%w: %s", dbErr, err)
			}
		}

		if count, _ := result.RowsAffected(); count > 0 {
			return nil
		}

		return frontiersession.ErrDeletingSession
	})
}

func (s *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	query, params, err := dialect.Delete(TABLE_SESSIONS).
		Where(
			goqu.Ex{
				"expires_at": goqu.Op{"lte": s.Now()},
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "DeleteAllExpired", func(ctx context.Context) error {
		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		count, _ := result.RowsAffected()
		s.log.Debug("deleted expired sessions", "expired_session_count", count)

		return nil
	})
}

func (s *SessionRepository) UpdateValidity(ctx context.Context, id uuid.UUID, validity time.Duration) error {
	query, params, err := dialect.Update(TABLE_SESSIONS).Set(
		goqu.Record{
			"expires_at": goqu.L("expires_at + INTERVAL '? hours'", validity.Hours()),
		}).Where(goqu.Ex{
		"id": id,
	}).ToSQL()

	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "Update", func(ctx context.Context) error {
		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		if count, _ := result.RowsAffected(); count > 0 {
			return nil
		}

		return fmt.Errorf("error updating session validity")
	})
}
