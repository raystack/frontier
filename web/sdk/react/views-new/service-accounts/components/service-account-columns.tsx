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
import { timestampToDayjs } from '../../../../utils/timestamp';
import { ProjectsCell } from './projects-cell';

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
  orgId: string;
}

export const getColumns = ({
  dateFormat,
  menuHandle,
  canUpdateWorkspace,
  orgId
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
    header: 'Projects',
    id: 'projects',
    accessorKey: 'id',
    enableSorting: false,
    cell: ({ getValue }) => {
      const serviceUserId = getValue() as string;
      return <ProjectsCell serviceUserId={serviceUserId} orgId={orgId} />;
    }
  },
  {
    header: 'Created On',
    accessorKey: 'createdAt',
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
      cell: { width: '48px' }
    },
    cell: ({ getValue }) => {
      const serviceAccountId = getValue() as string;

      if (!canUpdateWorkspace) return null;

      return (
        <Flex align="center" justify="center">
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
