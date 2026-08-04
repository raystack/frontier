package v1beta1connect

import (
	"context"
	"errors"
	"log/slog"
	"regexp"

	"connectrpc.com/connect"

	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	billingerrors "github.com/raystack/frontier/billing/errors"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
)

// provider-generated object ids: a short lowercase prefix, an optional
// test/live mode segment, and a long random alphanumeric part. The pattern is
// deliberately general so an id of an object type we haven't seen still gets
// masked; caller-supplied values like coupon codes don't match this shape.
var providerIDPattern = regexp.MustCompile(`\b([a-z]{2,10})_(?:(?:test|live)_)?[A-Za-z0-9]{8,}\b`)

// redactedProviderError hides provider object ids in a message shown to the
// caller, keeping the rest of the provider's text.
func redactedProviderError(providerErr *billingerrors.ProviderError) error {
	return errors.New(providerIDPattern.ReplaceAllString(providerErr.Error(), "${1}_*****"))
}

// mapBillingError is the fallback for billing handlers in place of a bare
// CodeInternal. Provider and account-state problems reach the caller as
// codes they can act on; everything else stays internal. Mapped errors are
// logged here with their full detail, because the caller-facing error is
// deliberately stripped of it and that error is all the logger interceptor
// sees.
func mapBillingError(ctx context.Context, err error) *connect.Error {
	mapped := mapBillingErrorCode(err)
	if mapped.Code() != connect.CodeInternal {
		args := []any{"error", err, "code", mapped.Code().String()}
		var providerErr *billingerrors.ProviderError
		if errors.As(err, &providerErr) && providerErr.RequestID != "" {
			args = append(args, "provider_request_id", providerErr.RequestID)
		}
		slog.WarnContext(ctx, "billing request failed", args...)
	}
	return mapped
}

func mapBillingErrorCode(err error) *connect.Error {
	switch {
	case errors.Is(err, billingerrors.ErrProviderResourceMissing):
		var providerErr *billingerrors.ProviderError
		if errors.As(err, &providerErr) {
			return connect.NewError(connect.CodeFailedPrecondition, redactedProviderError(providerErr))
		}
		return connect.NewError(connect.CodeFailedPrecondition, billingerrors.ErrProviderResourceMissing)
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
	case errors.Is(err, checkout.ErrPlanInactive), errors.Is(err, subscription.ErrPlanInactive):
		return connect.NewError(connect.CodeFailedPrecondition, ErrPlanInactive)
	case errors.Is(err, product.ErrProductNotFound):
		return connect.NewError(connect.CodeNotFound, product.ErrProductNotFound)
	case errors.Is(err, product.ErrFeatureNotFound):
		return connect.NewError(connect.CodeNotFound, product.ErrFeatureNotFound)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
