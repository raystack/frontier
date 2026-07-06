package invoice

import (
	"reflect"
	"testing"
	"time"
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
			lastInvoice: ptr(creditItem(&createdAtTruncated, &feb1)),
			wantStart:   feb1,
		},
		{
			name:        "first invoice anchors window when its start equals creation time",
			endRange:    mar1,
			lastInvoice: ptr(creditItem(&createdAt, &feb1)),
			wantStart:   feb1,
		},
		{
			name:        "later invoice anchors window at its end instead of customer creation",
			endRange:    jul1,
			lastInvoice: ptr(creditItem(&may1, &jun1)),
			wantStart:   jun1,
		},
		{
			name:                "range ending at window end is already invoiced",
			endRange:            jul1,
			lastInvoice:         ptr(creditItem(&may1, &jul1)),
			wantStart:           jul1,
			wantAlreadyInvoiced: true,
		},
		{
			name:        "item without range start does not panic",
			endRange:    jul1,
			lastInvoice: ptr(creditItem(nil, &jun1)),
			wantStart:   jun1,
		},
		{
			name:        "item without range end keeps window at customer creation",
			endRange:    jul1,
			lastInvoice: ptr(creditItem(&may1, nil)),
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

func ptr[T any](v T) *T {
	return &v
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
