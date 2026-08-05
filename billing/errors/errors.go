package errors

import (
	"errors"
	"net/http"

	stripe "github.com/stripe/stripe-go/v79"
)

var (
	ErrProviderResourceMissing = errors.New("record no longer exists on the billing provider")
	ErrPaymentFailed           = errors.New("payment failed")
	ErrProviderUnavailable     = errors.New("billing provider is unavailable")
)

// ProviderError is a billing provider failure classified as one of the
// Err* kinds above. Message carries the provider's human-readable message.
type ProviderError struct {
	Kind    error
	Message string
	cause   error
}

func (e *ProviderError) Error() string {
	if e.Message == "" {
		return e.Kind.Error()
	}
	return e.Kind.Error() + ": " + e.Message
}

func (e *ProviderError) Unwrap() []error {
	return []error{e.Kind, e.cause}
}

// TranslateStripeError converts a stripe error into a *ProviderError so
// callers can match it with errors.Is against the Err* kinds. Errors that
// don't match a known kind are returned unchanged.
func TranslateStripeError(err error) error {
	var stripeErr *stripe.Error
	if err == nil || !errors.As(err, &stripeErr) {
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
	return &ProviderError{Kind: kind, Message: stripeErr.Msg, cause: err}
}
