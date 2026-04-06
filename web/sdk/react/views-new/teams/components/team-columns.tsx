'use client';

import { DotsVerticalIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Menu,
  IconButton,
  DataTableColumnDef
} from '@raystack/apsara-v1';
import type { Group } from '@raystack/proton/frontier';

export interface TeamMenuPayload {
  teamId: string;
  title: string;
  canUpdate: boolean;
  canDelete: boolean;
}

type MenuHandle = ReturnType<typeof Menu.createHandle>;

export function getColumns({
  userAccessOnTeam,
  menuHandle
}: {
  userAccessOnTeam: Record<string, string[]>;
  menuHandle: MenuHandle;
}): DataTableColumnDef<Group, unknown>[] {
  return [
    {
      header: 'Title',
      accessorKey: 'title',
      cell: ({ getValue }) => (
        <Text size="regular">{getValue() as string}</Text>
      )
    },
    {
      header: 'Members',
      accessorKey: 'members_count',
      enableSorting: false,
      cell: ({ getValue }) => {
        const value = getValue() as number;
        return value ? (
          <Text size="regular" variant="secondary">
            {value} {value === 1 ? 'member' : 'members'}
          </Text>
        ) : null;
      }
    },
    {
      header: '',
      accessorKey: 'id',
      enableSorting: false,
      styles: {
        cell: { width: '48px' }
      },
      cell: ({ row }) => {
        const team = row.original as Group;
        const access = userAccessOnTeam[team.id!] ?? [];
        const canUpdate = access.includes('update');
        const canDelete = access.includes('delete');

        if (!canUpdate && !canDelete) return null;

        return (
          <Flex align="center" justify="center">
            <Menu.Trigger
              handle={menuHandle}
              payload={{
                teamId: team.id || '',
                title: team.title || '',
                canUpdate,
                canDelete
              }}
              render={
                <IconButton
                  size={3}
                  aria-label="Team actions"
                  data-test-id="frontier-sdk-team-actions-btn"
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
}
