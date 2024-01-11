import dayjs from 'dayjs';
import { BillingAccountAddress, V1Beta1Subscription } from '~/src';

const SUBSCRIPTION_STATES = {
  ACTIVE: 'active'
};

export const converBillingAddressToString = (
  address?: BillingAccountAddress
) => {
  if (!address) return '';
  const { line1, line2, city, state, country, postal_code } = address;
  return [line1, line2, city, state, country, postal_code]
    .filter(v => v)
    .join(', ');
};

export const getActiveSubscription = (subscriptions: V1Beta1Subscription[]) => {
  const activeSubscriptions = subscriptions
    .filter(sub => sub.state === SUBSCRIPTION_STATES.ACTIVE)
    .sort((a, b) => (dayjs(a.updated_at).isAfter(b.updated_at) ? -1 : 1));

  return activeSubscriptions[0];
};
