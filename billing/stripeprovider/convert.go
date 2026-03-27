package stripeprovider

import (
	"github.com/raystack/frontier/billing"
	"github.com/stripe/stripe-go/v79"
)

func toStripeAddress(a billing.ProviderAddress) *stripe.AddressParams {
	return &stripe.AddressParams{
		City:       &a.City,
		Country:    &a.Country,
		Line1:      &a.Line1,
		Line2:      &a.Line2,
		PostalCode: &a.PostalCode,
		State:      &a.State,
	}
}

func fromStripeCustomer(sc *stripe.Customer) *billing.ProviderCustomer {
	c := &billing.ProviderCustomer{
		ID:       sc.ID,
		Deleted:  sc.Deleted,
		Name:     sc.Name,
		Email:    sc.Email,
		Phone:    sc.Phone,
		Currency: string(sc.Currency),
	}
	if sc.Address != nil {
		c.Address = billing.ProviderAddress{
			City:       sc.Address.City,
			Country:    sc.Address.Country,
			Line1:      sc.Address.Line1,
			Line2:      sc.Address.Line2,
			PostalCode: sc.Address.PostalCode,
			State:      sc.Address.State,
		}
	}
	if sc.TaxIDs != nil {
		for _, tid := range sc.TaxIDs.Data {
			c.TaxIDs = append(c.TaxIDs, billing.ProviderTaxID{
				Type:  string(tid.Type),
				Value: tid.Value,
			})
		}
	}
	return c
}

func fromStripeSubscription(ss *stripe.Subscription) *billing.ProviderSubscription {
	ps := &billing.ProviderSubscription{
		ID:                 ss.ID,
		Status:             string(ss.Status),
		CanceledAt:         ss.CanceledAt,
		EndedAt:            ss.EndedAt,
		TrialEnd:           ss.TrialEnd,
		CurrentPeriodStart: ss.CurrentPeriodStart,
		CurrentPeriodEnd:   ss.CurrentPeriodEnd,
		BillingCycleAnchor: ss.BillingCycleAnchor,
		Livemode:           ss.Livemode,
		Metadata:           ss.Metadata,
	}
	if ss.AutomaticTax != nil {
		ps.AutomaticTaxEnabled = ss.AutomaticTax.Enabled
	}
	if ss.Items != nil {
		for _, item := range ss.Items.Data {
			si := billing.ProviderSubscriptionItem{
				ID:       item.ID,
				PriceID:  item.Price.ID,
				Quantity: item.Quantity,
				Metadata: item.Metadata,
			}
			if item.Price.Product != nil {
				si.ProductID = item.Price.Product.ID
			}
			if item.Price.Recurring != nil {
				si.Interval = string(item.Price.Recurring.Interval)
			}
			ps.Items = append(ps.Items, si)
		}
	}
	if ss.Schedule != nil && ss.Schedule.ID != "" {
		ps.Schedule = &billing.ProviderScheduleRef{
			ID: ss.Schedule.ID,
		}
		if ss.Schedule.CurrentPhase != nil {
			ps.Schedule.CurrentPhase = &billing.ProviderCurrentPhase{
				StartDate: ss.Schedule.CurrentPhase.StartDate,
				EndDate:   ss.Schedule.CurrentPhase.EndDate,
			}
		}
	}
	return ps
}

func fromStripeSchedule(ss *stripe.SubscriptionSchedule) *billing.ProviderSchedule {
	ps := &billing.ProviderSchedule{
		ID:          ss.ID,
		EndBehavior: string(ss.EndBehavior),
	}
	if ss.CurrentPhase != nil {
		ps.CurrentPhase = &billing.ProviderCurrentPhase{
			StartDate: ss.CurrentPhase.StartDate,
			EndDate:   ss.CurrentPhase.EndDate,
		}
	}
	for _, phase := range ss.Phases {
		pp := billing.ProviderPhase{
			StartDate: phase.StartDate,
			EndDate:   phase.EndDate,
			Currency:  string(phase.Currency),
			Metadata:  phase.Metadata,
			TrialEnd:  phase.TrialEnd,
		}
		if phase.Description != "" {
			pp.Description = phase.Description
		}
		if phase.ProrationBehavior != "" {
			pp.ProrationBehavior = string(phase.ProrationBehavior)
		}
		if phase.CollectionMethod != nil {
			pp.CollectionMethod = string(*phase.CollectionMethod)
		}
		if phase.AutomaticTax != nil {
			pp.AutomaticTaxEnabled = phase.AutomaticTax.Enabled
		}
		for _, item := range phase.Items {
			pi := billing.ProviderPhaseItem{
				PriceID:  item.Price.ID,
				Quantity: item.Quantity,
				Metadata: item.Metadata,
			}
			if item.Price.Product != nil {
				pi.ProductID = item.Price.Product.ID
			}
			pp.Items = append(pp.Items, pi)
		}
		ps.Phases = append(ps.Phases, pp)
	}
	return ps
}

func fromStripeCheckoutSession(cs *stripe.CheckoutSession) *billing.ProviderCheckoutSession {
	pcs := &billing.ProviderCheckoutSession{
		ID:            cs.ID,
		URL:           cs.URL,
		Status:        string(cs.Status),
		PaymentStatus: string(cs.PaymentStatus),
		ExpiresAt:     cs.ExpiresAt,
		AmountTotal:   cs.AmountTotal,
		Currency:      string(cs.Currency),
	}
	if cs.Subscription != nil {
		pcs.SubscriptionID = cs.Subscription.ID
	}
	if cs.LineItems != nil {
		for _, li := range cs.LineItems.Data {
			item := billing.ProviderCheckoutLineItem{
				Quantity: li.Quantity,
			}
			if li.Price != nil && li.Price.Product != nil {
				item.ProductID = li.Price.Product.ID
			}
			pcs.LineItems = append(pcs.LineItems, item)
		}
	}
	return pcs
}

func fromStripeInvoice(si *stripe.Invoice) *billing.ProviderInvoice {
	var customerProviderID string
	if si.Customer != nil {
		customerProviderID = si.Customer.ID
	}
	pi := &billing.ProviderInvoice{
		ID:                 si.ID,
		CustomerProviderID: customerProviderID,
		Status:             string(si.Status),
		EffectiveAt:        si.EffectiveAt,
		HostedURL:          si.HostedInvoiceURL,
		Total:              si.Total,
		Currency:           string(si.Currency),
		CreatedAt:          si.Created,
		DueDate:            si.DueDate,
		NextPaymentAttempt: si.NextPaymentAttempt,
		PeriodStart:        si.PeriodStart,
		PeriodEnd:          si.PeriodEnd,
		Metadata:           si.Metadata,
	}
	if si.Lines != nil {
		for _, line := range si.Lines.Data {
			li := billing.ProviderInvoiceLineItem{
				ID:          line.ID,
				Description: line.Description,
				Quantity:    line.Quantity,
				Metadata:    line.Metadata,
			}
			if line.Price != nil {
				li.UnitAmount = line.Price.UnitAmount
			}
			if line.Period != nil {
				li.PeriodStart = line.Period.Start
				li.PeriodEnd = line.Period.End
			}
			pi.LineItems = append(pi.LineItems, li)
		}
	}
	return pi
}

func toAnyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
