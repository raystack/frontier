package v1beta1connect

import (
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	stripe "github.com/stripe/stripe-go/v79"

	"github.com/raystack/frontier/billing/customer"
	billingerrors "github.com/raystack/frontier/billing/errors"
	"github.com/raystack/frontier/billing/subscription"
)

func TestMapBillingError(t *testing.T) {
	deadCustomer := billingerrors.TranslateStripeError(&stripe.Error{
		Code: stripe.ErrorCodeResourceMissing,
		Msg:  "No such customer: 'cus_123'",
	})
	cardDeclined := billingerrors.TranslateStripeError(&stripe.Error{
		Type: stripe.ErrorTypeCard,
		Msg:  "Your card was declined.",
	})
	rateLimited := billingerrors.TranslateStripeError(&stripe.Error{
		Code: stripe.ErrorCodeRateLimit,
	})

	tests := []struct {
		name     string
		err      error
		wantCode connect.Code
		wantMsg  string
	}{
		{
			name:     "provider resource missing",
			err:      fmt.Errorf("GetUpcomingInvoice: org_id=abc: %w", deadCustomer),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  ErrBillingProviderResourceMissing.Error(),
		},
		{
			name:     "payment failed keeps provider message",
			err:      fmt.Errorf("CreateCheckout: %w", cardDeclined),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  "payment failed: Your card was declined.",
		},
		{
			name:     "provider unavailable",
			err:      fmt.Errorf("ListInvoices: %w", rateLimited),
			wantCode: connect.CodeUnavailable,
			wantMsg:  ErrBillingProviderUnavailable.Error(),
		},
		{
			name:     "subscription gone on provider",
			err:      fmt.Errorf("ChangeSubscription: %w", subscription.ErrSubscriptionOnProviderNotFound),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  ErrSubscriptionProviderMissing.Error(),
		},
		{
			name:     "plan change in progress",
			err:      fmt.Errorf("ChangeSubscription: %w", subscription.ErrPhaseIsUpdating),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  subscription.ErrPhaseIsUpdating.Error(),
		},
		{
			name:     "pending dues",
			err:      fmt.Errorf("CreateBillingAccount: %w", customer.ErrExistingAccountWithPendingDues),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  customer.ErrExistingAccountWithPendingDues.Error(),
		},
		{
			name:     "unknown error stays internal",
			err:      fmt.Errorf("GetBillingAccount: %w", errors.New("db down")),
			wantCode: connect.CodeInternal,
			wantMsg:  "GetBillingAccount: db down",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapBillingError(tt.err)
			assert.Equal(t, tt.wantCode, got.Code())
			assert.Equal(t, tt.wantMsg, got.Message())
		})
	}
}
