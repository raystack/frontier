'use client';

import qs from 'query-string';
import { Button, Skeleton, Text, Flex, Tooltip, toastManager } from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import {
  PaymentMethod as PaymentMethodType,
  FrontierServiceQueries,
  CreateCheckoutRequestSchema
} from '@raystack/proton/frontier';
import { useMutation } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { AuthTooltipMessage } from '../../../utils';
import styles from '../billing-view.module.css';

interface PaymentMethodCardProps {
  paymentMethod?: PaymentMethodType;
  isLoading: boolean;
  isAllowed: boolean;
}

export function PaymentMethodCard({
  paymentMethod,
  isLoading,
  isAllowed
}: PaymentMethodCardProps) {
  const { config, activeOrganization } = useFrontier();

  const { mutateAsync: createCheckoutMutation, isPending: isActionLoading } =
    useMutation(FrontierServiceQueries.createCheckout, {
      onError: (err: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: err?.message,
          type: 'error'
        });
      }
    });

  const {
    cardLast4 = '',
    cardExpiryMonth,
    cardExpiryYear,
    cardBrand
  } = paymentMethod || {};

  const cardInfo = cardLast4
    ? `${cardBrand ? cardBrand.charAt(0).toUpperCase() + cardBrand.slice(1) : 'Card'} ending in ${cardLast4}`
    : 'N/A';

  const cardExp =
    cardExpiryMonth && cardExpiryYear
      ? `${cardExpiryMonth}/${cardExpiryYear}`
      : 'N/A';

  const isPaymentMethodAvailable = cardLast4 !== '';

  async function updatePaymentMethod() {
    const query = qs.stringify(
      {
        details: btoa(
          qs.stringify({
            organization_id: activeOrganization?.id,
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
        orgId: activeOrganization?.id || '',
        cancelUrl: cancel_url,
        successUrl: success_url,
        setupBody: {
          paymentMethod: true,
          customerPortal: false
        }
      })
    );
    const checkoutUrl = resp?.checkoutSession?.checkoutUrl;
    if (checkoutUrl) {
      window.location.href = checkoutUrl;
    }
  }

  const isBtnDisabled = isLoading || isActionLoading;

  return (
    <div className={styles.detailsBox}>
      <Flex align="center" justify="between">
        <Text size="regular" weight="medium">
          Payment method
        </Text>
        {isAllowed ? (
          <Tooltip>
            <Tooltip.Trigger render={<span />}>
              <Button
                variant="outline"
                color="neutral"
                size="small"
                onClick={updatePaymentMethod}
                disabled={isBtnDisabled}
                data-test-id="frontier-sdk-update-payment-method-btn"
              >
                {isPaymentMethodAvailable ? 'Update' : 'Add method'}
              </Button>
            </Tooltip.Trigger>
            <Tooltip.Content>
              {AuthTooltipMessage}
            </Tooltip.Content>
          </Tooltip>
        ) : null}
      </Flex>
      <Flex direction="column" gap={2}>
        <Text size="mini" weight="medium" variant="secondary">
          Card information
        </Text>
        <Text size="regular">
          {isLoading ? <Skeleton /> : cardInfo}
        </Text>
      </Flex>
      <Flex direction="column" gap={2}>
        <Text size="mini" weight="medium" variant="secondary">
          Expiry
        </Text>
        <Text size="regular">
          {isLoading ? <Skeleton /> : cardExp}
        </Text>
      </Flex>
    </div>
  );
}
