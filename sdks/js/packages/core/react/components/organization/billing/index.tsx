import {
  Button,
  Skeleton,
  Text,
  Flex,
  toast,
  Tooltip,
  Link
} from '@raystack/apsara';
import { Outlet } from '@tanstack/react-router';
import { PageHeader } from '~/react/components/common/page-header';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback, useEffect, useMemo } from 'react';
import {
  BillingAccount,
  ListInvoicesRequestSchema,
  FrontierServiceQueries,
  CreateCheckoutRequestSchema
} from '@raystack/proton/frontier';
import { useQuery as useConnectQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '~hooks';
import Invoices from './invoices';
import qs from 'query-string';
import sharedStyles from '../styles.module.css';
import billingStyles from './billing.module.css';

import { UpcomingBillingCycle } from './upcoming-billing-cycle';
import { PaymentIssue } from './payment-issue';
import { UpcomingPlanChangeBanner } from '../../common/upcoming-plan-change-banner';
import { PaymentMethod } from './payment-method';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';

interface BillingHeaderProps {
  billingSupportEmail?: string;
  isLoading?: boolean;
}

const BillingHeader = ({
  billingSupportEmail,
  isLoading
}: BillingHeaderProps) => {
  if (isLoading) {
    return (
      <Flex direction="column" gap={2} className={billingStyles.flex1}>
        <Skeleton />
        <Skeleton />
      </Flex>
    );
  }

  return (
    <PageHeader
      title="Billing"
      description={
        billingSupportEmail ? (
          <>
            Oversee your billing and invoices.{' '}
            For more details, contact{' '}
            <Link
              size="regular"
              href={`mailto:${billingSupportEmail}`}
              data-test-id="frontier-sdk-billing-email-link"
              external
              style={{ textDecoration: 'none' }}
            >
              {billingSupportEmail}
            </Link>
          </>
        ) : (
          'Oversee your billing and invoices.'
        )
      }
    />
  );
};


interface BillingDetailsProps {
  billingAccount?: BillingAccount;
  onAddDetailsClick?: () => void;
  isLoading: boolean;
  isAllowed: boolean;
  disabled?: boolean;
}

const BillingDetails = ({
  billingAccount,
  onAddDetailsClick = () => {},
  isLoading,
  isAllowed,
  disabled = false
}: BillingDetailsProps) => {
  const btnText =
    billingAccount?.email || billingAccount?.name ? 'Update' : 'Add details';
  const isButtonDisabled = isLoading || disabled;
  return (
    <div className={billingStyles.detailsBox}>
      <Flex align="center" justify="between" width="full">
        <Text className={billingStyles.detailsBoxHeading}>Billing Details</Text>
        {isAllowed ? (
          <Tooltip
            message="Contact support to update your billing address."
            side="bottom-right"
            disabled={!isButtonDisabled}
          >
            <Button
              data-test-id="frontier-sdk-billing-details-update-button"
              variant="outline"
              color="neutral"
              size="small"
              onClick={onAddDetailsClick}
              disabled={isButtonDisabled}
            >
              {btnText}
            </Button>
          </Tooltip>
        ) : null}
      </Flex>
      <Flex direction="column" gap={2}>
        <Text className={billingStyles.detailsBoxRowLabel}>Name</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : billingAccount?.name || 'N/A'}
        </Text>
      </Flex>
      <Flex direction="column" gap={2}>
        <Text className={billingStyles.detailsBoxRowLabel}>Email</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton count={2} /> : billingAccount?.email || 'N/A'}
        </Text>
      </Flex>
    </div>
  );
};

export default function Billing() {
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
      toast.error('Failed to load invoices', {
        description: invoicesError?.message
      });
    }
  }, [invoicesError]);

  const { mutateAsync: createCheckoutMutation } = useMutation(
    FrontierServiceQueries.createCheckout,
    {
      onError: (err: Error) => {
        console.error(err);
        toast.error('Something went wrong', {
          description: err?.message
        });
      }
    }
  );

  const onAddDetailsClick = useCallback(async () => {
    const orgId = activeOrganization?.id || '';
    if (!orgId) return;

    try {
      const query = qs.stringify(
        {
          details: btoa(
            qs.stringify({
              organization_id: activeOrganization?.id || '',
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
          orgId: activeOrganization?.id || '',
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
      toast.error('Something went wrong');
    }
  }, [
    activeOrganization?.id,
    createCheckoutMutation,
    config?.billing?.cancelUrl,
    config?.billing?.successUrl
  ]);

  const isLoading =
    isBillingAccountLoading ||
    isActiveSubscriptionLoading ||
    isInvoicesLoading ||
    isFetching ||
    isOrganizationKycLoading;

  const isOrganizationKycCompleted = organizationKyc?.status === true;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <BillingHeader
            isLoading={isLoading}
            billingSupportEmail={config.billing?.supportEmail}
          />
        </Flex>
        <Flex direction="column" gap={9}>
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
            <PaymentMethod
              paymentMethod={paymentMethod}
              isLoading={isLoading}
              isAllowed={isAllowed}
            />
            <BillingDetails
              billingAccount={billingAccount}
              onAddDetailsClick={onAddDetailsClick}
              isLoading={isLoading}
              isAllowed={isAllowed}
              disabled={isOrganizationKycCompleted}
            />
          </Flex>
          <UpcomingBillingCycle
            isAllowed={isAllowed}
            isPermissionLoading={isFetching}
          />
          <Invoices invoices={invoices} isLoading={isLoading} />
          </Flex>
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
