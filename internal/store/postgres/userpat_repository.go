package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"

	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/pkg/db"
)

type UserPATRepository struct {
	dbc *db.Client
}

func NewUserPATRepository(dbc *db.Client) *UserPATRepository {
	return &UserPATRepository{
		dbc: dbc,
	}
}

func (r UserPATRepository) Create(ctx context.Context, pat models.PAT) (models.PAT, error) {
	if strings.TrimSpace(pat.ID) == "" {
		pat.ID = uuid.New().String()
	}

	marshaledMetadata, err := json.Marshal(pat.Metadata)
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var model UserPAT
	query, params, err := dialect.Insert(TABLE_USER_PATS).Rows(
		goqu.Record{
			"id":          pat.ID,
			"user_id":     pat.UserID,
			"org_id":      pat.OrgID,
			"title":       pat.Title,
			"secret_hash": pat.SecretHash,
			"metadata":    marshaledMetadata,
			"expires_at":  pat.ExpiresAt,
		}).Returning(&UserPAT{}).ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&model)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, ErrDuplicateKey) {
			return models.PAT{}, paterrors.ErrConflict
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) CountActive(ctx context.Context, userID, orgID string) (int64, error) {
	now := time.Now()
	query, params, err := dialect.Select(goqu.COUNT("*")).From(TABLE_USER_PATS).Where(
		goqu.Ex{"user_id": userID},
		goqu.Ex{"org_id": orgID},
		goqu.Ex{"deleted_at": nil},
		goqu.C("expires_at").Gt(now),
	).ToSQL()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", queryErr, err)
	}

	var count int64
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "CountActive", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &count, query, params...)
	}); err != nil {
		return 0, fmt.Errorf("%w: %w", dbErr, err)
	}

	return count, nil
}

func (r UserPATRepository) GetByID(ctx context.Context, id string) (models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).
		Select(&UserPAT{}).
		Where(
			goqu.Ex{"id": id},
			goqu.Ex{"deleted_at": nil},
		).Limit(1).ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var model UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &model, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PAT{}, paterrors.ErrNotFound
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) List(ctx context.Context, userID, orgID string) ([]models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).
		Select(&UserPAT{}).
		Where(
			goqu.Ex{"user_id": userID},
			goqu.Ex{"org_id": orgID},
			goqu.Ex{"deleted_at": nil},
		).Order(goqu.C("created_at").Desc()).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", queryErr, err)
	}

	var rows []UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &rows, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", dbErr, err)
	}

	pats := make([]models.PAT, 0, len(rows))
	for _, row := range rows {
		pat, err := row.transform()
		if err != nil {
			return nil, err
		}
		pats = append(pats, pat)
	}
	return pats, nil
}

func (r UserPATRepository) GetByID(ctx context.Context, id string) (models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).
		Select(&UserPAT{}).
		Where(
			goqu.Ex{"id": id},
			goqu.Ex{"deleted_at": nil},
		).Limit(1).ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var model UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &model, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PAT{}, paterrors.ErrNotFound
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) GetBySecretHash(ctx context.Context, secretHash string) (models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).
		Select(&UserPAT{}).
		Where(
			goqu.Ex{"secret_hash": secretHash},
			goqu.Ex{"deleted_at": nil},
		).Limit(1).ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var model UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "GetBySecretHash", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &model, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PAT{}, paterrors.ErrNotFound
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) UpdateLastUsedAt(ctx context.Context, id string, at time.Time) error {
	query, params, err := dialect.Update(TABLE_USER_PATS).
		Set(goqu.Record{"last_used_at": at}).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "UpdateLastUsedAt", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		return err
	}); err != nil {
		return fmt.Errorf("%w: %w", dbErr, err)
	}

	return nil
}
