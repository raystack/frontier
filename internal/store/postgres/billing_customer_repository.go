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

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/billing/customer"
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
	ID         string `db:"id"`
	OrgID      string `db:"org_id"`
	ProviderID string `db:"provider_id"`

	Name     string             `db:"name"`
	Email    string             `db:"email"`
	Phone    *string            `db:"phone"`
	Currency string             `db:"currency"`
	Address  types.NullJSONText `db:"address"`
	Metadata types.NullJSONText `db:"metadata"`
	Tax      Tax                `db:"tax"`

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

	return customer.Customer{
		ID:         c.ID,
		OrgID:      c.OrgID,
		ProviderID: c.ProviderID,
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

	query, params, err := dialect.Insert(TABLE_BILLING_CUSTOMERS).Rows(
		goqu.Record{
			"org_id":      toCreate.OrgID,
			"provider_id": toCreate.ProviderID,
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
		return customer.Customer{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return customerModel.transform()
}

func (r BillingCustomerRepository) GetByID(ctx context.Context, id string) (customer.Customer, error) {
	stmt := dialect.Select().From(TABLE_BILLING_CUSTOMERS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %s", parseErr, err)
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
		return customer.Customer{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return customerModel.transform()
}

func (r BillingCustomerRepository) List(ctx context.Context, flt customer.Filter) ([]customer.Customer, error) {
	stmt := dialect.Select().From(TABLE_BILLING_CUSTOMERS)

	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	}
	if flt.State != "" {
		// where state is provided val or NULL or empty
		stmt = stmt.Where(goqu.L("state = ? OR state IS NULL OR state = ''", flt.State))
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

	var customerModels []Customer
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &customerModels, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []customer.Customer{}, nil
		}
		return nil, fmt.Errorf("%w: %s", dbErr, err)
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
		return customer.Customer{}, fmt.Errorf("%w: %s", parseErr, err)
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
	query, params, err := dialect.Update(TABLE_BILLING_CUSTOMERS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Customer{}).ToSQL()
	if err != nil {
		return customer.Customer{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var customerModel Customer
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.Customer{}, customer.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return customer.Customer{}, customer.ErrInvalidUUID
		default:
			return customer.Customer{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return customerModel.transform()
}

func (r BillingCustomerRepository) Delete(ctx context.Context, id string) error {
	stmt := dialect.Delete(TABLE_BILLING_CUSTOMERS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_CUSTOMERS, "Delete", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		return err
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return customer.ErrNotFound
		default:
			return fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return nil
}
