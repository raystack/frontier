package subscription

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound                       = fmt.Errorf("subscription not found")
	ErrInvalidUUID                    = fmt.Errorf("invalid syntax of uuid")
	ErrInvalidID                      = fmt.Errorf("invalid subscription id")
	ErrInvalidDetail                  = fmt.Errorf("invalid subscription detail")
	ErrAlreadyOnSamePlan              = fmt.Errorf("already on the same plan")
	ErrNoPhaseActive                  = fmt.Errorf("no phase active")
	ErrPhaseIsUpdating                = fmt.Errorf("phase is in the middle of a change, please try again later")
	ErrSubscriptionOnProviderNotFound = fmt.Errorf("failed to get subscription from billing provider")
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	StateActive   State = "active"
	StateTrialing State = "trialing"
	StatePastDue  State = "past_due"
	StateCanceled State = "canceled"
	StateEnded    State = "ended"
)

type PhaseReason string

func (s PhaseReason) String() string {
	return string(s)
}

const (
	SubscriptionCancel PhaseReason = "cancel"
	SubscriptionChange PhaseReason = "change"
)

type Phase struct {
	EffectiveAt time.Time
	EndsAt      time.Time
	PlanID      string
	Reason      string
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

func (s Subscription) IsCanceled() bool {
	return State(s.State) == StateCanceled || !s.DeletedAt.IsZero() || !s.CanceledAt.IsZero()
}

type Filter struct {
	CustomerID string
	ProviderID string
	PlanID     string
	State      string
}
