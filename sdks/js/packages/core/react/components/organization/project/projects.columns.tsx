import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { useEffect, useState } from 'react';
import { Link } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Project } from '~/src';

export const columns: ColumnDef<V1Beta1Project, any>[] = [
  {
    accessorKey: 'name',
    cell: ({ row, getValue }) => {
      return (
        <Link
          to={`/projects/$projectId`}
          params={{
            projectId: row.original.id || ''
          }}
          style={{ textDecoration: 'none', color: 'var(--foreground-base)' }}
        >
          {getValue()}
        </Link>
      );
    }
  },
  {
    accessorKey: 'privacy',
    cell: info => <Text>{info.getValue() ?? 'Public'}</Text>
  },
  {
    accessorKey: 'members',
    cell: ({ row, getValue }) => <ProjectMembers projectId={row.original.id} />
  },
  {
    header: '',
    accessorKey: 'id',
    meta: {
      style: {
        textAlign: 'end'
      }
    },
    cell: ({ row, getValue }) => (
      <ProjectActions project={row.original as V1Beta1Project} />
    )
  }
];

const ProjectActions = ({ project }: { project: V1Beta1Project }) => {
  return (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
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
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
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
