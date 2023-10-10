package feature

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

// Feature is a product feature and has a corresponding product in the billing engine
type Feature struct {
	ID     string
	PlanID string

	Name        string // a machine friendly name for the feature
	Title       string // a human friendly title for the feature
	Description string

	// Prices for the feature, return only, should not be set when creating a feature
	Prices []Price

	State    string
	Metadata metadata.Metadata

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type PriceType string

const (
	PriceTypeRecurring PriceType = "recurring"
	PriceTypeOnetime   PriceType = "onetime"
)

type PriceUsageType string

const (
	PriceUsageTypeLicensed PriceUsageType = "licensed"
	PriceUsageTypeMetered  PriceUsageType = "metered"
)

func (p PriceUsageType) ToStripe() string {
	switch p {
	case PriceUsageTypeLicensed:
		return "licensed"
	case PriceUsageTypeMetered:
		return "metered"
	}
	return string(p)
}

type BillingScheme string

const (
	BillingSchemeFlat   BillingScheme = "flat"
	BillingSchemeTiered BillingScheme = "tiered"
)

func (b BillingScheme) ToStripe() string {
	switch b {
	case BillingSchemeFlat:
		return "per_unit"
	case BillingSchemeTiered:
		return "tiered"
	}
	return string(b)
}

type PriceTierMode string

const (
	PriceTierModeGraduated PriceTierMode = "graduated"
	PriceTierModeVolume    PriceTierMode = "volume"
)

// Price is a product price and has a corresponding price in the billing engine
// when creating a price, the feature must already exist
// when subscribing to a plan, the price must already exist
type Price struct {
	ID         string
	FeatureID  string
	ProviderID string

	Name  string // a machine friendly name for the price
	Title string // a human friendly title for the price

	// BillingScheme specifies the billing scheme for the price
	// known schemes are "tiered" and "flat". Default is "flat"
	BillingScheme BillingScheme `default:"flat"`

	// Currency Three-letter ISO 4217 currency code in lower case
	// like "usd", "eur", "gbp"
	// https://www.six-group.com/en/products-services/financial-information/data-standards.html
	Currency string `default:"usd"`

	// Amount price in the minor currency unit
	// Minor unit is the smallest unit of a currency, e.g. 1 dollar equals 100 cents (with 2 decimals).
	Amount int64

	// Type of the price, e.g. "recurring" or "onetime"
	// Default is "onetime"
	Type PriceType `default:"onetime"`

	// UsageType specifies the usage type for the price
	// known types are "licensed" and "metered". Default is "licensed"
	UsageType PriceUsageType `default:"licensed"`

	// Interval specifies the billing interval for the feature
	// known intervals are "day", "week", "month", and "year". Default is "month
	RecurringInterval string `default:"month"`

	// MeteredAggregate specifies the aggregation method for the price
	// known aggregations are "sum", "last_during_period" and "max". Default is "sum"
	MeteredAggregate string `default:"sum"`

	Metadata metadata.Metadata

	// TierMode specifies the tier mode for the price
	// known modes are "graduated" and "volume". Default is "graduated"
	// In volume-based, the maximum quantity within a period determines the per-unit price
	// In graduated, pricing changes as the quantity increases to specific thresholds
	TierMode string `default:"graduated"`

	// Tiers specifies the optional tiers for the price
	// only applicable when BillingScheme is "tiered"
	Tiers []Tier

	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Tier struct {
	FlatAmount int64
	UpTo       int64
}
