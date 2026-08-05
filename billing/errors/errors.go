package errors

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"

	stripe "github.com/stripe/stripe-go/v79"
)

var (
	ErrProviderResourceMissing = errors.New("record no longer exists on the billing provider")
	ErrPaymentFailed           = errors.New("payment failed")
	ErrProviderUnavailable     = errors.New("billing provider is unavailable")
)

// ProviderError is a billing provider failure classified as one of the
// Err* kinds above. Message carries the provider's human-readable message,
// RequestID the provider's id for the failed request (for support tickets).
type ProviderError struct {
	Kind      error
	Message   string
	RequestID string
	cause     error
}

func (e *ProviderError) Error() string {
	if e.Message == "" {
		return e.Kind.Error()
	}
	return e.Kind.Error() + ": " + e.Message
}

func (e *ProviderError) Unwrap() []error {
	errs := []error{e.Kind}
	if e.cause != nil {
		errs = append(errs, e.cause)
	}
	return errs
}

// TranslateStripeError converts a stripe or network error into a
// *ProviderError so callers can match it with errors.Is against the Err*
// kinds. Errors that don't match a known kind are returned unchanged.
func TranslateStripeError(err error) error {
	if err == nil {
		return nil
	}

	var stripeErr *stripe.Error
	if !errors.As(err, &stripeErr) {
		// a canceled request is the caller's doing, not a provider outage
		if errors.Is(err, context.Canceled) {
			return err
		}
		var urlErr *url.Error
		var netErr net.Error
		if errors.As(err, &urlErr) || errors.As(err, &netErr) {
			return &ProviderError{
				Kind:    ErrProviderUnavailable,
				Message: "could not reach the billing provider",
				cause:   err,
			}
		}
		return err
	}

	var kind error
	switch {
	case stripeErr.Code == stripe.ErrorCodeResourceMissing:
		kind = ErrProviderResourceMissing
	case stripeErr.Type == stripe.ErrorTypeCard || stripeErr.DeclineCode != "":
		kind = ErrPaymentFailed
	case stripeErr.Code == stripe.ErrorCodeRateLimit,
		stripeErr.Type == stripe.ErrorTypeAPI,
		stripeErr.HTTPStatusCode == http.StatusTooManyRequests,
		stripeErr.HTTPStatusCode >= http.StatusInternalServerError:
		kind = ErrProviderUnavailable
	default:
		return err
	}
	return &ProviderError{
		Kind:      kind,
		Message:   stripeErr.Msg,
		RequestID: stripeErr.RequestID,
		cause:     err,
	}
}
