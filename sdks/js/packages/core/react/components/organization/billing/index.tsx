import { Button, Skeleton, Text } from '@raystack/apsara/v1';
import { Flex } from '@raystack/apsara';
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
import { toast } from '@raystack/apsara/v1';
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
    <Flex direction="column" gap="small">
      {isLoading ? (
        <Skeleton containerClassName={billingStyles.flex1} />
      ) : (
        <Text size="large>Billing</Text>
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
              <a
                data-test-id="frontier-sdk-billing-email-link"
                href={`mailto:${billingSupportEmail}`}
                target="_blank"
                style={{ fontWeight: 400, color: 'var(--foreground-accent)' }}
              >
                {billingSupportEmail}
              </a>
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
  hideUpdateBillingDetailsBtn: boolean;
}

const BillingDetails = ({
  billingAccount,
  onAddDetailsClick = () => {},
  isLoading,
  isAllowed,
  hideUpdateBillingDetailsBtn = false
}: BillingDetailsProps) => {
  // const addressStr = converBillingAddressToString(billingAccount?.address);
  const btnText =
    billingAccount?.email || billingAccount?.name ? 'Update' : 'Add details';
  return (
    <div className={billingStyles.detailsBox}>
      <Flex align={'center'} justify={'between'} style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Billing Details</Text>
        {isAllowed && !hideUpdateBillingDetailsBtn ? (
          <Button
            data-test-id="frontier-sdk-billing-details-update-button"
            variant="outline"
            color="neutral"
            size="small"
            onClick={onAddDetailsClick}
            disabled={isLoading}
          >
            {btnText}
          </Button>
        ) : null}
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Name</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : billingAccount?.name || 'N/A'}
        </Text>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
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
        <Text size="large>Billing</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
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
          <Flex style={{ gap: '24px' }}>
            <PaymentMethod
              paymentMethod={paymentMethod}
              isLoading={isLoading}
              isAllowed={isAllowed}
              hideUpdatePaymentMethodBtn={isProviderIdUnavailable}
            />
            <BillingDetails
              billingAccount={billingAccount}
              onAddDetailsClick={onAddDetailsClick}
              isLoading={isLoading}
              isAllowed={isAllowed}
              hideUpdateBillingDetailsBtn={isProviderIdUnavailable || isOrganizationKycCompleted}
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
