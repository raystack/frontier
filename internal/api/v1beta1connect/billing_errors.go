package v1beta1connect

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/raystack/frontier/billing/customer"
	billingerrors "github.com/raystack/frontier/billing/errors"
	"github.com/raystack/frontier/billing/subscription"
)

// mapBillingError is the fallback for billing handlers in place of a bare
// CodeInternal. Provider and account-state problems reach the caller as
// codes they can act on; everything else stays internal.
func mapBillingError(err error) *connect.Error {
	switch {
	case errors.Is(err, billingerrors.ErrProviderResourceMissing):
		return connect.NewError(connect.CodeFailedPrecondition, ErrBillingProviderResourceMissing)
	case errors.Is(err, billingerrors.ErrPaymentFailed):
		var providerErr *billingerrors.ProviderError
		if errors.As(err, &providerErr) {
			return connect.NewError(connect.CodeFailedPrecondition, providerErr)
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
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
