package product

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Behavior string

func (b Behavior) String() string {
	return string(b)
}

const (
	BasicBehavior     Behavior = "basic"
	CreditBehavior    Behavior = "credits"
	UserCountBehavior Behavior = "per_seat"
)

// Product is an item being sold by the platform and has a corresponding reference
// in the billing engine
type Product struct {
	ID         string   `json:"id" yaml:"id"`
	ProviderID string   `json:"provider_id" yaml:"provider_id"` // in case of stripe, provider id and id are same
	PlanIDs    []string // plans this feature belongs to, this is optional and can be empty

	Name        string `json:"name" yaml:"name"`   // a machine friendly name for the feature
	Title       string `json:"title" yaml:"title"` // a human friendly title for the feature
	Description string `json:"description" yaml:"description"`

	// Type is the type of the feature
	// known types are "credits" and "per_seat". Default is "basic"
	Behavior Behavior `json:"behavior" yaml:"behavior" default:"basic"`

	// CreditAmount is amount of credits that are awarded/consumed when buying/using this feature
	CreditAmount int64 `json:"credit_amount" yaml:"credit_amount"`

	// Prices for the product, return only, shouldn't be set while updating a product
	Prices []Price `json:"prices" yaml:"prices"`
	// Features for the product, return only, shouldn't be set while updating a product
	Features []Feature `json:"features" yaml:"features"`

	State    string            `json:"state" yaml:"state"`
	Metadata metadata.Metadata `json:"metadata" yaml:"metadata"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type PriceUsageType string

const (
	PriceUsageTypeLicensed PriceUsageType = "licensed"
	PriceUsageTypeMetered  PriceUsageType = "metered"
)

func BuildPriceUsageType(s string) PriceUsageType {
	switch s {
	case "licensed":
		return PriceUsageTypeLicensed
	case "metered":
		return PriceUsageTypeMetered
	}
	return PriceUsageTypeLicensed
}

func (p PriceUsageType) ToStripe() string {
	switch p {
	case PriceUsageTypeLicensed:
		return "licensed"
	case PriceUsageTypeMetered:
		return "metered"
	}
	return ""
}

type BillingScheme string

const (
	BillingSchemeFlat   BillingScheme = "flat"
	BillingSchemeTiered BillingScheme = "tiered"
)

func BuildBillingScheme(s string) BillingScheme {
	switch s {
	case "flat":
		return BillingSchemeFlat
	case "tiered":
		return BillingSchemeTiered
	}
	return BillingSchemeFlat
}

func (b BillingScheme) ToStripe() string {
	switch b {
	case BillingSchemeFlat:
		return "per_unit"
	case BillingSchemeTiered:
		return "tiered"
	}
	return ""
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
	ID         string `json:"id" yaml:"id"`
	ProductID  string `json:"feature_id" yaml:"feature_id"`
	ProviderID string `json:"provider_id" yaml:"provider_id"`

	Name string `json:"name" yaml:"name"` // a machine friendly name for the price

	// BillingScheme specifies the billing scheme for the price
	// known schemes are "tiered" and "flat". Default is "flat"
	BillingScheme BillingScheme `json:"billing_scheme" yaml:"billing_scheme" default:"flat"`

	// Currency Three-letter ISO 4217 currency code in lower case
	// like "usd", "eur", "gbp"
	// https://www.six-group.com/en/products-services/financial-information/data-standards.html
	Currency string `json:"currency" yaml:"currency" default:"usd"`

	// Amount price in the minor currency unit
	// Minor unit is the smallest unit of a currency, e.g. 1 dollar equals 100 cents (with 2 decimals).
	Amount int64 `json:"amount" yaml:"amount"`

	// UsageType specifies the usage type for the price
	// known types are "licensed" and "metered". Default is "licensed"
	UsageType PriceUsageType `json:"usage_type" yaml:"usage_type" default:"licensed"`

	// MeteredAggregate specifies the aggregation method for the price
	// known aggregations are "sum", "last_during_period", "last_ever" and "max". Default is "sum"
	MeteredAggregate string `json:"metered_aggregate" yaml:"metered_aggregate" default:"sum"`

	// Interval is the interval at which the plan is billed
	// e.g. day, week, month, year
	Interval string `json:"interval" yaml:"interval"`

	Metadata metadata.Metadata `json:"metadata" yaml:"metadata"`

	// TierMode specifies the tier mode for the price
	// known modes are "graduated" and "volume". Default is "graduated"
	// In volume-based, the maximum quantity within a period determines the per-unit price
	// In graduated, pricing changes as the quantity increases to specific thresholds
	TierMode string `json:"tier_mode" yaml:"tier_mode" default:"graduated"`

	// Tiers specifies the optional tiers for the price
	// only applicable when BillingScheme is "tiered"
	Tiers []Tier `json:"tiers" yaml:"tiers"`

	State     string `json:"state" yaml:"state"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Tier struct {
	FlatAmount int64
	UpTo       int64
}

// Feature are part of a product which allows for a more granular control on
// what is packed with the product. It is a platform specific concept and
// doesn't have a corresponding billing engine entity
type Feature struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"` // a human friendly name for the feature

	// products this feature belongs to, this is optional and can be empty
	// a product will have at least one feature with the same name as the product
	ProductIDs []string `json:"product_ids" yaml:"product_ids"`

	Metadata  metadata.Metadata `json:"metadata" yaml:"metadata"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
