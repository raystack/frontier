import { useCallback, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import qs from 'query-string';
import { toast } from '@raystack/apsara';
import { SubscriptionPhase } from '~/src';
import { SUBSCRIPTION_STATES } from '~/react/utils/constants';
import { PlanMetadata } from '~/src/types';
import { NIL as NIL_UUID } from 'uuid';
import { useMutation, FrontierServiceQueries } from '~hooks';
import { create } from '@bufbuild/protobuf';
import {
  type Plan,
  type CheckoutSession,
  CreateCheckoutRequestSchema,
  ChangeSubscriptionRequestSchema,
  CancelSubscriptionRequestSchema
} from '@raystack/proton/frontier';

interface checkoutPlanOptions {
  isTrial: boolean;
  planId: string;
  onSuccess: (data: CheckoutSession) => void;
}

interface changePlanOptions {
  planId: string;
  immediate?: boolean;
  onSuccess: () => void;
}

interface cancelSubscriptionOptions {
  onSuccess: () => void;
}

interface verifyPlanChangeOptions {
  planId: string;
  onSuccess?: (planPhase: SubscriptionPhase) => void;
}

interface verifyCancelSubscriptionOptions {
  onSuccess?: (planPhase: SubscriptionPhase) => void;
}

export const usePlans = () => {
  const [hasAlreadyTrialed, setHasAlreadyTrialed] = useState(false);
  const {
    activeOrganization,
    billingAccount,
    config,
    activeSubscription,
    subscriptions,
    fetchActiveSubsciption,
    allPlans,
    isAllPlansLoading
  } = useFrontier();

  // Setup mutations
  const { mutateAsync: createCheckoutMutation, isPending: isCheckoutPending } =
    useMutation(FrontierServiceQueries.createCheckout, {
      onError: (err: Error) => {
        console.error(err);
        toast.error('Checkout failed', {
          description: err?.message
        });
      }
    });
  const {
    mutateAsync: changeSubscriptionMutation,
    isPending: isChangePlanPending
  } = useMutation(FrontierServiceQueries.changeSubscription, {
    onError: (err: Error) => {
      console.error(err);
      toast.error('Failed to change plan', {
        description: err?.message
      });
    }
  });
  const {
    mutateAsync: cancelSubscriptionMutation,
    isPending: isCancelPending
  } = useMutation(FrontierServiceQueries.cancelSubscription, {
    onError: (err: Error) => {
      console.error(err);
      toast.error('Failed to cancel subscription', {
        description: err?.message
      });
    }
  });

  const isLoading = isCheckoutPending || isChangePlanPending || isCancelPending;

  const planMap = allPlans.reduce((acc, p) => {
    if (p.id) acc[p.id] = p;
    return acc;
  }, {} as Record<string, Plan>);

  const isCurrentlyTrialing = subscriptions?.some(
    sub => sub.state === SUBSCRIPTION_STATES.TRIALING
  );

  const checkoutPlan = useCallback(
    async ({ planId, onSuccess, isTrial }: checkoutPlanOptions) => {
      try {
        if (activeOrganization?.id && billingAccount?.id) {
          const query = qs.stringify(
            {
              details: btoa(
                qs.stringify({
                  billing_id: billingAccount?.id,
                  organization_id: activeOrganization?.id,
                  type: 'plans',
                  isTrial: isTrial
                })
              ),
              checkout_id: '{{.CheckoutID}}'
            },
            { encode: false }
          );
          const cancel_url = `${config?.billing?.cancelUrl}?${query}`;
          const success_url = `${config?.billing?.successUrl}?${query}`;

          let cancelAfterTrial = true;
          if (config?.billing?.cancelAfterTrial !== undefined) {
            cancelAfterTrial = config?.billing?.cancelAfterTrial;
          }

          const resp = await createCheckoutMutation(
            create(CreateCheckoutRequestSchema, {
              orgId: activeOrganization?.id,
              billingId: billingAccount?.id,
              cancelUrl: cancel_url,
              successUrl: success_url,
              subscriptionBody: {
                plan: planId,
                skipTrial: !isTrial,
                cancelAfterTrial: isTrial && cancelAfterTrial
              }
            })
          );
          if (resp?.checkoutSession?.checkoutUrl) {
            onSuccess(resp.checkoutSession);
          }
        }
      } catch (err: unknown) {
        console.error(err);
        toast.error('Failed to prepare checkout', {
          description: (err as Error)?.message
        });
      }
    },
    [
      activeOrganization?.id,
      billingAccount?.id,
      config?.billing?.cancelUrl,
      config?.billing?.successUrl,
      config?.billing?.cancelAfterTrial,
      createCheckoutMutation
    ]
  );

  const checkBasePlan = (planId: string) => {
    return planId === NIL_UUID;
  };

  const changePlan = useCallback(
    async ({ planId, onSuccess, immediate = false }: changePlanOptions) => {
      if (
        activeOrganization?.id &&
        billingAccount?.id &&
        activeSubscription?.id
      ) {
        const resp = await changeSubscriptionMutation(
          create(ChangeSubscriptionRequestSchema, {
            orgId: activeOrganization?.id,
            billingId: billingAccount?.id,
            id: activeSubscription?.id,
            change: {
              case: 'planChange',
              value: {
                plan: planId,
                immediate: immediate
              }
            }
          })
        );
        if (resp?.phase) {
          onSuccess();
        }
      }
    },
    [
      activeOrganization?.id,
      activeSubscription?.id,
      billingAccount?.id,
      changeSubscriptionMutation
    ]
  );

  const verifyPlanChange = useCallback(
    async ({ planId, onSuccess = () => {} }: verifyPlanChangeOptions) => {
      const activeSub = await fetchActiveSubsciption();
      if (activeSub) {
        const planPhase = activeSub.phases?.find(
          phase => phase?.planId === planId && phase.reason === 'change'
        );
        if (planPhase) {
          onSuccess(planPhase);
          return planPhase;
        }
      }
    },
    [fetchActiveSubsciption]
  );

  const verifySubscriptionCancel = useCallback(
    async ({ onSuccess = () => {} }: verifyCancelSubscriptionOptions) => {
      const activeSub = await fetchActiveSubsciption();
      if (activeSub) {
        const planPhase = activeSub.phases?.find(
          phase => phase?.planId === '' && phase.reason === 'cancel'
        );
        if (planPhase) {
          onSuccess(planPhase);
          return planPhase;
        }
      }
    },
    [fetchActiveSubsciption]
  );

  const getSubscribedPlans = useCallback(() => {
    return subscriptions
      .map(t => (t.planId ? planMap[t.planId] : null))
      .filter((plan): plan is Plan => !!plan);
  }, [planMap, subscriptions]);

  const getTrialedPlanMaxWeightage = (plans: Plan[]) => {
    return Math.max(
      ...plans
        .map(plan => {
          const metadata = plan?.metadata as PlanMetadata;
          return metadata?.weightage || 0;
        })
        .filter(w => w > 0)
    );
  };

  const getIneligiblePlansIdsSetForTrial = useCallback(
    (subscribedPlansIdsSet: Set<string>, maxTrialPlanWeightage = 0) => {
      return allPlans.reduce((acc, plan) => {
        const metadata = plan?.metadata as PlanMetadata;
        const weightage = metadata?.weightage || 0;
        const planId = plan?.id || '';
        if (
          (planId && subscribedPlansIdsSet.has(planId)) ||
          (maxTrialPlanWeightage > 0 && weightage < maxTrialPlanWeightage)
        ) {
          acc.add(planId);
        }
        return acc;
      }, new Set<string>());
    },
    [allPlans]
  );

  const checkAlreadyTrialed = useCallback(
    async (planIds: string[]) => {
      const subscribedPlans = getSubscribedPlans();
      const subscribedPlansIdsSet = new Set(
        subscribedPlans
          .map(plan => plan.id)
          .filter((planId): planId is string => !!planId)
      );
      const maxTrialPlanWeightage = getTrialedPlanMaxWeightage(subscribedPlans);
      const ineligiblePlansIdsSetForTrial = getIneligiblePlansIdsSetForTrial(
        subscribedPlansIdsSet,
        maxTrialPlanWeightage
      );
      const value = planIds.some(planId =>
        ineligiblePlansIdsSetForTrial.has(planId)
      );
      setHasAlreadyTrialed(value);
    },
    [getIneligiblePlansIdsSetForTrial, getSubscribedPlans]
  );

  const cancelSubscription = useCallback(
    async ({ onSuccess }: cancelSubscriptionOptions) => {
      if (
        activeOrganization?.id &&
        billingAccount?.id &&
        activeSubscription?.id
      ) {
        const resp = await cancelSubscriptionMutation(
          create(CancelSubscriptionRequestSchema, {
            orgId: activeOrganization?.id,
            billingId: billingAccount?.id,
            id: activeSubscription?.id,
            immediate: false
          })
        );
        if (resp) {
          onSuccess();
        }
      }
    },
    [
      activeOrganization?.id,
      billingAccount?.id,
      activeSubscription?.id,
      cancelSubscriptionMutation
    ]
  );

  return {
    checkoutPlan,
    isLoading,
    changePlan,
    verifyPlanChange,
    verifySubscriptionCancel,
    isTrialCheckLoading: isAllPlansLoading,
    hasAlreadyTrialed,
    isCurrentlyTrialing,
    checkAlreadyTrialed,
    subscriptions,
    cancelSubscription,
    checkBasePlan
  };
};
