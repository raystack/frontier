package credit

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound            = errors.New("transaction not found")
	ErrInvalidUUID         = errors.New("invalid syntax of uuid")
	ErrInvalidID           = errors.New("invalid transaction id")
	ErrInvalidDetail       = errors.New("invalid transaction detail")
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrAlreadyApplied      = errors.New("credits already applied")

	// TxNamespaceUUID is the namespace for generating transaction UUIDs deterministically
	TxNamespaceUUID = uuid.MustParse("967416d0-716e-4308-b58f-2468ac14f20a")

	SourceSystemBuyEvent     = "system.buy"
	SourceSystemAwardedEvent = "system.awarded"
	SourceSystemOnboardEvent = "system.starter"
	SourceSystemRevertEvent  = "system.revert"
)

type TransactionType string

func (t TransactionType) String() string {
	return string(t)
}

const (
	DebitType  TransactionType = "debit"
	CreditType TransactionType = "credit"
)

type Transaction struct {
	ID         string
	CustomerID string
	Amount     int64
	Type       TransactionType

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
	CustomerID  string
	Amount      int64
	UserID      string
	Source      string
	Description string

	Metadata metadata.Metadata
}

type Filter struct {
	CustomerID string
	Since      time.Time
}
