# Implementing a Billing Provider

This document describes the `billing.Provider` interface and the contract that
implementations must satisfy.

## Overview

`billing.Provider` is the interface between Frontier's billing services and
external billing APIs (Stripe, Polar, etc.). Services call this interface
instead of provider SDKs directly. The types in `billing/provider_types.go`
carry data between services and provider implementations.

The Stripe implementation lives in `billing/stripeprovider/` and serves as the
reference.

## Interface

```go
type Provider interface {
    // Customer CRUD
    CreateCustomer(ctx, CreateCustomerParams) (*ProviderCustomer, error)
    UpdateCustomer(ctx, providerID, UpdateCustomerParams) (*ProviderCustomer, error)
    DeleteCustomer(ctx, providerID) error
    GetCustomer(ctx, providerID) (*ProviderCustomer, error)
    ListPaymentMethods(ctx, customerProviderID) ([]ProviderPaymentMethod, error)

    // Product / Price catalog
    CreateProduct(ctx, CreateProductParams) error
    UpdateProduct(ctx, providerID, UpdateProductParams) error
    CreatePrice(ctx, CreatePriceParams) (providerID string, err error)
    UpdatePrice(ctx, providerID, UpdatePriceParams) error

    // Subscription lifecycle
    CreateSubscription(ctx, CreateSubscriptionParams) (*ProviderSubscription, error)
    GetSubscription(ctx, providerID) (*ProviderSubscription, error)
    CancelSubscription(ctx, providerID, CancelSubscriptionParams) (*ProviderSubscription, error)
    UpdateSubscriptionItems(ctx, providerID, UpdateSubscriptionItemsParams) error

    // Subscription scheduling
    GetSchedule(ctx, scheduleID) (*ProviderSchedule, error)
    CreateScheduleFromSubscription(ctx, subscriptionProviderID) (*ProviderSchedule, error)
    UpdateSchedule(ctx, scheduleID, UpdateScheduleParams) (*ProviderSchedule, error)

    // Checkout / billing portal
    CreateCheckoutSession(ctx, CreateCheckoutSessionParams) (*ProviderCheckoutSession, error)
    GetCheckoutSession(ctx, providerID) (*ProviderCheckoutSession, error)
    CreateBillingPortalSession(ctx, CreateBillingPortalParams) (url string, err error)

    // Invoice management
    ListInvoices(ctx, customerProviderID) ([]ProviderInvoice, error)
    GetUpcomingInvoice(ctx, customerProviderID) (*ProviderInvoice, error)
    CreateInvoice(ctx, CreateInvoiceParams) (*ProviderInvoice, error)
    CreateInvoiceItem(ctx, CreateInvoiceItemParams) error
    GetInvoice(ctx, providerID) (*ProviderInvoice, error)

    // Webhook verification
    VerifyWebhook(payload, signature, secrets) (*WebhookEvent, error)
}
```

## Sentinel Errors

Implementations must return these sentinel errors where appropriate:

| Error                     | When to return                                          |
|---------------------------|---------------------------------------------------------|
| `ErrNotFoundInProvider`   | Resource does not exist at the provider (deleted, never created, etc.) |
| `ErrNoUpcomingInvoice`    | `GetUpcomingInvoice` finds no upcoming invoice          |

Callers use `errors.Is()` to check these.

## Method Contracts

### Customer

- `CreateCustomer` registers a new billing customer. Returns the provider's
  customer record including its provider-assigned ID.
- `DeleteCustomer` must be idempotent — if the customer is already deleted at
  the provider, return nil.
- `GetCustomer` returns the current state. Used by the background sync to
  reconcile provider state into Frontier's database.
- `ListPaymentMethods` returns all payment methods. Set `IsDefault: true` on
  the customer's default method.

### Product / Price

- `CreateProduct` and `CreatePrice` register catalog entries. The `ID` field in
  `CreateProductParams` is Frontier's ID — use it as the provider's product ID
  if the provider supports client-assigned IDs.
- `CreatePrice` returns the provider-assigned price ID, which Frontier stores
  for later reference in subscriptions and checkout.

### Subscription

- `GetSubscription` must populate `Items` with price and product IDs, and
  `Schedule` with the schedule reference if one exists.
- `CancelSubscription` cancels immediately. The caller handles scheduling
  cancellation at period end via `UpdateSchedule`.

### Scheduling

Subscription schedules allow plan changes to take effect at the end of a billing
period. If your provider does not support schedules natively, you may need to
emulate them (e.g., by storing pending changes and applying them via webhook
when the period ends).

- `CreateScheduleFromSubscription` creates a schedule from an existing
  subscription. The schedule must include expanded phase and item data.
- `UpdateSchedule` replaces the schedule's phases. `SchedulePhaseInput` fields
  like `EndDateNow` and `Iterations` are hints — map them to your provider's
  equivalent.

### Checkout

- `CreateCheckoutSession` returns a URL the frontend redirects to. The `Mode`
  field is one of `"subscription"`, `"payment"`, or `"setup"`.
- `CreateBillingPortalSession` returns a URL for customer self-service
  (payment methods, invoices, etc.). If your provider has no equivalent, return
  an error.

### Invoice

- `ListInvoices` returns all invoices for a customer. Must include line items.
- `GetUpcomingInvoice` returns the next invoice that will be generated. Return
  `ErrNoUpcomingInvoice` if none exists.
- `CreateInvoice` + `CreateInvoiceItem` are used for credit overdraft invoicing.
  After creating items, the caller fetches the final invoice via `GetInvoice`.

### Webhook

- `VerifyWebhook` validates the webhook signature and parses the event. The
  `secrets` slice supports secret rotation — try each until one succeeds.
- Return `billing.Event*` constants for recognized event types. Unknown
  event types should be returned as-is — the caller ignores unrecognized types.

## Background Sync

Several services run background cron jobs that call provider methods
periodically (customer sync, subscription sync, invoice sync, checkout sync).
Your implementation should handle being called frequently and return consistent
state. Rate limiting and jitter are handled by the caller.

## Adding a New Provider

1. Create a new package under `billing/` (e.g., `billing/polarprovider/`).
2. Implement `billing.Provider`.
3. Add a compile-time check: `var _ billing.Provider = (*Provider)(nil)`.
4. Wire it in `cmd/serve.go` based on configuration.
5. Map your provider's webhook event types to the `billing.Event*` constants.

## Configuration

The current config in `billing.Config` has Stripe-specific fields
(`StripeKey`, `StripeAutoTax`, `StripeWebhookSecrets`). When adding a new
provider, extend the config with a provider selector and provider-specific
sub-configs:

```yaml
billing:
  provider: stripe  # or "polar"
  stripe:
    key: sk_...
    auto_tax: true
    webhook_secrets: [...]
  polar:
    key: pol_...
```
