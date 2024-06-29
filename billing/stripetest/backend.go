package stripetest

import (
	"bytes"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/form"
)

type Backend interface {
	Call(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) error
	CallStreaming(method, path, key string, params stripe.ParamsContainer, v stripe.StreamingLastResponseSetter) error
	CallRaw(method, path, key string, body *form.Values, params *stripe.Params, v stripe.LastResponseSetter) error
	CallMultipart(method, path, key, boundary string, body *bytes.Buffer, params *stripe.Params, v stripe.LastResponseSetter) error
	SetMaxNetworkRetries(maxNetworkRetries int64)
}
