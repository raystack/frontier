'use client';

import { useCallback } from 'react';
import { Button, Skeleton, Text, Flex } from '@raystack/apsara-v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { INVOICE_STATES, SUBSCRIPTION_STATES } from '../../../utils/constants';
import { Subscription, Invoice } from '@raystack/proton/frontier';
import { timestampToDayjs } from '../../../../utils/timestamp';
import styles from '../billing-view.module.css';

interface PaymentIssueProps {
  isLoading?: boolean;
  subscription?: Subscription;
  invoices: Invoice[];
}

export function PaymentIssue({
  isLoading,
  subscription,
  invoices
}: PaymentIssueProps) {
  const isPastDue = subscription?.state === SUBSCRIPTION_STATES.PAST_DUE;
  const openInvoices = invoices
    .filter(inv => inv.state === INVOICE_STATES.OPEN)
    .sort((a, b) => {
      const dateA = timestampToDayjs(a.dueDate);
      const dateB = timestampToDayjs(b.dueDate);
      if (!dateA || !dateB) return 0;
      return dateA.isAfter(dateB) ? -1 : 1;
    });

  const onRetryPayment = useCallback(() => {
    window.location.href = openInvoices[0]?.hostedUrl || '';
  }, [openInvoices]);

  if (isLoading) return <Skeleton />;
  if (!isPastDue) return null;

  return (
    <Flex className={styles.paymentIssueBox} justify="between" align="center">
      <Flex gap={3} align="center">
        <ExclamationTriangleIcon />
        <Text size="small">
          Your payment is due. Please try again.
        </Text>
      </Flex>
      <Button
        variant="text"
        color="neutral"
        size="small"
        onClick={onRetryPayment}
        data-test-id="frontier-sdk-retry-payment-btn"
      >
        Retry
      </Button>
    </Flex>
  );
}
