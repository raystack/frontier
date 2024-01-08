import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Text } from '@raystack/apsara';
import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import Skeleton from 'react-loading-skeleton';
import { V1Beta1Project } from '~/src';

export const getColumns: (
  userAccessOnProject: Record<string, string[]>,
  isLoading?: boolean
) => ColumnDef<V1Beta1Project, any>[] = (userAccessOnProject, isLoading) => [
  {
    header: 'Title',
    accessorKey: 'title',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          return (
            <Link
              to={`/projects/$projectId`}
              params={{
                projectId: row.original.id || ''
              }}
              style={{
                textDecoration: 'none',
                color: 'var(--foreground-base)'
              }}
            >
              {getValue()}
            </Link>
          );
        }
  },
  // {
  //   header: 'Privacy',
  //   accessorKey: 'privacy',
  //   cell: isLoading
  //     ? () => <Skeleton />
  //     : info => <Text>{info.getValue() ?? 'Public'}</Text>
  // },
  {
    header: 'Members',
    accessorKey: 'members_count',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          return <Text>{getValue()} members</Text>;
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => (
          <ProjectActions
            project={row.original as V1Beta1Project}
            userAccessOnProject={userAccessOnProject}
          />
        )
  }
];

const ProjectActions = ({
  project,
  userAccessOnProject
}: {
  project: V1Beta1Project;
  userAccessOnProject: Record<string, string[]>;
}) => {
  const canUpdateProject = (userAccessOnProject[project.id!] ?? []).includes(
    'update'
  );
  const canDeleteProject = (userAccessOnProject[project.id!] ?? []).includes(
    'delete'
  );
  const canDoActions = canUpdateProject || canDeleteProject;

  return canDoActions ? (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
          {canUpdateProject ? (
            <DropdownMenu.Item style={{ padding: 0 }}>
              <Link
                to={`/projects/$projectId`}
                params={{
                  projectId: project.id || ''
                }}
                style={{
                  gap: 'var(--pd-8)',
                  display: 'flex',
                  alignItems: 'center',
                  textDecoration: 'none',
                  color: 'var(--foreground-base)',
                  padding: 'var(--pd-8)'
                }}
              >
                <Pencil1Icon /> Rename
              </Link>
            </DropdownMenu.Item>
          ) : null}
          {canDeleteProject ? (
            <DropdownMenu.Item style={{ padding: 0 }}>
              <Link
                to={`/projects/$projectId/delete`}
                params={{
                  projectId: project.id || ''
                }}
                style={{
                  gap: 'var(--pd-8)',
                  display: 'flex',
                  alignItems: 'center',
                  textDecoration: 'none',
                  color: 'var(--foreground-base)',
                  padding: 'var(--pd-8)'
                }}
              >
                <TrashIcon /> Delete project
              </Link>
            </DropdownMenu.Item>
          ) : null}
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
