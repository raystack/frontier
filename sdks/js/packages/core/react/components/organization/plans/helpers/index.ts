import { V1Beta1Feature, V1Beta1Plan } from '~/src';
import {
  IntervalKeys,
  IntervalPricing,
  PlanIntervalPricing
} from '~/src/types';
import slugify from 'slugify';

function makePlanSlug(plan: V1Beta1Plan): string {
  const productIds = plan?.products
    ?.map(p => p.id)
    .sort()
    .join('-');
  const titleSlug = slugify(plan.title || '', { lower: true });
  return `${titleSlug}-${productIds}`;
}

export function groupPlansPricingByInterval(plans: V1Beta1Plan[]) {
  const plansMap: Record<string, PlanIntervalPricing> = {};
  plans.forEach(plan => {
    const slug = makePlanSlug(plan);
    plansMap[slug] = plansMap[slug] || {
      slug: slug,
      title: plan.title,
      description: plan?.description,
      intervals: {},
      features: {}
    };
    const planInterval = (plan?.interval || '') as IntervalKeys;
    const productPrices =
      plan?.products?.reduce((acc, product) => {
        product.prices?.forEach(price => {
          if (price.interval === planInterval) {
            acc.amount = Number(acc.amount || 0) + Number(price.amount);
            acc.currency = price.currency || '';
            acc.behavior = '';
          }
        });
        return acc;
      }, {} as IntervalPricing) || ({} as IntervalPricing);

    plan?.products?.forEach(product => {
      product.features?.forEach(feature => {
        plansMap[slug].features[feature?.id || ''] = feature;
      });
    }, {} as IntervalPricing) || ({} as IntervalPricing);
    plansMap[slug].intervals[planInterval] = {
      planId: plan?.id || '',
      planName: plan?.name || '',
      ...productPrices
    };
  });

  return Object.values(plansMap);
}

export function getAllPlansFeatuesMap(plans: V1Beta1Plan[]) {
  const featureMap: Record<string, V1Beta1Feature> = {};
  plans.forEach(plan => {
    plan?.products?.forEach(product => {
      product?.features?.forEach(feature => {
        featureMap[feature?.id || ''] = feature;
      });
    });
  });
  return featureMap;
}
