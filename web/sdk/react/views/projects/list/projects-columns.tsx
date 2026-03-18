'use client';

import {
    DotsHorizontalIcon,
    Pencil1Icon,
    TrashIcon
} from '@radix-ui/react-icons';
import { Text, DropdownMenu } from '@raystack/apsara';
import type { Project } from '@raystack/proton/frontier';
import type { DataTableColumnDef } from '@raystack/apsara';
import orgStyles from '../../../components/organization/organization.module.css';

export const getColumns = (
    userAccessOnProject: Record<string, string[]>,
    onProjectClick?: (projectId: string) => void,
    onDeleteProjectClick?: (projectId: string) => void
): DataTableColumnDef<Project, unknown>[] => [
  {
    header: 'Title',
    accessorKey: 'title',
    cell: ({ row, getValue }) => {
      return <Text>{getValue() as string}</Text>;
    }
  },
  {
    header: 'Members',
    accessorKey: 'membersCount',
    cell: ({ row, getValue }) => {
      const value = getValue() as string;
      return value ? <Text>{value} members</Text> : null;
    }
  },
  {
    header: '',
    accessorKey: 'id',
    meta: {
      style: {
        textAlign: 'end'
      }
    },
    enableSorting: false,
    cell: ({ row, getValue }) => (
      <ProjectActions
        project={row.original as Project}
        userAccessOnProject={userAccessOnProject}
        onProjectClick={onProjectClick}
        onDeleteProjectClick={onDeleteProjectClick}
      />
    )
  }
];

const ProjectActions = ({
    project,
    userAccessOnProject,
    onProjectClick,
    onDeleteProjectClick
}: {
    project: Project;
    userAccessOnProject: Record<string, string[]>;
    onProjectClick?: (projectId: string) => void;
    onDeleteProjectClick?: (projectId: string) => void;
}) => {
  const canUpdateProject = (userAccessOnProject[project.id!] ?? []).includes(
    'update'
  );
  const canDeleteProject = (userAccessOnProject[project.id!] ?? []).includes(
    'delete'
  );
  const canDoActions = canUpdateProject || canDeleteProject;

  function onDeleteClick(e: React.MouseEvent) {
    e.stopPropagation();
    onDeleteProjectClick?.(project.id || '');
  }

  function onRenameClick(e: React.MouseEvent) {
    e.stopPropagation();
    onProjectClick?.(project.id || '');
  }

  return canDoActions ? (
    <DropdownMenu placement="bottom-end">
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      {/* @ts-ignore */}
      <DropdownMenu.Content portal={false}>
        <DropdownMenu.Group>
          {canUpdateProject ? (
            <DropdownMenu.Item
              onClick={onRenameClick}
              className={orgStyles.dropdownActionItem}
              data-test-id="frontier-sdk-project-list-rename-link"
            >
              <Pencil1Icon /> Rename
            </DropdownMenu.Item>
          ) : null}
          {canDeleteProject ? (
            <DropdownMenu.Item
              onClick={onDeleteClick}
              className={orgStyles.dropdownActionItem}
              data-test-id="frontier-sdk-project-list-delete-link"
            >
              <TrashIcon /> Delete project
            </DropdownMenu.Item>
          ) : null}
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
