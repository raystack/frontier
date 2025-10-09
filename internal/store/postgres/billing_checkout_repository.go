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

	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/billing/checkout"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/pkg/db"
)

type SubscriptionConfig struct {
	SkipTrial        bool `db:"skip_trial" json:"skip_trial"`
	CancelAfterTrial bool `db:"cancel_after_trial" json:"cancel_after_trial"`
}

type checkoutWithContext struct {
	Checkout
	OrgName string `db:"org_name"`
}

func (s *SubscriptionConfig) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, s)
	case string:
		return json.Unmarshal([]byte(src), s)
	case nil:
		return nil
	}
	return fmt.Errorf("cannot convert %T to JsonB", src)
}

func (s SubscriptionConfig) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type Checkout struct {
	ID         string `db:"id"`
	CustomerID string `db:"customer_id"`
	ProviderID string `db:"provider_id"`

	PlanID             *string            `db:"plan_id"`
	SubscriptionConfig SubscriptionConfig `db:"subscription_config"`
	FeatureID          *string            `db:"feature_id"`

	SuccessUrl    *string            `db:"success_url"`
	CancelUrl     *string            `db:"cancel_url"`
	CheckoutUrl   string             `db:"checkout_url"`
	PaymentStatus *string            `db:"payment_status"`
	State         string             `db:"state"`
	Metadata      types.NullJSONText `db:"metadata"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	ExpireAt  *time.Time `db:"expire_at"`
}

func (c Checkout) transform() (checkout.Checkout, error) {
	var unmarshalledMetadata map[string]any
	if c.Metadata.Valid {
		if err := c.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return checkout.Checkout{}, err
		}
	}
	successUrl := ""
	if c.SuccessUrl != nil {
		successUrl = *c.SuccessUrl
	}
	cancelUrl := ""
	if c.CancelUrl != nil {
		cancelUrl = *c.CancelUrl
	}
	expireAt := time.Time{}
	if c.ExpireAt != nil {
		expireAt = *c.ExpireAt
	}
	paymentStatus := ""
	if c.PaymentStatus != nil {
		paymentStatus = *c.PaymentStatus
	}
	planID := ""
	if c.PlanID != nil {
		planID = *c.PlanID
	}
	featureID := ""
	if c.FeatureID != nil {
		featureID = *c.FeatureID
	}
	return checkout.Checkout{
		ID:               c.ID,
		ProviderID:       c.ProviderID,
		CustomerID:       c.CustomerID,
		PlanID:           planID,
		SkipTrial:        c.SubscriptionConfig.SkipTrial,
		CancelAfterTrial: c.SubscriptionConfig.CancelAfterTrial,
		ProductID:        featureID,
		CheckoutUrl:      c.CheckoutUrl,
		SuccessUrl:       successUrl,
		CancelUrl:        cancelUrl,
		State:            c.State,
		PaymentStatus:    paymentStatus,
		Metadata:         unmarshalledMetadata,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		ExpireAt:         expireAt,
	}, nil
}

type BillingCheckoutRepository struct {
	dbc *db.Client
}

func NewBillingCheckoutRepository(dbc *db.Client) *BillingCheckoutRepository {
	return &BillingCheckoutRepository{
		dbc: dbc,
	}
}

func (r BillingCheckoutRepository) Create(ctx context.Context, toCreate checkout.Checkout) (checkout.Checkout, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return checkout.Checkout{}, err
	}

	record := goqu.Record{
		"provider_id":    toCreate.ProviderID,
		"customer_id":    toCreate.CustomerID,
		"checkout_url":   toCreate.CheckoutUrl,
		"success_url":    toCreate.SuccessUrl,
		"cancel_url":     toCreate.CancelUrl,
		"state":          toCreate.State,
		"payment_status": toCreate.PaymentStatus,
		"metadata":       marshaledMetadata,
		"subscription_config": SubscriptionConfig{
			SkipTrial:        toCreate.SkipTrial,
			CancelAfterTrial: toCreate.CancelAfterTrial,
		},
		"updated_at": goqu.L("now()"),
		"expire_at":  toCreate.ExpireAt,
	}
	if toCreate.ID != "" {
		record["id"] = toCreate.ID
	}
	if toCreate.ProductID != "" {
		record["feature_id"] = toCreate.ProductID
	}
	if toCreate.PlanID != "" {
		record["plan_id"] = toCreate.PlanID
	}

	var result checkoutWithContext

	// Fetch org name via billing_customers -> organizations
	orgNameSubquery := dialect.From(TABLE_BILLING_CUSTOMERS).
		Join(
			goqu.T(TABLE_ORGANIZATIONS),
			goqu.On(goqu.I(TABLE_BILLING_CUSTOMERS+".org_id").Eq(goqu.I(TABLE_ORGANIZATIONS+".id"))),
		).
		Select(goqu.I(TABLE_ORGANIZATIONS + ".name")).
		Where(goqu.Ex{TABLE_BILLING_CUSTOMERS + ".id": goqu.I(TABLE_BILLING_CHECKOUTS + ".customer_id")})

	query, params, err := dialect.Insert(TABLE_BILLING_CHECKOUTS).
		Rows(record).
		Returning(
			goqu.I(TABLE_BILLING_CHECKOUTS+".*"),
			orgNameSubquery.As("org_name"),
		).ToSQL()
	if err != nil {
		return checkout.Checkout{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_CHECKOUTS, "Create", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&result); err != nil {
				return err
			}

			// Get org_id from billing_customers
			var orgID string
			orgIDQuery, orgIDParams, err := dialect.From(TABLE_BILLING_CUSTOMERS).
				Select("org_id").
				Where(goqu.Ex{"id": result.CustomerID}).
				ToSQL()
			if err != nil {
				return fmt.Errorf("failed to build org_id query: %w", err)
			}
			if err := tx.QueryRowContext(ctx, orgIDQuery, orgIDParams...).Scan(&orgID); err != nil {
				return fmt.Errorf("failed to get org_id: %w", err)
			}

			planID := ""
			if result.PlanID != nil {
				planID = *result.PlanID
			}
			featureID := ""
			if result.FeatureID != nil {
				featureID = *result.FeatureID
			}

			auditRecord := BuildAuditRecord(
				ctx,
				"billing_checkout.created",
				AuditResource{
					ID:   orgID,
					Type: "organization",
					Name: result.OrgName,
				},
				&AuditTarget{
					ID:   result.ID,
					Type: "billing_checkout",
					Metadata: map[string]interface{}{
						"plan_id":            planID,
						"feature_id":         featureID,
						"state":              result.State,
						"provider_id":        result.ProviderID,
						"customer_id":        result.CustomerID,
						"checkout_url":       result.CheckoutUrl,
						"skip_trial":         result.SubscriptionConfig.SkipTrial,
						"cancel_after_trial": result.SubscriptionConfig.CancelAfterTrial,
					},
				},
				orgID,
				nil,
				result.CreatedAt,
			)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		return checkout.Checkout{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return result.Checkout.transform()
}

func (r BillingCheckoutRepository) GetByID(ctx context.Context, id string) (checkout.Checkout, error) {
	stmt := dialect.Select().From(TABLE_BILLING_CHECKOUTS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return checkout.Checkout{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var checkoutModel Checkout
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CHECKOUTS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&checkoutModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return checkout.Checkout{}, checkout.ErrNotFound
		}
		return checkout.Checkout{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return checkoutModel.transform()
}

func (r BillingCheckoutRepository) GetByName(ctx context.Context, name string) (checkout.Checkout, error) {
	stmt := dialect.Select().From(TABLE_BILLING_CHECKOUTS).Where(goqu.Ex{
		"name": name,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return checkout.Checkout{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var checkoutModel Checkout
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CHECKOUTS, "GetByName", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&checkoutModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return checkout.Checkout{}, checkout.ErrNotFound
		}
		return checkout.Checkout{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return checkoutModel.transform()
}

func (r BillingCheckoutRepository) UpdateByID(ctx context.Context, toUpdate checkout.Checkout) (checkout.Checkout, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return checkout.Checkout{}, checkout.ErrInvalidID
	}
	if strings.TrimSpace(toUpdate.State) == "" {
		return checkout.Checkout{}, checkout.ErrInvalidDetail
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return checkout.Checkout{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	updateRecord := goqu.Record{
		"metadata":   marshaledMetadata,
		"updated_at": goqu.L("now()"),
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}
	if toUpdate.PaymentStatus != "" {
		updateRecord["payment_status"] = toUpdate.PaymentStatus
	}

	var result checkoutWithContext

	// Fetch org name via billing_customers -> organizations
	orgNameSubquery := dialect.From(TABLE_BILLING_CUSTOMERS).
		Select(goqu.I(TABLE_ORGANIZATIONS+".name")).
		Join(
			goqu.T(TABLE_ORGANIZATIONS),
			goqu.On(goqu.I(TABLE_BILLING_CUSTOMERS+".org_id").Eq(goqu.I(TABLE_ORGANIZATIONS+".id"))),
		).
		Where(goqu.Ex{TABLE_BILLING_CUSTOMERS + ".id": goqu.I(TABLE_BILLING_CHECKOUTS + ".customer_id")})

	query, params, err := dialect.Update(TABLE_BILLING_CHECKOUTS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(
		goqu.I(TABLE_BILLING_CHECKOUTS+".*"),
		orgNameSubquery.As("org_name"),
	).ToSQL()
	if err != nil {
		return checkout.Checkout{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_CHECKOUTS, "Update", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&result); err != nil {
				err = checkPostgresError(err)
				switch {
				case errors.Is(err, sql.ErrNoRows):
					return checkout.ErrNotFound
				case errors.Is(err, ErrInvalidTextRepresentation):
					return checkout.ErrInvalidUUID
				default:
					return fmt.Errorf("%w: %w", txnErr, err)
				}
			}

			// Get org_id from billing_customers
			var orgID string
			orgIDQuery, orgIDParams, err := dialect.From(TABLE_BILLING_CUSTOMERS).
				Select("org_id").
				Where(goqu.Ex{"id": result.CustomerID}).
				ToSQL()
			if err != nil {
				return fmt.Errorf("failed to build org_id query: %w", err)
			}
			if err := tx.QueryRowContext(ctx, orgIDQuery, orgIDParams...).Scan(&orgID); err != nil {
				return fmt.Errorf("failed to get org_id: %w", err)
			}

			planID := ""
			if result.PlanID != nil {
				planID = *result.PlanID
			}
			featureID := ""
			if result.FeatureID != nil {
				featureID = *result.FeatureID
			}
			paymentStatus := ""
			if result.PaymentStatus != nil {
				paymentStatus = *result.PaymentStatus
			}

			auditRecord := BuildAuditRecord(
				ctx,
				"billing_checkout.updated",
				AuditResource{
					ID:   orgID,
					Type: "organization",
					Name: result.OrgName,
				},
				&AuditTarget{
					ID:   result.ID,
					Type: "billing_checkout",
					Metadata: map[string]interface{}{
						"plan_id":            planID,
						"feature_id":         featureID,
						"state":              result.State,
						"payment_status":     paymentStatus,
						"provider_id":        result.ProviderID,
						"customer_id":        result.CustomerID,
						"checkout_url":       result.CheckoutUrl,
						"skip_trial":         result.SubscriptionConfig.SkipTrial,
						"cancel_after_trial": result.SubscriptionConfig.CancelAfterTrial,
					},
				},
				orgID,
				nil,
				result.UpdatedAt,
			)

			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		return checkout.Checkout{}, err
	}

	return result.Checkout.transform()
}

func (r BillingCheckoutRepository) List(ctx context.Context, flt checkout.Filter) ([]checkout.Checkout, error) {
	stmt := dialect.Select().From(TABLE_BILLING_CHECKOUTS).Order(goqu.I("created_at").Desc())
	if flt.CustomerID != "" {
		stmt = stmt.Where(goqu.Ex{
			"customer_id": flt.CustomerID,
		})
	}
	if flt.ProviderID != "" {
		stmt = stmt.Where(goqu.Ex{
			"provider_id": flt.ProviderID,
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var checkoutModels []Checkout
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CHECKOUTS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &checkoutModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	checkouts := make([]checkout.Checkout, 0, len(checkoutModels))
	for _, checkoutModel := range checkoutModels {
		checkout, err := checkoutModel.transform()
		if err != nil {
			return nil, err
		}
		checkouts = append(checkouts, checkout)
	}
	return checkouts, nil
}
