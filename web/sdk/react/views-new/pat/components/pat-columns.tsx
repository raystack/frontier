'use client';

import { Button, Text, DataTableColumnDef } from '@raystack/apsara-v1';
import type { PAT } from '@raystack/proton/frontier';
import { timestampToDayjs, isNullTimestamp } from '~/utils/timestamp';
import styles from '../pat-view.module.css';

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
        const d = timestampToDayjs(row.original.expiresAt);
        return d ? <Text size="regular">{d.format(dateFormat)}</Text> : null;
      }
    },
    {
      header: 'Last used',
      accessorKey: 'lastUsedAt',
      enableSorting: false,
      cell: ({ row }) => {
        const pat = row.original;
        if (!pat.lastUsedAt || isNullTimestamp(pat.lastUsedAt)) return null;
        const d = timestampToDayjs(pat.lastUsedAt);
        return d ? <Text size="regular">{d.fromNow()}</Text> : null;
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
