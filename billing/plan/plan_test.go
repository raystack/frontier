package plan_test

import (
	"testing"

	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
)

func TestPlan_IsFree(t *testing.T) {
	tests := []struct {
		name string
		plan plan.Plan
		want bool
	}{
		{
			name: "a plan with no products is free",
			plan: plan.Plan{},
			want: true,
		},
		{
			name: "an active paid price is not free",
			plan: plan.Plan{Products: []product.Product{{Prices: []product.Price{{Amount: 100}}}}},
			want: false,
		},
		{
			name: "an inactive paid price does not make the plan paid",
			plan: plan.Plan{Products: []product.Product{{Prices: []product.Price{{Amount: 100, State: product.PriceStateInactive}}}}},
			want: true,
		},
		{
			name: "an active zero price is free",
			plan: plan.Plan{Products: []product.Product{{Prices: []product.Price{{Amount: 0}}}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.plan.IsFree(); got != tt.want {
				t.Errorf("IsFree() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlan_IsInactive(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  bool
	}{
		{name: "inactive is inactive", state: "inactive", want: true},
		{name: "active is not inactive", state: "active", want: false},
		{name: "empty state is not inactive", state: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (plan.Plan{State: tt.state}).IsInactive(); got != tt.want {
				t.Errorf("IsInactive() = %v, want %v", got, tt.want)
			}
		})
	}
}
