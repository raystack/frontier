package invoice

import (
	"context"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/billing/customer"
	stripemock "github.com/raystack/frontier/billing/stripetest/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/client"
	"github.com/stripe/stripe-go/v79/form"
)

func TestService_computeOverdraftWindow(t *testing.T) {
	createdAt := time.Date(2024, 8, 12, 7, 59, 50, 544323000, time.UTC)
	// providers store item ranges in whole seconds, so the first invoice's
	// start is the creation time truncated to the second
	createdAtTruncated := createdAt.Truncate(time.Second)
	feb1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	mar1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	may1 := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	jun1 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	jul1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	creditItem := func(start, end *time.Time) Invoice {
		return Invoice{
			Items: []Item{
				{
					Type:           CreditItemType,
					TimeRangeStart: start,
					TimeRangeEnd:   end,
				},
			},
		}
	}

	tests := []struct {
		name                string
		endRange            time.Time
		lastInvoice         *Invoice
		wantStart           time.Time
		wantAlreadyInvoiced bool
	}{
		{
			name:      "no previous overdraft invoice starts window at customer creation",
			endRange:  jul1,
			wantStart: createdAt,
		},
		{
			name:        "first invoice anchors window at its end",
			endRange:    mar1,
			lastInvoice: new(creditItem(&createdAtTruncated, &feb1)),
			wantStart:   feb1,
		},
		{
			name:        "first invoice anchors window when its start equals creation time",
			endRange:    mar1,
			lastInvoice: new(creditItem(&createdAt, &feb1)),
			wantStart:   feb1,
		},
		{
			name:        "later invoice anchors window at its end instead of customer creation",
			endRange:    jul1,
			lastInvoice: new(creditItem(&may1, &jun1)),
			wantStart:   jun1,
		},
		{
			name:                "range ending at window end is already invoiced",
			endRange:            jul1,
			lastInvoice:         new(creditItem(&may1, &jul1)),
			wantStart:           jul1,
			wantAlreadyInvoiced: true,
		},
		{
			name:        "item without range start does not panic",
			endRange:    jul1,
			lastInvoice: new(creditItem(nil, &jun1)),
			wantStart:   jun1,
		},
		{
			name:        "item without range end keeps window at customer creation",
			endRange:    jul1,
			lastInvoice: new(creditItem(&may1, nil)),
			wantStart:   createdAt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, alreadyInvoiced := computeOverdraftWindow(createdAt, tt.endRange, tt.lastInvoice)
			if !start.Equal(tt.wantStart) {
				t.Errorf("computeOverdraftWindow() start = %v, want %v", start, tt.wantStart)
			}
			if alreadyInvoiced != tt.wantAlreadyInvoiced {
				t.Errorf("computeOverdraftWindow() alreadyInvoiced = %v, want %v", alreadyInvoiced, tt.wantAlreadyInvoiced)
			}
		})
	}
}

func TestService_getCreditOverdraftRange(t *testing.T) {
	tests := []struct {
		name    string
		shift   int
		current time.Time
		end     time.Time
	}{
		{
			name:    "start of year",
			current: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "between calender month",
			current: time.Date(2024, 10, 9, 8, 7, 6, 5, time.UTC),
			end:     time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "start of month",
			current: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			end:     time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "end of month",
			current: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
			end:     time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "between calender month with shift in future to currently running month",
			current: time.Date(2024, 10, 9, 8, 7, 6, 5, time.UTC),
			shift:   1,
			end:     time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "between calender month with shift in past",
			current: time.Date(2024, 10, 9, 8, 7, 6, 5, time.UTC),
			shift:   -1,
			end:     time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				creditOverdraftRangeShift:     tt.shift,
				creditOverdraftInvoiceDay:     1,
				creditOverdraftRangeOfInvoice: MonthCreditRangeForInvoice,
			}
			end := s.getCreditOverdraftEndDate(tt.current)
			if !reflect.DeepEqual(end, tt.end) {
				t.Errorf("getCreditOverdraftEndDate() end = %v, want %v", end, tt.end)
			}
		})
	}
}

func TestService_overdraftItemName(t *testing.T) {
	tests := []struct {
		name     string
		itemName string
		want     string
	}{
		{
			name:     "uses the overdraft product title when set",
			itemName: "For Satellite Images and Data Analytics Services",
			want:     "For Satellite Images and Data Analytics Services",
		},
		{
			name: "falls back to default when product title is empty",
			want: DefaultCreditOverdraftItemName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{creditOverdraftItemName: tt.itemName}
			if got := s.overdraftItemName(); got != tt.want {
				t.Errorf("overdraftItemName() = %v, want %v", got, tt.want)
			}
		})
	}
}

type recordingInvoiceRepository struct {
	Repository
	invoices []Invoice
	created  []Invoice
	updated  []Invoice
}

func (r *recordingInvoiceRepository) List(_ context.Context, flt Filter) ([]Invoice, error) {
	if flt.IncludeDeleted {
		return r.invoices, nil
	}
	var live []Invoice
	for _, inv := range r.invoices {
		if inv.DeletedAt.IsZero() {
			live = append(live, inv)
		}
	}
	return live, nil
}

func (r *recordingInvoiceRepository) Create(_ context.Context, inv Invoice) (Invoice, error) {
	r.created = append(r.created, inv)
	return inv, nil
}

func (r *recordingInvoiceRepository) UpdateByID(_ context.Context, inv Invoice) (Invoice, error) {
	r.updated = append(r.updated, inv)
	return inv, nil
}

func TestService_lastOverdraftInvoice(t *testing.T) {
	t.Run("a deleted overdraft invoice still anchors the window", func(t *testing.T) {
		end := time.Now().UTC()
		repo := &recordingInvoiceRepository{invoices: []Invoice{{
			ID:        "inv-gone",
			DeletedAt: end,
			Items:     []Item{{Type: CreditItemType, TimeRangeEnd: &end}},
		}}}
		svc := NewService(slog.Default(), nil, repo, nil, nil, nil, nil, billing.Config{})

		got, err := svc.lastOverdraftInvoice(context.Background(), "cust-1")
		assert.NoError(t, err)
		if assert.NotNil(t, got, "the deleted invoice must still be found") {
			assert.Equal(t, "inv-gone", got.ID)
		}
	})

	t.Run("an invoice without a credit item is not the anchor", func(t *testing.T) {
		repo := &recordingInvoiceRepository{invoices: []Invoice{{ID: "inv-plan"}}}
		svc := NewService(slog.Default(), nil, repo, nil, nil, nil, nil, billing.Config{})

		got, err := svc.lastOverdraftInvoice(context.Background(), "cust-1")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestService_SyncWithProvider(t *testing.T) {
	t.Run("a deleted local invoice is neither updated nor created again", func(t *testing.T) {
		backend := stripemock.NewBackend(t)
		backend.EXPECT().CallRaw("GET", "/v1/invoices", "key_123", mock.Anything, mock.Anything, mock.Anything).
			Run(func(_, _, _ string, _ *form.Values, _ *stripe.Params, v stripe.LastResponseSetter) {
				list := v.(*stripe.InvoiceList)
				list.Data = []*stripe.Invoice{{ID: "in_gone", Status: stripe.InvoiceStatusPaid, Total: 100}}
			}).Return(nil)
		stripeClient := client.New("key_123", &stripe.Backends{API: backend})

		repo := &recordingInvoiceRepository{invoices: []Invoice{{
			ID:         "inv-gone",
			ProviderID: "in_gone",
			CustomerID: "cust-1",
			State:      OpenState,
			DeletedAt:  time.Now(),
		}}}
		svc := NewService(slog.Default(), stripeClient, repo, nil, nil, nil, nil, billing.Config{})

		err := svc.SyncWithProvider(context.Background(), customer.Customer{ID: "cust-1", ProviderID: "cus_1"})
		assert.NoError(t, err)
		assert.Empty(t, repo.created, "the deleted invoice must not come back as a new row")
		assert.Empty(t, repo.updated, "the deleted invoice must not be touched")
	})

	t.Run("a live local invoice is still updated", func(t *testing.T) {
		backend := stripemock.NewBackend(t)
		backend.EXPECT().CallRaw("GET", "/v1/invoices", "key_123", mock.Anything, mock.Anything, mock.Anything).
			Run(func(_, _, _ string, _ *form.Values, _ *stripe.Params, v stripe.LastResponseSetter) {
				list := v.(*stripe.InvoiceList)
				list.Data = []*stripe.Invoice{{ID: "in_live", Status: stripe.InvoiceStatusPaid, Total: 100}}
			}).Return(nil)
		stripeClient := client.New("key_123", &stripe.Backends{API: backend})

		repo := &recordingInvoiceRepository{invoices: []Invoice{{
			ID:         "inv-live",
			ProviderID: "in_live",
			CustomerID: "cust-1",
			State:      OpenState,
			Amount:     100,
		}}}
		svc := NewService(slog.Default(), stripeClient, repo, nil, nil, nil, nil, billing.Config{})

		err := svc.SyncWithProvider(context.Background(), customer.Customer{ID: "cust-1", ProviderID: "cus_1"})
		assert.NoError(t, err)
		assert.Empty(t, repo.created)
		if assert.Len(t, repo.updated, 1) {
			assert.Equal(t, PaidState, repo.updated[0].State)
		}
	})
}
