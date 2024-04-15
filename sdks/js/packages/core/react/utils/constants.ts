export const DEFAULT_DATE_FORMAT = 'DD MMM YYYY';
export const DEFAULT_DATE_SHORT_FORMAT = 'DD MMM';
export const DEFAULT_TOKEN_PRODUCT_NAME = 'token';

export const SUBSCRIPTION_STATES = {
  ACTIVE: 'active',
  PAST_DUE: 'past_due',
  TRIALING: 'trialing'
} as const;

export const INVOICE_STATES = {
  OPEN: 'open',
  PAID: 'paid',
  DRAFT: 'draft'
} as const;
