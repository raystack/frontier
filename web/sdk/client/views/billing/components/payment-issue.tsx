'use client';

import { useCallback, useEffect } from 'react';
import { Button, Text, Flex, toastManager } from '@raystack/apsara';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  Subscription,
  RQLRequestSchema,
  RQLFilterSchema,
  RQLSortSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { INVOICE_STATES, SUBSCRIPTION_STATES } from '../../../utils/constants';
import { DEFAULT_PAGE_SIZE } from '../../../utils/connect-pagination';
import { useOrganizationInvoices } from '../../../hooks/useOrganizationInvoices';
import styles from '../billing-view.module.css';

// Open invoices with a non-zero amount, newest first — the invoice that needs
// payment when a subscription is past due.
const OPEN_INVOICES_QUERY = create(RQLRequestSchema, {
  filters: [
    create(RQLFilterSchema, {
      name: 'state',
      operator: 'eq',
      value: { case: 'stringValue', value: INVOICE_STATES.OPEN }
    }),
    create(RQLFilterSchema, {
      name: 'amount',
      operator: 'gt',
      value: { case: 'numberValue', value: 0 }
    })
  ],
  sort: [create(RQLSortSchema, { name: 'created_at', order: 'desc' })],
  offset: 0,
  limit: DEFAULT_PAGE_SIZE
});

interface PaymentIssueProps {
  isLoading?: boolean;
  subscription?: Subscription;
  hasAccess?: boolean;
}

export function PaymentIssue({
  isLoading,
  subscription,
  hasAccess
}: PaymentIssueProps) {
  const isPastDue = subscription?.state === SUBSCRIPTION_STATES.PAST_DUE;

  const {
    invoices,
    isLoading: isInvoicesLoading,
    error: invoicesError
  } = useOrganizationInvoices({
    query: OPEN_INVOICES_QUERY,
    enabled: hasAccess
  });

  useEffect(() => {
    if (invoicesError) {
      toastManager.add({
        title: 'Failed to load invoices',
        description: invoicesError?.message,
        type: 'error'
      });
    }
  }, [invoicesError]);

  const openInvoices = invoices.filter(
    inv => inv.state === INVOICE_STATES.OPEN
  );

  const onRetryPayment = useCallback(() => {
    window.location.href = openInvoices[0]?.invoiceLink || '';
  }, [openInvoices]);

  if (isLoading || isInvoicesLoading || !isPastDue) return null;

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
