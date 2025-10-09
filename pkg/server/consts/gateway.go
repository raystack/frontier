package consts

import (
	"context"
)

type contextKey struct {
	name string
}

func (c *contextKey) String() string { return "context value " + c.name }

var (
	// AuthenticatedPrincipalContextKey is context key that contains the principal object
	AuthenticatedPrincipalContextKey = contextKey{name: "auth-principal"}
	// AuthSuperUserContextKey is context key that contains super user flag
	AuthSuperUserContextKey = contextKey{name: "auth-superuser"}

	// SessionContextKey is context key that contains session metadata
	SessionContextKey = contextKey{name: "session-context"}

	// AuditRecordActorContextKey is context key that contains the audit record actor
	AuditRecordActorContextKey = contextKey{name: "audit-record-actor"}

	AuditActorContextKey    = contextKey{name: "audit-actor"}
	AuditMetadataContextKey = contextKey{name: "audit-metadata"}
	AuditServiceContextKey  = contextKey{name: "audit-service"}

	// BillingStripeTestClockContextKey is context key that contains the stripe test clock id
	BillingStripeTestClockContextKey = contextKey{name: "billing-stripe-test-clock"}

	// BillingStripeWebhookSignatureContextKey is context key that contains the stripe webhook signature
	BillingStripeWebhookSignatureContextKey = contextKey{name: "billing-stripe-webhook-signature"}

	// RequestIDContextKey is context key that contains the request id
	RequestIDContextKey = contextKey{name: "request-id"}
)

const (
	// const keys used to pass values from gRPC methods to http mux interface
	SessionIDGatewayKey     = "gateway-session-id"
	SessionDeleteGatewayKey = "gateway-session-delete"
	UserTokenGatewayKey     = "gateway-user-token"
	LocationGatewayKey      = "gateway-location"
	UserSecretGatewayKey    = "gateway-user-secret"

	// UserTokenRequestKey is returned from the application to client containing user details in
	// response headers
	UserTokenRequestKey = "x-user-token"

	// LocationRequestKey is used to set location response header for redirecting browser
	LocationRequestKey = "location"

	// ProjectRequestKey is used to set current project in jwt token
	ProjectRequestKey = "x-project"

	// SessionRequestKey is the key to store session value in browser
	SessionRequestKey = "sid"

	// StripeTestClockRequestKey is used to store stripe test clock id which will
	// be used to simulate a customer & subscription
	StripeTestClockRequestKey = "x-stripe-test-clock"

	// StripeWebhookSignature is used to store stripe webhook signature
	StripeWebhookSignature = "stripe-signature"

	// RequestIDHeader is the key to store request id from http headers
	RequestIDHeader = "x-request-id"

	AuditActorSuperUserKey  = "is_super_user"
	AuditSessionMetadataKey = "context"
)

func GetRequestIDFromCtx(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(RequestIDContextKey).(string)
	return u, ok
}

func WithRequestIDInCtx(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey, id)
}
