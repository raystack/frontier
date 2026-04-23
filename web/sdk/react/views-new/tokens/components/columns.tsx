import {
  Avatar,
  Flex,
  Text,
  Tooltip,
  getAvatarColor
} from '@raystack/apsara-v1';
import type { DataTableColumnDef } from '@raystack/apsara-v1';
import type { SearchOrganizationTokensResponse_OrganizationToken } from '@raystack/proton/frontier';
import { has, get } from 'lodash';
import { getInitials } from '~/utils';
import {
  isNullTimestamp,
  type TimeStamp,
  timestampToDate
} from '~/utils/timestamp';
import dayjs from 'dayjs';
import styles from '../tokens-view.module.css';

// Backend source constants mirror frontier/billing/credit/credit.go
const TxnEventSourceMap: Record<string, string> = {
  'system.starter': 'Starter tokens',
  'system.buy': 'Recharge',
  'system.awarded': 'Complimentary tokens',
  'system.revert': 'Refund',
  'system.overdraft': 'Overdraft'
};

const eventFilterOptions = Object.entries(TxnEventSourceMap).map(
  ([value, label]) => ({ value, label })
);

interface GetColumnsOptions {
  dateFormat: string;
}

export function getColumns({
  dateFormat
}: GetColumnsOptions): DataTableColumnDef<
  SearchOrganizationTokensResponse_OrganizationToken,
  unknown
>[] {
  return [
    {
      header: 'Date',
      accessorKey: 'createdAt',
      enableSorting: true,
      enableColumnFilter: true,
      filterType: 'date',
      styles: { cell: { flex: '0 0 140px' }, header: { flex: '0 0 140px' } },
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
      enableSorting: true,
      enableColumnFilter: true,
      filterType: 'number',
      styles: { cell: { flex: '0 0 200px' }, header: { flex: '0 0 200px' } },
      cell: ({ row, getValue }) => {
        const value = Number(getValue());
        const prefix = row?.original?.type === 'credit' ? '+' : '-';
        return (
          <Text size="regular" variant="secondary">
            {prefix}
            {value}
          </Text>
        );
      }
    },
    {
      header: 'Event',
      accessorKey: 'source',
      enableHiding: true,
      enableColumnFilter: true,
      filterType: 'multiselect',
      filterOptions: eventFilterOptions,
      cell: ({ row, getValue }) => {
        const value = getValue() as string;
        const eventName = (
          has(TxnEventSourceMap, value)
            ? get(TxnEventSourceMap, value)
            : row?.original?.description
        ) as string;
        if (!eventName) {
          return (
            <Text size="regular" variant="secondary">
              -
            </Text>
          );
        }
        return (
          <Tooltip message={eventName} delayDuration={500}>
            <Text
              size="regular"
              variant="secondary"
              className={styles.truncate}
            >
              {eventName}
            </Text>
          </Tooltip>
        );
      }
    },
    {
      header: 'Member',
      accessorKey: 'userTitle',
      enableSorting: true,
      enableHiding: true,
      cell: ({ row, getValue }) => {
        const userId = row?.original?.userId || '';
        const title = (getValue() as string) || '-';
        const avatarSrc = row?.original?.userAvatar;
        const color = getAvatarColor(userId);
        return (
          <Flex direction="row" gap={4} align="center">
            <Avatar
              src={avatarSrc}
              fallback={getInitials(title)}
              size={3}
              radius="full"
              color={color}
            />
            <Text size="regular">{title}</Text>
          </Flex>
        );
      }
    }
  ];
}
