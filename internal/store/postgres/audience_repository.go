package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/audience"
	"github.com/raystack/frontier/pkg/db"
)

type AudienceRepository struct {
	dbc *db.Client
}

func NewAudienceRepository(dbc *db.Client) *AudienceRepository {
	return &AudienceRepository{
		dbc: dbc,
	}
}

func (r AudienceRepository) Create(ctx context.Context, aud audience.Audience) (audience.Audience, error) {
	marshaledMetadata, err := json.Marshal(aud.Metadata)
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	insertRow := goqu.Record{
		"name":     aud.Name,
		"email":    aud.Email,
		"phone":    aud.Phone,
		"activity": aud.Activity,
		"status":   string(aud.Status),
		"source":   aud.Source,
		"verified": aud.Verified,
		"metadata": marshaledMetadata,
	}

	createQuery, params, err := dialect.Insert(TABLE_AUDIENCES).Rows(insertRow).Returning(&Audience{}).ToSQL()
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	tx, err := r.dbc.BeginTxx(ctx, nil)
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", beginTnxErr, err)
	}

	var audienceModel Audience
	if err = r.dbc.WithTimeout(ctx, TABLE_AUDIENCES, "Create", func(ctx context.Context) error {
		return tx.QueryRowxContext(ctx, createQuery, params...).StructScan(&audienceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return audience.Audience{}, audience.ErrEmailActivityAlreadyExists
		default:
			if rbErr := tx.Rollback(); rbErr != nil {
				return audience.Audience{}, rbErr
			}
			return audience.Audience{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return audience.Audience{}, err
	}
	transformedAudience, err := audienceModel.transformToAudience()
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedAudience, nil
}

func (r AudienceRepository) List(ctx context.Context, filters audience.Filter) ([]audience.Audience, error) {
	stmt := dialect.From(TABLE_AUDIENCES)
	if filters.Email != "" {
		stmt = stmt.Where(goqu.Ex{"email": strings.ToLower(filters.Email)})
	}
	if filters.Activity != "" {
		stmt = stmt.Where(goqu.Ex{"activity": filters.Activity})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return []audience.Audience{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var audienceModel []Audience
	if err = r.dbc.WithTimeout(ctx, TABLE_AUDIENCES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &audienceModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return []audience.Audience{}, audience.ErrNotExist
		}
		return []audience.Audience{}, err
	}

	var transformedAudiences []audience.Audience
	for _, a := range audienceModel {
		transformedAudience, err := a.transformToAudience()
		if err != nil {
			return []audience.Audience{}, fmt.Errorf("%w: %w", parseErr, err)
		}
		transformedAudiences = append(transformedAudiences, transformedAudience)
	}
	return transformedAudiences, nil
}

func (r AudienceRepository) Update(ctx context.Context, aud audience.Audience) (audience.Audience, error) {
	marshaledMetadata, err := json.Marshal(aud.Metadata)
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	updateRow := goqu.Record{
		"name":     aud.Name,
		"phone":    aud.Phone,
		"activity": aud.Activity,
		"status":   string(aud.Status),
		"source":   aud.Source,
		"verified": aud.Verified,
		"metadata": marshaledMetadata,
	}
	// succeeds only when both id and email are valid and belongs to same user.
	updateQuery, params, err := dialect.Update(TABLE_AUDIENCES).
		Set(updateRow).
		Where(goqu.Ex{"id": aud.ID, "email": aud.Email}).
		Returning(&Audience{}).
		ToSQL()
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	tx, err := r.dbc.BeginTxx(ctx, nil)
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", beginTnxErr, err)
	}

	var audienceModel Audience
	if err = r.dbc.WithTimeout(ctx, TABLE_AUDIENCES, "Update",
		func(ctx context.Context) error {
			return tx.QueryRowxContext(ctx, updateQuery, params...).StructScan(&audienceModel)
		}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return audience.Audience{}, audience.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return audience.Audience{}, audience.ErrEmailActivityAlreadyExists
		default:
			if rbErr := tx.Rollback(); rbErr != nil {
				return audience.Audience{}, rbErr
			}
			return audience.Audience{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return audience.Audience{}, err
	}
	transformedAudience, err := audienceModel.transformToAudience()
	if err != nil {
		return audience.Audience{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedAudience, nil
}
