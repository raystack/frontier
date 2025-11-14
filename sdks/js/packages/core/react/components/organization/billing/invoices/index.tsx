import type { DataTableColumnDef } from '@raystack/apsara';
import {
  EmptyState,
  Flex,
  Link,
  Text,
  DataTable,
  Amount
} from '@raystack/apsara';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT, INVOICE_STATES } from '~/react/utils/constants';
import type { Invoice } from '@raystack/proton/frontier';
import { capitalize } from '~/utils';
import { timestampToDayjs, type TimeStamp } from '~/utils/timestamp';
import styles from './invoice.module.css';

interface InvoicesProps {
  isLoading: boolean;
  invoices: Invoice[];
}

interface getColumnsOptions {
  dateFormat: string;
}

export const getColumns: (
  options: getColumnsOptions
) => DataTableColumnDef<Invoice, unknown>[] = ({ dateFormat }) => [
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
        <Flex direction="column">
          <Text>{date ? date.format(dateFormat) : '-'}</Text>
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
      const value = Number(getValue());
      return (
        <Flex direction="column">
          <Text>
            <Amount currency={row?.original?.currency} value={value} />
          </Text>
        </Flex>
      );
    }
  },
  {
    header: '',
    accessorKey: 'hostedUrl',
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
        classNames={{ header: styles.header }}
      />
    </DataTable>
  );
}
