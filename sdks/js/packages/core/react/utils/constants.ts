export const DEFAULT_DATE_FORMAT = 'DD MMM YYYY';
export const DEFAULT_DATE_SHORT_FORMAT = 'DD MMM';
export const DEFAULT_TOKEN_PRODUCT_NAME = 'token';
export const DEFAULT_PLAN_UPGRADE_MESSAGE =
  'Any remaining balance from your current plan will be prorated and credited to your account in future billing cycles.';

export const NEGATIVE_BALANCE_TOOLTIP_MESSAGE =
  'This negative amount shows a credit balance for prorated seats, which will be applied to future invoices.';

export const SUBSCRIPTION_STATES = {
  ACTIVE: 'active',
  PAST_DUE: 'past_due',
  TRIALING: 'trialing',
  CANCELED: 'canceled'
} as const;

export const INVOICE_STATES = {
  OPEN: 'open',
  PAID: 'paid',
  DRAFT: 'draft'
} as const;

export const DEFAULT_API_PLATFORM_APP_NAME = 'Frontier';
