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
	"github.com/raystack/frontier/pkg/utils"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/db"
)

type Tax struct {
	TaxIDs []customer.Tax `json:"taxids"`
}

func (t *Tax) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, t)
	case string:
		return json.Unmarshal([]byte(src), t)
	case nil:
		return nil
	}
	return fmt.Errorf("cannot convert %T to JsonB", src)
}

func (t Tax) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type Customer struct {
	ID         string  `db:"id"`
	OrgID      string  `db:"org_id"`
	ProviderID *string `db:"provider_id"` // this could be empty if the customer is created as offline

	Name      string             `db:"name"`
	Email     string             `db:"email"`
	Phone     *string            `db:"phone"`
	Currency  string             `db:"currency"`
	Address   types.NullJSONText `db:"address"`
	Metadata  types.NullJSONText `db:"metadata"`
	Tax       Tax                `db:"tax"`
	CreditMin int64              `db:"credit_min"`
	DueInDays int64              `db:"due_in_days"`

	State     string     `db:"state"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (c Customer) transform() (customer.Customer, error) {
	var unmarshalledMetadata map[string]any
	if c.Metadata.Valid {
		if err := c.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return customer.Customer{}, err
		}
	}
	var unmarshalledAddress customer.Address
	if c.Address.Valid {
		if err := c.Address.Unmarshal(&unmarshalledAddress); err != nil {
			return customer.Customer{}, err
		}
	}
	customerPhone := ""
	if c.Phone != nil {
		customerPhone = *c.Phone
	}
	var customerTax []customer.Tax
	if len(c.Tax.TaxIDs) > 0 {
		customerTax = c.Tax.TaxIDs
	}
	var providerID string
	if c.ProviderID != nil {
		providerID = *c.ProviderID
	}
	if c.State == "" {
		c.State = customer.ActiveState.String()
	}
	return customer.Customer{
		ID:         c.ID,
		OrgID:      c.OrgID,
		ProviderID: providerID,
		Name:       c.Name,
		Email:      c.Email,
		Phone:      customerPhone,
		Currency:   c.Currency,
		Address:    unmarshalledAddress,
		TaxData:    customerTax,
		Metadata:   unmarshalledMetadata,
		State:      customer.State(c.State),
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		DeletedAt:  c.DeletedAt,
	}, nil
}

type BillingCustomerRepository struct {
	dbc *db.Client
}

func NewBillingCustomerRepository(dbc *db.Client) *BillingCustomerRepository {
	return &BillingCustomerRepository{
		dbc: dbc,
	}
}

func (r BillingCustomerRepository) Create(ctx context.Context, toCreate customer.Customer) (customer.Customer, error) {
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return customer.Customer{}, err
	}
	marshaledAddress, err := json.Marshal(toCreate.Address)
	if err != nil {
		return customer.Customer{}, err
	}

	var providerID *string
	if toCreate.ProviderID != "" {
		providerID = &toCreate.ProviderID
	}
	query, params, err := dialect.Insert(TABLE_BILLING_CUSTOMERS).Rows(
		goqu.Record{
			"org_id":      toCreate.OrgID,
			"provider_id": providerID,
			"name":        toCreate.Name,
			"email":       toCreate.Email,
			"phone":       toCreate.Phone,
			"currency":    toCreate.Currency,
			"address":     marshaledAddress,
			"state":       toCreate.State,
			"tax": Tax{
				TaxIDs: toCreate.TaxData,
			},
			"metadata": marshaledMetadata,
		}).Returning(&Customer{}).ToSQL()
	if err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "Create", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&customerModel); err != nil {
				return err
			}

			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.BillingCustomerCreatedEvent.String(),
				AuditResource{
					ID:   customerModel.ID,
					Type: "billing_customer",
					Name: customerModel.Name,
					Metadata: map[string]interface{}{
						"email":       customerModel.Email,
						"currency":    customerModel.Currency,
						"address":     customerModel.Address,
						"credit_min":  customerModel.CreditMin,
						"due_in_days": customerModel.DueInDays,
						"provider_id": customerModel.ProviderID,
					},
				},
				nil,
				customerModel.OrgID,
				nil,
				customerModel.CreatedAt,
			)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return customerModel.transform()
}

func (r BillingCustomerRepository) GetByID(ctx context.Context, id string) (customer.Customer, error) {
	stmt := dialect.Select().From(TABLE_BILLING_CUSTOMERS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.Customer{}, customer.ErrNotFound
		}
		return customer.Customer{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return customerModel.transform()
}

func (r BillingCustomerRepository) List(ctx context.Context, flt customer.Filter) ([]customer.Customer, error) {
	stmt := dialect.Select().From(TABLE_BILLING_CUSTOMERS).Order(goqu.I("created_at").Desc())

	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	}
	if flt.State != "" {
		// where state is provided val or NULL or empty
		stmt = stmt.Where(goqu.L("(state = ? OR state IS NULL OR state = '')", flt.State))
	}
	if flt.ProviderID != "" {
		stmt = stmt.Where(goqu.Ex{
			"provider_id": flt.ProviderID,
		})
	}
	if utils.BoolValue(flt.Online) {
		stmt = stmt.Where(goqu.L("(provider_id IS NOT NULL AND provider_id != '')"))
	}
	if utils.BoolValue(flt.AllowedOverdraft) {
		stmt = stmt.Where(goqu.L("credit_min < 0"))
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", parseErr, err)
	}

	var customerModels []Customer
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &customerModels, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []customer.Customer{}, nil
		}
		return nil, fmt.Errorf("%w: %w", dbErr, err)
	}

	customers := make([]customer.Customer, 0, len(customerModels))
	for _, customerModel := range customerModels {
		customer, err := customerModel.transform()
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	return customers, nil
}

func (r BillingCustomerRepository) UpdateByID(ctx context.Context, toUpdate customer.Customer) (customer.Customer, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return customer.Customer{}, customer.ErrInvalidID
	}
	if strings.TrimSpace(toUpdate.Email) == "" {
		return customer.Customer{}, customer.ErrInvalidDetail
	}
	marshaledAddress, err := json.Marshal(toUpdate.Address)
	if err != nil {
		return customer.Customer{}, err
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	updateRecord := goqu.Record{
		"name":     toUpdate.Name,
		"email":    toUpdate.Email,
		"phone":    toUpdate.Phone,
		"currency": toUpdate.Currency,
		"address":  marshaledAddress,
		"tax": Tax{
			TaxIDs: toUpdate.TaxData,
		},
		"metadata":   marshaledMetadata,
		"updated_at": goqu.L("now()"),
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State
	}
	if toUpdate.ProviderID != "" {
		// useful when updating an offline customer
		updateRecord["provider_id"] = &toUpdate.ProviderID
	}
	query, params, err := dialect.Update(TABLE_BILLING_CUSTOMERS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Customer{}).ToSQL()
	if err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "Update", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&customerModel); err != nil {
				return err
			}

			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.BillingCustomerUpdatedEvent.String(),
				AuditResource{
					ID:   customerModel.ID,
					Type: "billing_customer",
					Name: customerModel.Name,
					Metadata: map[string]interface{}{
						"email":       customerModel.Email,
						"currency":    customerModel.Currency,
						"address":     customerModel.Address,
						"provider_id": customerModel.ProviderID,
					},
				},
				nil,
				customerModel.OrgID,
				nil,
				customerModel.UpdatedAt,
			)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.Customer{}, customer.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return customer.Customer{}, customer.ErrInvalidUUID
		default:
			return customer.Customer{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	return customerModel.transform()
}

func (r BillingCustomerRepository) UpdateCreditMinByID(ctx context.Context, customerID string, limit int64) (customer.Details, error) {
	if strings.TrimSpace(customerID) == "" {
		return customer.Details{}, customer.ErrInvalidID
	}
	updateRecord := goqu.Record{
		"credit_min": limit,
		"updated_at": goqu.L("now()"),
	}
	query, params, err := dialect.Update(TABLE_BILLING_CUSTOMERS).Set(updateRecord).Where(goqu.Ex{
		"id": customerID,
	}).Returning(&Customer{}).ToSQL()
	if err != nil {
		return customer.Details{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.Details{}, customer.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return customer.Details{}, customer.ErrInvalidUUID
		default:
			return customer.Details{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	return customer.Details{
		CreditMin: customerModel.CreditMin,
		DueInDays: customerModel.DueInDays,
	}, nil
}

func (r BillingCustomerRepository) GetDetailsByID(ctx context.Context, customerID string) (customer.Details, error) {
	stmt := dialect.Select("credit_min", "due_in_days").From(TABLE_BILLING_CUSTOMERS).Where(goqu.Ex{
		"id": customerID,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return customer.Details{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "GetDetailsByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.Details{}, customer.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return customer.Details{}, customer.ErrInvalidUUID
		default:
			return customer.Details{}, fmt.Errorf("%w: %w", dbErr, err)
		}
	}

	return customer.Details{
		CreditMin: customerModel.CreditMin,
		DueInDays: customerModel.DueInDays,
	}, nil
}

func (r BillingCustomerRepository) UpdateDetailsByID(ctx context.Context, customerID string, details customer.Details) (customer.Details, error) {
	if strings.TrimSpace(customerID) == "" {
		return customer.Details{}, customer.ErrInvalidID
	}
	updateRecord := goqu.Record{
		"credit_min":  details.CreditMin,
		"due_in_days": details.DueInDays,
		"updated_at":  goqu.L("now()"),
	}
	query, params, err := dialect.Update(TABLE_BILLING_CUSTOMERS).Set(updateRecord).Where(goqu.Ex{
		"id": customerID,
	}).Returning(&Customer{}).ToSQL()
	if err != nil {
		return customer.Details{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "UpdateDetailsByID", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&customerModel); err != nil {
				return err
			}

			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.BillingCustomerCreditUpdatedEvent.String(),
				AuditResource{
					ID:   customerModel.ID,
					Type: "billing_customer",
					Name: customerModel.Name,
					Metadata: map[string]interface{}{
						"credit_min":  customerModel.CreditMin,
						"due_in_days": customerModel.DueInDays,
					},
				},
				nil,
				customerModel.OrgID,
				nil,
				customerModel.UpdatedAt,
			)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.Details{}, customer.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return customer.Details{}, customer.ErrInvalidUUID
		default:
			return customer.Details{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	return customer.Details{
		CreditMin: customerModel.CreditMin,
		DueInDays: customerModel.DueInDays,
	}, nil
}

func (r BillingCustomerRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_BILLING_CUSTOMERS).
		Where(goqu.Ex{"id": id}).
		Returning(&Customer{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", parseErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "Delete", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&customerModel); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return customer.ErrNotFound
				}
				return err
			}

			deletedAt := time.Now()
			if customerModel.DeletedAt != nil {
				deletedAt = *customerModel.DeletedAt
			}

			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.BillingCustomerDeletedEvent.String(),
				AuditResource{
					ID:   customerModel.ID,
					Type: "billing_customer",
					Name: customerModel.Name,
					Metadata: map[string]interface{}{
						"email":    customerModel.Email,
						"currency": customerModel.Currency,
						"address":  customerModel.Address,
					},
				},
				nil,
				customerModel.OrgID,
				nil,
				deletedAt,
			)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.ErrNotFound
		default:
			return fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	return nil
}
