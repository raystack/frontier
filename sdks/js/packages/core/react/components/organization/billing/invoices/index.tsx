import type { DataTableColumnDef } from '@raystack/apsara/v1';
import { EmptyState, Flex, Link, Text, DataTable, Amount } from '@raystack/apsara/v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import dayjs from 'dayjs';
import { useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT, INVOICE_STATES } from '~/react/utils/constants';
import type { V1Beta1Invoice } from '~/src';
import { capitalize } from '~/utils';
import styles from './invoice.module.css';

interface InvoicesProps {
  isLoading: boolean;
  invoices: V1Beta1Invoice[];
}

interface getColumnsOptions {
  dateFormat: string;
}

export const getColumns: (
  options: getColumnsOptions
) => DataTableColumnDef<V1Beta1Invoice, unknown>[] = ({ dateFormat }) => [
  {
    header: 'Date',
    accessorKey: 'effective_at',
    cell: ({ row, getValue }) => {
      const value =
        row.original?.state === INVOICE_STATES.DRAFT
          ? row?.original?.due_date
          : (getValue() as string);
      return (
        <Flex direction="column">
          <Text>{value ? dayjs(value).format(dateFormat) : '-'}</Text>
        </Flex>
      );
    }
  },
  {
    header: 'Status',
    accessorKey: 'state',
    cell: ({ row, getValue }) => {
      return (
        <Flex direction="column">
          <Text>{capitalize(getValue() as string)}</Text>
        </Flex>
      );
    }
  },
  {
    header: 'Amount',
    accessorKey: 'amount',
    cell: ({ row, getValue }) => {
      return (
        <Flex direction="column">
          <Amount
            currency={row?.original?.currency}
            value={getValue() as number}
          />
        </Flex>
      );
    }
  },
  {
    header: '',
    accessorKey: 'hosted_url',
    classNames: {
      cell: styles.linkColumn
    },
    enableSorting: false,
    cell: ({ row, getValue }) => {
      const link = getValue() as string;
      return link ? (
        <Link
          href={link}
          target="_blank"
          rel="noopener noreferrer"
          data-test-id="frontier-sdk-view-invoice-link"
        >
          View invoice
        </Link>
      ) : null;
    }
  }
];

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading={'No previous invoices'}
  />
);

export default function Invoices({ isLoading, invoices }: InvoicesProps) {
  const { config } = useFrontier();

  const columns = getColumns({
    dateFormat: config?.dateFormat || DEFAULT_DATE_FORMAT
  });
  const tableStyle = useMemo(
    () =>
      invoices?.length ? { width: '100%' } : { width: '100%', height: '100%' },
    [invoices?.length]
  );

  const data = useMemo(() => {
    return invoices.sort((a, b) =>
      dayjs(a.effective_at).isAfter(b.effective_at) ? -1 : 1
    );
  }, [invoices]);

  return (
    <DataTable
      columns={columns}
      isLoading={isLoading}
      data={data}
      defaultSort={{ name: 'effective_at', order: 'desc' }}
      mode="client"
    >
      <DataTable.Content
        emptyState={noDataChildren}
        classNames={{ header: styles.header }}
      />
    </DataTable>
  );
}
