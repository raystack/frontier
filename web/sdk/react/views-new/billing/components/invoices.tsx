'use client';

import {
  Button,
  DataTable,
  EmptyState,
  Flex,
  Text,
  Amount
} from '@raystack/apsara-v1';
import type { DataTableColumnDef } from '@raystack/apsara-v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '../../../contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT, INVOICE_STATES } from '../../../utils/constants';
import type { Invoice } from '@raystack/proton/frontier';
import { capitalize } from '../../../../utils';
import { timestampToDayjs, type TimeStamp } from '../../../../utils/timestamp';
import styles from '../billing-view.module.css';

interface InvoicesProps {
  isLoading: boolean;
  invoices: Invoice[];
}

interface GetColumnsOptions {
  dateFormat: string;
}

const getColumns = ({
  dateFormat
}: GetColumnsOptions): DataTableColumnDef<Invoice, unknown>[] => [
  {
    header: 'Date',
    accessorKey: 'effectiveAt',
    cell: ({ row, getValue }) => {
      const value =
        row.original?.state === INVOICE_STATES.DRAFT
          ? row?.original?.dueDate
          : (getValue() as TimeStamp);
      const timestamp = value || row?.original?.createdAt;
      const date = timestampToDayjs(timestamp);
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
    cell: ({ getValue }) => {
      return (
        <Text size="regular" variant="secondary">
          {capitalize(getValue() as string)}
        </Text>
      );
    }
  },
  {
    header: 'Amount',
    accessorKey: 'amount',
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
    accessorKey: 'hostedUrl',
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

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No previous invoices"
  />
);

export function Invoices({ isLoading, invoices }: InvoicesProps) {
  const { config } = useFrontier();

  const columns = getColumns({
    dateFormat: config?.dateFormat || DEFAULT_DATE_FORMAT
  });

  return (
    <DataTable
      columns={columns}
      isLoading={isLoading}
      data={invoices}
      defaultSort={{ name: 'effectiveAt', order: 'desc' }}
      mode="client"
    >
      <DataTable.Content
        emptyState={noDataChildren}
      />
    </DataTable>
  );
}
