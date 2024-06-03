# Billing Configurations

Frontier provides billing and subscription related capabilities, which can be customized using various configs. Frontier uses Stripe as the billing engine to manage payments and subscriptions. For more details on concepts related to billing on Frontier, please [refer to this guide](../billing/introduction.md).

This document provides instructions on how to configure the billing settings for managing payment and subscriptions using Frontier.

## Prerequisites

- Stripe key (generated when a Stripe account is created)
- A file containing plans and products to be offered (Optional, but recommended)

## Configuration

| **Field**                            | **Description**                                                                                                                                                                     | **Example** | **Required**      |
| ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | ----------------- |
| **billing.stripe_key**                         | Developer key generated on Stripe                                                                                                                                                 | sk_test_abcdefghijklmnopqrstuvwxyz        | Yes               |
| **billing.stripe_auto_tax**                         | Set to true if you want Stripe to automatically apply tax on the invoices as per the customer's location                                                                                                                                                 | false        | No (default: false)               |
| **billing.stripe_webhook_secrets**                         | Webhook secrets to be used for validating stripe webhooks events                                                                                                                                                 | []        | No               |
| **billing.plans_path**                         | Path to a folder which has yaml files describing the products and plans that need to be created on the billing engine (Stripe). The plans and products in these files will be automatically created on Stripe as part of migration during application startup                                                                                                                                                 | "file:///tmp/plans"        | No (but recommended)               |
| **billing.default_plan**                         | Name of the plan that should be used subscribed automatically when the org is created. It also automatically creates an empty billing account under the org.<br/>**Note: The plan name provided here should exist in the billing engine.**                                                                                                                                                 | "standard_plan"        | No               |
| **billing.default_currency**                         | Default currency to be used for billing if not provided by the user                                                                                                                                                 | "USD"        | No (but recommended)               |
| **billing.plan_change.proration_behavior**                         | Proration behaviour to be used when a subscription is changed, or its quantity is updated. Can be one of "create_prorations", "always_invoice" or "none"                                                                                                                                                  | "create_prorations"        | No (default: create_prorations)               |
| **billing.plan_change.immediate_proration_behavior**                         | Proration behaviour to be used when a subscription is to be updated immediately, instead of waiting for the next billing cycle. Can be one of "create_prorations", "always_invoice" or "none"                                                                                                                                                  | "create_prorations"        | No (default: create_prorations)               |
| **billing.plan_change.collection_method**                         | The collection method determines how payment is to be processed for a product or subscription. It can take the following values: "charge_automatically", "send_invoice"                                                                                                                                                  | "create_prorations"        | No (default: charge_automatically)               |
| **billing.product.seat_change_behavior**                         | Determines how seat count needs to be adjusted with change in number of users in an org. It can tak two values: <br/>1. **"exact"** - This changes the seat count to the exact number of users within the organization (on both, increment as well as decrement of users)<br/>2. **"incremental"** - This changes the seat count to the number of users within the organization, but does not decrease the seat count if users are reduced. This can be used in scenarios where we want a "seat" based billing policy, where organizations purchase "seats" which can be filled by any user, and removal of such a user simply results in an empty seat (that can be filled by someone else)                                                                                                                                                | "exact"        | No (default: exact)               |