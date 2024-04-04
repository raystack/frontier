import { Button, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback, useEffect, useState } from 'react';
import billingStyles from './billing.module.css';
import {
  V1Beta1BillingAccount,
  V1Beta1Invoice,
  V1Beta1PaymentMethod
} from '~/src';
import * as _ from 'lodash';
import { toast } from 'sonner';
import Skeleton from 'react-loading-skeleton';
import { converBillingAddressToString } from '~/react/utils';
import Invoices from './invoices';

import { UpcomingBillingCycle } from './upcoming-billing-cycle';
import { PaymentIssue } from './payment-issue';
import { UpcomingPlanChangeBanner } from '../../common/upcoming-plan-change-banner';

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
        <Text size={6}>Billing</Text>
      )}
      {isLoading ? (
        <Skeleton containerClassName={billingStyles.flex1} />
      ) : (
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
          Oversee your billing and invoices.
          {billingSupportEmail ? (
            <>
              {' '}
              For more details, contact{' '}
              <a
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
}

const BillingDetails = ({
  billingAccount,
  onAddDetailsClick = () => {},
  isLoading
}: BillingDetailsProps) => {
  const addressStr = converBillingAddressToString(billingAccount?.address);
  // const btnText = addressStr || billingAccount?.name ? 'Update' : 'Add details';
  return (
    <div className={billingStyles.detailsBox}>
      <Flex align={'center'} justify={'between'} style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Billing Details</Text>
        {/* <Button variant={'secondary'} onClick={onAddDetailsClick}>
          {btnText}
        </Button> */}
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Name</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : billingAccount?.name || 'N/A'}
        </Text>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Address</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton count={2} /> : addressStr || 'N/A'}
        </Text>
      </Flex>
    </div>
  );
};

interface PaymentMethodProps {
  paymentMethod?: V1Beta1PaymentMethod;
  isLoading: boolean;
}

const PaymentMethod = ({
  paymentMethod = {},
  isLoading
}: PaymentMethodProps) => {
  const {
    card_last4 = '',
    card_expiry_month,
    card_expiry_year
  } = paymentMethod;
  // TODO: change card digit as per card type
  const cardDigit = 12;
  const cardNumber = card_last4 ? _.repeat('*', cardDigit) + card_last4 : 'N/A';
  const cardExp =
    card_expiry_month && card_expiry_year
      ? `${card_expiry_month}/${card_expiry_year}`
      : 'N/A';

  return (
    <div className={billingStyles.detailsBox}>
      <Flex align={'center'} justify={'between'} style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Payment method</Text>
        {/* <Button variant={'secondary'}>Add method</Button> */}
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Card Number</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : cardNumber}
        </Text>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Expiry</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : cardExp}
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
    paymentMethod
  } = useFrontier();
  const navigate = useNavigate({ from: '/billing' });

  const [invoices, setInvoices] = useState<V1Beta1Invoice[]>([]);
  const [isInvoicesLoading, setIsInvoicesLoading] = useState(false);

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

  const onAddDetailsClick = useCallback(() => {
    if (billingAccount?.id) {
      navigate({
        to: '/billing/$billingId/edit-address',
        params: { billingId: billingAccount?.id }
      });
    }
  }, [billingAccount?.id, navigate]);

  const isLoading =
    isBillingAccountLoading || isActiveSubscriptionLoading || isInvoicesLoading;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Billing</Text>
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
          />
          <Flex style={{ gap: '24px' }}>
            <PaymentMethod
              paymentMethod={paymentMethod}
              isLoading={isLoading}
            />
            <BillingDetails
              billingAccount={billingAccount}
              onAddDetailsClick={onAddDetailsClick}
              isLoading={isLoading}
            />
          </Flex>
          <UpcomingBillingCycle />
          <Invoices invoices={invoices} isLoading={isLoading} />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
