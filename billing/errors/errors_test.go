package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	stripe "github.com/stripe/stripe-go/v79"
)

func TestTranslateStripeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantKind error
		wantMsg  string
	}{
		{
			name: "nil stays nil",
			err:  nil,
		},
		{
			name: "non stripe error stays unchanged",
			err:  errors.New("db down"),
		},
		{
			name: "unknown stripe error stays unchanged",
			err:  &stripe.Error{Code: stripe.ErrorCodeParameterMissing, Type: stripe.ErrorTypeInvalidRequest},
		},
		{
			name:     "resource missing",
			err:      &stripe.Error{Code: stripe.ErrorCodeResourceMissing, Msg: "No such customer: 'cus_123'"},
			wantKind: ErrProviderResourceMissing,
			wantMsg:  "record no longer exists on the billing provider: No such customer: 'cus_123'",
		},
		{
			name:     "wrapped resource missing",
			err:      fmt.Errorf("get customer: %w", &stripe.Error{Code: stripe.ErrorCodeResourceMissing}),
			wantKind: ErrProviderResourceMissing,
			wantMsg:  "record no longer exists on the billing provider",
		},
		{
			name:     "card error",
			err:      &stripe.Error{Type: stripe.ErrorTypeCard, Code: stripe.ErrorCodeCardDeclined, Msg: "Your card was declined."},
			wantKind: ErrPaymentFailed,
			wantMsg:  "payment failed: Your card was declined.",
		},
		{
			name:     "decline code without card type",
			err:      &stripe.Error{Type: stripe.ErrorTypeInvalidRequest, DeclineCode: stripe.DeclineCodeInsufficientFunds, Msg: "Insufficient funds."},
			wantKind: ErrPaymentFailed,
			wantMsg:  "payment failed: Insufficient funds.",
		},
		{
			name:     "rate limited",
			err:      &stripe.Error{Code: stripe.ErrorCodeRateLimit, Type: stripe.ErrorTypeInvalidRequest},
			wantKind: ErrProviderUnavailable,
		},
		{
			name:     "stripe server error",
			err:      &stripe.Error{Type: stripe.ErrorTypeAPI, Msg: "An unknown error occurred."},
			wantKind: ErrProviderUnavailable,
			wantMsg:  "billing provider is unavailable: An unknown error occurred.",
		},
		{
			name:     "http 500 without api type",
			err:      &stripe.Error{Type: stripe.ErrorTypeInvalidRequest, HTTPStatusCode: 500},
			wantKind: ErrProviderUnavailable,
		},
		{
			name:     "http 429 without rate limit code",
			err:      &stripe.Error{Type: stripe.ErrorTypeInvalidRequest, HTTPStatusCode: 429},
			wantKind: ErrProviderUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateStripeError(tt.err)
			if tt.wantKind == nil {
				assert.Equal(t, tt.err, got)
				return
			}
			assert.ErrorIs(t, got, tt.wantKind)
			if tt.wantMsg != "" {
				assert.Equal(t, tt.wantMsg, got.Error())
			}

			var stripeErr *stripe.Error
			assert.ErrorAs(t, got, &stripeErr)
		})
	}
}

func TestTranslateStripeErrorKeepsOtherKindsApart(t *testing.T) {
	got := TranslateStripeError(&stripe.Error{Code: stripe.ErrorCodeResourceMissing})
	assert.NotErrorIs(t, got, ErrPaymentFailed)
	assert.NotErrorIs(t, got, ErrProviderUnavailable)
}
