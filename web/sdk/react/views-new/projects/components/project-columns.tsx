'use client';

import { DotsVerticalIcon, Pencil1Icon, TrashIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Menu,
  IconButton,
  DataTableColumnDef
} from '@raystack/apsara-v1';
import type { Project } from '@raystack/proton/frontier';
import { MembersCell } from './members-cell';

export interface ProjectMenuPayload {
  projectId: string;
  title: string;
  canUpdate: boolean;
  canDelete: boolean;
}

type MenuHandle = ReturnType<typeof Menu.createHandle>;

export const getColumns = ({
  userAccessOnProject,
  menuHandle
}: {
  userAccessOnProject: Record<string, string[]>;
  menuHandle: MenuHandle;
}): DataTableColumnDef<Project, unknown>[] => [
    {
      header: 'Name',
      accessorKey: 'title',
      maxSize: 400,
      minSize: 200,
      cell: ({ getValue }) => {
        return (
          <Text size="regular">
            {getValue() as string}
          </Text>
        );
      }
    },
    {
      header: 'Privacy',
      accessorKey: 'metadata',
      enableSorting: false,

      cell: () => {
        return (
          <Text size="regular">
            Private
          </Text>
        );
      }
    },
    {
      header: 'Members',
      accessorKey: 'membersCount',
      enableSorting: false,
      cell: ({ row }) => {
        const project = row.original as Project;
        return <MembersCell projectId={project.id || ''} />;
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
        const project = row.original as Project;
        const access = userAccessOnProject[project.id!] ?? [];
        const canUpdate = access.includes('update');
        const canDelete = access.includes('delete');

        if (!canUpdate && !canDelete) return null;

        return (
          <Flex align="center" justify="center">
            <Menu.Trigger
              handle={menuHandle}
              payload={{
                projectId: project.id || '',
                title: project.title || '',
                canUpdate,
                canDelete
              }}
              render={
                <IconButton
                  size={3}
                  aria-label="Project actions"
                  data-test-id="project-actions-btn"
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
