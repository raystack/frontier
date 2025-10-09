import {
  IntervalKeys,
  IntervalPricing,
  PlanIntervalPricing
} from '~/src/types';
import { getPlanPrice, makePlanSlug } from '~/react/utils';
import { Plan } from '@raystack/proton/frontier';

export function groupPlansPricingByInterval(plans: Plan[]) {
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
    const productPrices = getPlanPrice(plan);

    const planMetadata = (plan?.metadata as Record<string, string>) || {};
    plansMap[slug].intervals[planInterval] = {
      planId: plan?.id || '',
      planName: plan?.name || '',
      interval: planInterval,
      weightage: planMetadata?.weightage ? Number(planMetadata?.weightage) : 0,
      productNames: [],
      trial_days: plan?.trialDays || '',
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
