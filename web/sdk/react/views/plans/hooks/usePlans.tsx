import { useCallback, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import qs from 'query-string';
import { toast } from '@raystack/apsara';
import { SubscriptionPhase } from '~/src';
import { SUBSCRIPTION_STATES } from '~/react/utils/constants';
import { PlanMetadata } from '~/src/types';
import { NIL as NIL_UUID } from 'uuid';
import { useMutation, FrontierServiceQueries } from '~hooks';
import { handleConnectError } from '~/utils/error';
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
    config,
    activeSubscription,
    subscriptions,
    fetchActiveSubscription,
    allPlans,
    isAllPlansLoading
  } = useFrontier();

  // Setup mutations
  const { mutateAsync: createCheckoutMutation, isPending: isCheckoutPending } =
    useMutation(FrontierServiceQueries.createCheckout);
  const {
    mutateAsync: changeSubscriptionMutation,
    isPending: isChangePlanPending
  } = useMutation(FrontierServiceQueries.changeSubscription);
  const {
    mutateAsync: cancelSubscriptionMutation,
    isPending: isCancelPending
  } = useMutation(FrontierServiceQueries.cancelSubscription);

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
        if (activeOrganization?.id) {
          const query = qs.stringify(
            {
              details: btoa(
                qs.stringify({
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
        handleConnectError(err, {
          PermissionDenied: () =>
            toast.error("You don't have permission to perform this action"),
          InvalidArgument: e =>
            toast.error('Checkout failed', { description: e.message }),
          NotFound: e => toast.error('Not found', { description: e.message })
        });
      }
    },
    [
      activeOrganization?.id,
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
      if (activeSubscription?.id) {
        try {
          const resp = await changeSubscriptionMutation(
            create(ChangeSubscriptionRequestSchema, {
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
        } catch (error) {
          handleConnectError(error, {
            PermissionDenied: () =>
              toast.error("You don't have permission to perform this action"),
            InvalidArgument: err =>
              toast.error('Failed to change plan', {
                description: err.message
              }),
            NotFound: err =>
              toast.error('Not found', { description: err.message })
          });
        }
      }
    },
    [activeSubscription?.id, changeSubscriptionMutation]
  );

  const verifyPlanChange = useCallback(
    async ({ planId, onSuccess = () => {} }: verifyPlanChangeOptions) => {
      const activeSub = await fetchActiveSubscription();
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
    [fetchActiveSubscription]
  );

  const verifySubscriptionCancel = useCallback(
    async ({ onSuccess = () => {} }: verifyCancelSubscriptionOptions) => {
      const activeSub = await fetchActiveSubscription();
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
    [fetchActiveSubscription]
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
      if (activeSubscription?.id) {
        try {
          const resp = await cancelSubscriptionMutation(
            create(CancelSubscriptionRequestSchema, {
              id: activeSubscription?.id,
              immediate: false
            })
          );
          if (resp) {
            onSuccess();
          }
        } catch (error) {
          handleConnectError(error, {
            PermissionDenied: () =>
              toast.error("You don't have permission to perform this action"),
            NotFound: err =>
              toast.error('Not found', { description: err.message }),
            Default: err =>
              toast.error('Failed to cancel subscription', {
                description: err.message
              })
          });
        }
      }
    },
    [activeSubscription?.id, cancelSubscriptionMutation]
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
