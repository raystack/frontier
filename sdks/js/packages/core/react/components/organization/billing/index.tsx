import { Button, Skeleton, Text, Flex, toast, Link, Tooltip } from '@raystack/apsara/v1';
import { Outlet } from '@tanstack/react-router';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback, useEffect, useState } from 'react';
import billingStyles from './billing.module.css';
import {
  V1Beta1BillingAccount,
  V1Beta1CheckoutSetupBody,
  V1Beta1Invoice
} from '~/src';
// import { converBillingAddressToString } from '~/react/utils';
import Invoices from './invoices';
import qs from 'query-string';

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
  return (
    <Flex direction="column" gap={3}>
      {isLoading ? (
        <Skeleton containerClassName={billingStyles.flex1} />
      ) : (
        <Text size="large">Billing</Text>
      )}
      {isLoading ? (
        <Skeleton containerClassName={billingStyles.flex1} />
      ) : (
        <Text size="regular" variant="secondary">
          Oversee your billing and invoices.
          {billingSupportEmail ? (
            <>
              {' '}
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
          ) : null}
        </Text>
      )}
    </Flex>
  );
};

interface BillingDetailsProps {
  billingAccount?: V1Beta1BillingAccount;
  onAddDetailsClick?: () => void;
  isLoading: boolean;
  isAllowed: boolean;
  // hideUpdateBillingDetailsBtn: boolean;
  disabled?: boolean;
}

const BillingDetails = ({
  billingAccount,
  onAddDetailsClick = () => {},
  isLoading,
  isAllowed,
  // hideUpdateBillingDetailsBtn = false,
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
    if (billingAccount?.id && billingAccount?.org_id) {
      fetchInvoices(billingAccount?.org_id, billingAccount?.id);
    }
  }, [billingAccount?.id, billingAccount?.org_id, client, fetchInvoices]);

  const onAddDetailsClick = useCallback(async () => {
    const orgId = billingAccount?.org_id || '';
    const billingAccountId = billingAccount?.id || '';
    if (billingAccountId && orgId) {
      try {
        const query = qs.stringify(
          {
            details: btoa(
              qs.stringify({
                billing_id: billingAccount?.id,
                organization_id: billingAccount?.org_id,
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
          billingAccount?.org_id || '',
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
    billingAccount?.org_id,
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

  const isProviderIdUnavailable =
    billingAccount?.provider_id === undefined ||
    billingAccount?.provider_id === '';

  const isOrganizationKycCompleted = organizationKyc?.status === true;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size="large">Billing</Text>
      </Flex>
      <Flex direction="column" gap={9} style={styles.container}>
        <Flex direction="column" gap={7}>
          <BillingHeader
            isLoading={isLoading}
            billingSupportEmail={config.billing?.supportEmail}
          />
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
              // hideUpdatePaymentMethodBtn={isProviderIdUnavailable}
            />
            <BillingDetails
              billingAccount={billingAccount}
              onAddDetailsClick={onAddDetailsClick}
              isLoading={isLoading}
              isAllowed={isAllowed}
              // hideUpdateBillingDetailsBtn={isProviderIdUnavailable}
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
      <Outlet />
    </Flex>
  );
}
