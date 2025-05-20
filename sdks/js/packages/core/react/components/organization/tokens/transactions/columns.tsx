import { ApsaraColumnDef, Avatar, Flex } from '@raystack/apsara';
import { Text } from '@raystack/apsara/v1';
import dayjs from 'dayjs';
import { V1Beta1BillingTransaction } from '~/src';
import * as _ from 'lodash';
import tokenStyles from '../token.module.css';
import { getInitials } from '~/utils';

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
) => ApsaraColumnDef<V1Beta1BillingTransaction>[] = ({ dateFormat }) => [
  {
    header: 'Date',
    accessorKey: 'created_at',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: ({ row, getValue }) => {
      const value = getValue();
      return (
        <Flex direction="column">
          <Text size="regular" variant="secondary" className={tokenStyles.textMuted}>
            {dayjs(value).format(dateFormat)}
          </Text>
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
    cell: ({ row, getValue }) => {
      const value = getValue();
      const prefix = row?.original?.type === 'credit' ? '+' : '-';
      return (
        <Flex direction="column">
          <Text size="regular" variant="secondary" className={tokenStyles.textMuted}>
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
    cell: ({ row, getValue }) => {
      const value = getValue();
      const eventName = _.has(TxnEventSourceMap, value)
        ? _.get(TxnEventSourceMap, value)
        : row?.original?.description;
      return (
        <Flex direction="column">
          <Text size="regular" variant="secondary" className={tokenStyles.textMuted}>
            {eventName || '-'}
          </Text>
        </Flex>
      );
    }
  },
  {
    header: 'Member',
    accessorKey: 'user_id',
    meta: {
      style: {
        minHeight: '48px',
        padding: '12px 0'
      }
    },
    cell: ({ row }) => {
      const userTitle =
        row?.original?.user?.title || row?.original?.user?.email || '-';
      const avatarSrc = row?.original?.user?.avatar;
      return (
        <Flex direction="row" gap={'small'} align={'center'}>
          <Avatar
            shape={'square'}
            src={avatarSrc}
            fallback={getInitials(userTitle)}
            imageProps={{ width: '24px', height: '24px' }}
          />
          <Text size="regular">
            {userTitle}
          </Text>
        </Flex>
      );
    }
  }
];
