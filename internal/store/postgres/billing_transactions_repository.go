package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/raystack/frontier/billing/customer"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/jackc/pgconn"

	"github.com/jmoiron/sqlx"

	"github.com/raystack/frontier/billing/credit"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/pkg/db"
)

// Transaction represents a transaction entry in the database.
// A transaction can be of type credit or debit, every change creates two
// entry in records. At the moment transfer of funds are only between
// customer account and system account, we don't need a transaction id.
// If we do to support transfer of amount between more than two accounts
// we can add a transaction id which will be same for all entries in a
// single transaction.
type Transaction struct {
	ID          string             `db:"id"`
	AccountID   string             `db:"account_id"`
	Amount      int64              `db:"amount"`
	Type        string             `db:"type"`
	Source      string             `db:"source"`
	Description string             `db:"description"`
	UserID      *string            `db:"user_id"`
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
	userID := ""
	if c.UserID != nil {
		userID = *c.UserID
	}
	return credit.Transaction{
		ID:          c.ID,
		CustomerID:  c.AccountID,
		Amount:      c.Amount,
		Type:        credit.TransactionType(c.Type),
		Source:      c.Source,
		Description: c.Description,
		UserID:      userID,
		Metadata:    unmarshalledMetadata,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}, nil
}

type BillingTransactionRepository struct {
	dbc          *db.Client
	customerRepo *BillingCustomerRepository
}

func NewBillingTransactionRepository(dbc *db.Client) *BillingTransactionRepository {
	return &BillingTransactionRepository{
		dbc:          dbc,
		customerRepo: NewBillingCustomerRepository(dbc),
	}
}

var (
	maxRetries = 5
	// Error codes from https://www.postgresql.org/docs/current/errcodes-appendix.html
	serializationFailureCode = "40001"
	deadlockDetectedCode     = "40P01"
)

func (r BillingTransactionRepository) withRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := 0; i < maxRetries && ctx.Err() == nil; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) {
			// Retry on serialization failures or deadlocks
			if pqErr.Code == serializationFailureCode || pqErr.Code == deadlockDetectedCode {
				lastErr = err
				// Exponential backoff with jitter
				backoff := time.Duration(1<<uint(i)) * 50 * time.Millisecond
				jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
				time.Sleep(backoff + jitter)
				continue
			}
		}
		return err // Return immediately for other errors
	}
	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (r BillingTransactionRepository) CreateEntry(ctx context.Context, debitEntry credit.Transaction,
	creditEntry credit.Transaction) ([]credit.Transaction, error) {
	if creditEntry.ID == "" && debitEntry.ID == "" {
		return nil, credit.ErrInvalidID
	}
	txOpts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	var customerAcc customer.Customer
	var err error
	var debitModel Transaction
	var creditModel Transaction

	if debitEntry.CustomerID != schema.PlatformOrgID.String() {
		customerAcc, err = r.customerRepo.GetByID(ctx, debitEntry.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get customer account: %w", err)
		}
	}

	var creditReturnedEntry, debitReturnedEntry credit.Transaction
	err = r.withRetry(ctx, func() error {
		return r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
			if customerAcc.ID != "" {
				// check for balance only when deducting from customer account
				currentBalance, err := r.getBalanceInTx(ctx, tx, debitEntry.CustomerID, nil, nil)
				if err != nil {
					return fmt.Errorf("failed to get balance: %w", err)
				}

				if err := isSufficientBalance(customerAcc.CreditMin, currentBalance, debitEntry.Amount); err != nil {
					return err
				}
			}

			if err := r.createTransactionEntry(ctx, tx, debitEntry, &debitModel); err != nil {
				return fmt.Errorf("failed to create debit entry: %w", err)
			}
			if err := r.createTransactionEntry(ctx, tx, creditEntry, &creditModel); err != nil {
				return fmt.Errorf("failed to create credit entry: %w", err)
			}
			return nil
		})
	})
	if err != nil {
		if errors.Is(err, credit.ErrAlreadyApplied) {
			return nil, credit.ErrAlreadyApplied
		} else if errors.Is(err, credit.ErrInsufficientCredits) {
			return nil, credit.ErrInsufficientCredits
		}
		return nil, fmt.Errorf("failed to create transaction entry: %w", err)
	}

	creditReturnedEntry, err = creditModel.transform()
	if err != nil {
		return nil, fmt.Errorf("failed to transform credit entry: %w", err)
	}
	debitReturnedEntry, err = debitModel.transform()
	if err != nil {
		return nil, fmt.Errorf("failed to transform debit entry: %w", err)
	}
	return []credit.Transaction{debitReturnedEntry, creditReturnedEntry}, nil
}

func (r BillingTransactionRepository) createTransactionEntry(ctx context.Context, tx *sqlx.Tx, entry credit.Transaction, model *Transaction) error {
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]any)
	}
	metadata, err := json.Marshal(entry.Metadata)
	if err != nil {
		return err
	}

	record := goqu.Record{
		"account_id":  entry.CustomerID,
		"description": entry.Description,
		"type":        entry.Type,
		"source":      entry.Source,
		"amount":      entry.Amount,
		"user_id":     entry.UserID,
		"metadata":    metadata,
		"created_at":  goqu.L("now()"),
		"updated_at":  goqu.L("now()"),
	}
	if entry.ID != "" {
		record["id"] = entry.ID
	}

	query, params, err := dialect.Insert(TABLE_BILLING_TRANSACTIONS).Rows(record).Returning(&Transaction{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", parseErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "Create", func(ctx context.Context) error {
		return tx.QueryRowxContext(ctx, query, params...).StructScan(model)
	}); err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && (pqErr.Code == "23505") {
			if pqErr.ConstraintName == "billing_transactions_pkey" {
				return credit.ErrAlreadyApplied
			}
		}
		return fmt.Errorf("%w: %w", dbErr, err)
	}

	return nil
}

// isSufficientBalance checks if the customer has enough balance to perform the transaction.
// If the customer has a credit min limit set, then a negative balance means loaner/overdraft limit and
// a positive limit mean at least that much balance should be there in the account.
func isSufficientBalance(customerMinLimit int64, currentBalance int64, txAmount int64) error {
	if customerMinLimit < 0 {
		if currentBalance-customerMinLimit < txAmount {
			return credit.ErrInsufficientCredits
		}
	} else if currentBalance < txAmount+customerMinLimit {
		return credit.ErrInsufficientCredits
	}
	return nil
}

func (r BillingTransactionRepository) GetByID(ctx context.Context, id string) (credit.Transaction, error) {
	if strings.TrimSpace(id) == "" {
		return credit.Transaction{}, credit.ErrInvalidID
	}
	stmt := dialect.Select().From(TABLE_BILLING_TRANSACTIONS).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return credit.Transaction{}, fmt.Errorf("%w: %w", parseErr, err)
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
			return credit.Transaction{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	return customerModel.transform()
}

func (r BillingTransactionRepository) List(ctx context.Context, filter credit.Filter) ([]credit.Transaction, error) {
	stmt := dialect.Select().From(TABLE_BILLING_TRANSACTIONS).Order(goqu.I("created_at").Desc())
	if filter.CustomerID != "" {
		stmt = stmt.Where(goqu.Ex{
			"account_id": filter.CustomerID,
		})
	}
	if !filter.StartRange.IsZero() {
		stmt = stmt.Where(goqu.Ex{
			"created_at": goqu.Op{"gte": filter.StartRange},
		})
	}
	if !filter.EndRange.IsZero() {
		stmt = stmt.Where(goqu.Ex{
			"created_at": goqu.Op{"lte": filter.EndRange},
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

func (r BillingTransactionRepository) getDebitBalance(ctx context.Context, tx *sqlx.Tx, accountID string,
	start *time.Time, end *time.Time) (*int64, error) {
	stmt := dialect.Select(goqu.SUM("amount")).From(TABLE_BILLING_TRANSACTIONS).Where(goqu.Ex{
		"account_id": accountID,
		"type":       credit.DebitType,
	})
	if start != nil {
		stmt = stmt.Where(goqu.Ex{
			"created_at": goqu.Op{"gte": *start},
		})
	}
	if end != nil {
		stmt = stmt.Where(goqu.Ex{
			"created_at": goqu.Op{"lt": *end},
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var debitBalance *int64
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "GetDebitBalance", func(ctx context.Context) error {
		return tx.QueryRowxContext(ctx, query, params...).Scan(&debitBalance)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}
	return debitBalance, nil
}

func (r BillingTransactionRepository) getCreditBalance(ctx context.Context, tx *sqlx.Tx, accountID string,
	start *time.Time, end *time.Time) (*int64, error) {
	stmt := dialect.Select(goqu.SUM("amount")).From(TABLE_BILLING_TRANSACTIONS).Where(goqu.Ex{
		"account_id": accountID,
		"type":       credit.CreditType,
	})
	if start != nil {
		stmt = stmt.Where(goqu.Ex{
			"created_at": goqu.Op{"gte": *start},
		})
	}
	if end != nil {
		stmt = stmt.Where(goqu.Ex{
			"created_at": goqu.Op{"lt": *end},
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", parseErr, err)
	}

	var creditBalance *int64
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_TRANSACTIONS, "GetCreditBalance", func(ctx context.Context) error {
		return tx.QueryRowxContext(ctx, query, params...).Scan(&creditBalance)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}
	return creditBalance, nil
}

// getBalanceInTx returns the balance of the account in the given range.
// start time is inclusive and end time is exclusive.
// if nil, then it will consider all transactions.
func (r BillingTransactionRepository) getBalanceInTx(ctx context.Context, tx *sqlx.Tx, accountID string,
	start *time.Time, end *time.Time) (int64, error) {
	var creditBalance *int64
	var debitBalance *int64

	var err error
	if debitBalance, err = r.getDebitBalance(ctx, tx, accountID, start, end); err != nil {
		return 0, fmt.Errorf("failed to get debit balance: %w", err)
	}
	if creditBalance, err = r.getCreditBalance(ctx, tx, accountID, start, end); err != nil {
		return 0, fmt.Errorf("failed to get credit balance: %w", err)
	}
	if creditBalance == nil {
		creditBalance = new(int64)
	}
	if debitBalance == nil {
		debitBalance = new(int64)
	}
	return *creditBalance - *debitBalance, nil
}

// GetBalance currently sums all transactions for a customer and returns the balance.
// Ideally to speed this up we should create another table transaction_statement which
// will in batch compute the monthly summary for each customer, and then we can just
// query that table to get the balance since last month end date and add it to the entries
// in transaction table till now.
func (r BillingTransactionRepository) GetBalance(ctx context.Context, accountID string) (int64, error) {
	var amount int64

	err := r.withRetry(ctx, func() error {
		return r.dbc.WithTxn(ctx, sql.TxOptions{
			Isolation: sql.LevelSerializable,
			ReadOnly:  true,
		}, func(tx *sqlx.Tx) error {
			var err error
			amount, err = r.getBalanceInTx(ctx, tx, accountID, nil, nil)
			return err
		})
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return amount, nil
}

// GetTotalDebitedAmount sums all debited transactions for a customer including the reverted.
func (r BillingTransactionRepository) GetTotalDebitedAmount(ctx context.Context, accountID string) (int64, error) {
	transactions, err := r.List(ctx, credit.Filter{CustomerID: accountID})
	if err != nil {
		return 0, nil
	}
	var totalDebit int64
	for _, transaction := range transactions {
		switch transaction.Type {
		case credit.DebitType:
			totalDebit += transaction.Amount
		case credit.CreditType:
			if strings.Contains(transaction.Source, credit.SourceSystemRevertEvent) {
				totalDebit -= transaction.Amount
			}
		}
	}
	return totalDebit, nil
}

// GetBalanceForRange returns the balance of the account in the given range.
// start time is inclusive and end time is exclusive.
func (r BillingTransactionRepository) GetBalanceForRange(ctx context.Context, accountID string, start time.Time,
	end time.Time) (int64, error) {
	var amount int64
	if err := r.dbc.WithTxn(ctx, sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}, func(tx *sqlx.Tx) error {
		var err error
		amount, err = r.getBalanceInTx(ctx, tx, accountID, &start, &end)
		return err
	}); err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return amount, nil
}
