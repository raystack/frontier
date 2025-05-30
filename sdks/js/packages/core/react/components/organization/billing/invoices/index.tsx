import {
  ApsaraColumnDef,
  DataTable,
} from '@raystack/apsara';
import { EmptyState, Flex, Link, Text } from '@raystack/apsara/v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import dayjs from 'dayjs';
import { useMemo } from 'react';
import Amount from '~/react/components/helpers/Amount';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT, INVOICE_STATES } from '~/react/utils/constants';
import { V1Beta1Invoice } from '~/src';
import { capitalize } from '~/utils';

interface InvoicesProps {
  isLoading: boolean;
  invoices: V1Beta1Invoice[];
}

interface getColumnsOptions {
  dateFormat: string;
}

export const getColumns: (
  options: getColumnsOptions
) => ApsaraColumnDef<V1Beta1Invoice>[] = ({ dateFormat }) => [
  {
    header: 'Date',
    accessorKey: 'effective_at',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: ({ row, getValue }) => {
      const value =
        row.original?.state === INVOICE_STATES.DRAFT
          ? row?.original?.due_date
          : getValue();
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
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: ({ row, getValue }) => {
      return (
        <Flex direction="column">
          <Text>{capitalize(getValue())}</Text>
        </Flex>
      );
    }
  },
  {
    header: 'Amount',
    accessorKey: 'amount',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: ({ row, getValue }) => {
      return (
        <Flex direction="column">
          <Amount currency={row?.original?.currency} value={getValue()} />
        </Flex>
      );
    }
  },
  {
    header: '',
    accessorKey: 'hosted_url',
    meta: {
      style: {
        paddingLeft: 0,
        textAlign: 'end'
      }
    },
    enableSorting: false,
    cell: ({ row, getValue }) => {
      const link = getValue();
      return link ? (
        <Flex direction="column">
          <Link
            href={link}
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--foreground-accent)' }}
            data-test-id="frontier-sdk-view-invoice-link"
          >
            View invoice
          </Link>
        </Flex>
      ) : null;
    }
  }
];

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading={"No previous invoices"}
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
    <div>
      <DataTable
        columns={columns}
        isLoading={isLoading}
        data={data}
        style={tableStyle}
        emptyState={noDataChildren}
      />
    </div>
  );
}
