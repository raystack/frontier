package entitlement_test

import (
	"context"
	"testing"

	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
)

// A feature detached from its last product has an empty ProductIDs. Check must
// deny it: an empty ProductIDs filter would otherwise match every product in the
// store and wrongly grant the feature to any active subscriber. The ProductService
// List mock is intentionally left unset, so the test fails if Check ever calls it
// (i.e. if the guard is removed).
func TestService_Check_FeatureWithNoProducts(t *testing.T) {
	ctx := context.Background()
	s, mockSubscription, mockProduct, _, _ := mockService(t)

	mockSubscription.EXPECT().List(ctx, subscription.Filter{CustomerID: "cust-1"}).
		Return([]subscription.Subscription{
			{PlanID: "plan-1", State: subscription.StateActive.String()},
		}, nil)
	mockProduct.EXPECT().GetFeatureByID(ctx, "feat-detached").
		Return(product.Feature{ProductIDs: []string{}}, nil)

	got, err := s.Check(ctx, "cust-1", "feat-detached")

	if err != nil {
		t.Fatalf("Check() unexpected error: %v", err)
	}
	if got {
		t.Errorf("Check() = true, want false for a feature attached to no product")
	}
}
