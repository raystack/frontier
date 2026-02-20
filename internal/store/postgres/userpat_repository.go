package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"

	"github.com/raystack/frontier/core/userpat"
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

func (r UserPATRepository) Create(ctx context.Context, pat userpat.PersonalAccessToken) (userpat.PersonalAccessToken, error) {
	if strings.TrimSpace(pat.ID) == "" {
		pat.ID = uuid.New().String()
	}

	marshaledMetadata, err := json.Marshal(pat.Metadata)
	if err != nil {
		return userpat.PersonalAccessToken{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var model UserToken
	query, params, err := dialect.Insert(TABLE_USER_TOKENS).Rows(
		goqu.Record{
			"id":          pat.ID,
			"user_id":     pat.UserID,
			"org_id":      pat.OrgID,
			"title":       pat.Title,
			"secret_hash": pat.SecretHash,
			"metadata":    marshaledMetadata,
			"expires_at":  pat.ExpiresAt,
		}).Returning(&UserToken{}).ToSQL()
	if err != nil {
		return userpat.PersonalAccessToken{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USER_TOKENS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&model)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, ErrDuplicateKey) {
			return userpat.PersonalAccessToken{}, userpat.ErrConflict
		}
		return userpat.PersonalAccessToken{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) CountActive(ctx context.Context, userID, orgID string) (int64, error) {
	now := time.Now()
	query, params, err := dialect.Select(goqu.COUNT("*")).From(TABLE_USER_TOKENS).Where(
		goqu.Ex{"user_id": userID},
		goqu.Ex{"org_id": orgID},
		goqu.Ex{"deleted_at": nil},
		goqu.C("expires_at").Gt(now),
	).ToSQL()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", queryErr, err)
	}

	var count int64
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_TOKENS, "CountActive", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &count, query, params...)
	}); err != nil {
		return 0, fmt.Errorf("%w: %w", dbErr, err)
	}

	return count, nil
}
