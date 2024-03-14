import { Flex, Text } from '@raystack/apsara';
import { ColumnDef } from '@tanstack/react-table';
import dayjs from 'dayjs';
import Skeleton from 'react-loading-skeleton';
import { V1Beta1BillingTransaction } from '~/src';

interface getColumnsOptions {
  isLoading: boolean;
  dateFormat: string;
}

export const getColumns: (
  options: getColumnsOptions
) => ColumnDef<V1Beta1BillingTransaction, any>[] = ({
  isLoading,
  dateFormat
}) => [
  {
    header: 'Date',
    accessorKey: 'created_at',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const value = getValue();
          return (
            <Flex direction="column">
              <Text>{dayjs(value).format(dateFormat)}</Text>
            </Flex>
          );
        }
  },
  {
    header: 'Tokens',
    accessorKey: 'amount',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const value = getValue();
          const prefix = row?.original?.type === 'credit' ? '+' : '-';
          return (
            <Flex direction="column">
              <Text>
                {prefix}
                {value}
              </Text>
            </Flex>
          );
        }
  }
];
