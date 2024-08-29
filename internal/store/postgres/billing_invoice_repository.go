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

	"github.com/raystack/frontier/billing/invoice"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/pkg/db"
)

type Invoice struct {
	ID         string `db:"id"`
	ProviderID string `db:"provider_id"`
	CustomerID string `db:"customer_id"`
	State      string `db:"state"`
	Currency   string `db:"currency"`
	Amount     int64  `db:"amount"`
	HostedURL  string `db:"hosted_url"`

	Metadata types.NullJSONText `db:"metadata"`

	PeriodStartAt *time.Time `db:"period_start_at"`
	PeriodEndAt   *time.Time `db:"period_end_at"`
	DueAt         *time.Time `db:"due_at"`
	EffectiveAt   *time.Time `db:"effective_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (i Invoice) transform() (invoice.Invoice, error) {
	var unmarshalledMetadata map[string]any
	if i.Metadata.Valid {
		if err := i.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return invoice.Invoice{}, err
		}
	}
	dueAt := time.Time{}
	if i.DueAt != nil {
		dueAt = *i.DueAt
	}
	effectiveAt := time.Time{}
	if i.EffectiveAt != nil {
		effectiveAt = *i.EffectiveAt
	}
	var periodStartAt time.Time
	if i.PeriodStartAt != nil {
		periodStartAt = *i.PeriodStartAt
	}
	var periodEndAt time.Time
	if i.PeriodEndAt != nil {
		periodEndAt = *i.PeriodEndAt
	}
	return invoice.Invoice{
		ID:            i.ID,
		ProviderID:    i.ProviderID,
		CustomerID:    i.CustomerID,
		State:         i.State,
		Currency:      i.Currency,
		Amount:        i.Amount,
		HostedURL:     i.HostedURL,
		Metadata:      unmarshalledMetadata,
		DueAt:         dueAt,
		EffectiveAt:   effectiveAt,
		CreatedAt:     i.CreatedAt,
		PeriodStartAt: periodStartAt,
		PeriodEndAt:   periodEndAt,
	}, nil
}

type BillingInvoiceRepository struct {
	dbc *db.Client
}

func NewBillingInvoiceRepository(dbc *db.Client) *BillingInvoiceRepository {
	return &BillingInvoiceRepository{
		dbc: dbc,
	}
}

func (r BillingInvoiceRepository) Create(ctx context.Context, toCreate invoice.Invoice) (invoice.Invoice, error) {
	if toCreate.ID == "" {
		toCreate.ID = uuid.New().String()
	}
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return invoice.Invoice{}, err
	}

	query, params, err := dialect.Insert(TABLE_BILLING_INVOICES).Rows(
		goqu.Record{
			"id":              toCreate.ID,
			"provider_id":     toCreate.ProviderID,
			"customer_id":     toCreate.CustomerID,
			"state":           toCreate.State,
			"currency":        toCreate.Currency,
			"amount":          toCreate.Amount,
			"hosted_url":      toCreate.HostedURL,
			"due_at":          toCreate.DueAt,
			"effective_at":    toCreate.EffectiveAt,
			"metadata":        marshaledMetadata,
			"period_start_at": toCreate.PeriodStartAt,
			"period_end_at":   toCreate.PeriodEndAt,
			"created_at":      goqu.L("now()"),
			"updated_at":      goqu.L("now()"),
		}).Returning(&Invoice{}).ToSQL()
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var invoiceModel Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&invoiceModel)
	}); err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return invoiceModel.transform()
}

func (r BillingInvoiceRepository) GetByID(ctx context.Context, id string) (invoice.Invoice, error) {
	stmt := dialect.Select().From(TABLE_BILLING_INVOICES).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var invoiceModel Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&invoiceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return invoice.Invoice{}, invoice.ErrNotFound
		}
		return invoice.Invoice{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return invoiceModel.transform()
}

func (r BillingInvoiceRepository) List(ctx context.Context, flt invoice.Filter) ([]invoice.Invoice, error) {
	stmt := dialect.Select().From(TABLE_BILLING_INVOICES)
	if flt.CustomerID != "" {
		stmt = stmt.Where(goqu.Ex{
			"customer_id": flt.CustomerID,
		})
	}
	if flt.NonZeroOnly {
		stmt = stmt.Where(goqu.Ex{
			"amount": goqu.Op{"gt": 0},
		})
	}

	if flt.Pagination != nil {
		offset := flt.Pagination.Offset()
		limit := flt.Pagination.PageSize

		// always make this call after all the filters have been applied
		totalCountStmt := stmt.Select(goqu.COUNT("*"))
		totalCountQuery, _, err := totalCountStmt.ToSQL()

		if err != nil {
			return []invoice.Invoice{}, fmt.Errorf("%w: %s", queryErr, err)
		}

		var totalCount int32
		if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "Count", func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &totalCount, totalCountQuery)
		}); err != nil {
			return nil, fmt.Errorf("%w: %s", dbErr, err)
		}

		flt.Pagination.SetCount(totalCount)
		stmt = stmt.Limit(uint(limit)).Offset(uint(offset))
	}

	stmt = stmt.Order(goqu.I("created_at").Desc())
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var invoiceModels []Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &invoiceModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	invoices := make([]invoice.Invoice, 0, len(invoiceModels))
	for _, invoiceModel := range invoiceModels {
		invoice, err := invoiceModel.transform()
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (r BillingInvoiceRepository) UpdateByID(ctx context.Context, toUpdate invoice.Invoice) (invoice.Invoice, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return invoice.Invoice{}, invoice.ErrInvalidDetail
	}

	updateRecord := goqu.Record{
		"updated_at": goqu.L("now()"),
	}
	if toUpdate.Metadata != nil {
		marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
		if err != nil {
			return invoice.Invoice{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		updateRecord["metadata"] = marshaledMetadata
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}
	if !toUpdate.EffectiveAt.IsZero() {
		updateRecord["effective_at"] = toUpdate.EffectiveAt
	}
	if toUpdate.HostedURL != "" {
		updateRecord["hosted_url"] = toUpdate.HostedURL
	}

	query, params, err := dialect.Update(TABLE_BILLING_INVOICES).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Invoice{}).ToSQL()
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var invoiceModel Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "UpdateByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&invoiceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return invoice.Invoice{}, invoice.ErrNotFound
		default:
			return invoice.Invoice{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return invoiceModel.transform()
}

func (r BillingInvoiceRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_BILLING_INVOICES).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "Delete", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		return err
	}); err != nil {
		return fmt.Errorf("%s: %w", txnErr, err)
	}
	return nil
}
