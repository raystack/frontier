import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Text } from '@raystack/apsara';
import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import { useEffect, useState } from 'react';
import Skeleton from 'react-loading-skeleton';
import { useFrontier } from '~/react/contexts/FrontierContext';
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
    accessorKey: 'members',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => <ProjectMembers projectId={row.original.id} />
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

interface ProjectMembersProps {
  projectId?: string;
}
const ProjectMembers = ({ projectId }: ProjectMembersProps) => {
  const { client } = useFrontier();
  const [members, setMembers] = useState([]);

  useEffect(() => {
    async function getProjectMembers() {
      if (!projectId) return;

      const {
        // @ts-ignore
        data: { users }
      } = await client?.frontierServiceListProjectUsers(projectId);
      setMembers(users);
    }
    getProjectMembers();
  }, [client, projectId]);

  return <Text>{members.length} members</Text>;
};
