package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

type BehaviorConfig struct {
	CreditAmount int64 `json:"credit_amount"`
	SeatLimit    int64 `json:"seat_limit"`
}

func (b *BehaviorConfig) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, b)
	case string:
		return json.Unmarshal([]byte(src), b)
	case nil:
		return nil
	}
	return fmt.Errorf("cannot convert %T to JsonB", src)
}

func (b BehaviorConfig) Value() (driver.Value, error) {
	return json.Marshal(b)
}

type Product struct {
	ID          string         `db:"id"`
	ProviderID  string         `db:"provider_id"`
	PlanIDs     pq.StringArray `db:"plan_ids"`
	Name        string         `db:"name"`
	Title       *string        `db:"title"`
	Description *string        `db:"description"`

	Behavior string             `db:"behavior"`
	Config   BehaviorConfig     `db:"config"`
	State    string             `db:"state"`
	Metadata types.NullJSONText `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (p Product) transform() (product.Product, error) {
	var unmarshalledMetadata map[string]any
	if p.Metadata.Valid {
		if err := p.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return product.Product{}, err
		}
	}
	featureTitle := ""
	if p.Title != nil {
		featureTitle = *p.Title
	}
	featureDescription := ""
	if p.Description != nil {
		featureDescription = *p.Description
	}
	return product.Product{
		ID:          p.ID,
		ProviderID:  p.ProviderID,
		PlanIDs:     p.PlanIDs,
		Name:        p.Name,
		Title:       featureTitle,
		Description: featureDescription,
		State:       p.State,
		Config: product.BehaviorConfig{
			CreditAmount: p.Config.CreditAmount,
			SeatLimit:    p.Config.SeatLimit,
		},
		Behavior:  product.Behavior(p.Behavior),
		Metadata:  unmarshalledMetadata,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: p.DeletedAt,
	}, nil
}

type BillingProductRepository struct {
	dbc *db.Client
}

func NewBillingProductRepository(dbc *db.Client) *BillingProductRepository {
	return &BillingProductRepository{
		dbc: dbc,
	}
}

func (r BillingProductRepository) Create(ctx context.Context, toCreate product.Product) (product.Product, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return product.Product{}, err
	}
	if toCreate.ProviderID == "" {
		toCreate.ProviderID = toCreate.ID
	}

	query, params, err := dialect.Insert(TABLE_BILLING_PRODUCTS).Rows(
		goqu.Record{
			"id":          toCreate.ID,
			"provider_id": toCreate.ProviderID,
			"plan_ids":    pq.StringArray(toCreate.PlanIDs),
			"name":        toCreate.Name,
			"title":       toCreate.Title,
			"description": toCreate.Description,
			"state":       toCreate.State,
			"config":      BehaviorConfig(toCreate.Config),
			"behavior":    toCreate.Behavior,
			"metadata":    marshaledMetadata,
		}).Returning(&Product{}).Prepared(true).ToSQL()
	if err != nil {
		return product.Product{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var productModel Product
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRODUCTS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&productModel)
	}); err != nil {
		return product.Product{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return productModel.transform()
}

func (r BillingProductRepository) GetByID(ctx context.Context, id string) (product.Product, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PRODUCTS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return product.Product{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var productModel Product
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRODUCTS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&productModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Product{}, product.ErrProductNotFound
		}
		return product.Product{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return productModel.transform()
}

func (r BillingProductRepository) GetByName(ctx context.Context, name string) (product.Product, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PRODUCTS).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return product.Product{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var productModel Product
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRODUCTS, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&productModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Product{}, product.ErrProductNotFound
		}
		return product.Product{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return productModel.transform()
}

func (r BillingProductRepository) List(ctx context.Context, flt product.Filter) ([]product.Product, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PRODUCTS)
	if flt.PlanID != "" {
		stmt = stmt.Where(goqu.L("plan_ids @> ?", pq.StringArray{flt.PlanID}))
	}
	if len(flt.ProductIDs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"id": goqu.Op{"in": flt.ProductIDs},
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var productModels []Product
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRODUCTS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &productModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	features := make([]product.Product, 0, len(productModels))
	for _, productModel := range productModels {
		feature, err := productModel.transform()
		if err != nil {
			return nil, err
		}
		features = append(features, feature)
	}
	return features, nil
}

func (r BillingProductRepository) UpdateByName(ctx context.Context, toUpdate product.Product) (product.Product, error) {
	if strings.TrimSpace(toUpdate.Name) == "" {
		return product.Product{}, product.ErrInvalidDetail
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return product.Product{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	updateRecord := goqu.Record{
		"title":       toUpdate.Title,
		"description": toUpdate.Description,
		"config":      BehaviorConfig(toUpdate.Config),
		"metadata":    marshaledMetadata,
		"updated_at":  goqu.L("now()"),
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}
	if len(toUpdate.PlanIDs) > 0 {
		updateRecord["plan_ids"] = pq.StringArray(toUpdate.PlanIDs)
	}
	query, params, err := dialect.Update(TABLE_BILLING_PRODUCTS).Set(updateRecord).Where(goqu.Ex{
		"name": toUpdate.Name,
	}).Returning(&Product{}).ToSQL()
	if err != nil {
		return product.Product{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var productModel Product
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRODUCTS, "UpdateByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&productModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Product{}, product.ErrProductNotFound
		default:
			return product.Product{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return productModel.transform()
}
