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
import { V1Beta1Group } from '~/src';
import Skeleton from 'react-loading-skeleton';

export const getColumns: (
  isLoading?: boolean
) => ColumnDef<V1Beta1Group, any>[] = isLoading => [
  {
    header: 'Title',
    accessorKey: 'name',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => (
          <Link
            to={'/teams/$teamId'}
            params={{
              teamId: row.original.id || ''
            }}
            style={{ textDecoration: 'none', color: 'var(--foreground-base)' }}
          >
            {getValue()}
          </Link>
        )
  },
  {
    accessorKey: 'members',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => (
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => (
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
              to={'/teams/$teamId'}
              params={{
                teamId: team.id || ''
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
              to={'/teams/$teamId/delete'}
              params={{
                teamId: team.id || ''
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
