package plan

import (
	"errors"
	"time"

	"github.com/raystack/frontier/billing/product"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound      = errors.New("plan not found")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidName   = errors.New("plan name is invalid")
	ErrInvalidDetail = errors.New("invalid plan detail")
	// ErrPlanInactive is the single sentinel for "a retired plan cannot take a new
	// subscription". checkout and subscription both surface it, so a handler maps
	// one error to one client code.
	ErrPlanInactive = errors.New("plan is inactive and cannot be subscribed to")
)

const (
	StateActive   = "active"
	StateInactive = "inactive"
	// StateAll is a Filter.State sentinel that matches plans in any state. An
	// empty Filter.State defaults to active, so this is the way to list every plan.
	StateAll = "all"
)

// IsInactive reports whether the plan is retired: hidden from ListPlans and
// closed to new subscriptions. Existing subscriptions on it keep working.
func (p Plan) IsInactive() bool {
	return p.State == StateInactive
}

// Plan is a collection of products
// it is a logical grouping of products and doesn't have
// a corresponding billing engine entity
type Plan struct {
	ID string `json:"id" yaml:"id"`

	Name        string            `json:"name" yaml:"name"`   // a machine friendly name for the feature
	Title       string            `json:"title" yaml:"title"` // a human friendly title
	Description string            `json:"description" yaml:"description"`
	Metadata    metadata.Metadata `json:"metadata" yaml:"metadata"`

	// Interval is the interval at which the plan is billed
	// e.g. day, week, month, year
	// This is just used to group related product prices and has no
	// immediate effect on the billing engine
	Interval string `json:"interval" yaml:"interval"`

	// OnStartCredits is the number of credits that are awarded when a subscription is started
	OnStartCredits int64 `json:"on_start_credits" yaml:"on_start_credits"`

	// Products for the plan, return only, should not be set when creating a plan
	Products []product.Product `json:"products" yaml:"products"`

	// TrialDays is the number of days a subscription is in trial
	TrialDays int64 `json:"trial_days" yaml:"trial_days"`

	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (p Plan) GetUserSeatProduct() (product.Product, bool) {
	for _, f := range p.Products {
		if f.Behavior == product.PerSeatBehavior {
			return f, true
		}
	}
	return product.Product{}, false
}

func (p Plan) IsFree() bool {
	for _, prod := range p.Products {
		for _, price := range prod.Prices {
			// an inactive price is not charged, so it must not make the plan paid
			if !price.IsActive() {
				continue
			}
			if price.Amount > 0 {
				return false
			}
		}
	}
	return true
}

type Filter struct {
	IDs      []string
	Interval string
	State    string
}

type File struct {
	Plans    []Plan            `json:"plans" yaml:"plans"`
	Products []product.Product `json:"products" yaml:"products"`
	Features []product.Feature `json:"features" yaml:"features"`
}
