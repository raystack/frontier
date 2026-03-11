# Billing Subscriptions

Billing subscriptions enable billing customers to subscribe to recurring plans on the billing engine(currently, only Stripe). Subscriptions can be availed on a trial basis as well, where a trial period can be set for a plan, and customers can use the subscription for the trial period without paying any charges.
On Frontier, plans and subscriptions are not weighted, and do not have any hierarchy. Thus, there is no inherent concept of upgrades and downgrades when it comes to subscripitons. Whenever a customer chooses to change their plan, the plan amounts are prorated as per the configuration in Frontier. Frontier provides various configurations around trials, default subscriptions, prorations etc. which are described in the next section.


## Subscription States

A subscription can have multiple states which are as follows:

- **Active**: When a subscription is active
- **Trialing**: When a subscription is in trial phase
- **Past due**: When a subscription is unpaid, and the due date of the invoice has already passed
- **Canceled**: When a subscription has been cancelled


## Configuration

Frontier provides various configurations for subscription management, which are nested under the `plan_change` key as follows:

- `proration_behavior`: Specifies what kind of behaviour we want for prorations, when a customer changes their subscription midway through the subscription tenure. Can take values of `create_prorations`, `none` and `always_invoice`

- `immediate_proration_behavior`: Specifies the proration behaviour to apply when the plan is changed immediately, instead of waiting for the next billing cycle (This is triggered by setting the `immediate` flag in plan update API). Can be one of `create_prorations`, `none` and `always_invoice`

- `collection_method`: Specifies the behaviour of payment collection when a plan is changed. It can take values of `charge_automatically` and `send_invoice`

## Syncing billing subscription data

The SyncWithProvider method ensures that the subscription state in the application is synchronized with the state maintained by the billing provider (e.g. Stripe). This process involves validating the state of each subscription, reconciling discrepancies, and updating subscription details to reflect the current state accurately.

The billing sync is done by fetching all subscriptions for each customers. Each subscription is processed sequentially. 
For each active subscription:

- There is an attempt to reconcile the subscription with the provider.
- If a subscription is not found on the provider:
  - Test mode resources are marked as canceled.
  - Non-test subscriptions are logged as errors.
- If successful, we compare the subscription’s metadata against the provider's data to determine if updates are required.
- Fields that are syncronized for updation are as follows:

  - State: e.g., active, canceled, trialing.
  - Timestamps: e.g., CanceledAt, EndedAt, TrialEndsAt.
  -  Plan ID: Update the subscription plan and append it to the plan history if changed.
  - Phase: Modify the subscription's phase and reason if the plan or schedule has changed.

- If the subscription is active and has a valid plan ID:
  - Retrieve the plan details that exist in Frontier.
  - Update the quantity of the subscription product if it uses per-seat pricing.
  - Ensure the subscription is complemented with applicable free credits using ensureCreditsForPlan.

The subscription updates are thread safe and done via locking.
