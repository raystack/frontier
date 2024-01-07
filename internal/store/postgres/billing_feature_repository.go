package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/db"
)

type Feature struct {
	ID         string         `db:"id"`
	Name       string         `db:"name"`
	ProductIDs pq.StringArray `db:"product_ids"`

	Metadata types.NullJSONText `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (p Feature) transform() (product.Feature, error) {
	var unmarshalledMetadata map[string]any
	if p.Metadata.Valid {
		if err := p.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return product.Feature{}, err
		}
	}
	return product.Feature{
		ID:         p.ID,
		Name:       p.Name,
		ProductIDs: p.ProductIDs,
		Metadata:   unmarshalledMetadata,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		DeletedAt:  p.DeletedAt,
	}, nil
}

type BillingFeatureRepository struct {
	dbc *db.Client
}

func NewBillingFeatureRepository(dbc *db.Client) *BillingFeatureRepository {
	return &BillingFeatureRepository{
		dbc: dbc,
	}
}

func (r BillingFeatureRepository) Create(ctx context.Context, toCreate product.Feature) (product.Feature, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return product.Feature{}, err
	}

	query, params, err := dialect.Insert(TABLE_BILLING_FEATURES).Rows(
		goqu.Record{
			"id":          toCreate.ID,
			"product_ids": pq.StringArray(toCreate.ProductIDs),
			"name":        toCreate.Name,
			"metadata":    marshaledMetadata,
			"created_at":  goqu.L("now()"),
			"updated_at":  goqu.L("now()"),
		}).Returning(&Feature{}).ToSQL()
	if err != nil {
		return product.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		return product.Feature{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return featureModel.transform()
}

func (r BillingFeatureRepository) GetByID(ctx context.Context, id string) (product.Feature, error) {
	stmt := dialect.Select().From(TABLE_BILLING_FEATURES).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return product.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Feature{}, product.ErrFeatureNotFound
		}
		return product.Feature{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return featureModel.transform()
}

func (r BillingFeatureRepository) GetByName(ctx context.Context, name string) (product.Feature, error) {
	stmt := dialect.Select().From(TABLE_BILLING_FEATURES).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return product.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Feature{}, product.ErrFeatureNotFound
		}
		return product.Feature{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return featureModel.transform()
}

func (r BillingFeatureRepository) List(ctx context.Context, flt product.Filter) ([]product.Feature, error) {
	stmt := dialect.Select().From(TABLE_BILLING_FEATURES)
	if flt.ProductID != "" {
		stmt = stmt.Where(goqu.L("product_ids @> ?", pq.StringArray{flt.ProductID}))
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModels []Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &featureModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var features []product.Feature
	for _, featureModel := range featureModels {
		feature, err := featureModel.transform()
		if err != nil {
			return nil, err
		}
		features = append(features, feature)
	}
	return features, nil
}

func (r BillingFeatureRepository) UpdateByName(ctx context.Context, toUpdate product.Feature) (product.Feature, error) {
	if strings.TrimSpace(toUpdate.Name) == "" {
		return product.Feature{}, product.ErrInvalidDetail
	}

	updateRecord := goqu.Record{
		"updated_at": goqu.L("now()"),
	}
	if toUpdate.Metadata != nil {
		marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
		if err != nil {
			return product.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		updateRecord["metadata"] = marshaledMetadata
	}
	updateRecord["product_ids"] = pq.StringArray(toUpdate.ProductIDs)

	query, params, err := dialect.Update(TABLE_BILLING_FEATURES).Set(updateRecord).Where(goqu.Ex{
		"name": toUpdate.Name,
	}).Returning(&Feature{}).ToSQL()
	if err != nil {
		return product.Feature{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "UpdateByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Feature{}, product.ErrFeatureNotFound
		default:
			return product.Feature{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return featureModel.transform()
}
