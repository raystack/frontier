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

// on_start_credits and trial_days are bounded non-negative by the proto, so a
// negative value could never round-trip through reconcile export either. Create
// rejects a negative value before it touches the database, for the same boot-time
// loader reason as the interval guard.
func TestBillingPlanRepository_Create_RejectsNegativeCreditsAndTrial(t *testing.T) {
	r := postgres.NewBillingPlanRepository(nil)

	t.Run("negative on_start_credits", func(t *testing.T) {
		_, err := r.Create(context.Background(), plan.Plan{Name: "neg_credits", Interval: "month", OnStartCredits: -5})
		require.Error(t, err)
		require.Contains(t, err.Error(), "on_start_credits")
	})
	t.Run("negative trial_days", func(t *testing.T) {
		_, err := r.Create(context.Background(), plan.Plan{Name: "neg_trial", Interval: "month", TrialDays: -3})
		require.Error(t, err)
		require.Contains(t, err.Error(), "trial_days")
	})
}
