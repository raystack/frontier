import { BillingAccountAddress } from '~/src';
import {
  BasePlan,
  IntervalKeys,
  IntervalLabelMap,
  IntervalPricing,
  PaymentMethodMetadata
} from '~/src/types';
import { SUBSCRIPTION_STATES } from './constants';
import slugify from 'slugify';
import { NIL as NIL_UUID } from 'uuid';
import type { GooglerpcStatus } from '~/src';
import {
  FeatureSchema,
  PaymentMethod,
  Plan,
  PlanSchema,
  Subscription
} from '@raystack/proton/frontier';
import { timestampToDayjs } from '../../utils/timestamp';
import { create } from '@bufbuild/protobuf';

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

export const getActiveSubscription = (subscriptions: Subscription[]) => {
  const activeSubscriptions = subscriptions
    .filter(
      sub =>
        sub.state === SUBSCRIPTION_STATES.ACTIVE ||
        sub.state === SUBSCRIPTION_STATES.PAST_DUE
    )
    .sort((a, b) =>
      timestampToDayjs(a.updatedAt)?.isAfter(timestampToDayjs(b.updatedAt))
        ? -1
        : 1
    );

  return activeSubscriptions[0];
};

export const getTrialingSubscription = (subscriptions: Subscription[]) => {
  const activeSubscriptions = subscriptions
    .filter(sub => sub.state === SUBSCRIPTION_STATES.TRIALING)
    .sort((a, b) =>
      timestampToDayjs(a.updatedAt)?.isAfter(timestampToDayjs(b.updatedAt))
        ? -1
        : 1
    );

  return activeSubscriptions[0];
};

export interface PlanChangeAction {
  btnLabel: string;
  btnDoneLabel: string;
  btnLoadingLabel: string;
  showModal?: boolean;
  disabled?: boolean;
  immediate?: boolean;
  btnVariant: 'outline' | 'solid';
  btnColor: 'neutral' | 'accent';
  btnSize: 'small';
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
      btnVariant: 'solid',
      btnColor: 'accent',
      btnSize: 'small',
      immediate: true,
      showModal: true
    };
  } else if (diff < 0 || nextPlanWeightage === undefined) {
    return {
      btnLabel: 'Downgrade',
      btnDoneLabel: 'Downgraded',
      btnLoadingLabel: 'Downgrading',
      btnVariant: 'outline',
      btnColor: 'neutral',
      btnSize: 'small',
      showModal: true
    };
  } else {
    return {
      btnLabel: 'Change',
      btnDoneLabel: 'Changed',
      btnLoadingLabel: 'Changing',
      btnVariant: 'solid',
      btnColor: 'accent',
      btnSize: 'small',
      immediate: true
    };
  }
};

export const checkSimilarPlans = (
  plan1: Plan = create(PlanSchema, {}),
  plan2: Plan = create(PlanSchema, {})
) => {
  const plan1Metadata = (plan1.metadata as Record<string, string>) || {};
  const plan2Metadata = (plan2.metadata as Record<string, string>) || {};
  const plan1Slug = plan1Metadata?.plan_group_id || makePlanSlug(plan1);
  const plan2Slug = plan2Metadata?.plan_group_id || makePlanSlug(plan2);
  return plan1Slug === plan2Slug;
};

export function getFormattedNumberString(num: Number = 0) {
  const numString = num.toString();
  const isNegative = numString.startsWith('-');
  const sign = isNegative ? '-' : '';
  const withoutSign = isNegative ? numString.slice(1) : numString;
  const [integerPart, decimalPart] = withoutSign.split('.');

  const length = integerPart.length;
  const formattedInteger = integerPart.split('').reduce((acc, val, i) => {
    const diff = length - i;
    if (diff % 3 === 0 && diff < length) {
      return acc + ',' + val;
    }
    return acc + val;
  }, '');

  return (
    sign +
    formattedInteger +
    (decimalPart !== undefined ? '.' + decimalPart : '')
  );
}

interface getPlanNameWithIntervalOptions {
  hyphenSeperated?: boolean;
}

export function getPlanIntervalName(plan: Plan = create(PlanSchema, {})) {
  return IntervalLabelMap[plan?.interval as IntervalKeys];
}

export function getPlanNameWithInterval(
  plan: Plan = create(PlanSchema, {}),
  { hyphenSeperated }: getPlanNameWithIntervalOptions = {}
) {
  const interval = getPlanIntervalName(plan);
  return hyphenSeperated
    ? `${plan?.title} - ${interval}`
    : `${plan?.title} (${interval})`;
}

export function makePlanSlug(plan: Plan): string {
  const productIds = plan?.products
    ?.map(p => p.id)
    .sort()
    .join('-');
  const titleSlug = slugify(plan.title || '', { lower: true });
  return `${titleSlug}-${productIds}`;
}

export function getPlanPrice(plan: Plan) {
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

export function getDefaultPaymentMethod(paymentMethods: PaymentMethod[] = []) {
  const defaultMethod = paymentMethods.find(pm => {
    const metadata = pm.metadata as PaymentMethodMetadata;
    return metadata.default;
  });

  return defaultMethod ? defaultMethod : paymentMethods[0];
}

export const enrichBasePlan = (plan?: BasePlan): Plan | undefined => {
  const features = Object.entries(plan?.features || {}).map(([key, value]) => {
    return create(FeatureSchema, {
      title: key,
      metadata: {
        [plan?.title || '']: value
      }
    });
  });
  return plan
    ? {
        ...plan,
        id: NIL_UUID,
        interval: 'year',
        products: [
          {
            ...plan.products?.[0],
            name: plan.title,
            features: features
          }
        ]
      }
    : undefined;
};

export function generateHashFromString(input: string, hashLength = 6): string {
  if (!input || input.length === 0) {
    throw new Error('Input string cannot be empty');
  }
  if (hashLength < 1 || hashLength > 12) {
    throw new Error('Hash length must be between 1 and 12');
  }

  let hash = 5381;
  for (let i = 0; i < input.length; i++) {
    hash = (hash * 33) ^ input.charCodeAt(i);
  }
  return Math.abs(hash)
    .toString(36)
    .padStart(hashLength, '0')
    .substring(0, hashLength);
}

export interface HttpErrorResponse extends Response {
  data: unknown;
  error: GooglerpcStatus;
}

export const handleSelectValueChange = (onChange: (value: string) => void) => {
  // WORKAROUND FOR: https://github.com/radix-ui/primitives/issues/3135
  return (value: string) => {
    if (value !== '') {
      onChange(value);
    }
  };
};

/**
 * Formats location from structured object to display string
 * @param location - Location object with city, country, latitude, longitude
 * @returns Formatted string like "City, Country" or "Unknown" if empty
 * Note: This function also exists in the Admin UI sessions/index.ts file.
 * If you make any changes here, please update the Admin UI sessions/index.ts file as well.
 */
export const formatLocation = (location?: {
  city?: string;
  country?: string;
  latitude?: string;
  longitude?: string;
}): string => {
  if (!location) return 'Unknown';

  const city = location.city?.trim() || '';
  const country = location.country?.trim() || '';

  if (city && country) {
    return `${city}, ${country}`;
  }
  if (city) {
    return city;
  }
  if (country) {
    return country;
  }

  return 'Unknown';
};
