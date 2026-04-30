'use client';

import { DotsVerticalIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Menu,
  IconButton,
  DataTableColumnDef
} from '@raystack/apsara-v1';
import type { ServiceUser } from '@raystack/proton/frontier';
import type { Timestamp } from '@bufbuild/protobuf/wkt';
import { timestampToDayjs } from '~/utils/timestamp';
import styles from './service-account-columns.module.css';

export interface ServiceAccountMenuPayload {
  serviceAccountId: string;
  canManageAccess: boolean;
  canDelete: boolean;
}

type MenuHandle = ReturnType<typeof Menu.createHandle>;

interface GetColumnsOptions {
  dateFormat: string;
  menuHandle: MenuHandle;
  canUpdateWorkspace: boolean;
}

export const getColumns = ({
  dateFormat,
  menuHandle,
  canUpdateWorkspace
}: GetColumnsOptions): DataTableColumnDef<ServiceUser, unknown>[] => [
  {
    header: 'Name',
    accessorKey: 'title',
    cell: ({ getValue }) => {
      const value = getValue() as string;
      return <Text size="regular">{value}</Text>;
    }
  },
  {
    header: 'Created On',
    accessorKey: 'createdAt',
    styles: {
      cell: { maxWidth: '300px' },
      header: { maxWidth: '300px' }
    },
    cell: ({ getValue }) => {
      const value = getValue() as Timestamp | undefined;
      return (
        <Text size="regular">
          {timestampToDayjs(value)?.format(dateFormat) ?? '-'}
        </Text>
      );
    }
  },
  {
    header: '',
    id: 'actions',
    accessorKey: 'id',
    enableSorting: false,
    styles: {
      cell: { width: '48px', minWidth: '48px', maxWidth: '48px' },
      header: { width: '48px', minWidth: '48px', maxWidth: '48px' }
    },
    cell: ({ getValue }) => {
      const serviceAccountId = getValue() as string;

      if (!canUpdateWorkspace) return null;

      return (
        <Flex align="center" justify="center" className={styles.actionsCell}>
          <Menu.Trigger
            handle={menuHandle}
            payload={{
              serviceAccountId,
              canManageAccess: canUpdateWorkspace,
              canDelete: canUpdateWorkspace
            }}
            render={
              <IconButton
                size={3}
                aria-label="Service account actions"
                data-test-id="frontier-sdk-service-account-actions-btn"
                className={styles.actionButton}
              />
            }
          >
            <DotsVerticalIcon />
          </Menu.Trigger>
        </Flex>
      );
    }
  }
];
