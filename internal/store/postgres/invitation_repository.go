package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/raystack/frontier/pkg/auditrecord"

	"github.com/raystack/frontier/core/invitation"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
)

type InvitationRepository struct {
	log log.Logger
	dbc *db.Client
	Now func() time.Time
}

func NewInvitationRepository(logger log.Logger, dbc *db.Client) *InvitationRepository {
	return &InvitationRepository{
		dbc: dbc,
		log: logger,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *InvitationRepository) Set(ctx context.Context, invite invitation.Invitation) (invitation.Invitation, error) {
	if invite.ID == uuid.Nil {
		return invitation.Invitation{}, ErrInvalidID
	}
	if invite.Metadata == nil {
		invite.Metadata = make(map[string]any)
	}
	invite.Metadata["group_ids"] = invite.GroupIDs
	invite.Metadata["role_ids"] = invite.RoleIDs
	marshaledMetadata, err := json.Marshal(invite.Metadata)
	if err != nil {
		return invitation.Invitation{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	invite.CreatedAt = s.Now()
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = s.Now().Add(invitation.DefaultExpiryDuration)
	}

	// Struct to hold invitation data with additional context
	type invitationWithContext struct {
		Invitation
		OrgName string `db:"org_name"`
	}

	orgNameSubquery := dialect.From(TABLE_ORGANIZATIONS).
		Select("name").
		Where(goqu.Ex{"id": invite.OrgID})

	query, params, err := dialect.Insert(TABLE_INVITATIONS).Rows(
		goqu.Record{
			"id":         invite.ID,
			"user_id":    strings.ToLower(invite.UserEmailID),
			"org_id":     invite.OrgID,
			"metadata":   marshaledMetadata,
			"created_at": invite.CreatedAt,
			"expires_at": invite.ExpiresAt,
		}).OnConflict(goqu.DoUpdate("id", goqu.Record{
		"user_id":  strings.ToLower(invite.UserEmailID),
		"org_id":   invite.OrgID,
		"metadata": marshaledMetadata,
	})).Returning(
		goqu.I(TABLE_INVITATIONS+".*"),
		orgNameSubquery.As("org_name"),
	).ToSQL()
	if err != nil {
		return invitation.Invitation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var result invitationWithContext
	if err = s.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return s.dbc.WithTimeout(ctx, TABLE_INVITATIONS, "Set", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&result); err != nil {
				return err
			}

			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.OrganizationInvitedEvent,
				AuditResource{
					ID:   result.OrgID,
					Type: auditrecord.OrganizationType,
					Name: result.OrgName,
				},
				&AuditTarget{
					ID:   result.ID.String(),
					Type: auditrecord.InvitationType,
					Metadata: map[string]interface{}{
						"email":     invite.UserEmailID,
						"group_ids": invite.GroupIDs,
						"role_ids":  invite.RoleIDs,
					},
				},
				result.OrgID,
				nil,
				result.CreatedAt,
			)

			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		return invitation.Invitation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return result.Invitation.transformToInvitation()
}

func (s *InvitationRepository) Get(ctx context.Context, id uuid.UUID) (invitation.Invitation, error) {
	var inviteModel Invitation
	query, params, err := dialect.From(TABLE_INVITATIONS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return invitation.Invitation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_INVITATIONS, "Get", func(ctx context.Context) error {
		return s.dbc.QueryRowxContext(ctx, query, params...).StructScan(&inviteModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return invitation.Invitation{}, invitation.ErrNotFound
		}
		return invitation.Invitation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return inviteModel.transformToInvitation()
}

func (s *InvitationRepository) List(ctx context.Context, flt invitation.Filter) ([]invitation.Invitation, error) {
	var fetchedInvitations []Invitation
	stmt := dialect.From(TABLE_INVITATIONS)
	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	}
	if flt.UserID != "" {
		stmt = stmt.Where(goqu.Ex{
			"user_id": strings.ToLower(flt.UserID),
		})
	}

	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_INVITATIONS, "List", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &fetchedInvitations, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedInvitations []invitation.Invitation
	for _, o := range fetchedInvitations {
		transPerm, err := o.transformToInvitation()
		if err != nil {
			return []invitation.Invitation{}, fmt.Errorf("failed to transform invitation model: %w", err)
		}
		transformedInvitations = append(transformedInvitations, transPerm)
	}

	return transformedInvitations, nil
}

func (s *InvitationRepository) ListByUser(ctx context.Context, id string) ([]invitation.Invitation, error) {
	var fetchedInvitations []Invitation
	query, params, err := dialect.From(TABLE_INVITATIONS).Where(
		goqu.Ex{
			"user_id": strings.ToLower(id),
		},
	).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = s.dbc.WithTimeout(ctx, TABLE_INVITATIONS, "ListByUser", func(ctx context.Context) error {
		return s.dbc.SelectContext(ctx, &fetchedInvitations, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedInvitations []invitation.Invitation
	for _, o := range fetchedInvitations {
		transPerm, err := o.transformToInvitation()
		if err != nil {
			return nil, fmt.Errorf("failed to transform invitation model: %w", err)
		}
		transformedInvitations = append(transformedInvitations, transPerm)
	}

	return transformedInvitations, nil
}

func (s *InvitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, params, err := dialect.Delete(TABLE_INVITATIONS).
		Where(
			goqu.Ex{
				"id": id,
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_INVITATIONS, "Delete", func(ctx context.Context) error {
		_, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		return nil
	})
}

// TODO(kushsharma): schedue a cron for it
func (s *InvitationRepository) GarbageCollect(ctx context.Context) error {
	query, params, err := dialect.Delete(TABLE_INVITATIONS).
		Where(
			goqu.Ex{
				"expires_at": goqu.Op{"lte": s.Now()},
			},
		).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return s.dbc.WithTimeout(ctx, TABLE_INVITATIONS, "GarbageCollect", func(ctx context.Context) error {
		result, err := s.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		count, _ := result.RowsAffected()
		s.log.Debug("deleted expired invitation", "expired_invitations_count", count)
		return nil
	})
}
