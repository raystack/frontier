package customer

import (
	"context"
	"time"

	"github.com/raystack/frontier/pkg/server/consts"

	"github.com/raystack/frontier/pkg/metadata"
)

type Provider string

const (
	ProviderStripe Provider = "stripe"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	ActiveState   State = "active"
	DisabledState State = "disabled"
)

type Customer struct {
	ID         string
	OrgID      string
	ProviderID string // identifier set by the billing engine provider

	Name    string
	Email   string
	Phone   string
	Address Address
	TaxData []Tax
	// Currency Three-letter ISO 4217 currency code in lower case
	Currency string `default:"usd"`
	Metadata metadata.Metadata

	// Stripe specific fields
	// StripeTestClockID is used for testing purposes only to simulate a subscription
	StripeTestClockID *string

	State     State
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Address struct {
	City       string `json:"city"`
	Country    string `json:"country"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	PostalCode string `json:"postal_code"`
	State      string `json:"state"`
}

type Tax struct {
	// Type like "vat", "gst", "sales_tax" or if it's
	// provider specific us_ein, uk_vat, in_gst, etc
	Type string
	// ID is the tax identifier
	ID string
}

type Filter struct {
	OrgID string
	State State
}

type PaymentMethod struct {
	ID         string
	CustomerID string
	ProviderID string
	Type       string

	CardLast4       string
	CardBrand       string
	CardExpiryYear  int64
	CardExpiryMonth int64

	Metadata  metadata.Metadata
	CreatedAt time.Time
}

// GetStripeTestClockFromContext returns the stripe test clock id from the context
func GetStripeTestClockFromContext(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(consts.BillingStripeTestClockContextKey).(string)
	return u, ok
}

// SetStripeTestClockInContext sets the stripe test clock id in the context
func SetStripeTestClockInContext(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, consts.BillingStripeTestClockContextKey, s)
}
