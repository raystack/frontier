package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/server/consts"

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
			"updated_at":       session.CreatedAt,
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
			"id": id,
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
			return nil, fmt.Errorf("%w: %w", dbErr, frontiersession.ErrNoSession)
		default:
			return nil, fmt.Errorf("%w: %w", dbErr, err)
		}
	}

	return session.transformToSession()
}

// sessionWithUserName extends Session to include user_name for audit logging
type sessionWithUserName struct {
	Session
	UserName string `db:"user_name"`
}

// Delete marks a session as deleted by setting deleted_at timestamp
func (s *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var result sessionWithUserName

	// Subquery to fetch user_name (title)
	userNameSubquery := dialect.From(TABLE_USERS).
		Select("title").
		Where(goqu.Ex{"id": goqu.I(TABLE_SESSIONS + ".user_id")})

	query, params, err := dialect.Update(TABLE_SESSIONS).Set(
		goqu.Record{
			"deleted_at": time.Now().UTC(),
		},
	).Where(goqu.Ex{"id": id}).
		Returning(
			goqu.I(TABLE_SESSIONS+".*"),
			userNameSubquery.As("user_name"),
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "Delete", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&result); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return frontiersession.ErrNoSession
				}
				return err
			}

			return s.createSessionRevokeAuditRecord(ctx, tx, result)
		})
	}); err != nil {
		return fmt.Errorf("%w: %w", dbErr, err)
	}

	return nil
}

// createSessionRevokeAuditRecord creates an audit record if admin is revoking someone else's session
func (s *SessionRepository) createSessionRevokeAuditRecord(ctx context.Context, tx *sqlx.Tx, session sessionWithUserName) error {
	actorID := extractActorIDFromContext(ctx)
	userID := session.UserID.String()

	// Only create audit record if actor is different from the session owner (admin revoking someone else's session)
	if actorID == "" || actorID == userID {
		return nil
	}

	deletedAt := time.Now().UTC()
	if session.DeletedAt != nil {
		deletedAt = *session.DeletedAt
	}

	auditRecord := BuildAuditRecord(
		ctx,
		pkgAuditRecord.SessionRevokedEvent,
		AuditResource{
			ID:   userID,
			Type: pkgAuditRecord.UserType,
			Name: session.UserName,
		},
		&AuditTarget{
			ID:   session.ID.String(),
			Type: pkgAuditRecord.SessionType,
		},
		"", // sessions are not org-specific, will be stored as zero uuid
		nil,
		deletedAt,
	)

	return InsertAuditRecordInTx(ctx, tx, auditRecord)
}

func extractActorIDFromContext(ctx context.Context) string {
	if val := ctx.Value(consts.AuditRecordActorContextKey); val != nil {
		if actorMap, ok := val.(map[string]interface{}); ok {
			if id, ok := actorMap["id"].(string); ok {
				return id
			}
		}
	}
	return ""
}

func (s *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	now := s.Now()

	// Delete sessions that have been expired OR soft-deleted for 24+ hours
	query, params, err := dialect.Delete(TABLE_SESSIONS).
		Where(
			goqu.Or(
				goqu.Ex{"expires_at": goqu.Op{"lte": now.Add(-24 * time.Hour)}},
				goqu.Ex{"deleted_at": goqu.Op{"lte": now.Add(-24 * time.Hour)}},
			),
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "DeleteExpiredSessions", func(ctx context.Context) error {
		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		count, _ := result.RowsAffected()
		s.log.Debug("deleted expired and old soft-deleted sessions", "deleted_count", count)

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

func (s *SessionRepository) UpdateSessionMetadata(ctx context.Context, id uuid.UUID, metadata frontiersession.SessionMetadata, updatedAt time.Time) error {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshaling session metadata: %w", err)
	}

	query, params, err := dialect.Update(TABLE_SESSIONS).Set(
		goqu.Record{
			"metadata":   metadataBytes,
			"updated_at": updatedAt,
		},
	).Where(goqu.Ex{"id": id}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}
	return s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "UpdateSessionMetadata", func(ctx context.Context) error {
		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			return fmt.Errorf("%w: %s", dbErr, err)
		}
		if count, _ := result.RowsAffected(); count == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func (s *SessionRepository) List(ctx context.Context, userID string) ([]*frontiersession.Session, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("error parsing user id: %w", err)
	}

	query, params, err := dialect.From(TABLE_SESSIONS).Where(
		goqu.Ex{
			"user_id":    uid,
			"deleted_at": nil,
		}).Order(goqu.I("created_at").Desc()).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var sessions []*Session
	if err := s.dbc.WithTimeout(ctx, TABLE_SESSIONS, "List", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &sessions, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var domainSessions []*frontiersession.Session
	for _, session := range sessions {
		domainSession, err := session.transformToSession()
		if err != nil {
			return nil, fmt.Errorf("error transforming session: %w", err)
		}
		domainSessions = append(domainSessions, domainSession)
	}

	return domainSessions, nil
}
