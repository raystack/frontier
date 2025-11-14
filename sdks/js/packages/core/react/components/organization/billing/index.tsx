import {
  Button,
  Skeleton,
  Text,
  Flex,
  toast,
  Link,
  Tooltip
} from '@raystack/apsara';
import { Outlet } from '@tanstack/react-router';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback, useEffect } from 'react';
import billingStyles from './billing.module.css';
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
    isOrganizationKycLoading
  } = useFrontier();

  const { isAllowed, isFetching } = useBillingPermission();

  const {
    data: invoices = [],
    isLoading: isInvoicesLoading,
    error: invoicesError
  } = useConnectQuery(
    FrontierServiceQueries.listInvoices,
    create(ListInvoicesRequestSchema, {
      orgId: billingAccount?.orgId || '',
      billingId: billingAccount?.id || '',
      nonzeroAmountOnly: true
    }),
    {
      enabled: !!billingAccount?.id && !!billingAccount?.orgId,
      select: data => data?.invoices || []
    }
  );

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
    const orgId = billingAccount?.orgId || '';
    const billingAccountId = billingAccount?.id || '';
    if (!billingAccountId || !orgId) return;

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

      const resp = await createCheckoutMutation(
        create(CreateCheckoutRequestSchema, {
          orgId: billingAccount?.orgId || '',
          billingId: billingAccount?.id || '',
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
    billingAccount?.id,
    billingAccount?.orgId,
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
    <Flex direction="column" width="full">
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
      <Outlet />
    </Flex>
  );
}
