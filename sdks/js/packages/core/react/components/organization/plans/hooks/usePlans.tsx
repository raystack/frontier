import { useCallback, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import qs from 'query-string';
import { toast } from '@raystack/apsara/v1';
import { SubscriptionPhase, V1Beta1CheckoutSession, V1Beta1Plan } from '~/src';
import { SUBSCRIPTION_STATES } from '~/react/utils/constants';
import { PlanMetadata } from '~/src/types';
import { NIL as NIL_UUID } from 'uuid';

interface checkoutPlanOptions {
  isTrial: boolean;
  planId: string;
  onSuccess: (data: V1Beta1CheckoutSession) => void;
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
  const [isLoading, setIsLoading] = useState(false);
  const [hasAlreadyTrialed, setHasAlreadyTrialed] = useState(false);
  const {
    client,
    activeOrganization,
    billingAccount,
    config,
    activeSubscription,
    subscriptions,
    fetchActiveSubsciption,
    allPlans,
    isAllPlansLoading
  } = useFrontier();

  const planMap = allPlans.reduce((acc, p) => {
    if (p.id) acc[p.id] = p;
    return acc;
  }, {} as Record<string, V1Beta1Plan>);

  const isCurrentlyTrialing = subscriptions?.some(
    sub => sub.state === SUBSCRIPTION_STATES.TRIALING
  );

  const checkoutPlan = useCallback(
    async ({ planId, onSuccess, isTrial }: checkoutPlanOptions) => {
      setIsLoading(true);
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

          const resp = await client?.frontierServiceCreateCheckout(
            activeOrganization?.id,
            billingAccount?.id,
            {
              cancel_url: cancel_url,
              success_url: success_url,
              subscription_body: {
                plan: planId,
                skip_trial: !isTrial,
                cancel_after_trial: isTrial && cancelAfterTrial
              }
            }
          );
          if (resp?.data?.checkout_session?.checkout_url) {
            onSuccess(resp?.data?.checkout_session);
          }
        }
      } catch (err: any) {
        console.error(err);
        toast.error('Something went wrong', {
          description: err?.message
        });
      } finally {
        setIsLoading(false);
      }
    },
    [
      activeOrganization?.id,
      billingAccount?.id,
      config?.billing?.cancelUrl,
      config?.billing?.successUrl,
      config?.billing?.cancelAfterTrial,
      client
    ]
  );

  const checkBasePlan = (planId: string) => {
    return planId === NIL_UUID;
  };

  const changePlan = useCallback(
    async ({ planId, onSuccess, immediate = false }: changePlanOptions) => {
      setIsLoading(true);
      try {
        if (
          activeOrganization?.id &&
          billingAccount?.id &&
          activeSubscription?.id
        ) {
          const resp = await client?.frontierServiceChangeSubscription(
            activeOrganization?.id,
            billingAccount?.id,
            activeSubscription?.id,
            {
              plan_change: {
                plan: planId,
                immediate: immediate
              }
            }
          );
          if (resp?.data?.phase) {
            onSuccess();
          }
        }
      } catch (err: any) {
        console.error(err);
        toast.error('Something went wrong', {
          description: err?.message
        });
      } finally {
        setIsLoading(false);
      }
    },
    [activeOrganization?.id, activeSubscription?.id, billingAccount?.id, client]
  );

  const verifyPlanChange = useCallback(
    async ({ planId, onSuccess = () => {} }: verifyPlanChangeOptions) => {
      const activeSub = await fetchActiveSubsciption();
      if (activeSub) {
        const planPhase = activeSub.phases?.find(
          phase => phase?.plan_id === planId && phase.reason === 'change'
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
          phase => phase?.plan_id === '' && phase.reason === 'cancel'
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
      .map(t => (t.plan_id ? planMap[t.plan_id] : null))
      .filter((plan): plan is V1Beta1Plan => !!plan);
  }, [planMap, subscriptions]);

  const getTrialedPlanMaxWeightage = (plans: V1Beta1Plan[]) => {
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
      setIsLoading(true);
      try {
        if (
          activeOrganization?.id &&
          billingAccount?.id &&
          activeSubscription?.id
        ) {
          const resp = await client?.frontierServiceCancelSubscription(
            activeOrganization?.id,
            billingAccount?.id,
            activeSubscription?.id,
            { immediate: false }
          );
          if (resp?.data) {
            onSuccess();
          }
        }
      } catch (err: any) {
        console.error(err);
        toast.error('Something went wrong', {
          description: err?.message
        });
      } finally {
        setIsLoading(false);
      }
    },
    [activeOrganization?.id, billingAccount?.id, activeSubscription?.id, client]
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
