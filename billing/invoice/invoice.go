package invoice

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/pagination"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound      = fmt.Errorf("invoice not found")
	ErrInvalidDetail = fmt.Errorf("invalid invoice detail")
	ErrBadInput      = fmt.Errorf("invalid input")
)

const (
	// ItemIDMetadataKey is used to uniquely identify the item in the invoice
	// this is useful when reconciling the invoice items for payments and
	// avoid creating duplicate credits
	ItemIDMetadataKey = "item_id"
	// ItemTypeMetadataKey is used to identify the item type in the invoice
	// this is useful when reconciling the invoice items for payments and
	// avoid creating duplicate invoices
	ItemTypeMetadataKey = "item_type"

	// ReconciledMetadataKey marks the invoice as reconciled and avoid processing it again
	// as an optimization
	ReconciledMetadataKey = "reconciled"

	// GenerateForCreditLockKey is used to lock the invoice generation within current application
	GenerateForCreditLockKey = "generate_for_credit"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	DraftState State = "draft"
	OpenState  State = "open"
	PaidState  State = "paid"
)

type Invoice struct {
	ID         string
	CustomerID string
	ProviderID string
	// State could be one of draft, open, paid, uncollectible, void
	State         State
	Currency      string
	Amount        int64
	HostedURL     string
	DueAt         time.Time
	EffectiveAt   time.Time
	CreatedAt     time.Time
	PeriodStartAt time.Time
	PeriodEndAt   time.Time

	Items    []Item
	Metadata metadata.Metadata
}

type InvoiceWithOrganization struct {
	ID          string
	Amount      int64
	Currency    string
	State       State
	InvoiceLink string
	CreatedAt   time.Time
	OrgID       string
	OrgName     string
	OrgTitle    string
}

type ItemType string

func (t ItemType) String() string {
	return string(t)
}

const (
	// CreditItemType is used to charge for the credits used in the system
	// as overdraft
	CreditItemType ItemType = "credit"
)

type Item struct {
	ID         string `json:"id"`
	ProviderID string `json:"provider_id"`
	// Name is the item name
	Name string `json:"name"`
	// Type is the item type
	Type ItemType `json:"type"`
	// UnitAmount is per unit cost
	UnitAmount int64 `json:"unit_amount"`
	// Quantity is the number of units
	Quantity int64 `json:"quantity"`

	// TimeRangeStart is the start time of the item since it's effective
	TimeRangeStart *time.Time `json:"range_start"`
	// TimeRangeEnd is the end time of the item since it's effective
	TimeRangeEnd *time.Time `json:"range_end"`
}

type Filter struct {
	CustomerID  string
	NonZeroOnly bool
	State       State

	Pagination *pagination.Pagination
}
