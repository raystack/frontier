'use client';

import { DotsVerticalIcon, TrashIcon, UpdateIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Avatar,
  Menu,
  IconButton,
  DataTableColumnDef,
  getAvatarColor
} from '@raystack/apsara-v1';
import type { User, Group, Role } from '@raystack/proton/frontier';
import { getInitials } from '~/utils';
import teamIcon from '~/react/assets/users.svg';

export type MemberRow = (Group & { isTeam: true }) | (User & { isTeam?: false });

export interface MemberMenuPayload {
  memberId: string;
  isTeam: boolean;
  excludedRoles: Role[];
}

type MenuHandle = ReturnType<typeof Menu.createHandle<MemberMenuPayload>>;

interface GetColumnsOptions {
  memberRoles: Record<string, Role[]>;
  groupRoles: Record<string, Role[]>;
  roles: Role[];
  canUpdateProject: boolean;
  menuHandle: MenuHandle;
}

export function getColumns({
  memberRoles,
  groupRoles,
  roles,
  canUpdateProject,
  menuHandle
}: GetColumnsOptions): DataTableColumnDef<MemberRow, unknown>[] {
  return [
    {
      header: 'Name',
      accessorKey: 'title',
      cell: ({ row }) => {
        const member = row.original;
        const fallback = member.isTeam
          ? getInitials(member.title || member.name)
          : getInitials(member.title || member.email);
        const color = getAvatarColor(member.id || '');
        const label = member.isTeam ? member.title : member.title;
        const subLabel = member.isTeam ? member.name : member.email;
        return (
          <Flex align="center" gap={3}>
            <Avatar
              src={member.isTeam ? teamIcon as unknown as string : member.avatar}
              fallback={fallback}
              size={5}
              radius="small"
              color={color}
            />
            <Flex direction="column" gap={1}>
              <Text size="regular" weight="medium">
                {label}
              </Text>
              <Text size="small" variant="secondary">
                {subLabel}
              </Text>
            </Flex>
          </Flex>
        );
      }
    },
    {
      header: 'Role',
      accessorKey: 'email',
      cell: ({ row }) => {
        const member = row.original;
        const roleList = member.isTeam
          ? (member.id && groupRoles[member.id]) || []
          : (member.id && memberRoles[member.id]) || [];
        const roleText =
          roleList.map((r: Role) => r.title || r.name).join(', ') ||
          (member.isTeam ? 'Project Viewer' : 'Inherited role');
        return (
          <Text size="regular" variant="secondary">
            {roleText}
          </Text>
        );
      }
    },
    {
      header: '',
      accessorKey: 'id',
      enableSorting: false,
      styles: {
        cell: { width: '32px' }
      },
      cell: ({ row }) => {
        if (!canUpdateProject) return null;

        const member = row.original;
        const currentRoles = member.isTeam
          ? (member.id && groupRoles[member.id]) || []
          : (member.id && memberRoles[member.id]) || [];
        const currentRoleIds = new Set(currentRoles.map(r => r.id));
        const excludedRoles = roles.filter(r => !currentRoleIds.has(r.id));

        return (
          <Flex align="center" justify="center">
            <Menu.Trigger
              handle={menuHandle}
              payload={{
                memberId: member.id || '',
                isTeam: !!member.isTeam,
                excludedRoles
              }}
              render={
                <IconButton
                  size={3}
                  aria-label="Member actions"
                  data-test-id="frontier-sdk-member-actions-btn"
                />
              }
            >
              <DotsVerticalIcon />
            </Menu.Trigger>
          </Flex>
        );
      }
    }
  ];
}
