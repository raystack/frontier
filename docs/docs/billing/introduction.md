# Introduction

## Overview
The Billing Service is a comprehensive solution for managing all billing-related operations in your Go service. It provides a robust Pricing Engine that allows you to create and manage various pricing plans and features, onboard customers, and handle all billing transactions.

The service is designed to be flexible and scalable, accommodating a wide range of billing scenarios. It supports the creation of features and plans based on a variety of monthly and yearly pricing, and it integrates with popular billing engines like Stripe for seamless payment processing.

## Features
### Billing Accounts
Billing accounts are the primary entities in the Billing Service. They represent the billing relationship between an organization and a customer. You can create, update, delete, and retrieve billing accounts, as well as list all billing accounts for an organization.

### Subscriptions
Subscriptions represent a customer's commitment to pay for a specific plan on a recurring basis. You can create, update, and cancel subscriptions, as well as list all subscriptions for a billing account.

### Features and Plans
Features and plans are the building blocks of your pricing model. Features represent individual capabilities or resources that you charge for, while plans are collections of features offered at a specific price. You can create, update, and retrieve features and plans, as well as list all features and plans.

### Checkouts
Checkouts represent the process of a customer agreeing to a subscription or purchasing a feature. You can create checkouts and list all checkouts for a billing account.

### Billing Balance and Transactions
The Billing Service provides functionality to check the balance of a billing account and list all transactions for a billing account.

## API Capabilities:

- Billing Account Management: Create, update, and delete accounts, managing information like name, email, address, and currency.
- Subscription Management: Create, update, cancel, and list subscriptions associated with specific billing accounts, including metadata for custom information.
- Feature Management: Create, update, and list features with configurable pricing models and metadata for custom information.
- Checkout Session Management: Create checkout sessions for users to purchase features or start subscriptions, with support for both types and customizable success and cancellation URLs.
- Billing Usage Reporting: Report platform usage for features with information like feature ID, amount, and timestamp for accurate billing calculations.
- Entitlement Verification: Check user access to specific features based on their account for efficient access control and restriction enforcement.
- Billing Balance Access: View the current balance of a specific billing account, providing insight into outstanding charges and payment requirements.
- Plan Management: Create, update, and list plans that define features and pricing for subscriptions.

## Enabling the Billing Service
Here are the steps to enable the billing service on the platform:

- The platform admin creates features using the CreateFeature RPC.
- The platform admin creates plans using the CreatePlan RPC. These plans can be based on a variety of monthly or yearly pricing.
- The platform admin configures the billing engine (like Stripe).

## Onboarding a Customer
Here are the steps to onboard a customer on this service:

- Create a billing account for the customer's organization using the CreateBillingAccount RPC.
- The customer subscribes to a plan using the CreateCheckout RPC with a CheckoutSubscriptionBody.
- The customer can also buy virtual credits using the CreateCheckout RPC with a CheckoutFeatureBody.
- The customer can check their balance using the GetBillingBalance RPC.

## Onboarding

### Create a Billing Account
Endpoint: POST `/v1beta1/organizations/{org_id}/billing`

RPC: `CreateBillingAccount`

Request:

```json

{
    "org_id": "org123",
    "body": {
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phone": "+1234567890",
        "address": {
            "line1": "123 Main St",
            "line2": "Apt 4B",
            "city": "New York",
            "state": "NY",
            "postal_code": "10001",
            "country": "USA"
        },
        "currency": "usd"
    }
}
```

Response:
```json
{
    "billing_account": {
        "id": "ba123",
        "org_id": "org123",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "phone": "+1234567890",
        "address": {
        "line1": "123 Main St",
        "line2": "Apt 4B",
        "city": "New York",
        "state": "NY",
        "postal_code": "10001",
        "country": "USA"
    },
    "currency": "usd",
    "created_at": "2022-01-01T00:00:00Z",
    "updated_at": "2022-01-01T00:00:00Z"
    }
}
```

### Create a Checkout for a Subscription
Endpoint: POST `/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts`

RPC: `CreateCheckout`

Request:

```json
{
    "org_id": "org123",
    "billing_id": "ba123",
    "success_url": "https://example.com/success",
    "cancel_url": "https://example.com/cancel",
    "subscription_body": {
        "plan": "plan123",
        "trail_days": 14
    }
}
```
Response:
```json
{
    "checkout_session": {
        "id": "cs123",
        "checkout_url": "https://checkout.stripe.com/pay/cs_test_123",
        "success_url": "https://example.com/success",
        "cancel_url": "https://example.com/cancel",
        "created_at": "2022-01-01T00:00:00Z",
        "updated_at": "2022-01-01T00:00:00Z",
        "expire_at": "2022-01-02T00:00:00Z"
    }
}
```

### Create a Checkout for a Feature
Endpoint: POST `/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts`

RPC: `CreateCheckout`

Request:

```json
{
    "org_id": "org123",
    "billing_id": "ba123",
    "success_url": "https://example.com/success",
    "cancel_url": "https://example.com/cancel",
    "feature_body": {
      "feature": "feature123"
    }
}
```
Response:
```json
{
    "checkout_session": {
        "id": "cs123",
        "checkout_url": "https://checkout.stripe.com/pay/cs_test_123",
        "success_url": "https://example.com/success",
        "cancel_url": "https://example.com/cancel",
        "created_at": "2022-01-01T00:00:00Z",
        "updated_at": "2022-01-01T00:00:00Z",
        "expire_at": "2022-01-02T00:00:00Z"
    }
}
```

### Get Billing Balance
Endpoint: GET `/v1beta1/organizations/{org_id}/billing/{id}/balance`

RPC: `GetBillingBalance`

Request:

```json
{
    "id": "ba123",
    "org_id": "org123"
}
```
Response:
```json
{
    "balance": {
        "amount": 10000,
        "currency": "usd",
        "updated_at": "2022-01-01T00:00:00Z"
    }
}
```

## Sample Plan and Feature structure

```yaml
features:
  - name: support_credits
    title: Support Credits
    description: Support for enterprise help
    credit_amount: 100
    prices:
      - name: default
        amount: 20000
        currency: inr
#  - name: basic_access
#    title: Basic base access
#    description: Base access to the platform
#    prices:
#      - name: monthly
#        interval: month
#        amount: 50000
#        currency: inr
  - name: starter_plan_credits
    title: Starter Plan Credits
    description: One time credits for joining Starter Plan
    credit_amount: 50
    prices:
      - name: default
        amount: 10000
        currency: inr
  - name: starter_access
    title: Starter base access
    description: Base access to the platform
    prices:
      - name: monthly
        interval: month
        amount: 1000
        currency: inr
      - name: yearly
        interval: year
        amount: 5000
        currency: inr
#  - name: enterprise_access
#    title: Enterprise base access for year
#    description: Base access to the platform
#    prices:
#      - name: default
#        interval: year
#        amount: 8000
#        currency: inr
plans:
#  - name: basic_monthly
#    title: Basic Monthly Plan
#    description: Basic Monthly Plan
#    interval: month
#    features:
#      - name: basic_access
  - name: starter_yearly
    title: Starter Plan
    description: Starter Plan
    interval: year
    features:
      - name: starter_access
  - name: starter_monthly
    title: Starter Plan
    description: Starter Plan
    interval: month
    features:
      - name: starter_access
      - name: starter_plan_credits
#  - name: enterprise_yearly
#    title: Enterprise Plan
#    description: Enterprise Plan
#    interval: year
#    features:
#      - name: enterprise_access
```