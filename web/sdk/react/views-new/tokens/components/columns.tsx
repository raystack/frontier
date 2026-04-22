import {
  Avatar,
  Flex,
  Text,
  Tooltip,
  getAvatarColor
} from '@raystack/apsara-v1';
import type { DataTableColumnDef } from '@raystack/apsara-v1';
import type { SearchOrganizationTokensResponse_OrganizationToken } from '@raystack/proton/frontier';
import { getInitials } from '~/utils';
import {
  isNullTimestamp,
  type TimeStamp,
  timestampToDate
} from '~/utils/timestamp';
import dayjs from 'dayjs';
import styles from '../tokens-view.module.css';

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
      header: 'Events',
      accessorKey: 'description',
      enableHiding: true,
      cell: ({ getValue }) => {
        const text = (getValue() as string) ?? '';
        if (!text) {
          return (
            <Text size="regular" variant="secondary">
              -
            </Text>
          );
        }
        return (
          <Tooltip message={text} delayDuration={500}>
            <Text
              size="regular"
              variant="secondary"
              className={styles.truncate}
            >
              {text}
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
        const title = (getValue() as string) || userId || '-';
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
