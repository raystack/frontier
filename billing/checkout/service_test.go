package checkout

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/stretchr/testify/require"
)

// The fakes embed the service interfaces so only the methods Create reaches
// before the inactive-plan gate need real behavior; any other call would panic,
// which keeps the test honest about what the gate depends on.
type fakeCustomerSvc struct{ CustomerService }

func (fakeCustomerSvc) RegisterToProviderIfRequired(_ context.Context, id string) (customer.Customer, error) {
	return customer.Customer{ID: id, OrgID: "org-1"}, nil
}

type fakeAuthnSvc struct{ AuthnService }

func (fakeAuthnSvc) GetPrincipal(_ context.Context, _ ...authenticate.ClientAssertion) (authenticate.Principal, error) {
	return authenticate.Principal{}, nil
}

type fakePlanSvc struct {
	PlanService
	p plan.Plan
}

func (f fakePlanSvc) GetByID(_ context.Context, _ string) (plan.Plan, error) {
	return f.p, nil
}

type fakeSubscriptionSvc struct{ SubscriptionService }

func (fakeSubscriptionSvc) List(_ context.Context, _ subscription.Filter) ([]subscription.Subscription, error) {
	return nil, nil
}

// A retired (inactive) plan is closed to new checkouts: Create must reject it
// with ErrPlanInactive instead of building a subscription. The subscription
// ChangePlan path has its own test; this covers the checkout path.
func TestService_Create_RejectsInactivePlan(t *testing.T) {
	s := &Service{
		log:                 slog.New(slog.NewTextHandler(io.Discard, nil)),
		customerService:     fakeCustomerSvc{},
		authnService:        fakeAuthnSvc{},
		planService:         fakePlanSvc{p: plan.Plan{ID: "plan-1", Name: "retired", State: plan.StateInactive}},
		subscriptionService: fakeSubscriptionSvc{},
	}

	_, err := s.Create(context.Background(), Checkout{CustomerID: "cust-1", PlanID: "plan-1"})

	require.ErrorIs(t, err, plan.ErrPlanInactive)
}
