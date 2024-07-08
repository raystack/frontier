import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { ApsaraColumnDef, DropdownMenu, Text } from '@raystack/apsara';
import { Link } from '@tanstack/react-router';
import { V1Beta1Group } from '~/src';
import styles from '../organization.module.css';

export const getColumns: (
  userAccessOnTeam: Record<string, string[]>
) => ApsaraColumnDef<V1Beta1Group>[] = userAccessOnTeam => [
  {
    header: 'Title',
    accessorKey: 'title',
    cell: ({ row, getValue }) => (
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
    cell: ({ row, getValue }) => <Text>{getValue()} members</Text>
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
