package postgres_test

import (
	"context"
	"testing"

	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/stretchr/testify/require"
)

// A plan's interval is set only at creation and has no default. A plan with no
// interval could never round-trip through reconcile export, so the store must
// never hold one. Create rejects an empty interval before it touches the
// database, which is why a nil client is safe here: the guard short-circuits
// first. The connect API already blocks this, but the boot-time plan loader
// writes through this repository directly and bypasses that check.
func TestBillingPlanRepository_Create_RejectsEmptyInterval(t *testing.T) {
	r := postgres.NewBillingPlanRepository(nil)

	_, err := r.Create(context.Background(), plan.Plan{Name: "no_interval", Interval: ""})

	require.Error(t, err)
	require.Contains(t, err.Error(), "interval")
}
