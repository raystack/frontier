import { DataTable, EmptyState, Flex, Link, Text } from '@raystack/apsara';
import { ColumnDef } from '@tanstack/react-table';
import dayjs from 'dayjs';
import { useMemo } from 'react';
import Skeleton from 'react-loading-skeleton';
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
  isLoading?: boolean;
  dateFormat: string;
}

export const getColumns: (
  options: getColumnsOptions
) => ColumnDef<V1Beta1Invoice, any>[] = ({ isLoading, dateFormat }) => [
  {
    header: 'Date',
    accessorKey: 'effective_at',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const value =
            row.original?.state === INVOICE_STATES.DRAFT
              ? row?.original?.due_date
              : getValue();
          return (
            <Flex direction="column">
              <Text>{dayjs(value).format(dateFormat)}</Text>
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const link = getValue();
          return link ? (
            <Flex direction="column">
              <Link
                href={link}
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: 'var(--foreground-accent)' }}
              >
                View invoice
              </Link>
            </Flex>
          ) : null;
        }
  }
];

const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <div className="pera">No previous invoices</div>
  </EmptyState>
);

export default function Invoices({ isLoading, invoices }: InvoicesProps) {
  const { config } = useFrontier();

  const columns = getColumns({
    isLoading: isLoading,
    dateFormat: config?.dateFormat || DEFAULT_DATE_FORMAT
  });
  const tableStyle = useMemo(
    () =>
      invoices?.length ? { width: '100%' } : { width: '100%', height: '100%' },
    [invoices?.length]
  );

  const data = useMemo(() => {
    return isLoading
      ? [...new Array(3)].map<V1Beta1Invoice>((_, i) => ({
          id: i.toString()
        }))
      : invoices.sort((a, b) =>
          dayjs(a.effective_at).isAfter(b.effective_at) ? -1 : 1
        );
  }, [invoices, isLoading]);

  return (
    <div>
      <DataTable
        columns={columns}
        data={data}
        style={tableStyle}
        emptyState={noDataChildren}
      />
    </div>
  );
}
