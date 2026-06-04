'use client';

import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Menu,
  IconButton,
  DataTableColumnDef
} from '@raystack/apsara';
import type { Project } from '@raystack/proton/frontier';
import { MembersCell } from './members-cell';
import styles from './project-columns.module.css';

export interface ProjectMenuPayload {
  projectId: string;
  name: string;
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
      header: 'Members',
      accessorKey: 'membersCount',
      enableSorting: false,
      styles: {
        cell: { maxWidth: '300px' },
        header: { maxWidth: '300px' }
      },
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
        cell: { width: '48px' },
        header: { width: '48px' }
      },
      cell: ({ row }) => {
        const project = row.original as Project;
        const access = userAccessOnProject[project.id!] ?? [];
        const canUpdate = access.includes('update');
        const canDelete = access.includes('delete');

        if (!canUpdate && !canDelete) return null;

        return (
          <Flex align="center" justify="center" className={styles.actionsCell}>
            <Menu.Trigger
              handle={menuHandle}
              payload={{
                projectId: project.id || '',
                name: project.name || '',
                title: project.title || '',
                canUpdate,
                canDelete
              }}
              render={
                <IconButton
                  size={2}
                  aria-label="Project actions"
                  data-test-id="project-actions-btn"
                />
              }
            >
              <DotsHorizontalIcon />
            </Menu.Trigger>
          </Flex>
        );
      }
    }
  ];
