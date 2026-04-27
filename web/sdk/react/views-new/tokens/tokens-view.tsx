'use client';

import { useEffect, useMemo } from 'react';
import {
  Button,
  Callout,
  DataTable,
  Dialog,
  EmptyState,
  Flex,
  Headline,
  Image,
  Skeleton,
  Text,
  Tooltip
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import { InfoCircledIcon, PlusIcon, ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import { useTokens } from '~/react/hooks/useTokens';
import { AuthTooltipMessage, getFormattedNumberString } from '~/react/utils';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import {
  FrontierServiceQueries,
  ListBillingTransactionsRequestSchema
} from '@raystack/proton/frontier';
import { ViewContainer } from '~/react/components/view-container';
import { ViewHeader } from '~/react/components/view-header';
import coin from '~/react/assets/coin.svg';
import { getColumns } from './components/columns';
import { AddTokensDialog } from './components/add-tokens-dialog';
import styles from './tokens-view.module.css';

const addTokensDialogHandle = Dialog.createHandle();

export function TokensView() {
  const {
    config,
    activeOrganization,
    billingDetails,
    isActiveOrganizationLoading,
    isBillingAccountLoading
  } = useFrontier();

  const { isAllowed, isFetching: isPermissionsFetching } =
    useBillingPermission();
  const { tokenBalance, isTokensLoading } = useTokens();

  const {
    data: transactionsRawData,
    isLoading: isTransactionsListLoading,
    error: transactionsError
  } = useQuery(
    FrontierServiceQueries.listBillingTransactions,
    create(ListBillingTransactionsRequestSchema, {
      orgId: activeOrganization?.id ?? '',
      expand: ['user']
    }),
    {
      enabled: !!activeOrganization?.id
    }
  );

  const transactionsData = useMemo(
    () => transactionsRawData?.transactions ?? [],
    [transactionsRawData]
  );

  useEffect(() => {
    if (transactionsError) {
      console.error(transactionsError);
      toastManager.add({
        title: 'Unable to fetch transactions',
        type: 'error'
      });
    }
  }, [transactionsError]);

  const isLoading =
    !activeOrganization?.id ||
    isActiveOrganizationLoading ||
    isBillingAccountLoading ||
    isTokensLoading ||
    isPermissionsFetching;

  const isTxnDataLoading = isLoading || isTransactionsListLoading;

  const columns = useMemo(
    () =>
      getColumns({
        dateFormat: config.dateFormat || DEFAULT_DATE_FORMAT
      }),
    [config.dateFormat]
  );

  const formattedBalance = getFormattedNumberString(
    tokenBalance,
    config?.locale
  );

  const isPostpaid =
    billingDetails?.creditMin && billingDetails.creditMin < BigInt(0);

  const billingSupportEmail = config.billing?.supportEmail;
  const description = billingSupportEmail
    ? `Oversee your billing and invoices. For more details, contact ${billingSupportEmail}`
    : 'Oversee your billing and invoices.';

  const disableAddTokensBtn = isLoading || !isAllowed;

  return (
    <ViewContainer>
      <ViewHeader title="Tokens" description={description} />

      {isPostpaid && isAllowed ? (
        <Callout type="accent" icon={<InfoCircledIcon />} className={styles.callout}>
          You can now add tokens anytime to reduce next month&apos;s invoice. But
          this won&apos;t settle any existing or overdue invoices.
        </Callout>
      ) : null}

      <Flex
        className={styles.balancePanel}
        justify="between"
        align="center"
      >
        <Flex gap={4} align="center">
          <Image
            src={coin as unknown as string}
            alt="coin"
            className={styles.coinIcon}
          />
          <Flex direction="column" gap={2}>
            <Text size="mini" weight="medium" variant="secondary">
              Available tokens
            </Text>
            {isLoading ? (
              <Skeleton height="28px" width="120px" />
            ) : (
              <Headline size="t2">
                {formattedBalance}
              </Headline>
            )}
          </Flex>
        </Flex>
        {isLoading ? (
          <Skeleton height="28px" width="100px" />
        ) : (
          <Tooltip>
            <Tooltip.Trigger
              disabled={isAllowed}
              render={<span />}
            >
              <Button
                variant="outline"
                color="neutral"
                size="small"
                leadingIcon={<PlusIcon />}
                onClick={() => addTokensDialogHandle.open(null)}
                disabled={disableAddTokensBtn}
                data-test-id="frontier-sdk-add-tokens-btn"
              >
                Add tokens
              </Button>
            </Tooltip.Trigger>
            {!isAllowed && (
              <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
            )}
          </Tooltip>
        )}
      </Flex>

      <DataTable
        data={transactionsData}
        columns={columns}
        isLoading={isTxnDataLoading}
        mode="client"
        defaultSort={{ name: 'createdAt', order: 'desc' }}
      >
        <Flex direction="column" gap={5}>
          <Text size="small" weight="medium">
            Token transactions
          </Text>
          <DataTable.Filters />
          <DataTable.VirtualizedContent
            classNames={{ root: styles.tableRoot }}
            rowHeight={48}
            overscan={10}
            emptyState={
              <EmptyState
                icon={<ExclamationTriangleIcon />}
                heading="No Transactions"
              />
            }
          />
        </Flex>
      </DataTable>

      <AddTokensDialog handle={addTokensDialogHandle} />
    </ViewContainer>
  );
}
