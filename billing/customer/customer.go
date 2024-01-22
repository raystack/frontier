package customer

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Provider string

const (
	ProviderStripe Provider = "stripe"
)

type Customer struct {
	ID         string
	OrgID      string
	ProviderID string // identifier set by the billing engine provider

	Name    string
	Email   string
	Phone   string
	Address Address
	// Currency Three-letter ISO 4217 currency code in lower case
	Currency string `default:"usd"`
	Metadata metadata.Metadata

	// Stripe specific fields
	// StripeTestClockID is used for testing purposes only to simulate a subscription
	StripeTestClockID *string

	State     string
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

type Filter struct {
	OrgID string
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
