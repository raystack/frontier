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
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/pkg/db"
)

type Subscription struct {
	ID             string `db:"id"`
	ProviderID     string `db:"provider_id"`
	SubscriptionID string `db:"customer_id"`
	PlanID         string `db:"plan_id"`

	Metadata types.NullJSONText `db:"metadata"`
	State    string             `db:"state"`

	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	CanceledAt *time.Time `db:"canceled_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

func (c Subscription) transform() (subscription.Subscription, error) {
	var unmarshalledMetadata map[string]any
	if c.Metadata.Valid {
		if err := c.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return subscription.Subscription{}, err
		}
	}

	return subscription.Subscription{
		ID:         c.ID,
		ProviderID: c.ProviderID,
		CustomerID: c.SubscriptionID,
		State:      c.State,
		PlanID:     c.PlanID,

		Metadata:   unmarshalledMetadata,
		CreatedAt:  c.CreatedAt,
		CanceledAt: c.CanceledAt,
		UpdatedAt:  c.UpdatedAt,
		DeletedAt:  c.DeletedAt,
	}, nil
}

type BillingSubscriptionRepository struct {
	dbc *db.Client
}

func NewBillingSubscriptionRepository(dbc *db.Client) *BillingSubscriptionRepository {
	return &BillingSubscriptionRepository{
		dbc: dbc,
	}
}

func (r BillingSubscriptionRepository) Create(ctx context.Context, toCreate subscription.Subscription) (subscription.Subscription, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return subscription.Subscription{}, err
	}

	query, params, err := dialect.Insert(TABLE_BILLING_SUBSCRIPTIONS).Rows(
		goqu.Record{
			"id":          toCreate.ID,
			"provider_id": toCreate.ProviderID,
			"customer_id": toCreate.CustomerID,
			"plan_id":     toCreate.PlanID,
			"state":       toCreate.State,
			"metadata":    marshaledMetadata,
			"updated_at":  goqu.L("now()"),
		}).Returning(&Subscription{}).ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&subscriptionModel)
	}); err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return subscriptionModel.transform()
}

func (r BillingSubscriptionRepository) GetByID(ctx context.Context, id string) (subscription.Subscription, error) {
	stmt := dialect.Select().From(TABLE_BILLING_SUBSCRIPTIONS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&subscriptionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return subscriptionModel.transform()
}

func (r BillingSubscriptionRepository) GetByName(ctx context.Context, name string) (subscription.Subscription, error) {
	stmt := dialect.Select().From(TABLE_BILLING_SUBSCRIPTIONS).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&subscriptionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return subscriptionModel.transform()
}

func (r BillingSubscriptionRepository) UpdateByID(ctx context.Context, toUpdate subscription.Subscription) (subscription.Subscription, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return subscription.Subscription{}, subscription.ErrInvalidID
	}
	if strings.TrimSpace(toUpdate.State) == "" {
		return subscription.Subscription{}, subscription.ErrInvalidDetail
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	updateRecord := goqu.Record{
		"metadata":   marshaledMetadata,
		"updated_at": goqu.L("now()"),
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}
	if toUpdate.CanceledAt != nil {
		updateRecord["canceled_at"] = toUpdate.CanceledAt
	}
	query, params, err := dialect.Update(TABLE_BILLING_SUBSCRIPTIONS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Subscription{}).ToSQL()
	if err != nil {
		return subscription.Subscription{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var customerModel Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return subscription.Subscription{}, subscription.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return subscription.Subscription{}, subscription.ErrInvalidUUID
		default:
			return subscription.Subscription{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return customerModel.transform()
}

func (r BillingSubscriptionRepository) List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error) {
	stmt := dialect.Select().From(TABLE_BILLING_SUBSCRIPTIONS)
	if filter.CustomerID != "" {
		stmt = stmt.Where(goqu.Ex{
			"customer_id": filter.CustomerID,
		})
	}
	if filter.State != "" {
		stmt = stmt.Where(goqu.Ex{
			"state": filter.State,
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var subscriptionModels []Subscription
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_SUBSCRIPTIONS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &subscriptionModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var subscriptions []subscription.Subscription
	for _, subscriptionModel := range subscriptionModels {
		subscription, err := subscriptionModel.transform()
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}
