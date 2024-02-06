import { useCallback, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import qs from 'query-string';
import { toast } from 'sonner';
import { V1Beta1CheckoutSession } from '~/src';

interface checkoutPlanOptions {
  planId: string;
  onSuccess: (data: V1Beta1CheckoutSession) => void;
}

interface changePlanOptions {
  planId: string;
  immediate?: boolean;
  onSuccess: () => void;
}

export const usePlans = () => {
  const [isLoading, setIsLoading] = useState(false);
  const {
    client,
    activeOrganization,
    billingAccount,
    config,
    activeSubscription
  } = useFrontier();

  const checkoutPlan = useCallback(
    async ({ planId, onSuccess }: checkoutPlanOptions) => {
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
                plan: planId
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
              plan: planId,
              immediate: immediate
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

  return { checkoutPlan, isLoading, changePlan };
};
