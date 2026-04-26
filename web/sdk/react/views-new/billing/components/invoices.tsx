'use client';

import { useMemo, useState } from 'react';
import {
  Button,
  DataTable,
  EmptyState,
  Flex,
  Text,
  Amount
} from '@raystack/apsara-v1';
import type {
  DataTableColumnDef,
  DataTableQuery,
  DataTableSort
} from '@raystack/apsara-v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useInfiniteQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  type SearchOrganizationInvoicesResponse_OrganizationInvoice
} from '@raystack/proton/frontier';
import { useDebounceValue } from 'usehooks-ts';
import { useFrontier } from '../../../contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT, INVOICE_STATES } from '../../../utils/constants';
import {
  DEFAULT_PAGE_SIZE,
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage
} from '../../../utils/connect-pagination';
import { transformDataTableQueryToRQLRequest } from '../../../utils/transform-query';
import { timestampToDayjs, type TimeStamp } from '../../../../utils/timestamp';
import { capitalize } from '../../../../utils';
import styles from '../billing-view.module.css';

const DEFAULT_SORT: DataTableSort = { name: 'created_at', order: 'desc' };

const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE
};

type InvoiceStatus = (typeof INVOICE_STATES)[keyof typeof INVOICE_STATES];

const InvoiceStatusesMap: Record<InvoiceStatus, string> = Object.fromEntries(
  Object.values(INVOICE_STATES).map(state => [state, capitalize(state)])
) as Record<InvoiceStatus, string>;

interface GetColumnsOptions {
  dateFormat: string;
  groupCountMap: Record<string, Record<string, number>>;
}

const getColumns = ({
  dateFormat,
  groupCountMap
}: GetColumnsOptions): DataTableColumnDef<
  SearchOrganizationInvoicesResponse_OrganizationInvoice,
  unknown
>[] => [
  {
    header: 'Date',
    id: 'created_at',
    accessorKey: 'createdAt',
    enableSorting: true,
    cell: ({ getValue }) => {
      const value = getValue() as TimeStamp;
      const date = timestampToDayjs(value);
      return (
        <Text size="regular" variant="secondary">
          {date ? date.format(dateFormat) : '-'}
        </Text>
      );
    }
  },
  {
    header: 'Status',
    accessorKey: 'state',
    enableSorting: true,
    enableHiding: true,
    enableGrouping: true,
    groupLabelsMap: InvoiceStatusesMap,
    showGroupCount: true,
    groupCountMap: groupCountMap['state'] || {},
    cell: ({ getValue }) => {
      const value = getValue() as keyof typeof InvoiceStatusesMap;
      return (
        <Text size="regular" variant="secondary">
          {InvoiceStatusesMap[value] ?? capitalize(value as string)}
        </Text>
      );
    }
  },
  {
    header: 'Amount',
    accessorKey: 'amount',
    enableSorting: true,
    enableHiding: true,
    cell: ({ row, getValue }) => {
      const value = Number(getValue());
      return (
        <Text size="regular" variant="secondary">
          <Amount currency={row?.original?.currency} value={value} />
        </Text>
      );
    }
  },
  {
    header: '',
    accessorKey: 'invoiceLink',
    enableSorting: false,
    classNames: {
      cell: styles.linkColumn
    },
    cell: ({ getValue }) => {
      const link = getValue() as string;
      if (!link) return null;
      return (
        <Button
          variant="text"
          color="neutral"
          size="small"
          onClick={() => window.open(link, '_blank', 'noopener,noreferrer')}
          data-test-id="frontier-sdk-view-invoice-link"
        >
          View invoice
        </Button>
      );
    }
  }
];

const NoInvoices = () => (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No previous invoices"
  />
);

const ErrorState = () => (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="Error loading invoices"
    subHeading="Something went wrong. Please try refreshing the page."
  />
);

export function Invoices() {
  const { activeOrganization, config } = useFrontier();
  const organizationId = activeOrganization?.id || '';

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const computedQuery = useMemo(
    () => transformDataTableQueryToRQLRequest(tableQuery),
    [tableQuery]
  );

  const [query] = useDebounceValue(computedQuery, 200);

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError
  } = useInfiniteQuery(
    FrontierServiceQueries.searchOrganizationInvoices,
    { id: organizationId, query },
    {
      enabled: !!organizationId,
      pageParamKey: 'query',
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query }, 'organizationInvoices'),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000
    }
  );

  const invoices = useMemo(
    () =>
      infiniteData?.pages?.flatMap(page => page.organizationInvoices) ?? [],
    [infiniteData]
  );
  const loading = (isLoading || isFetchingNextPage) && !isError;

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  const fetchMore = async () => {
    if (hasNextPage && !isFetchingNextPage && !isError) {
      await fetchNextPage();
    }
  };

  const groupCountMap = useMemo(
    () => (infiniteData ? getGroupCountMapFromFirstPage(infiniteData) : {}),
    [infiniteData]
  );

  const columns = getColumns({
    dateFormat: config?.dateFormat || DEFAULT_DATE_FORMAT,
    groupCountMap
  });

  return (
    <DataTable
      columns={columns}
      data={invoices}
      isLoading={loading}
      defaultSort={DEFAULT_SORT}
      mode="server"
      onTableQueryChange={onTableQueryChange}
      onLoadMore={fetchMore}
      query={tableQuery}
    >
      <Flex direction="column" style={{ width: '100%' }}>
        <DataTable.Toolbar />
        <DataTable.Content
          emptyState={isError ? <ErrorState /> : <NoInvoices />}
        />
      </Flex>
    </DataTable>
  );
}
