import { Avatar, Flex, Text, getAvatarColor } from '@raystack/apsara-v1';
import type { DataTableColumnDef } from '@raystack/apsara-v1';
import type { BillingTransaction } from '@raystack/proton/frontier';
import * as _ from 'lodash';
import { getInitials } from '~/utils';
import {
  isNullTimestamp,
  type TimeStamp,
  timestampToDate
} from '~/utils/timestamp';
import dayjs from 'dayjs';
import isSameOrBefore from 'dayjs/plugin/isSameOrBefore';
import isSameOrAfter from 'dayjs/plugin/isSameOrAfter';

dayjs.extend(isSameOrBefore);
dayjs.extend(isSameOrAfter);

interface GetColumnsOptions {
  dateFormat: string;
}

const TxnEventSourceMap: Record<string, string> = {
  'system.starter': 'Starter tokens',
  'system.buy': 'Recharge',
  'system.awarded': 'Complimentary tokens',
  'system.revert': 'Refund'
};

const eventFilterOptions = Object.entries(TxnEventSourceMap).map(
  ([value, label]) => ({ value, label })
);

export function getColumns({
  dateFormat
}: GetColumnsOptions): DataTableColumnDef<BillingTransaction, unknown>[] {
  return [
    {
      header: 'Date',
      accessorKey: 'createdAt',
      enableColumnFilter: true,
      filterType: 'date',
      styles: { cell: { flex: '0 0 140px' }, header: { flex: '0 0 140px' } },
      filterFn: (row, columnId, filterValue) => {
        const ts = row.getValue(columnId) as unknown as TimeStamp;
        if (isNullTimestamp(ts)) return false;
        const rowDate = dayjs(timestampToDate(ts));
        const filterDate = dayjs(filterValue.date);
        if (!rowDate.isValid() || !filterDate.isValid()) return false;
        const op = filterValue.operator || 'eq';
        switch (op) {
          case 'eq': return rowDate.isSame(filterDate, 'day');
          case 'neq': return !rowDate.isSame(filterDate, 'day');
          case 'lt': return rowDate.isBefore(filterDate, 'day');
          case 'lte': return rowDate.isSameOrBefore(filterDate, 'day');
          case 'gt': return rowDate.isAfter(filterDate, 'day');
          case 'gte': return rowDate.isSameOrAfter(filterDate, 'day');
          default: return true;
        }
      },
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? '-'
          : dayjs(timestampToDate(value)).format(dateFormat);
        return (
          <Text size="regular" variant="secondary">
            {date}
          </Text>
        );
      }
    },
    {
      header: 'Tokens',
      accessorKey: 'amount',
      styles: { cell: { flex: '0 0 200px' }, header: { flex: '0 0 200px' } },
      cell: ({ row, getValue }) => {
        const value = getValue() as bigint;
        const prefix = row?.original?.type === 'credit' ? '+' : '-';
        return (
          <Text size="regular" variant="secondary">
            {prefix}
            {Number(value)}
          </Text>
        );
      }
    },
    {
      header: 'Event',
      accessorKey: 'source',
      enableColumnFilter: true,
      filterType: 'multiselect',
      filterOptions: eventFilterOptions,
      cell: ({ row, getValue }) => {
        const value = getValue() as string;
        const eventName = (
          _.has(TxnEventSourceMap, value)
            ? _.get(TxnEventSourceMap, value)
            : row?.original?.description
        ) as string;
        return (
          <Text
            size="regular"
            variant="secondary"
            style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', display: 'block' }}
          >
            {eventName || '-'}
          </Text>
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
          <Flex direction="row" gap={4} align="center">
            <Avatar
              src={avatarSrc}
              fallback={getInitials(userTitle)}
              size={3}
              radius="full"
              color={color}
            />
            <Text size="regular">{userTitle}</Text>
          </Flex>
        );
      }
    }
  ];
}
