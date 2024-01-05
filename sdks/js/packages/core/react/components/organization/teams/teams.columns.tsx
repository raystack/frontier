import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Text } from '@raystack/apsara';
import { Link } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import Skeleton from 'react-loading-skeleton';
import { V1Beta1Group } from '~/src';
import styles from '../organization.module.css';

export const getColumns: (
  userAccessOnTeam: Record<string, string[]>,
  isLoading?: boolean
) => ColumnDef<V1Beta1Group, any>[] = (userAccessOnTeam, isLoading) => [
  {
    header: 'Title',
    accessorKey: 'title',
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
    header: 'Members',
    accessorKey: 'members_count',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => <Text>{getValue()} members</Text>
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
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
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
                className={styles.dropdownActionItem}
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
                className={styles.dropdownActionItem}
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
