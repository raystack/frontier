import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { Text, DropdownMenu, DataTableColumnDef } from '@raystack/apsara';
import { Link } from '@tanstack/react-router';
import type { V1Beta1Group } from '~/src';
import styles from '../organization.module.css';

export const getColumns: (
  userAccessOnTeam: Record<string, string[]>
) => DataTableColumnDef<V1Beta1Group, unknown>[] = userAccessOnTeam => [
  {
    header: 'Title',
    accessorKey: 'title',
    cell: ({ row, getValue }) => (
      <Link
        to={'/teams/$teamId'}
        params={{
          teamId: row.original.id || ''
        }}
        style={{
          textDecoration: 'none',
          color: 'var(--rs-color-foreground-base-primary)',
          fontSize: 'var(--rs-font-size-small)'
        }}
      >
        {getValue() as string}
      </Link>
    )
  },
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
    enableSorting: false,
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
    <DropdownMenu placement="bottom-end">
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      {/* @ts-ignore */}
      <DropdownMenu.Content portal={false}>
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
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
