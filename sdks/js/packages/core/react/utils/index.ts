import dayjs from 'dayjs';
import { V1Beta1Subscription, BillingAccountAddress } from '~/src';
import { IntervalPricingWithPlan } from '~/src/types';
import { SUBSCRIPTION_STATES } from './constants';

export const AuthTooltipMessage =
  'You don’t have access to perform this action';

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
    .filter(
      sub =>
        sub.state === SUBSCRIPTION_STATES.ACTIVE ||
        sub.state === SUBSCRIPTION_STATES.PAST_DUE
    )
    .sort((a, b) => (dayjs(a.updated_at).isAfter(b.updated_at) ? -1 : 1));

  return activeSubscriptions[0];
};

export const getPlanChangeAction = (
  nextPlan: IntervalPricingWithPlan,
  currentPlan?: IntervalPricingWithPlan
) => {
  const diff = nextPlan.weightage - (currentPlan?.weightage || 0);
  if (diff > 0) {
    return {
      btnLabel: 'Upgrade',
      btnLoadingLabel: 'Upgrading'
    };
  } else if (diff < 0) {
    return {
      btnLabel: 'Downgrade',
      btnLoadingLabel: 'Downgrading'
    };
  } else {
    return {
      btnLabel: 'Change',
      btnLoadingLabel: 'Changing'
    };
  }
};
