import dayjs from 'dayjs';
import {
  V1Beta1Subscription,
  BillingAccountAddress,
  V1Beta1Plan,
  V1Beta1PaymentMethod
} from '~/src';
import {
  IntervalKeys,
  IntervalLabelMap,
  IntervalPricing,
  PaymentMethodMetadata
} from '~/src/types';
import { SUBSCRIPTION_STATES } from './constants';
import slugify from 'slugify';

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

export interface PlanChangeAction {
  btnLabel: string;
  btnDoneLabel: string;
  btnLoadingLabel: string;
  showModal?: boolean;
  disabled?: boolean;
  immediate?: boolean;
  btnVariant: 'secondary' | 'primary';
}

export const getPlanChangeAction = (
  nextPlanWeightage?: number,
  currentPlanWeightage?: number
): PlanChangeAction => {
  const diff = (nextPlanWeightage || 0) - (currentPlanWeightage || 0);

  if (diff > 0 || !currentPlanWeightage) {
    return {
      btnLabel: 'Upgrade',
      btnDoneLabel: 'Upgraded',
      btnLoadingLabel: 'Upgrading',
      btnVariant: 'primary',
      immediate: true,
      showModal: true
    };
  } else if (diff < 0 || nextPlanWeightage === undefined) {
    return {
      btnLabel: 'Downgrade',
      btnDoneLabel: 'Downgraded',
      btnLoadingLabel: 'Downgrading',
      btnVariant: 'secondary',
      showModal: true
    };
  } else {
    return {
      btnLabel: 'Change',
      btnDoneLabel: 'Changed',
      btnLoadingLabel: 'Changing',
      btnVariant: 'primary',
      immediate: true
    };
  }
};

export const checkSimilarPlans = (plan1: V1Beta1Plan, plan2: V1Beta1Plan) => {
  const plan1Metadata = (plan1.metadata as Record<string, string>) || {};
  const plan2Metadata = (plan2.metadata as Record<string, string>) || {};
  const plan1Slug = plan1Metadata?.plan_group_id || makePlanSlug(plan1);
  const plan2Slug = plan2Metadata?.plan_group_id || makePlanSlug(plan2);
  return plan1Slug === plan2Slug;
};

export function getFormattedNumberString(num: Number = 0) {
  const numString = num.toString();
  const length = numString.length;

  return numString.split('').reduce((acc, val, i) => {
    const diff = length - i;
    if (diff % 3 === 0 && diff < length) {
      return acc + ',' + val;
    }
    return acc + val;
  }, '');
}

interface getPlanNameWithIntervalOptions {
  hyphenSeperated?: boolean;
}

export function getPlanIntervalName(plan: V1Beta1Plan = {}) {
  return IntervalLabelMap[plan?.interval as IntervalKeys];
}

export function getPlanNameWithInterval(
  plan: V1Beta1Plan = {},
  { hyphenSeperated }: getPlanNameWithIntervalOptions = {}
) {
  const interval = getPlanIntervalName(plan);
  return hyphenSeperated
    ? `${plan?.title} - ${interval}`
    : `${plan?.title} (${interval})`;
}

export function makePlanSlug(plan: V1Beta1Plan): string {
  const productIds = plan?.products
    ?.map(p => p.id)
    .sort()
    .join('-');
  const titleSlug = slugify(plan.title || '', { lower: true });
  return `${titleSlug}-${productIds}`;
}

export function getPlanPrice(plan: V1Beta1Plan) {
  const planInterval = (plan?.interval || '') as IntervalKeys;
  return (
    plan?.products?.reduce((acc, product) => {
      product.prices?.forEach(price => {
        if (price.interval === planInterval) {
          acc.amount = Number(acc.amount || 0) + Number(price.amount);
          acc.currency = price.currency || '';
        }
      });
      return acc;
    }, {} as IntervalPricing) || ({} as IntervalPricing)
  );
}

export function getDefaultPaymentMethod(
  paymentMethods: V1Beta1PaymentMethod[] = []
) {
  const defaultMethod = paymentMethods.find(pm => {
    const metadata = pm.metadata as PaymentMethodMetadata;
    return metadata.default;
  });

  return defaultMethod ? defaultMethod : paymentMethods[0];
}
