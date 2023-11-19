package credit

import (
	"errors"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound       = errors.New("transaction not found")
	ErrInvalidUUID    = errors.New("invalid syntax of uuid")
	ErrInvalidID      = errors.New("invalid transaction id")
	ErrInvalidDetail  = errors.New("invalid transaction detail")
	ErrNotEnough      = errors.New("not enough credits")
	ErrAlreadyApplied = errors.New("credits already applied")
)

type TransactionType string

const (
	TypeDebit  TransactionType = "debit"
	TypeCredit TransactionType = "credit"
)

type Transaction struct {
	ID          string
	AccountID   string
	Amount      int64
	Type        TransactionType
	Source      string
	Description string
	Metadata    metadata.Metadata
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Credit struct {
	ID          string
	AccountID   string
	Amount      int64
	Description string

	Metadata metadata.Metadata
}

type Filter struct {
	AccountID string
	Since     time.Time
}
