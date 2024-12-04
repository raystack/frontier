package invoice

import (
	"reflect"
	"testing"
	"time"
)

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
