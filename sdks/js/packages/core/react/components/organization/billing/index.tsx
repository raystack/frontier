import { Button, Flex, Text } from '@raystack/apsara';
import { Outlet } from '@tanstack/react-router';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useEffect } from 'react';
import billingStyles from './billing.module.css';
import { V1Beta1BillingAccount } from '~/src';
import { converBillingAddressToString } from './helper';
import { InfoCircledIcon } from '@radix-ui/react-icons';

interface BillingHeaderProps {
  billingSupportEmail?: string;
}

const BillingHeader = ({ billingSupportEmail }: BillingHeaderProps) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>Billing</Text>
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
      </Flex>
    </Flex>
  );
};

interface BillingDetailsProps {
  billingAccount?: V1Beta1BillingAccount;
}

const BillingDetails = ({ billingAccount }: BillingDetailsProps) => {
  const addressStr = converBillingAddressToString(billingAccount?.address);
  return (
    <div className={billingStyles.detailsBox}>
      <Flex align={'center'} justify={'between'} style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Billing Details</Text>
        <Button variant={'secondary'}>Add details</Button>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Name</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {billingAccount?.name || 'NA'}
        </Text>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Address</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {addressStr || 'NA'}
        </Text>
      </Flex>
    </div>
  );
};

const PaymentMethod = () => {
  return (
    <div className={billingStyles.detailsBox}>
      <Flex align={'center'} justify={'between'} style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Payment method</Text>
        <Button variant={'secondary'}>Add method</Button>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>
          Card information
        </Text>
        <Text className={billingStyles.detailsBoxRowValue}>NA</Text>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Name on card</Text>
        <Text className={billingStyles.detailsBoxRowValue}>NA</Text>
      </Flex>
    </div>
  );
};

const CurrentPlanInfo = () => {
  return (
    <Flex
      className={billingStyles.currentPlanInfoBox}
      align={'center'}
      justify={'between'}
      gap={'small'}
    >
      <Flex gap={'small'}>
        <InfoCircledIcon className={billingStyles.currentPlanInfoText} />
        <Text size={2} className={billingStyles.currentPlanInfoText}>
          You are on starter plan
        </Text>
      </Flex>
      <Button variant={'secondary'}>Upgrade plan</Button>
    </Flex>
  );
};

export default function Billing() {
  const { billingAccount, client, config } = useFrontier();

  useEffect(() => {
    async function getPaymentMethod(orgId: string, billingId: string) {
      const resp = await client?.frontierServiceGetBillingAccount(
        orgId,
        billingId,
        { with_payment_methods: true }
      );
      console.log(resp);
    }

    async function getSubscription(orgId: string, billingId: string) {
      const resp = await client?.frontierServiceListSubscriptions(
        orgId,
        billingId
      );
      console.log(resp);
    }
    if (billingAccount?.id && billingAccount?.org_id) {
      getPaymentMethod(billingAccount?.org_id, billingAccount?.id);
      getSubscription(billingAccount?.org_id, billingAccount?.id);
    }
  }, [billingAccount?.id, billingAccount?.org_id, client]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Billing</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <BillingHeader billingSupportEmail={config.billingSupportEmail} />
          <Flex style={{ gap: '24px' }}>
            <PaymentMethod />
            <BillingDetails billingAccount={billingAccount} />
          </Flex>
          <CurrentPlanInfo />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
