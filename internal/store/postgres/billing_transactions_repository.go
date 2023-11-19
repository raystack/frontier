package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/raystack/frontier/billing/credit"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/pkg/db"
)

type Transaction struct {
	ID          string             `db:"id"`
	AccountID   string             `db:"account_id"`
	Amount      int64              `db:"amount"`
	Type        string             `db:"type"`
	Source      string             `db:"source"`
	Description string             `db:"description"`
	Metadata    types.NullJSONText `db:"metadata"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (c Transaction) transform() (credit.Transaction, error) {
	var unmarshalledMetadata map[string]any
	if c.Metadata.Valid {
		if err := c.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return credit.Transaction{}, err
		}
	}
	return credit.Transaction{
		ID:          c.ID,
		AccountID:   c.AccountID,
		Amount:      c.Amount,
		Type:        credit.TransactionType(c.Type),
		Source:      c.Source,
		Description: c.Description,
		Metadata:    unmarshalledMetadata,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}, nil
}

type BillingTransactionRepository struct {
	dbc *db.Client
}

func NewBillingTransactionRepository(dbc *db.Client) *BillingTransactionRepository {
	return &BillingTransactionRepository{
		dbc: dbc,
	}
}

func (r BillingTransactionRepository) CreateEntry(ctx context.Context, debitEntry credit.Transaction,
	creditEntry credit.Transaction) ([]credit.Transaction, error) {
	if debitEntry.Metadata == nil {
		debitEntry.Metadata = make(map[string]any)
	}
	debitMetadata, err := json.Marshal(debitEntry.Metadata)
	if err != nil {
		return nil, err
	}
	debitRecord := goqu.Record{
		"account_id":  debitEntry.AccountID,
		"description": debitEntry.Description,
		"type":        debitEntry.Type,
		"source":      debitEntry.Source,
		"amount":      debitEntry.Amount,
		"metadata":    debitMetadata,
		"created_at":  goqu.L("now()"),
		"updated_at":  goqu.L("now()"),
	}
	if debitEntry.ID != "" {
		debitRecord["id"] = debitEntry.ID
	}

	if creditEntry.Metadata == nil {
		creditEntry.Metadata = make(map[string]any)
	}
	creditMetadata, err := json.Marshal(creditEntry.Metadata)
	if err != nil {
		return nil, err
	}
	creditRecord := goqu.Record{
		"account_id":  creditEntry.AccountID,
		"description": creditEntry.Description,
		"type":        creditEntry.Type,
		"source":      creditEntry.Source,
		"amount":      creditEntry.Amount,
		"metadata":    creditMetadata,
		"created_at":  goqu.L("now()"),
		"updated_at":  goqu.L("now()"),
	}
	if creditEntry.ID != "" {
		creditRecord["id"] = creditEntry.ID
	}

	var creditReturnedEntry, debitReturnedEntry credit.Transaction
	if err := r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var debitModel Transaction
		var creditModel Transaction
		query, params, err := dialect.Insert(TABLE_BILLING_TRANSACTIONS).Rows(debitRecord).Returning(&Transaction{}).ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", parseErr, err)
		}
		if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "Create", func(ctx context.Context) error {
			return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&debitModel)
		}); err != nil {
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		query, params, err = dialect.Insert(TABLE_BILLING_TRANSACTIONS).Rows(creditRecord).Returning(&Transaction{}).ToSQL()
		if err != nil {
			return fmt.Errorf("%w: %s", parseErr, err)
		}
		if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "Create", func(ctx context.Context) error {
			return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&creditModel)
		}); err != nil {
			return fmt.Errorf("%w: %s", dbErr, err)
		}

		creditReturnedEntry, err = creditModel.transform()
		if err != nil {
			return fmt.Errorf("failed to transform credit entry: %w", err)
		}
		debitReturnedEntry, err = debitModel.transform()
		if err != nil {
			return fmt.Errorf("failed to transform debit entry: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create transaction entry: %w", err)
	}

	return []credit.Transaction{debitReturnedEntry, creditReturnedEntry}, nil
}

func (r BillingTransactionRepository) GetByID(ctx context.Context, id string) (credit.Transaction, error) {
	stmt := dialect.Select().From(TABLE_BILLING_TRANSACTIONS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return credit.Transaction{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var transactionModel Transaction
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&transactionModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return credit.Transaction{}, credit.ErrNotFound
		}
		return credit.Transaction{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return transactionModel.transform()
}

func (r BillingTransactionRepository) UpdateByID(ctx context.Context, toUpdate credit.Transaction) (credit.Transaction, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return credit.Transaction{}, credit.ErrInvalidID
	}
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return credit.Transaction{}, fmt.Errorf("%w: %s", parseErr, err)
	}
	updateRecord := goqu.Record{
		"metadata":   marshaledMetadata,
		"updated_at": goqu.L("now()"),
	}
	query, params, err := dialect.Update(TABLE_BILLING_TRANSACTIONS).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Transaction{}).ToSQL()
	if err != nil {
		return credit.Transaction{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var customerModel Transaction
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&customerModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return credit.Transaction{}, credit.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return credit.Transaction{}, credit.ErrInvalidUUID
		default:
			return credit.Transaction{}, fmt.Errorf("%s: %w", txnErr, err)
		}
	}

	return customerModel.transform()
}

func (r BillingTransactionRepository) List(ctx context.Context, filter credit.Filter) ([]credit.Transaction, error) {
	stmt := dialect.Select().From(TABLE_BILLING_TRANSACTIONS)
	if filter.AccountID != "" {
		stmt = stmt.Where(goqu.Ex{
			"account_id": filter.AccountID,
		})
	}
	if !filter.Since.IsZero() {
		stmt = stmt.Where(goqu.Ex{
			"created_at": goqu.Op{"gt": filter.Since},
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var transactionModels []Transaction
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &transactionModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transactions []credit.Transaction
	for _, transactionModel := range transactionModels {
		transaction, err := transactionModel.transform()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// GetBalance currently sums all transactions for a customer and returns the balance.
// Ideally to speed this up we should create another table transaction_statement which
// will in batch compute the monthly summary for each customer, and then we can just
// query that table to get the balance since last month end date and add it to the entries
// in transaction table till now.
func (r BillingTransactionRepository) GetBalance(ctx context.Context, accountID string) (int64, error) {
	stmt := dialect.Select(goqu.SUM("amount")).From(TABLE_BILLING_TRANSACTIONS).Where(goqu.Ex{
		"account_id": accountID,
		"type":       credit.TypeDebit,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("%w: %s", parseErr, err)
	}

	var debitBalance *int64
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "GetBalance", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).Scan(&debitBalance)
	}); err != nil {
		return 0, fmt.Errorf("%w: %s", dbErr, err)
	}

	stmt = dialect.Select(goqu.SUM("amount")).From(TABLE_BILLING_TRANSACTIONS).Where(goqu.Ex{
		"account_id": accountID,
		"type":       credit.TypeCredit,
	})
	query, params, err = stmt.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("%w: %s", parseErr, err)
	}

	var creditBalance *int64
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "GetBalance", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).Scan(&creditBalance)
	}); err != nil {
		return 0, fmt.Errorf("%w: %s", dbErr, err)
	}

	if creditBalance == nil {
		creditBalance = new(int64)
	}
	if debitBalance == nil {
		debitBalance = new(int64)
	}
	return max(*creditBalance-*debitBalance, 0), nil
}
