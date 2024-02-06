package subscription

import (
	"errors"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound      = errors.New("subscription not found")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidID     = errors.New("invalid subscription id")
	ErrInvalidDetail = errors.New("invalid subscription detail")
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	StateActive  State = "active"
	StatePastDue State = "past_due"
)

type Phase struct {
	EffectiveAt time.Time
	PlanID      string
}

type Subscription struct {
	ID         string
	ProviderID string // identifier set by the billing engine provider
	CustomerID string
	PlanID     string

	State string

	Metadata metadata.Metadata

	Phase Phase

	CreatedAt   time.Time
	UpdatedAt   time.Time
	CanceledAt  time.Time
	DeletedAt   time.Time
	EndedAt     time.Time
	TrialEndsAt time.Time
}

type Filter struct {
	CustomerID string
	PlanID     string
	State      string
}
