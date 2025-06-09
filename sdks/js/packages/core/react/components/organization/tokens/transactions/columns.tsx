import {
  Avatar,
  Text,
  Flex,
  type DataTableColumnDef,
  getAvatarColor
} from '@raystack/apsara/v1';
import dayjs from 'dayjs';
import type { V1Beta1BillingTransaction } from '~/src';
import * as _ from 'lodash';
import { getInitials } from '~/utils';
import tokenStyles from '../token.module.css';

interface getColumnsOptions {
  dateFormat: string;
}

const TxnEventSourceMap = {
  'system.starter': 'Starter tokens',
  'system.buy': 'Recharge',
  'system.awarded': 'Complimentary tokens',
  'system.revert': 'Refund'
};

export const getColumns: (
  options: getColumnsOptions
) => DataTableColumnDef<V1Beta1BillingTransaction, unknown>[] = ({
  dateFormat
}) => [
  {
    header: 'Date',
    accessorKey: 'created_at',
    cell: ({ row, getValue }) => {
      const value = getValue() as string;
      return (
        <Flex direction="column">
          <Text variant="secondary" size="regular">
            {dayjs(value).format(dateFormat)}
          </Text>
        </Flex>
      );
    }
  },
  {
    header: 'Tokens',
    accessorKey: 'amount',
    cell: ({ row, getValue }) => {
      const value = getValue() as number;
      const prefix = row?.original?.type === 'credit' ? '+' : '-';
      return (
        <Flex direction="column">
          <Text variant="secondary" size="regular">
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
          <Text variant="secondary" size="regular">
            {eventName || '-'}
          </Text>
        </Flex>
      );
    }
  },
  {
    header: 'Member',
    accessorKey: 'user_id',
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
          <Text size="regular">{userTitle}</Text>
        </Flex>
      );
    }
  }
];
