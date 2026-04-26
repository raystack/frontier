'use client';

import { useEffect, useMemo } from 'react';
import { Flex, Dialog, toastManager } from '@raystack/apsara-v1';
import {
  ListInvoicesRequestSchema,
  FrontierServiceQueries
} from '@raystack/proton/frontier';
import { useQuery as useConnectQuery } from '@connectrpc/connect-query';
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
import { BillingDetailsDialog } from './components/billing-details-dialog';

const cycleSwitchDialogHandle =
  Dialog.createHandle<ConfirmCycleSwitchPayload>();
const billingDetailsDialogHandle = Dialog.createHandle();

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

  const { isAllowed, isFetching } = useBillingPermission();

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
      enabled: !!activeOrganization?.id
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

  const isLoading =
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

  function handleBillingDetailsUpdateClick() {
    billingDetailsDialogHandle.open(null);
  }

  return (
    <ViewContainer>
      <ViewHeader title="Billing" description={description} />

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
            onUpdateClick={handleBillingDetailsUpdateClick}
          />
        </Flex>

        <UpcomingBillingCycle
          isAllowed={isAllowed}
          isPermissionLoading={isFetching}
          onCycleSwitchClick={handleCycleSwitchClick}
          onNavigateToPlans={onNavigateToPlans}
        />

        <Invoices />
      </Flex>

      <ConfirmCycleSwitchDialog handle={cycleSwitchDialogHandle} />
      <BillingDetailsDialog handle={billingDetailsDialogHandle} />
    </ViewContainer>
  );
}
