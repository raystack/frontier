package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	if err = r.dbc.WithTimeout(ctx, TABLE_AUDIENCES, OPERATION_CREATE, func(ctx context.Context) error {
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
		return audience.Audience{}, err
	}
	return transformedAudience, nil
}
