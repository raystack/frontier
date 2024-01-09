import { Flex, Text } from '@raystack/apsara';
import { Outlet } from '@tanstack/react-router';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useEffect } from 'react';

interface BillingHeaderProps {
  billingSupportEmail?: string;
}

export const BillingHeader = ({ billingSupportEmail }: BillingHeaderProps) => {
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
    if (billingAccount?.id && billingAccount?.org_id) {
      getPaymentMethod(billingAccount?.org_id, billingAccount?.id);
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
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
