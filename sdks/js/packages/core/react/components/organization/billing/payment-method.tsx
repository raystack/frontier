import qs from 'query-string';
import { useFrontier } from '~/react/contexts/FrontierContext';
import * as _ from 'lodash';
import { Button, Skeleton, Text, Flex } from '@raystack/apsara';
import billingStyles from './billing.module.css';
import { V1Beta1CheckoutSetupBody } from '~/src';
import { PaymentMethod as PaymentMethodType } from '@raystack/proton/frontier';
import { toast } from '@raystack/apsara';
import { useState } from 'react';

interface PaymentMethodProps {
  paymentMethod?: PaymentMethodType;
  isLoading: boolean;
  isAllowed: boolean;
}

export const PaymentMethod = ({
  paymentMethod,
  isLoading,
  isAllowed
}: PaymentMethodProps) => {
  const { client, config, billingAccount } = useFrontier();
  const [isActionLoading, setIsActionLoading] = useState(false);
  const {
    cardLast4 = '',
    cardExpiryMonth,
    cardExpiryYear
  } = paymentMethod || {};
  // TODO: change card digit as per card type
  const cardDigit = 12;
  const cardNumber = cardLast4 ? _.repeat('*', cardDigit) + cardLast4 : 'N/A';
  const cardExp =
    cardExpiryMonth && cardExpiryYear
      ? `${cardExpiryMonth}/${cardExpiryYear}`
      : 'N/A';

  const isPaymentMethodAvailable = cardLast4 !== '';

  const updatePaymentMethod = async () => {
    const orgId = billingAccount?.orgId || '';
    const billingAccountId = billingAccount?.id || '';
    if (billingAccountId && orgId) {
      setIsActionLoading(true);
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

        const setup_body: V1Beta1CheckoutSetupBody = {
          payment_method: true
        };

        const resp = await client?.frontierServiceCreateCheckout(
          billingAccount?.orgId || '',
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
      } finally {
        setIsActionLoading(false);
      }
    }
  };

  function onClick() {
    updatePaymentMethod();
  }

  const isBtnDisabled = isLoading || isActionLoading;

  return (
    <div className={billingStyles.detailsBox}>
      <Flex align="center" justify="between" style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Payment method</Text>
        {isAllowed ? (
          <Button
            variant="outline"
            color="neutral"
            size="small"
            onClick={onClick}
            disabled={isBtnDisabled}
            data-test-id="frontier-sdk-update-payment-method-btn"
          >
            {isPaymentMethodAvailable ? 'Update' : 'Add method'}
          </Button>
        ) : null}
      </Flex>
      <Flex direction="column" gap={2}>
        <Text
          size="small"
          weight="medium"
          className={billingStyles.detailsBoxRowLabel}
        >
          Card Number
        </Text>
        <Text
          size="small"
          variant="secondary"
          className={billingStyles.detailsBoxRowValue}
        >
          {isLoading ? <Skeleton /> : cardNumber}
        </Text>
      </Flex>
      <Flex direction="column" gap={2}>
        <Text
          size="small"
          weight="medium"
          className={billingStyles.detailsBoxRowLabel}
        >
          Expiry
        </Text>
        <Text
          size="small"
          variant="secondary"
          className={billingStyles.detailsBoxRowValue}
        >
          {isLoading ? <Skeleton /> : cardExp}
        </Text>
      </Flex>
    </div>
  );
};
