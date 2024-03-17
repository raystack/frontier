package credit

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound       = errors.New("transaction not found")
	ErrInvalidUUID    = errors.New("invalid syntax of uuid")
	ErrInvalidID      = errors.New("invalid transaction id")
	ErrInvalidDetail  = errors.New("invalid transaction detail")
	ErrNotEnough      = errors.New("not enough credits")
	ErrAlreadyApplied = errors.New("credits already applied")

	// TxNamespaceUUID is the namespace for generating transaction UUIDs deterministically
	TxNamespaceUUID = uuid.MustParse("967416d0-716e-4308-b58f-2468ac14f20a")

	SourceSystemBuyEvent     = "system.buy"
	SourceSystemOnboardEvent = "system.starter"
)

type TransactionType string

const (
	TypeDebit  TransactionType = "debit"
	TypeCredit TransactionType = "credit"
)

type Transaction struct {
	ID        string
	AccountID string
	Amount    int64
	Type      TransactionType

	// Source is the source app or event that caused the transaction
	Source      string
	Description string

	// UserID is the user who initiated the transaction
	UserID string

	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Credit struct {
	ID          string
	AccountID   string
	Amount      int64
	UserID      string
	Source      string
	Description string

	Metadata metadata.Metadata
}

type Filter struct {
	AccountID string
	Since     time.Time
}
