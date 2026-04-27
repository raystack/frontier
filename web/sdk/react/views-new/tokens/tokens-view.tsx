'use client';

import { useEffect, useMemo, useState } from 'react';
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
import type { DataTableQuery, DataTableSort } from '@raystack/apsara-v1';
import {
  InfoCircledIcon,
  PlusIcon,
  ExclamationTriangleIcon
} from '@radix-ui/react-icons';
import { useInfiniteQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useDebounceValue } from 'usehooks-ts';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import { useTokens } from '~/react/hooks/useTokens';
import { AuthTooltipMessage, getFormattedNumberString } from '~/react/utils';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import {
  DEFAULT_PAGE_SIZE,
  getConnectNextPageParam
} from '~/react/utils/connect-pagination';
import { transformDataTableQueryToRQLRequest } from '~/react/utils/transform-query';
import { ViewContainer } from '~/react/components/view-container';
import { ViewHeader } from '~/react/components/view-header';
import coin from '~/react/assets/coin.svg';
import { getColumns } from './components/columns';
import { AddTokensDialog } from './components/add-tokens-dialog';
import styles from './tokens-view.module.css';

const addTokensDialogHandle = Dialog.createHandle();

const DEFAULT_SORT: DataTableSort = { name: 'createdAt', order: 'desc' };

const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE
};

const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: 'created_at',
    userTitle: 'user_title'
  }
};

const NoTokens = () => (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No Transactions"
  />
);

const ErrorState = () => (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="Error loading transactions"
    subHeading="Something went wrong. Please try refreshing the page."
  />
);

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

  const organizationId = activeOrganization?.id ?? '';

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const computedQuery = useMemo(
    () => transformDataTableQueryToRQLRequest(tableQuery, TRANSFORM_OPTIONS),
    [tableQuery]
  );

  const [query] = useDebounceValue(computedQuery, 200);

  const {
    data: infiniteData,
    isLoading: isTransactionsListLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError,
    error: transactionsError
  } = useInfiniteQuery(
    FrontierServiceQueries.searchOrganizationTokens,
    { id: organizationId, query },
    {
      enabled: !!organizationId,
      pageParamKey: 'query',
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query }, 'organizationTokens'),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000
    }
  );

  useEffect(() => {
    if (transactionsError) {
      console.error('Unable to fetch transactions', transactionsError);
    }
  }, [transactionsError]);

  const transactionsData = useMemo(
    () =>
      infiniteData?.pages?.flatMap(page => page.organizationTokens) ?? [],
    [infiniteData]
  );

  const isLoading =
    isActiveOrganizationLoading ||
    isBillingAccountLoading ||
    isTokensLoading ||
    isPermissionsFetching;

  const isTxnDataLoading =
    (isLoading || isTransactionsListLoading || isFetchingNextPage) && !isError;

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  const fetchMore = async () => {
    if (hasNextPage && !isFetchingNextPage && !isError) {
      try {
        await fetchNextPage();
      } catch (err) {
        console.error('Unable to fetch next page of transactions', err);
      }
    }
  };

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
        <Callout
          type="accent"
          icon={<InfoCircledIcon />}
          className={styles.callout}
        >
          You can now add tokens anytime to reduce next month&apos;s invoice.
          But this won&apos;t settle any existing or overdue invoices.
        </Callout>
      ) : null}

      <Flex className={styles.balancePanel} justify="between" align="center">
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
              <Headline size="t2">{formattedBalance}</Headline>
            )}
          </Flex>
        </Flex>
        {isLoading ? (
          <Skeleton height="28px" width="100px" />
        ) : (
          <Tooltip>
            <Tooltip.Trigger disabled={isAllowed} render={<span />}>
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
        columns={columns}
        data={transactionsData}
        isLoading={isTxnDataLoading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMore}
        query={tableQuery}
      >
        <Flex direction="column" gap={5} style={{ width: '100%' }}>
          <Text size="small" weight="medium">
            Token transactions
          </Text>
          <DataTable.Toolbar />
          <DataTable.VirtualizedContent
            classNames={{ root: styles.tableRoot }}
            rowHeight={48}
            overscan={10}
            emptyState={isError ? <ErrorState /> : <NoTokens />}
          />
        </Flex>
      </DataTable>

      <AddTokensDialog handle={addTokensDialogHandle} />
    </ViewContainer>
  );
}
