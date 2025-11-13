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
import { useCallback, useEffect, useState } from 'react';
import {
  V1Beta1CheckoutSetupBody,
  V1Beta1Invoice
} from '~/src';
import { BillingAccount } from '@raystack/proton/frontier';
// import { converBillingAddressToString } from '~/react/utils';
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
  // const addressStr = converBillingAddressToString(billingAccount?.address);
  const btnText =
    billingAccount?.email || billingAccount?.name ? 'Update' : 'Add details';
  const isButtonDisabled = isLoading || disabled;
  return (
    <div className={billingStyles.detailsBox}>
      <Flex align="center" justify="between" style={{ width: '100%' }}>
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
    client,
    config,
    activeSubscription,
    isActiveSubscriptionLoading,
    paymentMethod,
    organizationKyc,
    isOrganizationKycLoading
  } = useFrontier();

  const [invoices, setInvoices] = useState<V1Beta1Invoice[]>([]);
  const [isInvoicesLoading, setIsInvoicesLoading] = useState(false);
  const { isAllowed, isFetching } = useBillingPermission();

  const fetchInvoices = useCallback(
    async (organizationId: string, billingId: string) => {
      setIsInvoicesLoading(true);
      try {
        const resp = await client?.frontierServiceListInvoices(
          organizationId,
          billingId,
          { nonzero_amount_only: true }
        );
        const newInvoices = resp?.data?.invoices || [];
        setInvoices(newInvoices);
      } catch (err) {
        console.error(err);
      } finally {
        setIsInvoicesLoading(false);
      }
    },
    [client]
  );

  useEffect(() => {
    if (billingAccount?.id && billingAccount?.orgId) {
      fetchInvoices(billingAccount?.orgId, billingAccount?.id);
    }
  }, [billingAccount?.id, billingAccount?.orgId, client, fetchInvoices]);

  const onAddDetailsClick = useCallback(async () => {
    const orgId = billingAccount?.orgId || '';
    const billingAccountId = billingAccount?.id || '';
    if (billingAccountId && orgId) {
      try {
        const query = qs.stringify(
          {
            details: btoa(
              qs.stringify({
                billing_id: billingAccount?.id,
                organization_id: billingAccount?.orgId,
                type: 'billing'
              })
            ),
            checkout_id: '{{.CheckoutID}}'
          },
          { encode: false }
        );
        const cancel_url = `${config?.billing?.cancelUrl}?${query}`;
        const success_url = `${config?.billing?.successUrl}?${query}`;

        const setup_body: V1Beta1CheckoutSetupBody = {
          customer_portal: true
        };

        const resp = await client?.frontierServiceCreateCheckout(
          billingAccount?.orgId || '',
          billingAccount?.id || '',
          {
            cancel_url,
            success_url,
            setup_body
          }
        );
        const checkout_url = resp?.data?.checkout_session?.checkout_url;
        if (checkout_url) {
          window.location.href = checkout_url;
        }
      } catch (err) {
        console.error(err);
        toast.error('Something went wrong');
      }
    }
  }, [
    billingAccount?.id,
    billingAccount?.orgId,
    client,
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
