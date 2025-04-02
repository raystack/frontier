package checkout

import (
	"errors"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type State string

const (
	StatePending  State = "pending"
	StateExpired  State = "expired"
	StateComplete State = "complete"
)

func (s State) String() string {
	return string(s)
}

var (
	ErrNotFound      = errors.New("checkout not found")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidID     = errors.New("invalid checkout id")
	ErrInvalidDetail = errors.New("invalid checkout detail")
	ErrKycCompleted  = errors.New("organization kyc completed")
)

type Checkout struct {
	ID         string
	ProviderID string // identifier set by the billing engine provider
	CustomerID string

	PlanID           string // uuid of plan if resource type is subscription
	SkipTrial        bool   // if set, no trial period
	CancelAfterTrial bool   // if set, cancel subscription after trial period
	ProviderCouponID string // coupon identifier set by the billing engine provider
	ProductID        string
	Quantity         int64 // product quantity if any

	// CancelUrl is the URL to which provider sends customers when payment is canceled
	CancelUrl string
	// SuccessUrl is the URL to which provider sends customers when payment is complete
	SuccessUrl string
	// CheckoutUrl is the URL to which provider sends customers to finish payment
	CheckoutUrl string

	State         string
	PaymentStatus string

	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpireAt  time.Time
}

type Filter struct {
	CustomerID string
	ProviderID string
}
