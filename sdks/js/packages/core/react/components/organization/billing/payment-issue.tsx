import { Button, Skeleton, Image, Text, Flex } from '@raystack/apsara/v1';
import { INVOICE_STATES, SUBSCRIPTION_STATES } from '~/react/utils/constants';
import { V1Beta1Invoice, V1Beta1Subscription } from '~/src';
import billingStyles from './billing.module.css';
import exclamationTriangle from '~/react/assets/exclamation-triangle.svg';
import dayjs from 'dayjs';
import { useCallback } from 'react';

interface PaymentIssueProps {
  isLoading?: boolean;
  subscription?: V1Beta1Subscription;
  invoices: V1Beta1Invoice[];
}

export function PaymentIssue({
  isLoading,
  subscription,
  invoices
}: PaymentIssueProps) {
  const isPastDue = subscription?.state === SUBSCRIPTION_STATES.PAST_DUE;
  const openInvoices = invoices
    .filter(inv => inv.state === INVOICE_STATES.OPEN)
    .sort((a, b) => (dayjs(a.due_date).isAfter(b.due_date) ? -1 : 1));

  const onRetryPayment = useCallback(() => {
    window.location.href = openInvoices[0]?.hosted_url || '';
  }, [openInvoices]);

  return isLoading ? (
    <Skeleton />
  ) : isPastDue ? (
    <Flex className={billingStyles.paymentIssueBox} justify="between">
      <Flex gap={3} className={billingStyles.flex1}>
        <Image src={exclamationTriangle as unknown as string} alt="Exclamation Triangle" />
        <Text className={billingStyles.paymentIssueText}>
          Your Payment is due. Please try again
        </Text>
      </Flex>
      <Flex className={billingStyles.flex1} justify="end">
        <Button
          className={billingStyles.retryPaymentBtn}
          onClick={onRetryPayment}
          data-test-id="frontier-sdk-retry-payment-btn"
          variant="text"
          color="neutral"
          size="small"
        >
          Retry
        </Button>
      </Flex>
    </Flex>
  ) : null;
}
