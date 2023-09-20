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
import { V1Beta1Group } from '~/src';

export const getColumns: (
  userAccessOnTeam: Record<string, string[]>,
  isLoading?: boolean
) => ColumnDef<V1Beta1Group, any>[] = (userAccessOnTeam, isLoading) => [
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
          <TeamActions
            team={row.original as V1Beta1Group}
            userAccessOnTeam={userAccessOnTeam}
          />
        )
  }
];

const TeamActions = ({
  team,
  userAccessOnTeam
}: {
  team: V1Beta1Group;
  userAccessOnTeam: Record<string, string[]>;
}) => {
  const canUpdateTeam = (userAccessOnTeam[team.id!] ?? []).includes('update');
  const canDeleteTeam = (userAccessOnTeam[team.id!] ?? []).includes('delete');
  const canDoActions = canUpdateTeam || canDeleteTeam;

  return canDoActions ? (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
          {canUpdateTeam ? (
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
          ) : null}
          {canDeleteTeam ? (
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
          ) : null}
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
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
