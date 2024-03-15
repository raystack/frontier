package subscription

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound          = fmt.Errorf("subscription not found")
	ErrInvalidUUID       = fmt.Errorf("invalid syntax of uuid")
	ErrInvalidID         = fmt.Errorf("invalid subscription id")
	ErrInvalidDetail     = fmt.Errorf("invalid subscription detail")
	ErrAlreadyOnSamePlan = fmt.Errorf("already on the same plan")
	ErrNoPhaseActive     = fmt.Errorf("no phase active")
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
	EndsAt      time.Time
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

	Phase       Phase
	PlanHistory []Phase

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
