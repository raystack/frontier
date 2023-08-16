import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group } from '~/src';

export const columns: ColumnDef<V1Beta1Group, any>[] = [
  {
    header: 'Title',
    accessorKey: 'name',
    cell: ({ row, getValue }) => (
      <Link
        to={`/teams/${row.original.id}`}
        style={{ textDecoration: 'none', color: 'var(--foreground-base)' }}
      >
        {getValue()}
      </Link>
    )
  },
  {
    accessorKey: 'members',
    cell: ({ row, getValue }) => (
      <TeamMembers teamId={row.original.id} orgId={row.original.org_id} />
    )
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
      <TeamActions team={row.original as V1Beta1Group} />
    )
  }
];

const TeamActions = ({ team }: { team: V1Beta1Group }) => {
  const { client, user } = useFrontier();
  return (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
          <DropdownMenu.Item style={{ padding: 0 }}>
            <Link
              to={`/teams/${team.id}`}
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
              to={`/teams/${team.id}/delete`}
              style={{
                gap: 'var(--pd-8)',
                display: 'flex',
                alignItems: 'center',
                textDecoration: 'none',
                color: 'var(--foreground-base)',
                padding: 'var(--pd-8)'
              }}
            >
              <TrashIcon /> Delete team
            </Link>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
};

interface TeamMembersProps {
  teamId?: string;
  orgId?: string;
}
const TeamMembers = ({ teamId, orgId }: TeamMembersProps) => {
  const { client, user } = useFrontier();
  const [members, setMembers] = useState([]);

  useEffect(() => {
    async function getTeamMembers() {
      if (!teamId) return;
      if (!orgId) return;

      const {
        // @ts-ignore
        data: { users }
      } = await client?.frontierServiceListGroupUsers(orgId, teamId);
      setMembers(users);
    }
    getTeamMembers();
  }, [client, orgId, teamId]);

  return <Text>{members.length} members</Text>;
};
