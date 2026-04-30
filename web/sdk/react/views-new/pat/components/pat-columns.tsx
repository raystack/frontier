'use client';

import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { Button, Text, DataTableColumnDef } from '@raystack/apsara-v1';
import type { PAT } from '@raystack/proton/frontier';
import { timestampToDayjs, isNullTimestamp } from '~/utils/timestamp';
import styles from '../pat-view.module.css';

dayjs.extend(relativeTime);

export function getColumns({
  dateFormat,
  onRevoke
}: {
  dateFormat: string;
  onRevoke: (patId: string) => void;
}): DataTableColumnDef<PAT, unknown>[] {
  return [
    {
      header: 'Title',
      accessorKey: 'title',
      cell: ({ getValue }) => (
        <Text size="regular">{getValue() as string}</Text>
      )
    },
    {
      header: 'Expiry Date',
      accessorKey: 'expiresAt',
      enableSorting: false,
      cell: ({ row }) => {
        const date = timestampToDayjs(row.original.expiresAt);
        return date ? <Text size="regular">{date.format(dateFormat)}</Text> : null;
      }
    },
    {
      header: 'Last used',
      accessorKey: 'usedAt',
      enableSorting: false,
      cell: ({ row }) => {
        const pat = row.original;
        if (!pat.usedAt || isNullTimestamp(pat.usedAt)) return null;
        const date = timestampToDayjs(pat.usedAt);
        return date ? <Text size="regular">{date.fromNow()}</Text> : null;
      }
    },
    {
      header: '',
      accessorKey: 'id',
      enableSorting: false,
      styles: {
        cell: { width: '73px' }
      },
      cell: ({ row }) => (
        <Button
          variant="text"
          color="neutral"
          size="small"
          className={styles.revokeButton}
          onClick={e => {
            e.stopPropagation();
            onRevoke(row.original.id);
          }}
          data-test-id="frontier-sdk-revoke-pat-btn"
        >
          Revoke
        </Button>
      )
    }
  ];
}
