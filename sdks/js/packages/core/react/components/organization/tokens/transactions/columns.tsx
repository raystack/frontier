import {
  Avatar,
  Text,
  Flex,
  type DataTableColumnDef,
  getAvatarColor
} from '@raystack/apsara';
import type { BillingTransaction } from '~/src';
import * as _ from 'lodash';
import { getInitials } from '~/utils';
import {
  isNullTimestamp,
  type TimeStamp,
  timestampToDate
} from '~/utils/timestamp';
import tokenStyles from '../token.module.css';
import dayjs from 'dayjs';

interface getColumnsOptions {
  dateFormat: string;
}
const TxnEventSourceMap = {
  'system.starter': 'Starter tokens',
  'system.buy': 'Recharge',
  'system.awarded': 'Complimentary tokens',
  'system.revert': 'Refund'
};

export const getColumns = ({
  dateFormat
}: getColumnsOptions): DataTableColumnDef<BillingTransaction, unknown>[] => [
  {
    header: 'Date',
    accessorKey: 'createdAt',
    cell: ({ getValue }) => {
      const value = getValue() as TimeStamp;
      const date = isNullTimestamp(value)
        ? '-'
        : dayjs(timestampToDate(value)).format(dateFormat);
      return <Text variant="secondary">{date}</Text>;
    }
  },
  {
    header: 'Tokens',
    accessorKey: 'amount',
    cell: ({ row, getValue }) => {
      const value = getValue() as bigint;
      const prefix = row?.original?.type === 'credit' ? '+' : '-';
      return (
        <Flex direction="column">
          <Text variant="secondary">
            {prefix}
            {Number(value)}
          </Text>
        </Flex>
      );
    }
  },
  {
    header: 'Event',
    accessorKey: 'source',
    classNames: {
      cell: tokenStyles.txnTableEventColumn
    },
    cell: ({ row, getValue }) => {
      const value = getValue() as string;
      const eventName = (
        _.has(TxnEventSourceMap, value)
          ? _.get(TxnEventSourceMap, value)
          : row?.original?.description
      ) as string;
      return (
        <Flex direction="column">
          <Text variant="secondary">{eventName || '-'}</Text>
        </Flex>
      );
    }
  },
  {
    header: 'Member',
    accessorKey: 'userId',
    cell: ({ row, getValue }) => {
      const userTitle =
        row?.original?.user?.title || row?.original?.user?.email || '-';
      const avatarSrc = row?.original?.user?.avatar;
      const color = getAvatarColor(getValue() as string);
      return (
        <Flex direction="row" gap={'small'} align={'center'}>
          <Avatar
            src={avatarSrc}
            fallback={getInitials(userTitle)}
            size={3}
            radius="small"
            color={color}
          />
          <Text>{userTitle}</Text>
        </Flex>
      );
    }
  }
];
