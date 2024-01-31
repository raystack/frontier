export const DEFAULT_DATE_FORMAT = 'DD MMM YYYY';

export const SUBSCRIPTION_STATES = {
  ACTIVE: 'active',
  PAST_DUE: 'past_due'
} as const;

export const INVOICE_STATES = {
  OPEN: 'open',
  PAID: 'paid',
  DRAFT: 'draft'
} as const;
