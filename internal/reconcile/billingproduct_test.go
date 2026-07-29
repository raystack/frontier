package reconcile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// newBillingProduct returns a fresh valid product spec for tests to mutate.
func newBillingProduct() BillingProductSpec {
	return BillingProductSpec{
		Name:     "token",
		Title:    "Tokens",
		Behavior: "credits",
		Config:   BillingProductConfig{CreditAmount: 1, MinQuantity: 1, MaxQuantity: 100000},
		Prices:   []BillingPriceSpec{{Name: "default", Amount: 100, Currency: "usd", Interval: "month"}},
		Features: []BillingFeatureRef{{Name: "f1"}},
	}
}

// curToken is the current server state matching newBillingProduct, with the
// price fields the server fills in on create.
func curToken() currentBillingProduct {
	return currentBillingProduct{
		ID:       "p1",
		Name:     "token",
		Title:    "Tokens",
		Behavior: "credits",
		Config:   BillingProductConfig{CreditAmount: 1, MinQuantity: 1, MaxQuantity: 100000},
		Prices:   []BillingPriceSpec{{Name: "default", Amount: 100, Currency: "usd", Interval: "month", UsageType: "licensed", BillingScheme: "flat"}},
		Features: []string{"f1"},
	}
}

func TestValidateBillingProductSpec(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*BillingProductSpec)
		wantErr string
	}{
		{"valid", func(*BillingProductSpec) {}, ""},
		{"missing name", func(s *BillingProductSpec) { s.Name = "" }, "name is required"},
		{"delete is rejected", func(s *BillingProductSpec) { s.Delete = true }, "cannot be deleted"},
		{"unknown behavior", func(s *BillingProductSpec) { s.Behavior = "weird" }, "unknown behavior"},
		{"empty price name", func(s *BillingProductSpec) { s.Prices[0].Name = "" }, "price with no name"},
		{"negative amount", func(s *BillingProductSpec) { s.Prices[0].Amount = -1 }, "negative amount"},
		{"unknown interval", func(s *BillingProductSpec) { s.Prices[0].Interval = "fortnight" }, "unknown interval"},
		{"unknown usage type", func(s *BillingProductSpec) { s.Prices[0].UsageType = "weird" }, "unknown usage type"},
		{"unknown billing scheme", func(s *BillingProductSpec) { s.Prices[0].BillingScheme = "weird" }, "unknown billing scheme"},
		{"duplicate price name", func(s *BillingProductSpec) {
			s.Prices = append(s.Prices, BillingPriceSpec{Name: "default", Amount: 200})
		}, "more than once"},
		{"duplicate feature", func(s *BillingProductSpec) {
			s.Features = []BillingFeatureRef{{Name: "f"}, {Name: "f"}}
		}, "more than once"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newBillingProduct()
			c.mutate(&s)
			err := validateBillingProductSpec(s)
			if c.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, c.wantErr)
		})
	}
}

func TestDiffBillingProducts(t *testing.T) {
	t.Run("adds a product the server does not have", func(t *testing.T) {
		desired := []BillingProductSpec{
			newBillingProduct(),
			{Name: "seat", Behavior: "per_seat", Prices: []BillingPriceSpec{{Name: "monthly", Amount: 15000, Currency: "usd", Interval: "month"}}},
		}
		ops, err := diffBillingProducts(desired, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opAdd, ops[0].action)
			assert.Equal(t, "seat", ops[0].spec.Name)
		}
	})

	t.Run("is a no-op when the product already matches", func(t *testing.T) {
		ops, err := diffBillingProducts([]BillingProductSpec{newBillingProduct()}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("updates a product whose title differs", func(t *testing.T) {
		s := newBillingProduct()
		s.Title = "Renamed tokens"
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opUpdate, ops[0].action)
			assert.Equal(t, "p1", ops[0].id)
			assert.Contains(t, ops[0].detail, "title")
		}
	})

	t.Run("updates when a price is added", func(t *testing.T) {
		s := newBillingProduct()
		s.Prices = append(s.Prices, BillingPriceSpec{Name: "yearly", Amount: 1000, Currency: "usd", Interval: "year"})
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Contains(t, ops[0].detail, "prices")
		}
	})

	t.Run("updates when features differ", func(t *testing.T) {
		s := newBillingProduct()
		s.Features = []BillingFeatureRef{{Name: "f1"}, {Name: "f2"}}
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Contains(t, ops[0].detail, "features")
		}
	})

	t.Run("fails when a server product is missing from the file", func(t *testing.T) {
		_, err := diffBillingProducts(nil, []currentBillingProduct{curToken()})
		assert.ErrorContains(t, err, "not in the file")
	})

	t.Run("rejects a delete flag", func(t *testing.T) {
		s := newBillingProduct()
		s.Delete = true
		_, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.ErrorContains(t, err, "cannot be deleted")
	})

	t.Run("rejects a product listed more than once", func(t *testing.T) {
		_, err := diffBillingProducts([]BillingProductSpec{newBillingProduct(), newBillingProduct()}, []currentBillingProduct{curToken()})
		assert.ErrorContains(t, err, "listed more than once")
	})
}
