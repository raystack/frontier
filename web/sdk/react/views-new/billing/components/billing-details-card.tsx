'use client';

import { Button, Skeleton, Text, Flex, Tooltip } from '@raystack/apsara-v1';
import type { BillingAccount } from '@raystack/proton/frontier';
import { converBillingAddressToString } from '../../../utils';
import styles from '../billing-view.module.css';

interface BillingDetailsCardProps {
  billingAccount?: BillingAccount;
  isLoading: boolean;
  isAllowed: boolean;
  disabled?: boolean;
  onUpdateClick?: () => void;
}

export function BillingDetailsCard({
  billingAccount,
  isLoading,
  isAllowed,
  disabled = false,
  onUpdateClick
}: BillingDetailsCardProps) {
  const btnText =
    billingAccount?.email || billingAccount?.name ? 'Update' : 'Add details';
  const isButtonDisabled = isLoading || disabled;

  const address = converBillingAddressToString(billingAccount?.address);

  return (
    <div className={styles.detailsBox}>
      <Flex align="center" justify="between">
        <Text size="regular" weight="medium">
          Billing details
        </Text>
        {isAllowed ? (
          <Tooltip>
            <Tooltip.Trigger
              disabled={!isButtonDisabled}
              render={<span />}
            >
              <Button
                variant="outline"
                color="neutral"
                size="small"
                onClick={onUpdateClick}
                disabled={isButtonDisabled}
                data-test-id="frontier-sdk-billing-details-update-button"
              >
                {btnText}
              </Button>
            </Tooltip.Trigger>
            {isButtonDisabled && (
              <Tooltip.Content>
                Contact support to update your billing address.
              </Tooltip.Content>
            )}
          </Tooltip>
        ) : null}
      </Flex>
      <Flex direction="column" gap={2}>
        <Text size="mini" weight="medium" variant="secondary">
          Name
        </Text>
        <Text size="regular">
          {isLoading ? <Skeleton /> : billingAccount?.name || 'N/A'}
        </Text>
      </Flex>
      <Flex direction="column" gap={2}>
        <Text size="mini" weight="medium" variant="secondary">
          Address
        </Text>
        <Text size="regular">
          {isLoading ? <Skeleton /> : address || 'N/A'}
        </Text>
      </Flex>
    </div>
  );
}
