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
	StateActive   State = "active"
	StateTrialing State = "trialing"
	StatePastDue  State = "past_due"
)

type Phase struct {
	EffectiveAt time.Time
	PlanID      string
}

type ChangeRequest struct {
	PlanID    string
	Immediate bool

	CancelUpcoming bool
}

type Subscription struct {
	ID         string
	ProviderID string // identifier set by the billing engine provider
	CustomerID string
	PlanID     string

	State string

	Metadata metadata.Metadata

	Phase Phase

	CreatedAt            time.Time
	UpdatedAt            time.Time
	CanceledAt           time.Time
	DeletedAt            time.Time
	EndedAt              time.Time
	TrialEndsAt          time.Time
	CurrentPeriodStartAt time.Time
	CurrentPeriodEndAt   time.Time
	BillingCycleAnchorAt time.Time
}

func (s Subscription) IsActive() bool {
	return State(s.State) == StateActive || State(s.State) == StateTrialing
}

type Filter struct {
	CustomerID string
	PlanID     string
	State      string
}
