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
	"github.com/raystack/frontier/billing/feature"
	"github.com/raystack/frontier/pkg/db"
)

type Feature struct {
	ID          string         `db:"id"`
	ProviderID  string         `db:"provider_id"`
	PlanIDs     pq.StringArray `db:"plan_ids"`
	Name        string         `db:"name"`
	Title       *string        `db:"title"`
	Description *string        `db:"description"`

	Behavior     string             `db:"behavior"`
	CreditAmount int64              `db:"credit_amount"`
	State        string             `db:"state"`
	Metadata     types.NullJSONText `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (f Feature) transform() (feature.Feature, error) {
	var unmarshalledMetadata map[string]any
	if f.Metadata.Valid {
		if err := f.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return feature.Feature{}, err
		}
	}
	featureTitle := ""
	if f.Title != nil {
		featureTitle = *f.Title
	}
	featureDescription := ""
	if f.Description != nil {
		featureDescription = *f.Description
	}
	return feature.Feature{
		ID:           f.ID,
		ProviderID:   f.ProviderID,
		PlanIDs:      f.PlanIDs,
		Name:         f.Name,
		Title:        featureTitle,
		Description:  featureDescription,
		State:        f.State,
		CreditAmount: f.CreditAmount,
		Behavior:     feature.Behavior(f.Behavior),
		Metadata:     unmarshalledMetadata,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		DeletedAt:    f.DeletedAt,
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

func (r BillingFeatureRepository) Create(ctx context.Context, toCreate feature.Feature) (feature.Feature, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return feature.Feature{}, err
	}
	if toCreate.ProviderID == "" {
		toCreate.ProviderID = toCreate.ID
	}

	query, params, err := dialect.Insert(TABLE_BILLING_FEATURES).Rows(
		goqu.Record{
			"id":            toCreate.ID,
			"provider_id":   toCreate.ProviderID,
			"plan_ids":      pq.StringArray(toCreate.PlanIDs),
			"name":          toCreate.Name,
			"title":         toCreate.Title,
			"description":   toCreate.Description,
			"state":         toCreate.State,
			"credit_amount": toCreate.CreditAmount,
			"behavior":      toCreate.Behavior,
			"metadata":      marshaledMetadata,
		}).Returning(&Feature{}).ToSQL()
	if err != nil {
		return feature.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		return feature.Feature{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return featureModel.transform()
}

func (r BillingFeatureRepository) GetByID(ctx context.Context, id string) (feature.Feature, error) {
	stmt := dialect.Select().From(TABLE_BILLING_FEATURES).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return feature.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return feature.Feature{}, feature.ErrFeatureNotFound
		}
		return feature.Feature{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return featureModel.transform()
}

func (r BillingFeatureRepository) GetByName(ctx context.Context, name string) (feature.Feature, error) {
	stmt := dialect.Select().From(TABLE_BILLING_FEATURES).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return feature.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return feature.Feature{}, feature.ErrFeatureNotFound
		}
		return feature.Feature{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return featureModel.transform()
}

func (r BillingFeatureRepository) List(ctx context.Context, flt feature.Filter) ([]feature.Feature, error) {
	stmt := dialect.Select().From(TABLE_BILLING_FEATURES)
	if flt.PlanID != "" {
		stmt = stmt.Where(goqu.L("plan_ids @> ?", pq.StringArray{flt.PlanID}))
	}
	if len(flt.FeatureIDs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"id": goqu.Op{"in": flt.FeatureIDs},
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var featureModels []Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "GetByPlanID", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &featureModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var features []feature.Feature
	for _, featureModel := range featureModels {
		feature, err := featureModel.transform()
		if err != nil {
			return nil, err
		}
		features = append(features, feature)
	}
	return features, nil
}

func (r BillingFeatureRepository) UpdateByName(ctx context.Context, toUpdate feature.Feature) (feature.Feature, error) {
	if strings.TrimSpace(toUpdate.Name) == "" {
		return feature.Feature{}, feature.ErrInvalidDetail
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return feature.Feature{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	updateRecord := goqu.Record{
		"title":       toUpdate.Title,
		"description": toUpdate.Description,
		"metadata":    marshaledMetadata,
		"updated_at":  goqu.L("now()"),
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}
	if len(toUpdate.PlanIDs) > 0 {
		updateRecord["plan_ids"] = pq.StringArray(toUpdate.PlanIDs)
	}
	query, params, err := dialect.Update(TABLE_BILLING_FEATURES).Set(updateRecord).Where(goqu.Ex{
		"name": toUpdate.Name,
	}).Returning(&Feature{}).ToSQL()
	if err != nil {
		return feature.Feature{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var featureModel Feature
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_FEATURES, "UpdateByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&featureModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return feature.Feature{}, feature.ErrFeatureNotFound
		default:
			return feature.Feature{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return featureModel.transform()
}
