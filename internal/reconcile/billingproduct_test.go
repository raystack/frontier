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
		{"name too short", func(s *BillingProductSpec) { s.Name = "ab" }, "at least three characters"},
		{"delete is rejected", func(s *BillingProductSpec) { s.Delete = true }, "cannot be deleted"},
		{"empty price name", func(s *BillingProductSpec) { s.Prices[0].Name = "" }, "price with no name"},
		{"negative amount", func(s *BillingProductSpec) { s.Prices[0].Amount = -1 }, "negative amount"},
		{"tiered scheme rejected", func(s *BillingProductSpec) { s.Prices[0].BillingScheme = "tiered" }, "does not support"},
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

	// The server merges rather than full-writes, so the diff must not plan
	// changes the server would silently drop, or it would loop forever.
	t.Run("does not plan a change for an omitted title or zero credit amount", func(t *testing.T) {
		s := newBillingProduct()
		s.Title = ""              // server keeps its title when the file omits it
		s.Config.CreditAmount = 0 // server keeps credit_amount when it is zero
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("does not plan a behavior change", func(t *testing.T) {
		s := newBillingProduct()
		s.Behavior = "basic" // differs from server "credits", but behavior is create-only
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("does not plan a price change for an empty price list", func(t *testing.T) {
		s := newBillingProduct()
		s.Prices = nil // the server ignores an empty list, so dropping prices is not plannable
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("does not plan a change for a currency case difference", func(t *testing.T) {
		s := newBillingProduct()
		s.Prices[0].Currency = "USD" // server stored "usd"
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("does not plan a change for an interval or usage-type case difference", func(t *testing.T) {
		s := newBillingProduct()
		s.Prices[0].Interval = "Month"     // server stored "month"
		s.Prices[0].UsageType = "Licensed" // server stored "licensed"
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("fails the plan when a retired price name is reused with different fields", func(t *testing.T) {
		cur := curToken()
		cur.RetiredPrices = []BillingPriceSpec{{Name: "promo", Amount: 50, Currency: "usd", Interval: "month", UsageType: "licensed", BillingScheme: "flat"}}
		s := newBillingProduct()
		s.Prices = append(s.Prices, BillingPriceSpec{Name: "promo", Amount: 99, Currency: "usd", Interval: "month"})
		_, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{cur})
		assert.ErrorContains(t, err, "retired")
	})

	t.Run("plans a reactivation when a retired price name is reused with the same fields", func(t *testing.T) {
		cur := curToken()
		cur.RetiredPrices = []BillingPriceSpec{{Name: "promo", Amount: 50, Currency: "usd", Interval: "month", UsageType: "licensed", BillingScheme: "flat"}}
		s := newBillingProduct()
		s.Prices = append(s.Prices, BillingPriceSpec{Name: "promo", Amount: 50, Currency: "usd", Interval: "month"})
		ops, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{cur})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Contains(t, ops[0].detail, "prices")
		}
	})

	t.Run("fails the plan on an in-place price amount change", func(t *testing.T) {
		s := newBillingProduct()
		s.Prices[0].Amount = 200 // same name "default", different amount
		_, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
		assert.ErrorContains(t, err, "immutable")
	})

	t.Run("fails the plan deterministically when an add and an immutable change coexist", func(t *testing.T) {
		s := newBillingProduct()
		s.Prices[0].Amount = 200 // an immutable change on the existing "default"
		s.Prices = append(s.Prices, BillingPriceSpec{Name: "yearly", Amount: 1000, Currency: "usd", Interval: "year"})
		// map iteration order must not change the outcome: the immutable check
		// runs over every name before any add is considered, so this always fails.
		for i := 0; i < 25; i++ {
			_, err := diffBillingProducts([]BillingProductSpec{s}, []currentBillingProduct{curToken()})
			assert.ErrorContains(t, err, "immutable")
		}
	})
}
