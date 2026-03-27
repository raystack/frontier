package stripeprovider

import (
	"testing"

	"github.com/raystack/frontier/billing"
	"github.com/stripe/stripe-go/v79"
)

func TestFromStripeCustomer(t *testing.T) {
	sc := &stripe.Customer{
		ID:       "cus_123",
		Deleted:  false,
		Name:     "Acme Corp",
		Email:    "billing@acme.com",
		Phone:    "+1555000",
		Currency: "usd",
		Address: &stripe.Address{
			City:       "Portland",
			Country:    "US",
			Line1:      "123 Main St",
			Line2:      "Suite 4",
			PostalCode: "97201",
			State:      "OR",
		},
		TaxIDs: &stripe.TaxIDList{
			Data: []*stripe.TaxID{
				{Type: "us_ein", Value: "12-3456789"},
			},
		},
	}

	pc := fromStripeCustomer(sc)

	assertEqual(t, "ID", pc.ID, "cus_123")
	assertEqual(t, "Deleted", pc.Deleted, false)
	assertEqual(t, "Name", pc.Name, "Acme Corp")
	assertEqual(t, "Email", pc.Email, "billing@acme.com")
	assertEqual(t, "Phone", pc.Phone, "+1555000")
	assertEqual(t, "Currency", pc.Currency, "usd")
	assertEqual(t, "Address.City", pc.Address.City, "Portland")
	assertEqual(t, "Address.Country", pc.Address.Country, "US")
	assertEqual(t, "Address.PostalCode", pc.Address.PostalCode, "97201")
	if len(pc.TaxIDs) != 1 {
		t.Fatalf("expected 1 tax ID, got %d", len(pc.TaxIDs))
	}
	assertEqual(t, "TaxIDs[0].Type", pc.TaxIDs[0].Type, "us_ein")
	assertEqual(t, "TaxIDs[0].Value", pc.TaxIDs[0].Value, "12-3456789")
}

func TestFromStripeCustomer_NilAddress(t *testing.T) {
	sc := &stripe.Customer{ID: "cus_456"}
	pc := fromStripeCustomer(sc)
	assertEqual(t, "Address.City", pc.Address.City, "")
}

func TestFromStripeCustomer_Deleted(t *testing.T) {
	sc := &stripe.Customer{ID: "cus_789", Deleted: true}
	pc := fromStripeCustomer(sc)
	assertEqual(t, "Deleted", pc.Deleted, true)
}

func TestFromStripeSubscription(t *testing.T) {
	ss := &stripe.Subscription{
		ID:                 "sub_123",
		Status:             stripe.SubscriptionStatusActive,
		CanceledAt:         1700000000,
		EndedAt:            0,
		TrialEnd:           1700100000,
		CurrentPeriodStart: 1699900000,
		CurrentPeriodEnd:   1702500000,
		BillingCycleAnchor: 1699800000,
		Livemode:           true,
		AutomaticTax:       &stripe.SubscriptionAutomaticTax{Enabled: true},
		Items: &stripe.SubscriptionItemList{
			Data: []*stripe.SubscriptionItem{
				{
					ID:       "si_1",
					Quantity: 5,
					Price: &stripe.Price{
						ID:      "price_1",
						Product: &stripe.Product{ID: "prod_1"},
						Recurring: &stripe.PriceRecurring{
							Interval: stripe.PriceRecurringIntervalMonth,
						},
					},
					Metadata: map[string]string{"key": "val"},
				},
			},
		},
		Schedule: &stripe.SubscriptionSchedule{
			ID: "sub_sched_1",
			CurrentPhase: &stripe.SubscriptionScheduleCurrentPhase{
				StartDate: 1699900000,
				EndDate:   1702500000,
			},
		},
		Metadata: map[string]string{"org_id": "org_1"},
	}

	ps := fromStripeSubscription(ss)

	assertEqual(t, "ID", ps.ID, "sub_123")
	assertEqual(t, "Status", ps.Status, "active")
	assertEqual(t, "CanceledAt", ps.CanceledAt, int64(1700000000))
	assertEqual(t, "TrialEnd", ps.TrialEnd, int64(1700100000))
	assertEqual(t, "Livemode", ps.Livemode, true)
	assertEqual(t, "AutomaticTaxEnabled", ps.AutomaticTaxEnabled, true)

	if len(ps.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(ps.Items))
	}
	assertEqual(t, "Items[0].ID", ps.Items[0].ID, "si_1")
	assertEqual(t, "Items[0].PriceID", ps.Items[0].PriceID, "price_1")
	assertEqual(t, "Items[0].ProductID", ps.Items[0].ProductID, "prod_1")
	assertEqual(t, "Items[0].Quantity", ps.Items[0].Quantity, int64(5))
	assertEqual(t, "Items[0].Interval", ps.Items[0].Interval, "month")

	if ps.Schedule == nil {
		t.Fatal("expected schedule ref")
	}
	assertEqual(t, "Schedule.ID", ps.Schedule.ID, "sub_sched_1")
	assertEqual(t, "Schedule.CurrentPhase.StartDate", ps.Schedule.CurrentPhase.StartDate, int64(1699900000))
}

func TestFromStripeSubscription_NoSchedule(t *testing.T) {
	ss := &stripe.Subscription{
		ID:     "sub_456",
		Status: stripe.SubscriptionStatusCanceled,
		Items:  &stripe.SubscriptionItemList{},
	}
	ps := fromStripeSubscription(ss)
	if ps.Schedule != nil {
		t.Fatal("expected nil schedule")
	}
}

func TestFromStripeSchedule(t *testing.T) {
	ss := &stripe.SubscriptionSchedule{
		ID:          "sub_sched_1",
		EndBehavior: stripe.SubscriptionScheduleEndBehaviorRelease,
		CurrentPhase: &stripe.SubscriptionScheduleCurrentPhase{
			StartDate: 100,
			EndDate:   200,
		},
		Phases: []*stripe.SubscriptionSchedulePhase{
			{
				StartDate:   100,
				EndDate:     200,
				Currency:    "usd",
				Description: "Phase 1",
				TrialEnd:    150,
				Metadata:    map[string]string{"plan_id": "plan_1"},
				AutomaticTax: &stripe.SubscriptionAutomaticTax{
					Enabled: true,
				},
				Items: []*stripe.SubscriptionSchedulePhaseItem{
					{
						Price: &stripe.Price{
							ID:        "price_1",
							Product:   &stripe.Product{ID: "prod_1"},
							Recurring: &stripe.PriceRecurring{Interval: stripe.PriceRecurringIntervalMonth},
						},
						Quantity: 3,
						Metadata: map[string]string{"managed_by": "frontier"},
					},
				},
			},
		},
	}

	ps := fromStripeSchedule(ss)

	assertEqual(t, "ID", ps.ID, "sub_sched_1")
	assertEqual(t, "EndBehavior", ps.EndBehavior, "release")
	assertEqual(t, "CurrentPhase.StartDate", ps.CurrentPhase.StartDate, int64(100))

	if len(ps.Phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(ps.Phases))
	}
	p := ps.Phases[0]
	assertEqual(t, "Phase.Currency", p.Currency, "usd")
	assertEqual(t, "Phase.Description", p.Description, "Phase 1")
	assertEqual(t, "Phase.TrialEnd", p.TrialEnd, int64(150))
	assertEqual(t, "Phase.AutomaticTaxEnabled", p.AutomaticTaxEnabled, true)
	assertEqual(t, "Phase.Metadata[plan_id]", p.Metadata["plan_id"], "plan_1")

	if len(p.Items) != 1 {
		t.Fatalf("expected 1 phase item, got %d", len(p.Items))
	}
	assertEqual(t, "PhaseItem.PriceID", p.Items[0].PriceID, "price_1")
	assertEqual(t, "PhaseItem.ProductID", p.Items[0].ProductID, "prod_1")
	assertEqual(t, "PhaseItem.Quantity", p.Items[0].Quantity, int64(3))
	assertEqual(t, "PhaseItem.Interval", p.Items[0].Interval, "month")
}

func TestFromStripeCheckoutSession(t *testing.T) {
	cs := &stripe.CheckoutSession{
		ID:            "cs_123",
		URL:           "https://checkout.stripe.com/pay/cs_123",
		Status:        stripe.CheckoutSessionStatusOpen,
		PaymentStatus: stripe.CheckoutSessionPaymentStatusUnpaid,
		ExpiresAt:     1700000000,
		AmountTotal:   9900,
		Currency:      "usd",
		Subscription:  &stripe.Subscription{ID: "sub_123"},
		LineItems: &stripe.LineItemList{
			Data: []*stripe.LineItem{
				{
					Quantity: 2,
					Price:    &stripe.Price{Product: &stripe.Product{ID: "prod_1"}},
				},
			},
		},
	}

	pcs := fromStripeCheckoutSession(cs)

	assertEqual(t, "ID", pcs.ID, "cs_123")
	assertEqual(t, "URL", pcs.URL, "https://checkout.stripe.com/pay/cs_123")
	assertEqual(t, "Status", pcs.Status, "open")
	assertEqual(t, "PaymentStatus", pcs.PaymentStatus, "unpaid")
	assertEqual(t, "AmountTotal", pcs.AmountTotal, int64(9900))
	assertEqual(t, "SubscriptionID", pcs.SubscriptionID, "sub_123")
	if len(pcs.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(pcs.LineItems))
	}
	assertEqual(t, "LineItems[0].ProductID", pcs.LineItems[0].ProductID, "prod_1")
	assertEqual(t, "LineItems[0].Quantity", pcs.LineItems[0].Quantity, int64(2))
}

func TestFromStripeCheckoutSession_NoSubscription(t *testing.T) {
	cs := &stripe.CheckoutSession{ID: "cs_456", Status: "complete"}
	pcs := fromStripeCheckoutSession(cs)
	assertEqual(t, "SubscriptionID", pcs.SubscriptionID, "")
}

func TestFromStripeInvoice(t *testing.T) {
	si := &stripe.Invoice{
		ID:                 "in_123",
		Customer:           &stripe.Customer{ID: "cus_123"},
		Status:             stripe.InvoiceStatusPaid,
		EffectiveAt:        1700000000,
		HostedInvoiceURL:   "https://invoice.stripe.com/i/in_123",
		Total:              4200,
		Currency:           "usd",
		Created:            1699900000,
		DueDate:            1700100000,
		NextPaymentAttempt: 0,
		PeriodStart:        1699800000,
		PeriodEnd:          1702400000,
		Metadata:           map[string]string{"org_id": "org_1"},
		Lines: &stripe.InvoiceLineItemList{
			Data: []*stripe.InvoiceLineItem{
				{
					ID:          "il_1",
					Description: "Pro plan",
					Quantity:    1,
					Price:       &stripe.Price{UnitAmount: 4200},
					Period: &stripe.Period{
						Start: 1699800000,
						End:   1702400000,
					},
					Metadata: map[string]string{"item_id": "itm_1"},
				},
			},
		},
	}

	pi := fromStripeInvoice(si)

	assertEqual(t, "ID", pi.ID, "in_123")
	assertEqual(t, "CustomerProviderID", pi.CustomerProviderID, "cus_123")
	assertEqual(t, "Status", pi.Status, "paid")
	assertEqual(t, "HostedURL", pi.HostedURL, "https://invoice.stripe.com/i/in_123")
	assertEqual(t, "Total", pi.Total, int64(4200))
	assertEqual(t, "Currency", pi.Currency, "usd")
	assertEqual(t, "Metadata[org_id]", pi.Metadata["org_id"], "org_1")

	if len(pi.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(pi.LineItems))
	}
	li := pi.LineItems[0]
	assertEqual(t, "LineItem.ID", li.ID, "il_1")
	assertEqual(t, "LineItem.Description", li.Description, "Pro plan")
	assertEqual(t, "LineItem.UnitAmount", li.UnitAmount, int64(4200))
	assertEqual(t, "LineItem.PeriodStart", li.PeriodStart, int64(1699800000))
	assertEqual(t, "LineItem.PeriodEnd", li.PeriodEnd, int64(1702400000))
}

func TestFromStripeInvoice_NilCustomer(t *testing.T) {
	si := &stripe.Invoice{ID: "in_456"}
	pi := fromStripeInvoice(si)
	assertEqual(t, "CustomerProviderID", pi.CustomerProviderID, "")
}

func TestToAnyMap(t *testing.T) {
	m := map[string]string{"a": "1", "b": "2"}
	result := toAnyMap(m)
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["a"] != "1" || result["b"] != "2" {
		t.Fatalf("unexpected values: %v", result)
	}
}

func TestToAnyMap_Nil(t *testing.T) {
	if toAnyMap(nil) != nil {
		t.Fatal("expected nil for nil input")
	}
}

func assertEqual[T comparable](t *testing.T, field string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", field, got, want)
	}
}

func TestMapStripeEventType(t *testing.T) {
	tests := []struct {
		stripe string
		want   string
	}{
		{string(stripe.EventTypeCheckoutSessionCompleted), billing.EventCheckoutCompleted},
		{string(stripe.EventTypeCheckoutSessionAsyncPaymentSucceeded), billing.EventCheckoutPaymentSucceeded},
		{string(stripe.EventTypeCustomerCreated), billing.EventCustomerCreated},
		{string(stripe.EventTypeCustomerUpdated), billing.EventCustomerUpdated},
		{string(stripe.EventTypeCustomerSourceCreated), billing.EventCustomerSourceCreated},
		{string(stripe.EventTypeCustomerSourceUpdated), billing.EventCustomerSourceUpdated},
		{string(stripe.EventTypeCustomerSubscriptionCreated), billing.EventSubscriptionCreated},
		{string(stripe.EventTypeCustomerSubscriptionUpdated), billing.EventSubscriptionUpdated},
		{string(stripe.EventTypeCustomerSubscriptionDeleted), billing.EventSubscriptionDeleted},
		{string(stripe.EventTypeInvoicePaid), billing.EventInvoicePaid},
		{"unknown.event", "unknown.event"},
	}
	for _, tt := range tests {
		got := mapStripeEventType(tt.stripe)
		if got != tt.want {
			t.Errorf("mapStripeEventType(%q) = %q, want %q", tt.stripe, got, tt.want)
		}
	}
}
