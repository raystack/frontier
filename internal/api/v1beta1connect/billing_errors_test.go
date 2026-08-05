package v1beta1connect

import (
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	stripe "github.com/stripe/stripe-go/v79"

	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	billingerrors "github.com/raystack/frontier/billing/errors"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
)

func TestMapBillingError(t *testing.T) {
	deadCustomer := billingerrors.TranslateStripeError(&stripe.Error{
		Code: stripe.ErrorCodeResourceMissing,
		Msg:  "No such customer: 'cus_QhBNKtbzOZzumU'",
	})
	deadCoupon := billingerrors.TranslateStripeError(&stripe.Error{
		Code: stripe.ErrorCodeResourceMissing,
		Msg:  "No such coupon: 'SUMMER20'",
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
			name:     "provider resource missing masks the provider id",
			err:      fmt.Errorf("GetUpcomingInvoice: org_id=abc: %w", deadCustomer),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  "record no longer exists on the billing provider: No such customer: 'cus_*****'",
		},
		{
			name:     "provider resource missing keeps a caller-supplied value",
			err:      fmt.Errorf("DelegatedCheckout: %w", deadCoupon),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  "record no longer exists on the billing provider: No such coupon: 'SUMMER20'",
		},
		{
			name: "provider resource missing masks a test-mode checkout session id",
			err: fmt.Errorf("GetCheckout: %w", billingerrors.TranslateStripeError(&stripe.Error{
				Code: stripe.ErrorCodeResourceMissing,
				Msg:  "No such checkout.session: 'cs_test_c1GSMJhe9lzCEkJAVj3R5ife'",
			})),
			wantCode: connect.CodeFailedPrecondition,
			wantMsg:  "record no longer exists on the billing provider: No such checkout.session: 'cs_*****'",
		},
		{
			name:     "provider resource missing without provider error",
			err:      fmt.Errorf("GetUpcomingInvoice: %w", billingerrors.ErrProviderResourceMissing),
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
			name:     "customer not found",
			err:      fmt.Errorf("CreateCheckout.GetBillingAccountFromOrgID: org_id=abc: %w", customer.ErrNotFound),
			wantCode: connect.CodeNotFound,
			wantMsg:  ErrCustomerNotFound.Error(),
		},
		{
			name:     "already subscribed to the plan",
			err:      fmt.Errorf("CreateCheckout.Create: %w", checkout.ErrAlreadySubscribed),
			wantCode: connect.CodeAlreadyExists,
			wantMsg:  checkout.ErrAlreadySubscribed.Error(),
		},
		{
			name:     "product not found",
			err:      fmt.Errorf("GetProduct.GetByID: product_id=abc: %w", product.ErrProductNotFound),
			wantCode: connect.CodeNotFound,
			wantMsg:  product.ErrProductNotFound.Error(),
		},
		{
			name:     "feature not found",
			err:      fmt.Errorf("CheckFeatureEntitlement: feature=abc: %w", product.ErrFeatureNotFound),
			wantCode: connect.CodeNotFound,
			wantMsg:  product.ErrFeatureNotFound.Error(),
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
