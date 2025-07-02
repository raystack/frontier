import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { Text, DropdownMenu } from '@raystack/apsara/v1';
import { Link } from '@tanstack/react-router';
import type { V1Beta1Project } from '~/src';
import type { DataTableColumnDef } from '@raystack/apsara/v1';

export const getColumns: (
  userAccessOnProject: Record<string, string[]>
) => DataTableColumnDef<V1Beta1Project, unknown>[] = userAccessOnProject => [
  {
    header: 'Title',
    accessorKey: 'title',
    cell: ({ row, getValue }) => {
      return (
        <Link
          to={`/projects/$projectId`}
          params={{
            projectId: row.original.id || ''
          }}
          style={{
            textDecoration: 'none',
            color: 'var(--rs-color-foreground-base-primary)',
            fontSize: 'var(--rs-font-size-small)'
          }}
        >
          {getValue() as string}
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
    <DropdownMenu placement="bottom-end">
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      {/* @ts-ignore */}
      <DropdownMenu.Content portal={false}>
        <DropdownMenu.Group>
          {canUpdateProject ? (
            <DropdownMenu.Item>
              <Link
                to={`/projects/$projectId`}
                params={{
                  projectId: project.id || ''
                }}
                style={{
                  gap: 'var(--rs-space-3)',
                  display: 'flex',
                  alignItems: 'center',
                  textDecoration: 'none',
                  color: 'var(--rs-color-foreground-base-primary)',
                  flex: 1
                }}
              >
                <Pencil1Icon /> Rename
              </Link>
            </DropdownMenu.Item>
          ) : null}
          {canDeleteProject ? (
            <DropdownMenu.Item>
              <Link
                to={`/projects/$projectId/delete`}
                params={{
                  projectId: project.id || ''
                }}
                style={{
                  gap: 'var(--rs-space-3)',
                  display: 'flex',
                  alignItems: 'center',
                  textDecoration: 'none',
                  color: 'var(--rs-color-foreground-base-primary)',
                  flex: 1
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
