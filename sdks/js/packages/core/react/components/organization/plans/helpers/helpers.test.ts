import { V1Beta1Plan } from '~/src';
import { groupPlansPricingByInterval } from './index';

describe('Plans:helpers:groupPlansPricingByInterval', () => {
  test('should return empty array for no plans', () => {
    const result = groupPlansPricingByInterval([]);
    expect(result).toEqual([]);
  });
  test('should merge plan based on name and productIds', () => {
    const plans: V1Beta1Plan[] = [
      {
        id: 'plan-1',
        name: 'starter_plan_plan-1',
        title: 'Starter Plan',
        description: 'Starter Plan',
        interval: 'year',
        products: [
          {
            id: 'product-1',
            prices: [
              {
                amount: '0',
                interval: 'year',
                currency: 'INR'
              },
              {
                amount: '0',
                interval: 'month',
                currency: 'INR'
              }
            ]
          },
          {
            id: 'product-2',
            prices: [
              {
                amount: '0',
                interval: 'year',
                currency: 'INR'
              },
              {
                amount: '0',
                interval: 'month',
                currency: 'INR'
              }
            ]
          }
        ]
      },
      {
        id: 'plan-2',
        name: 'starter_plan_plan-2',
        title: 'Starter Plan',
        description: 'Starter Plan',
        interval: 'month',
        products: [
          {
            id: 'product-1',
            prices: [
              {
                amount: '0',
                interval: 'year',
                currency: 'INR'
              },
              {
                amount: '0',
                interval: 'month',
                currency: 'INR'
              }
            ]
          },
          {
            id: 'product-2',
            prices: [
              {
                amount: '0',
                interval: 'year',
                currency: 'INR'
              },
              {
                amount: '0',
                interval: 'month',
                currency: 'INR'
              }
            ]
          }
        ]
      },
      {
        id: 'plan-3',
        name: 'starter_plan_plan-3',
        title: 'Starter Plan 3',
        description: 'Starter Plan 3',
        interval: 'month',
        products: [
          {
            id: 'product-1',
            prices: [
              {
                amount: '0',
                interval: 'year',
                currency: 'INR'
              },
              {
                amount: '0',
                interval: 'month',
                currency: 'INR'
              }
            ]
          },
          {
            id: 'product-3',
            prices: [
              {
                amount: '100',
                interval: 'year',
                currency: 'INR'
              },
              {
                amount: '500',
                interval: 'month',
                currency: 'INR'
              }
            ]
          }
        ]
      }
    ];

    const result = groupPlansPricingByInterval(plans);
    expect(result).toEqual([
      {
        slug: 'starter-plan-product-1-product-2',
        title: 'Starter Plan',
        description: 'Starter Plan',
        intervals: {
          year: {
            planId: 'plan-1',
            planName: 'starter_plan_plan-1',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'year'
          },
          month: {
            planId: 'plan-2',
            planName: 'starter_plan_plan-2',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'month'
          }
        },
        features: {}
      },
      {
        slug: 'starter-plan-3-product-1-product-3',
        title: 'Starter Plan 3',
        description: 'Starter Plan 3',
        intervals: {
          month: {
            amount: 500,
            behavior: '',
            currency: 'INR',
            planId: 'plan-3',
            planName: 'starter_plan_plan-3',
            interval: 'month'
          }
        },
        features: {}
      }
    ]);
  });
});
