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
    const metaData = (plan?.metadata as Record<string, string>) || {};
    const slug = metaData?.plan_group_id || makePlanSlug(plan);
    plansMap[slug] = plansMap[slug] || {
      slug: slug,
      title: plan.title,
      description: plan?.description,
      weightage: 0,
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

    const planMetadata = (plan?.metadata as Record<string, string>) || {};
    plansMap[slug].intervals[planInterval] = {
      planId: plan?.id || '',
      planName: plan?.name || '',
      interval: planInterval,
      weightage: planMetadata?.weightage ? Number(planMetadata?.weightage) : 0,
      productNames: [],
      features: {},
      ...productPrices
    };

    plan?.products?.forEach(product => {
      plansMap[slug].intervals[planInterval].productNames = [
        ...plansMap[slug].intervals[planInterval].productNames,
        product.name || ''
      ];
      product.features?.forEach(feature => {
        plansMap[slug].intervals[planInterval].features[feature?.title || ''] =
          feature;
      });
    }, {} as IntervalPricing) || ({} as IntervalPricing);

    plansMap[slug].weightage = Object.values(plansMap[slug].intervals).reduce(
      (acc, data) => acc + data.weightage,
      0
    );
  });

  return Object.values(plansMap);
}
