package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/db"
)

type Price struct {
	ID            string `db:"id"`
	ProductID     string `db:"product_id"`
	ProviderID    string `db:"provider_id"`
	Name          string `db:"name"`
	BillingScheme string `db:"billing_scheme"`
	Currency      string `db:"currency"`
	Amount        int64  `db:"amount"`

	Interval string `db:"interval"`

	UsageType        string  `db:"usage_type"`
	MeteredAggregate *string `db:"metered_aggregate"`
	TierMode         *string `db:"tier_mode"`

	State    string             `db:"state"`
	Metadata types.NullJSONText `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (p Price) transform() (product.Price, error) {
	var unmarshalledMetadata map[string]any
	if p.Metadata.Valid {
		if err := p.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return product.Price{}, err
		}
	}
	meteredAggregate := ""
	if p.MeteredAggregate != nil {
		meteredAggregate = *p.MeteredAggregate
	}
	tierMode := ""
	if p.TierMode != nil {
		tierMode = *p.TierMode
	}
	return product.Price{
		ID:               p.ID,
		ProductID:        p.ProductID,
		ProviderID:       p.ProviderID,
		Name:             p.Name,
		BillingScheme:    product.BillingScheme(p.BillingScheme),
		Currency:         p.Currency,
		Amount:           p.Amount,
		UsageType:        product.PriceUsageType(p.UsageType),
		MeteredAggregate: meteredAggregate,
		TierMode:         tierMode,
		Interval:         p.Interval,

		State:     p.State,
		Metadata:  unmarshalledMetadata,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: p.DeletedAt,
	}, nil
}

type BillingPriceRepository struct {
	dbc *db.Client
}

func NewBillingPriceRepository(dbc *db.Client) *BillingPriceRepository {
	return &BillingPriceRepository{
		dbc: dbc,
	}
}

func (r BillingPriceRepository) Create(ctx context.Context, toCreate product.Price) (product.Price, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return product.Price{}, err
	}
	if toCreate.ID == "" {
		toCreate.ID = uuid.New().String()
	}

	query, params, err := dialect.Insert(TABLE_BILLING_PRICES).Rows(
		goqu.Record{
			"id":                toCreate.ID,
			"name":              toCreate.Name,
			"product_id":        toCreate.ProductID,
			"provider_id":       toCreate.ProviderID,
			"billing_scheme":    toCreate.BillingScheme,
			"currency":          toCreate.Currency,
			"amount":            toCreate.Amount,
			"usage_type":        toCreate.UsageType,
			"interval":          toCreate.Interval,
			"metered_aggregate": toCreate.MeteredAggregate,
			"tier_mode":         toCreate.TierMode,
			"metadata":          marshaledMetadata,
		}).Returning(&Price{}).ToSQL()
	if err != nil {
		return product.Price{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var priceModel Price
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRICES, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&priceModel)
	}); err != nil {
		return product.Price{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return priceModel.transform()
}

func (r BillingPriceRepository) GetByID(ctx context.Context, id string) (product.Price, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PRICES).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return product.Price{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var priceModel Price
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRICES, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&priceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Price{}, product.ErrPriceNotFound
		}
		return product.Price{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return priceModel.transform()
}

func (r BillingPriceRepository) GetByName(ctx context.Context, name string) (product.Price, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PRICES).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return product.Price{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var priceModel Price
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRICES, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&priceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Price{}, product.ErrPriceNotFound
		}
		return product.Price{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return priceModel.transform()
}

func (r BillingPriceRepository) UpdateByID(ctx context.Context, toUpdate product.Price) (product.Price, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return product.Price{}, product.ErrInvalidDetail
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return product.Price{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	updateRecord := goqu.Record{
		"name":       toUpdate.Name,
		"metadata":   marshaledMetadata,
		"updated_at": goqu.L("now()"),
	}
	query, params, err := dialect.Update(TABLE_BILLING_PRICES).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Price{}).ToSQL()
	if err != nil {
		return product.Price{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var priceModel Price
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRICES, "UpdateByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&priceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return product.Price{}, product.ErrPriceNotFound
		default:
			return product.Price{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return priceModel.transform()
}

func (r BillingPriceRepository) List(ctx context.Context, filter product.Filter) ([]product.Price, error) {
	stmt := dialect.Select().From(TABLE_BILLING_PRICES)
	if len(filter.ProductIDs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"product_id": goqu.Op{"in": filter.ProductIDs},
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var priceModels []Price
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_PRICES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &priceModels, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []product.Price{}, nil
		}
		return nil, fmt.Errorf("%s: %w", err, dbErr)
	}

	prices := make([]product.Price, 0, len(priceModels))
	for _, priceModel := range priceModels {
		price, err := priceModel.transform()
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}
