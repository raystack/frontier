package errors

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
			name:     "connection failure",
			err:      &url.Error{Op: "Post", URL: "https://api.stripe.com/v1/customers", Err: errors.New("connection refused")},
			wantKind: ErrProviderUnavailable,
			wantMsg:  "billing provider is unavailable: could not reach the billing provider",
		},
		{
			name: "canceled request stays unchanged",
			err:  fmt.Errorf("get customer: %w", context.Canceled),
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
				if tt.err == nil {
					assert.Nil(t, got)
				} else {
					assert.Same(t, tt.err, got)
				}
				return
			}
			assert.ErrorIs(t, got, tt.wantKind)
			assert.ErrorIs(t, got, tt.err)
			if tt.wantMsg != "" {
				assert.Equal(t, tt.wantMsg, got.Error())
			}

			var inputStripeErr *stripe.Error
			if errors.As(tt.err, &inputStripeErr) {
				var stripeErr *stripe.Error
				assert.ErrorAs(t, got, &stripeErr)
			}
		})
	}
}

func TestTranslateStripeErrorKeepsOtherKindsApart(t *testing.T) {
	got := TranslateStripeError(&stripe.Error{Code: stripe.ErrorCodeResourceMissing})
	assert.NotErrorIs(t, got, ErrPaymentFailed)
	assert.NotErrorIs(t, got, ErrProviderUnavailable)
}

func TestTranslateStripeErrorKeepsRequestID(t *testing.T) {
	got := TranslateStripeError(&stripe.Error{
		Code:      stripe.ErrorCodeResourceMissing,
		RequestID: "req_AbCdEf123",
	})
	var providerErr *ProviderError
	assert.ErrorAs(t, got, &providerErr)
	assert.Equal(t, "req_AbCdEf123", providerErr.RequestID)
}

func TestProviderErrorUnwrapWithoutCause(t *testing.T) {
	err := &ProviderError{Kind: ErrPaymentFailed}
	for _, unwrapped := range err.Unwrap() {
		assert.NotNil(t, unwrapped)
	}
	assert.ErrorIs(t, err, ErrPaymentFailed)
}
