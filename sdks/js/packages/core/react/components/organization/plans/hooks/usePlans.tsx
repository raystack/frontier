import { useCallback, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import qs from 'query-string';
import { toast } from 'sonner';
import { SubscriptionPhase, V1Beta1CheckoutSession } from '~/src';
import { SUBSCRIPTION_STATES } from '~/react/utils/constants';

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

interface verifyPlanChangeOptions {
  planId: string;
  onSuccess?: (planPhase: SubscriptionPhase) => void;
}

export const usePlans = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [isTrailCheckLoading, setIsTrailCheckLoading] = useState(false);
  const [hasAlreadyTrailed, setHasAlreadyTrialed] = useState(false);
  const {
    client,
    activeOrganization,
    billingAccount,
    config,
    activeSubscription,
    subscriptions,
    fetchActiveSubsciption
  } = useFrontier();

  const trailSubscription = subscriptions.find(
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
                  type: 'plans'
                })
              ),
              checkout_id: '{{.CheckoutID}}'
            },
            { encode: false }
          );
          const cancel_url = `${config?.billing?.cancelUrl}?${query}`;
          const success_url = `${config?.billing?.successUrl}?${query}`;

          const resp = await client?.frontierServiceCreateCheckout(
            activeOrganization?.id,
            billingAccount?.id,
            {
              cancel_url: cancel_url,
              success_url: success_url,
              subscription_body: {
                plan: planId,
                skip_trial: !isTrial,
                cancel_after_trial: config?.billing?.cancelAfterTrial || false
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
      client
    ]
  );

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
          phase => phase?.plan_id === planId
        );
        if (planPhase) {
          onSuccess(planPhase);
          return planPhase;
        }
      }
    },
    [fetchActiveSubsciption]
  );

  const checkAlreadyTrialed = useCallback(
    async (planIds: string[]) => {
      try {
        setIsTrailCheckLoading(true);
        if (activeOrganization?.id && billingAccount?.id) {
          const resps = await Promise.all(
            planIds.map(planId =>
              client?.frontierServiceHasTrialed(
                activeOrganization?.id || '',
                billingAccount?.id || '',
                planId
              )
            )
          );
          const value = resps.some(resp => resp?.data?.trialed);
          setHasAlreadyTrialed(value);
        }
      } catch (err: any) {
        console.error(err);
      } finally {
        setIsTrailCheckLoading(false);
      }
    },
    [activeOrganization?.id, billingAccount?.id, client]
  );

  return {
    checkoutPlan,
    isLoading,
    changePlan,
    verifyPlanChange,
    isTrailCheckLoading,
    hasAlreadyTrailed,
    checkAlreadyTrialed,
    trailSubscription
  };
};
