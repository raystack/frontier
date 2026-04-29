'use client';

import { useCallback, useEffect, useMemo } from 'react';
import qs from 'query-string';
import { Flex, Dialog, EmptyState, toastManager } from '@raystack/apsara-v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  CreateCheckoutRequestSchema,
  ListInvoicesRequestSchema,
  FrontierServiceQueries
} from '@raystack/proton/frontier';
import {
  useMutation,
  useQuery as useConnectQuery
} from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '../../contexts/FrontierContext';
import { useBillingPermission } from '../../hooks/useBillingPermission';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { PaymentMethodCard } from './components/payment-method-card';
import { BillingDetailsCard } from './components/billing-details-card';
import { PaymentIssue } from './components/payment-issue';
import { UpcomingPlanChangeBanner } from './components/upcoming-plan-change-banner';
import { UpcomingBillingCycle } from './components/upcoming-billing-cycle';
import { Invoices } from './components/invoices';
import {
  ConfirmCycleSwitchDialog,
  type ConfirmCycleSwitchPayload
} from './components/confirm-cycle-switch-dialog';

const cycleSwitchDialogHandle =
  Dialog.createHandle<ConfirmCycleSwitchPayload>();

export interface BillingViewProps {
  onNavigateToPlans?: () => void;
}

export function BillingView({ onNavigateToPlans }: BillingViewProps) {
  const {
    billingAccount,
    isBillingAccountLoading,
    config,
    activeSubscription,
    isActiveSubscriptionLoading,
    paymentMethod,
    organizationKyc,
    isOrganizationKycLoading,
    activeOrganization
  } = useFrontier();

  const { isAllowed, canSeeBilling, isFetching } = useBillingPermission();

  const isPermissionsLoading = !activeOrganization?.id || isFetching;
  const hasNoAccess = !canSeeBilling && !isPermissionsLoading;

  const {
    data: invoicesData,
    isLoading: isInvoicesLoading,
    error: invoicesError
  } = useConnectQuery(
    FrontierServiceQueries.listInvoices,
    create(ListInvoicesRequestSchema, {
      orgId: activeOrganization?.id || '',
      nonzeroAmountOnly: true
    }),
    {
      enabled: !!activeOrganization?.id && canSeeBilling
    }
  );

  const invoices = useMemo(() => invoicesData?.invoices || [], [invoicesData]);

  useEffect(() => {
    if (invoicesError) {
      toastManager.add({
        title: 'Failed to load invoices',
        description: invoicesError?.message,
        type: 'error'
      });
    }
  }, [invoicesError]);

  const isLoading = !activeOrganization?.id ||
    isBillingAccountLoading ||
    isActiveSubscriptionLoading ||
    isInvoicesLoading ||
    isFetching ||
    isOrganizationKycLoading;

  const isOrganizationKycCompleted = organizationKyc?.status === true;

  const billingSupportEmail = config.billing?.supportEmail;

  const description = billingSupportEmail
    ? `Oversee your billing and invoices. For more details, contact ${billingSupportEmail}`
    : 'Oversee your billing and invoices.';

  function handleCycleSwitchClick(planId: string) {
    cycleSwitchDialogHandle.openWithPayload({ planId });
  }

  const { mutateAsync: createCheckoutMutation, isPending: isCheckoutPending } =
    useMutation(FrontierServiceQueries.createCheckout, {
      onError: (err: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: err?.message,
          type: 'error'
        });
      }
    });

  const handleBillingDetailsUpdateClick = useCallback(async () => {
    const orgId = activeOrganization?.id || '';
    if (!orgId) return;

    try {
      const query = qs.stringify(
        {
          details: btoa(
            qs.stringify({
              organization_id: orgId,
              type: 'billing'
            })
          ),
          checkout_id: '{{.CheckoutID}}'
        },
        { encode: false }
      );
      const cancel_url = `${config?.billing?.cancelUrl}?${query}`;
      const success_url = `${config?.billing?.successUrl}?${query}`;

      const resp = await createCheckoutMutation(
        create(CreateCheckoutRequestSchema, {
          orgId,
          cancelUrl: cancel_url,
          successUrl: success_url,
          setupBody: {
            paymentMethod: false,
            customerPortal: true
          }
        })
      );
      const checkoutUrl = resp?.checkoutSession?.checkoutUrl;
      if (checkoutUrl) {
        window.location.href = checkoutUrl;
      }
    } catch (err) {
      console.error(err);
    }
  }, [
    activeOrganization?.id,
    createCheckoutMutation,
    config?.billing?.cancelUrl,
    config?.billing?.successUrl
  ]);

  return (
    <ViewContainer>
      <ViewHeader title="Billing" description={description} />

      {hasNoAccess ? (
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Restricted Access"
          subHeading="Admin access required, please reach out to your admin to view billing."
        />
      ) : (
        <>
          <Flex direction="column" gap={7}>
            <PaymentIssue
              isLoading={isLoading}
              subscription={activeSubscription}
              invoices={invoices}
            />

            <UpcomingPlanChangeBanner
              isLoading={isLoading}
              subscription={activeSubscription}
              isAllowed={isAllowed}
            />

            <Flex gap={7}>
              <PaymentMethodCard
                paymentMethod={paymentMethod}
                isLoading={isLoading}
                isAllowed={isAllowed}
              />
              <BillingDetailsCard
                billingAccount={billingAccount}
                isLoading={isLoading}
                isAllowed={isAllowed}
                disabled={isOrganizationKycCompleted}
                isActionLoading={isCheckoutPending}
                onUpdateClick={handleBillingDetailsUpdateClick}
              />
            </Flex>

            <UpcomingBillingCycle
              isAllowed={isAllowed}
              isPermissionLoading={isPermissionsLoading}
              onCycleSwitchClick={handleCycleSwitchClick}
              onNavigateToPlans={onNavigateToPlans}
            />

            <Invoices />
          </Flex>

          <ConfirmCycleSwitchDialog handle={cycleSwitchDialogHandle} />
        </>
      )}
    </ViewContainer>
  );
}
