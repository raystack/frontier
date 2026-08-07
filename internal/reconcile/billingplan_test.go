package reconcile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

// newBillingPlan returns a fresh valid plan spec for tests to mutate.
func newBillingPlan() BillingPlanSpec {
	return BillingPlanSpec{
		Name:           "starter",
		Title:          "Starter",
		Description:    "Starter plan",
		Interval:       "month",
		OnStartCredits: 100,
		TrialDays:      14,
		State:          "active",
		Products:       []BillingPlanProductRef{{Name: "starter_product"}},
	}
}

// curStarter is the current server state matching newBillingPlan.
func curStarter() currentBillingPlan {
	return currentBillingPlan{
		ID:             "p1",
		Name:           "starter",
		Title:          "Starter",
		Description:    "Starter plan",
		Interval:       "month",
		OnStartCredits: 100,
		TrialDays:      14,
		State:          "active",
		Products:       []string{"starter_product"},
	}
}

func TestValidateBillingPlanSpec(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*BillingPlanSpec)
		wantErr string
	}{
		{"valid", func(*BillingPlanSpec) {}, ""},
		{"missing name", func(s *BillingPlanSpec) { s.Name = "" }, "name is required"},
		{"name too short", func(s *BillingPlanSpec) { s.Name = "ab" }, "at least three characters"},
		{"empty title", func(s *BillingPlanSpec) { s.Title = "" }, "must have a title"},
		{"delete is rejected", func(s *BillingPlanSpec) { s.Delete = true }, "cannot be deleted"},
		{"empty product name", func(s *BillingPlanSpec) { s.Products[0].Name = "" }, "product with no name"},
		{"duplicate product", func(s *BillingPlanSpec) {
			s.Products = []BillingPlanProductRef{{Name: "p"}, {Name: "p"}}
		}, "more than once"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newBillingPlan()
			c.mutate(&s)
			err := validateBillingPlanSpec(s)
			if c.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, c.wantErr)
		})
	}
}

// The docs and the diff both treat an omitted state as active, so a file that
// leaves state out is valid and must pass the up-front check. Before the fix the
// create and update bodies sent an empty state, which the proto's state enum
// rejected, so a valid file failed at Validate even though the diff and the store
// would have defaulted it to active.
func TestValidateBillingPlanRequest_DefaultsOmittedStateToActive(t *testing.T) {
	s := newBillingPlan()
	s.State = "" // omitted

	t.Run("create", func(t *testing.T) {
		assert.NoError(t, validateBillingPlanRequest(billingPlanOp{action: opAdd, spec: s}))
	})
	t.Run("update", func(t *testing.T) {
		assert.NoError(t, validateBillingPlanRequest(billingPlanOp{action: opUpdate, id: "p1", spec: s}))
	})
}

// Validate is the up-front, server-free check Run calls on every document before
// anything applies. A hand-written file that omits state relies on the documented
// default of active, so Validate must accept it.
func TestBillingPlanReconciler_Validate_AcceptsOmittedState(t *testing.T) {
	spec := []byte("- name: standard_monthly\n  title: Standard\n  interval: month\n  products:\n    - name: prod_a\n")
	r := NewBillingPlanReconciler(nil, "")
	assert.NoError(t, r.Validate(spec))
}

func TestDiffBillingPlans(t *testing.T) {
	t.Run("adds a plan the server does not have", func(t *testing.T) {
		desired := []BillingPlanSpec{
			newBillingPlan(),
			{Name: "pro", Title: "Pro", Interval: "year", State: "active", Products: []BillingPlanProductRef{{Name: "pro_product"}}},
		}
		ops, err := diffBillingPlans(desired, []currentBillingPlan{curStarter()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opAdd, ops[0].action)
			assert.Equal(t, "pro", ops[0].spec.Name)
		}
	})

	t.Run("is a no-op when the plan already matches", func(t *testing.T) {
		ops, err := diffBillingPlans([]BillingPlanSpec{newBillingPlan()}, []currentBillingPlan{curStarter()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("updates a plan whose title differs", func(t *testing.T) {
		s := newBillingPlan()
		s.Title = "Renamed"
		ops, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opUpdate, ops[0].action)
			assert.Equal(t, "p1", ops[0].id)
			assert.Contains(t, ops[0].detail, "title")
		}
	})

	t.Run("updates when the state changes to inactive", func(t *testing.T) {
		s := newBillingPlan()
		s.State = "inactive"
		ops, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Contains(t, ops[0].detail, "state")
		}
	})

	t.Run("does not plan a change for a state case difference", func(t *testing.T) {
		s := newBillingPlan()
		s.State = "Active" // server stored "active"
		ops, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("treats an empty file state as active", func(t *testing.T) {
		s := newBillingPlan()
		s.State = "" // omitted; the server default is active
		ops, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("carries the current metadata onto an update so it is preserved", func(t *testing.T) {
		cur := curStarter()
		cur.Metadata, _ = structpb.NewStruct(map[string]any{"plan_group_id": "starter"})
		s := newBillingPlan()
		s.Title = "Renamed"
		ops, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{cur})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, cur.Metadata, ops[0].metadata)
		}
	})

	t.Run("fails the plan when the interval changes", func(t *testing.T) {
		s := newBillingPlan()
		s.Interval = "year"
		_, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.ErrorContains(t, err, "interval cannot change")
	})

	t.Run("fails the plan when the product set changes", func(t *testing.T) {
		s := newBillingPlan()
		s.Products = []BillingPlanProductRef{{Name: "different_product"}}
		_, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.ErrorContains(t, err, "products cannot change")
	})

	t.Run("does not plan a change for a product-name case difference", func(t *testing.T) {
		s := newBillingPlan()
		s.Products = []BillingPlanProductRef{{Name: "Starter_Product"}} // server stored "starter_product"
		ops, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("fails when a server plan is missing from the file", func(t *testing.T) {
		_, err := diffBillingPlans(nil, []currentBillingPlan{curStarter()})
		assert.ErrorContains(t, err, "not in the file")
	})

	t.Run("rejects a delete flag", func(t *testing.T) {
		s := newBillingPlan()
		s.Delete = true
		_, err := diffBillingPlans([]BillingPlanSpec{s}, []currentBillingPlan{curStarter()})
		assert.ErrorContains(t, err, "cannot be deleted")
	})

	t.Run("rejects a plan listed more than once", func(t *testing.T) {
		_, err := diffBillingPlans([]BillingPlanSpec{newBillingPlan(), newBillingPlan()}, []currentBillingPlan{curStarter()})
		assert.ErrorContains(t, err, "listed more than once")
	})
}
