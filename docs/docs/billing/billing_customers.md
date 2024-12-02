# Billing Customers

Billing customers represent a customer entity with fields for storing billing related customer data like ID, organization ID (OrgID), currency etc. It also includes a field called `provider_id` which represents the ID of the customer in a billing engine (Frontier supports Stripe as the default billing engine).

## Configuration

This configuration section streamlines the customer onboarding process by automating account creation, plan assignment, and credit allocation. It provides flexibility in managing customer accounts and billing preferences.

### Key Settings:

- `auto_create_with_org`: Determines whether a default customer account should be automatically created when a new organization is created.
  - Value: true or false
  - Default: true

- `default_plan`: Specifies the default plan that should be automatically subscribed to when a new organization is created. This also triggers the creation of an empty billing account under the organization.
  - Value: Plan name (string)
  - Default: Empty string

- `default_offline`: Controls the default offline status for customer accounts. If set to true, the customer account will not be registered with the billing provider.
  - Value: true or false
  - Default: false

- `onboard_credits_with_org`: Specifies the amount of free credits to be added to a customer account when it's created as part of an organization.
  - Value: Integer (number of credits)
  - Default: 0


### Billing Customer Creation

Using the configurations described above, the creation process of a billing account can be customized. Billing customers are billing counterparts to the "organization" entity in Frontier, and are created as soon as an organization is created.
During billing account creation, we check for existing active billing accounts within the same organization. If they exist, the new account is not created. New billing accounts are also not created when the organization has existing accounts with negative credit balance. 
In order to ensure that billing accounts are created on Stripe only when a customer actually tries to purchase something, we can set the `default_offline` flag to true in the config. This makes sure that billing accounts are created in Frontier without a counterpart in Stripe. A Stripe account is created during the checkout flow in such cases.

## Syncing Billing Customer Data

In order to sync changes that have been made directly on Stripe (instead of via Frontier), Frontier has a background syncer that priodically syncs customer data to ensure that the data on Frontier is consistent with that on Stripe.
This is done by initialising a background worker that runs periodically. The frequency of this worker is configurable in the `refresh_interval` configuration parameter of Frontier

The working of the syncer is as follows:

1. Acquire a lock (s.mu) on the syncer's run (in case a previous syncer is still running) in Frontier to prevent race conditions when accessing shared resources.
2. Fetch customer details from Stripe using the provided customer ID.
3. Check if the customer is marked as deleted in Stripe. If so, and the local customer is active, disable the local customer.
4. Selective Updates (Active Customers Only): Process updates only if the local customer is active. Various customer fields are compared between the local customr object and the retrieved Stripe customer object such as:
    - Tax data (using a custom comparison function)
    - Phone number
    - Email (if not empty)
    - Name
    - Currency
    - Address details (city, country, address lines, postal code, state)
5. If any discrepancies are found in the customer data between Stripe and Frontier, the necessary changes are synced and saved to the database, and the lock acquired in (1) is released so that the syncer can run again in the next iteration





