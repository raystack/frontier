'use client';

import {
    DotsHorizontalIcon,
    Pencil1Icon,
    TrashIcon
} from '@radix-ui/react-icons';
import { Text, DropdownMenu, DataTableColumnDef } from '@raystack/apsara';
import type { Group } from '@raystack/proton/frontier';
import orgStyles from '../../../components/organization/organization.module.css';
import React from 'react';

export const getColumns = (
    userAccessOnTeam: Record<string, string[]>,
    onTeamClick?: (teamId: string) => void,
    onDeleteTeamClick?: (teamId: string) => void
): DataTableColumnDef<Group, unknown>[] => [
  {
    header: 'Title',
    accessorKey: 'title',
    cell: ({ row, getValue }) => <Text>{getValue() as string}</Text>
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
        team={row.original as Group}
        userAccessOnTeam={userAccessOnTeam}
        onTeamClick={onTeamClick}
        onDeleteTeamClick={onDeleteTeamClick}
      />
    )
  }
];

const TeamActions = ({
    team,
    userAccessOnTeam,
    onTeamClick,
    onDeleteTeamClick
}: {
    team: Group;
    userAccessOnTeam: Record<string, string[]>;
    onTeamClick?: (teamId: string) => void;
    onDeleteTeamClick?: (teamId: string) => void;
}) => {
    const canUpdateTeam = (userAccessOnTeam[team.id!] ?? []).includes('update');
    const canDeleteTeam = (userAccessOnTeam[team.id!] ?? []).includes('delete');
    const canDoActions = canUpdateTeam || canDeleteTeam;

  function onDeleteClick(e: React.MouseEvent) {
    e.stopPropagation();
    onDeleteTeamClick?.(team.id || '');
  }

  function onRenameClick(e: React.MouseEvent) {
    e.stopPropagation();
    onTeamClick?.(team.id || '');
  }

  return canDoActions ? (
    <DropdownMenu placement="bottom-end">
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      {/* @ts-ignore */}
      <DropdownMenu.Content portal={false}>
        {canUpdateTeam ? (
          <DropdownMenu.Item
            onClick={onRenameClick}
            className={orgStyles.dropdownActionItem}
            data-test-id="frontier-sdk-teams-list-rename-link"
          >
            <Pencil1Icon /> Rename
          </DropdownMenu.Item>
        ) : null}
        {canDeleteTeam ? (
          <DropdownMenu.Item
            onClick={onDeleteClick}
            className={orgStyles.dropdownActionItem}
            data-test-id="frontier-sdk-teams-list-delete-link"
          >
            <TrashIcon /> Delete team
          </DropdownMenu.Item>
        ) : null}
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
