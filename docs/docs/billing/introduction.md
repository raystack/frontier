# Introduction

## Overview
The Billing Service is a comprehensive solution for managing all billing-related operations in your Go service. It provides a robust Pricing Engine that allows you to create and manage various pricing plans and features, onboard customers, and handle all billing transactions.

The service is designed to be flexible and scalable, accommodating a wide range of billing scenarios. It supports the creation of features and plans based on a variety of monthly and yearly pricing, and it integrates with popular billing engines like Stripe for seamless payment processing.

## Features
### Billing Accounts
Billing accounts are the primary entities in the Billing Service. They represent the billing relationship between an organization and a customer. You can create, update, delete, and retrieve billing accounts, as well as list all billing accounts for an organization.

### Subscriptions
Subscriptions represent a customer's commitment to pay for a specific plan on a recurring basis. You can create, update, and cancel subscriptions, as well as list all subscriptions for a billing account.

### Products and Plans
Products and plans are the building blocks of your pricing model. Products represent individual capabilities or resources that you charge for, while plans are collections of products offered at a specific price. You can create, update, and retrieve products and plans, as well as list all products and plans.

### Features (Product features)
Features are individual functionalities that a product offers. They cannot be individually purchased, but are offerings of a product itself. When a product is purchased by a customer, the customer is entitled to all features offered by that product.

### Entitlement checks
Certain functionalities can be restricted for customers, depending on the plans and products they have purchased. Frontier offers entitlement checks, where we can check a customer's entitlement to a particular feature depending on their subscriptions and purchased products.

### Checkouts
Checkouts represent the process of a customer agreeing to a subscription or purchasing a feature. You can create checkouts and list all checkouts for a billing account.

### Billing Balance and Transactions
The Billing Service provides functionality to check the balance of a billing account and list all transactions for a billing account.

## API Capabilities:

- Billing Account Management: Create, update, and delete accounts, managing information like name, email, address, and currency.
- Subscription Management: Create, update, cancel, and list subscriptions associated with specific billing accounts, including metadata for custom information.
- Product Management: Create, update, and list products with configurable pricing models and metadata for custom information.
- Feature Management: Create, update and list features, which act as building blocks of products, specifying functionalities offered by the product
- Checkout Session Management: Create checkout sessions for users to purchase features or start subscriptions, with support for both types and customizable success and cancellation URLs.
- Billing Usage Reporting: Report platform usage for features with information like feature ID, amount, and timestamp for accurate billing calculations.
- Entitlement Verification: Check user access to specific features based on their account for efficient access control and restriction enforcement.
- Billing Balance Access: View the current balance of a specific billing account, providing insight into outstanding charges and payment requirements.
- Plan Management: Create, update, and list plans that define products and pricing for subscriptions.

## Enabling the Billing Service
Here are the steps to enable the billing service on the platform:

- The platform admin creates features using the CreateFeature RPC (optional).
- The platform admin creates products using the CreateProduct RPC.
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

### Create a Checkout for a Product
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

## Sample Plan and Product structure

```yaml
products:
  - name: support_credits
    title: Support Credits
    description: Support for enterprise help
    behavior: credits
    config:
      credit_amount: 100
    prices:
      - name: default
        amount: 20000
        currency: inr
  - name: basic_access
    title: Basic base access
    description: Base access to the platform
    prices:
      - name: monthly
        interval: month
        amount: 100
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
    features:
      - name: starter_feature_1
      - name: starter_feature_2
  - name: starter_per_seat
    title: Starter per seat
    description: Per seat access cost to the platform
    behavior: per_seat
    config:
      seat_limit: 3
    prices:
      - name: monthly
        interval: month
        amount: 20
        currency: inr
      - name: yearly
        interval: year
        amount: 15
        currency: inr
plans:
  - name: basic_monthly
    title: Basic Monthly Plan
    description: Basic Monthly Plan
    interval: month
    products:
      - name: basic_access
  - name: starter_yearly
    title: Starter Plan
    description: Starter Plan
    interval: year
    products:
      - name: starter_access
  - name: starter_monthly
    title: Starter Plan
    description: Starter Plan
    interval: month
    on_start_credits: 50
    products:
      - name: starter_access
      - name: starter_per_seat
```

### Stripe Test clocks

Stripe allows simulating test clocks to test subscriptions, payments, and invoices. This clock needs to be created from the stripe 
dashboard and the clock id must be passed as a request header while creating a new billing customer account for the stripe to 
use this simulated clock. Once the customer is created, there is no need to pass this header and all of its subsequent 
subscriptions will automatically be part of the simulated clock.

Remember that once the clock expires, all of its resources also expire. Only platform users can use this header to test frontier mechanics.

Example:
```
X-Stripe-Test-Clock: clk_123
```

### Product Customizations

Frontier offers different types of product customizations which can be set while creating the product. These are controlled using the `behavior` and `behavior_config` fields on products. 

Behavior can be of three types:
1. `basic` - This is the default behavior of products
2. `credits` - The behavior is set to `credits` when we want a product to offer virtual credits offered by Frontier. When such a product is purchased, virtual credits are automatically credited to the organization's account
3. `per_seat` - Behavior is set to `per_seat` in case of products/subscriptions where have a seat based pricing. When such a product/subscription is purchased, the organization is automatically charged on the basis of number of users they have in an organization. Proration for user quantity changes are handled automatically by Frontier on the basis of proration settings in config.

Once a behavior is set, a `behavior_config` can be defined on the product for more granular control. The `behavior_config` object has the following properties:
1. `credit_amount` - To be used in case the `behavior` is set to `credits`. This denotes the amount of virtual credits to be credited to an organization when the product is purchased.
2. `seat_limit` - To be used in combination with `per_seat` behavior. This restricts the number of users that an organization can have.
3. `min_quantity` - Specifies the minimum quantity of a product that must be purchased
3. `max_quantity` - Specifies the maximum quantity of a product that can be purchased

## Virtual Credits Management

Virtual credits are a form of currency that can be used to consume services based on usage cost. They are typically 
used to provide a pay-as-you-go model for services where the user is charged based on the usage of the service.
This is apart from subscriptions where the user is charged a fixed amount for allowing access to a set of features.

### Virtual Credit Purchase

Virtual credits can be purchased by the user using the `CreateCheckout` RPC with a `CheckoutProductBody` request body.
Product needs to be defined with a behavior of `credits` and a price defined for the product. Product also states how
much credit is provided for the price defined. Once the checkout is successful, the user's account will be credited with the
amount of credits defined in the product.

### Virtual Credit Usage

When reporting usage, the user can specify the amount of credits consumed for the usage. The billing service will then
deduct the credits from the user's account balance based on the usage reported. For example, if the user triggers a
machine learning model that runs for 5 mins, the system user can report the usage as 20 units of credits. The billing
service will then deduct 20 credits from the user's account balance. If the user does not have enough credits, the
usage will be rejected. The user can then purchase more credits to continue using the service.
Usage is reported via `CreateBillingUsage` RPC(`/v1beta1/organizations/{org_id}/billing/{billing_id}/usages`).

### Virtual Credit Balance

The user can check their credit balance using the `GetBillingBalance` RPC(`/v1beta1/organizations/{org_id}/billing/{billing_id}/balance`).
The response will include the amount of credits available in the user's account. The user can then decide to purchase more
credits if needed.

### Reverting Virtual Credit Usage

In case of any issues with the usage reported, the user can revert the usage by using the `RevertBillingUsage` RPC(`/v1beta1/organizations/{org_id}/billing/{billing_id}/usages/{usage_id}/revert`).
This will add the credits back to the user's account balance. The user can then use these credits for future usage.
An example use case is if the machine learning model fails to run due to an error, the usage can be reverted and the credits
can be used for the next run. Revert can be full or partial based on the requirement.

### VC Internals

Virtual credits are maintained using a double entry bookkeeping system. When credits are purchased, a credit transaction is
created with the amount of credits purchased. When usage is reported, a debit transaction is created with the amount of credits
consumed. The balance is calculated by summing up all the credit transactions and subtracting the sum of all the debit transactions.
Each entry creates two transactions, one for the credit side and one for the debit side. This ensures that the balance is always
accurate and consistent. For example, when user purchases 100 credits, a credit transaction is created with 100 credits
to customer account and a debit transaction is created with 100 credits to the system account.
