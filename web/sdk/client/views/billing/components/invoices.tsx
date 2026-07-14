'use client';

import { useMemo, useState } from 'react';
import {
  Button,
  DataTable,
  EmptyState,
  Flex,
  Text,
  Amount
} from '@raystack/apsara';
import type {
  DataTableColumnDef,
  DataTableQuery,
  DataTableSort
} from '@raystack/apsara';
import {
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { type SearchOrganizationInvoicesResponse_OrganizationInvoice } from '@raystack/proton/frontier';
import { useDebouncedValue } from '~hooks';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useOrganizationInvoices } from '../../../hooks/useOrganizationInvoices';
import { DEFAULT_DATE_FORMAT, INVOICE_STATES } from '../../../utils/constants';
import { DEFAULT_PAGE_SIZE } from '../../../utils/connect-pagination';
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
      styles: {
        cell: { width: '120px', maxWidth: '120px' },
        header: { width: '120px', maxWidth: '120px' }
      },
      cell: ({ getValue }) => {
        const link = getValue() as string;
        if (!link) return null;
        return (
          <Button
            variant="text"
            color="neutral"
            size="small"
            className={styles.viewInvoiceBtn}
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

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const computedQuery = useMemo(
    () => transformDataTableQueryToRQLRequest(tableQuery),
    [tableQuery]
  );

  const query = useDebouncedValue(computedQuery, 200);

  const {
    invoices,
    groupCountMap,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError
  } = useOrganizationInvoices({ query });

  const loading = (!activeOrganization?.id || isLoading || isFetchingNextPage) && !isError;

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  const fetchMore = async () => {
    if (hasNextPage && !isFetchingNextPage && !isError) {
      await fetchNextPage();
    }
  };

  const columns = getColumns({
    dateFormat: config?.dateFormat || DEFAULT_DATE_FORMAT,
    groupCountMap
  });

  return (
    <DataTable
      columns={columns}
      data={invoices}
      isLoading={loading}
      loadingRowCount={5}
      defaultSort={DEFAULT_SORT}
      mode="server"
      onTableQueryChange={onTableQueryChange}
      onLoadMore={fetchMore}
      query={tableQuery}
    >
      <Flex direction="column" gap={4} style={{ width: '100%' }}>
        <Flex justify="between" align="center">
          <Text size="small" weight="medium">
            Billing transactions
          </Text>
          <DataTable.DisplayControls
          />
        </Flex>
        <DataTable.VirtualizedContent
          rowHeight={48}
          groupHeaderHeight={48}
          emptyState={isError ? <ErrorState /> : <NoInvoices />}
        />
      </Flex>
    </DataTable>
  );
}
