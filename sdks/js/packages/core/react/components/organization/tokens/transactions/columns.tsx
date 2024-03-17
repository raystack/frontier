import { Flex, Text } from '@raystack/apsara';
import { ColumnDef } from '@tanstack/react-table';
import dayjs from 'dayjs';
import Skeleton from 'react-loading-skeleton';
import { V1Beta1BillingTransaction } from '~/src';
import * as _ from 'lodash';

interface getColumnsOptions {
  isLoading: boolean;
  dateFormat: string;
}

const TxnEventSourceMap = {
  'system.starter': 'Starter Credits',
  'system.buy': 'Recharge'
};

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
  },
  {
    header: 'Event',
    accessorKey: 'source',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const value = getValue();
          const eventName = _.has(TxnEventSourceMap, value)
            ? _.get(TxnEventSourceMap, value)
            : '-';
          return (
            <Flex direction="column">
              <Text>{eventName}</Text>
            </Flex>
          );
        }
  },
  {
    header: 'Member',
    accessorKey: 'user_id',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const userId = getValue() || '-';
          return (
            <Flex direction="column">
              <Text>{userId}</Text>
            </Flex>
          );
        }
  }
];
