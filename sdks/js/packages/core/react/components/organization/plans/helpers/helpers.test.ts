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
        weightage: 0,
        intervals: {
          year: {
            planId: 'plan-1',
            planName: 'starter_plan_plan-1',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'year',
            weightage: 0
          },
          month: {
            planId: 'plan-2',
            planName: 'starter_plan_plan-2',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'month',
            weightage: 0
          }
        },
        features: {}
      },
      {
        slug: 'starter-plan-3-product-1-product-3',
        title: 'Starter Plan 3',
        description: 'Starter Plan 3',
        weightage: 0,
        intervals: {
          month: {
            amount: 500,
            behavior: '',
            currency: 'INR',
            planId: 'plan-3',
            planName: 'starter_plan_plan-3',
            interval: 'month',
            weightage: 0
          }
        },
        features: {}
      }
    ]);
  });

  test('should add plans weightage', () => {
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
        ],
        metadata: {
          weightage: '1'
        }
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
        ],
        metadata: {
          weightage: '2'
        }
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
        ],
        metadata: {
          weightage: '5'
        }
      }
    ];

    const result = groupPlansPricingByInterval(plans);
    expect(result).toEqual([
      {
        slug: 'starter-plan-product-1-product-2',
        title: 'Starter Plan',
        description: 'Starter Plan',
        weightage: 3,
        intervals: {
          year: {
            planId: 'plan-1',
            planName: 'starter_plan_plan-1',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'year',
            weightage: 1
          },
          month: {
            planId: 'plan-2',
            planName: 'starter_plan_plan-2',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'month',
            weightage: 2
          }
        },
        features: {}
      },
      {
        slug: 'starter-plan-3-product-1-product-3',
        title: 'Starter Plan 3',
        description: 'Starter Plan 3',
        weightage: 5,
        intervals: {
          month: {
            amount: 500,
            behavior: '',
            currency: 'INR',
            planId: 'plan-3',
            planName: 'starter_plan_plan-3',
            interval: 'month',
            weightage: 5
          }
        },
        features: {}
      }
    ]);
  });

  test('should group plans based on `plan_group_id`', () => {
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
        ],
        metadata: {
          weightage: '1',
          plan_group_id: 'group-1'
        }
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
        ],
        metadata: {
          weightage: '2',
          plan_group_id: 'group-1'
        }
      },
      {
        id: 'plan-3',
        name: 'starter_plan_plan-3',
        title: 'Starter Plan 3',
        description: 'Starter Plan 3',
        interval: 'week',
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
              },
              {
                amount: '500',
                interval: 'week',
                currency: 'INR'
              }
            ]
          }
        ],
        metadata: {
          weightage: '5',
          plan_group_id: 'group-1'
        }
      }
    ];

    const result = groupPlansPricingByInterval(plans);
    expect(result).toEqual([
      {
        slug: 'group-1',
        title: 'Starter Plan',
        description: 'Starter Plan',
        weightage: 8,
        intervals: {
          year: {
            planId: 'plan-1',
            planName: 'starter_plan_plan-1',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'year',
            weightage: 1
          },
          month: {
            planId: 'plan-2',
            planName: 'starter_plan_plan-2',
            amount: 0,
            behavior: '',
            currency: 'INR',
            interval: 'month',
            weightage: 2
          },
          week: {
            amount: 500,
            behavior: '',
            currency: 'INR',
            planId: 'plan-3',
            planName: 'starter_plan_plan-3',
            interval: 'week',
            weightage: 5
          }
        },
        features: {}
      }
    ]);
  });
});
