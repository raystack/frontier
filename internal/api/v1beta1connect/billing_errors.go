package v1beta1connect

import (
	"errors"
	"regexp"

	"connectrpc.com/connect"

	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	billingerrors "github.com/raystack/frontier/billing/errors"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
)

// provider-generated object ids (random alphanumeric after the prefix).
// Caller-supplied values like coupon codes don't match this shape.
var providerIDPattern = regexp.MustCompile(`\b(cus|sub_sched|sub|in|ii|il|pi|pm|seti|si|price|prod|cs|ch|coup|promo)_(?:(?:test|live)_)?[A-Za-z0-9]{8,}\b`)

// redactedProviderError hides provider object ids in a message shown to the
// caller, keeping the rest of the provider's text.
func redactedProviderError(providerErr *billingerrors.ProviderError) error {
	return errors.New(providerIDPattern.ReplaceAllString(providerErr.Error(), "${1}_*****"))
}

// mapBillingError is the fallback for billing handlers in place of a bare
// CodeInternal. Provider and account-state problems reach the caller as
// codes they can act on; everything else stays internal.
func mapBillingError(err error) *connect.Error {
	switch {
	case errors.Is(err, billingerrors.ErrProviderResourceMissing):
		var providerErr *billingerrors.ProviderError
		if errors.As(err, &providerErr) {
			return connect.NewError(connect.CodeFailedPrecondition, redactedProviderError(providerErr))
		}
		return connect.NewError(connect.CodeFailedPrecondition, ErrBillingProviderResourceMissing)
	case errors.Is(err, billingerrors.ErrPaymentFailed):
		var providerErr *billingerrors.ProviderError
		if errors.As(err, &providerErr) {
			return connect.NewError(connect.CodeFailedPrecondition, redactedProviderError(providerErr))
		}
		return connect.NewError(connect.CodeFailedPrecondition, billingerrors.ErrPaymentFailed)
	case errors.Is(err, billingerrors.ErrProviderUnavailable):
		return connect.NewError(connect.CodeUnavailable, ErrBillingProviderUnavailable)
	case errors.Is(err, subscription.ErrSubscriptionOnProviderNotFound):
		return connect.NewError(connect.CodeFailedPrecondition, ErrSubscriptionProviderMissing)
	case errors.Is(err, subscription.ErrPhaseIsUpdating):
		return connect.NewError(connect.CodeFailedPrecondition, subscription.ErrPhaseIsUpdating)
	case errors.Is(err, customer.ErrExistingAccountWithPendingDues):
		return connect.NewError(connect.CodeFailedPrecondition, customer.ErrExistingAccountWithPendingDues)
	case errors.Is(err, customer.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, ErrCustomerNotFound)
	case errors.Is(err, checkout.ErrAlreadySubscribed):
		return connect.NewError(connect.CodeAlreadyExists, checkout.ErrAlreadySubscribed)
	case errors.Is(err, product.ErrProductNotFound):
		return connect.NewError(connect.CodeNotFound, product.ErrProductNotFound)
	case errors.Is(err, product.ErrFeatureNotFound):
		return connect.NewError(connect.CodeNotFound, product.ErrFeatureNotFound)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
