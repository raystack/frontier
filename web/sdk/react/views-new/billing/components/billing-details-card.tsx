'use client';

import { Button, Skeleton, Text, Flex, Tooltip } from '@raystack/apsara-v1';
import type { BillingAccount } from '@raystack/proton/frontier';
import { convertBillingAddressToString } from '../../../utils';
import styles from '../billing-view.module.css';

interface BillingDetailsCardProps {
  billingAccount?: BillingAccount;
  isLoading: boolean;
  isAllowed: boolean;
  disabled?: boolean;
  isActionLoading?: boolean;
  onUpdateClick?: () => void;
}

export function BillingDetailsCard({
  billingAccount,
  isLoading,
  isAllowed,
  disabled = false,
  isActionLoading = false,
  onUpdateClick
}: BillingDetailsCardProps) {
  const btnText =
    billingAccount?.email || billingAccount?.name ? 'Update' : 'Add details';
  const isRestricted = isLoading || disabled;
  const isButtonDisabled = isRestricted || isActionLoading;

  const address = convertBillingAddressToString(billingAccount?.address);

  return (
    <div className={styles.detailsBox}>
      <Flex align="center" justify="between">
        <Text size="regular" weight="medium">
          Billing details
        </Text>
        {!isLoading && isAllowed ? (
          <Tooltip>
            <Tooltip.Trigger
              disabled={!isRestricted}
              render={<span />}
            >
              <Button
                variant="outline"
                color="neutral"
                size="small"
                onClick={onUpdateClick}
                disabled={isButtonDisabled}
                loading={isActionLoading}
                loaderText={btnText}
                data-test-id="frontier-sdk-billing-details-update-button"
              >
                {btnText}
              </Button>
            </Tooltip.Trigger>
            {isRestricted && (
              <Tooltip.Content>
                Contact support to update your billing address.
              </Tooltip.Content>
            )}
          </Tooltip>
        ) : null}
      </Flex>
      {isLoading ? (
        <Flex direction="column" gap={2}>
          <Skeleton height={16} width={60} />
          <Skeleton height={20} width={250} />
        </Flex>
      ) : (
        <Flex direction="column" gap={2}>
          <Text size="mini" weight="medium" variant="secondary">
            Name
          </Text>
          <Text size="regular">{billingAccount?.name || 'N/A'}</Text>
        </Flex>
      )}
      {isLoading ? (
        <Flex direction="column" gap={2}>
          <Skeleton height={16} width={80} />
          <Skeleton height={20} width={300} />
        </Flex>
      ) : (
        <Flex direction="column" gap={2}>
          <Text size="mini" weight="medium" variant="secondary">
            Address
          </Text>
          <Text size="regular">{address || 'N/A'}</Text>
        </Flex>
      )}
    </div>
  );
}
