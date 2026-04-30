'use client';

import { DotsVerticalIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Avatar,
  Menu,
  IconButton,
  DataTableColumnDef,
  getAvatarColor
} from '@raystack/apsara-v1';
import type { User, Role } from '@raystack/proton/frontier';
import { getInitials } from '~/utils';

export interface MemberMenuPayload {
  memberId: string;
  excludedRoles: Role[];
}

type MenuHandle = ReturnType<typeof Menu.createHandle<MemberMenuPayload>>;

interface GetColumnsOptions {
  memberRoles: Record<string, Role[]>;
  roles: Role[];
  canUpdateGroup: boolean;
  menuHandle: MenuHandle;
}

export function getColumns({
  memberRoles,
  roles,
  canUpdateGroup,
  menuHandle
}: GetColumnsOptions): DataTableColumnDef<User, unknown>[] {
  return [
    {
      header: 'Name',
      accessorKey: 'title',
      cell: ({ row }) => {
        const member = row.original;
        const fallback = getInitials(member.title || member.email);
        const color = getAvatarColor(member.id || '');
        return (
          <Flex align="center" gap={3}>
            <Avatar
              src={member.avatar}
              fallback={fallback}
              size={5}
              radius="full"
              color={color}
            />
            <Flex direction="column" gap={1}>
              <Text size="regular" weight="medium">
                {member.title}
              </Text>
              <Text size="small" variant="secondary">
                {member.email}
              </Text>
            </Flex>
          </Flex>
        );
      }
    },
    {
      header: 'Role',
      accessorKey: 'email',
      styles: {
        cell: { maxWidth: '300px' },
        header: { maxWidth: '300px' }
      },
      cell: ({ row }) => {
        const member = row.original;
        const roleList =
          (member.id && memberRoles[member.id]) || [];
        const roleText =
          roleList.map((r: Role) => r.title || r.name).join(', ') ||
          'Inherited role';
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
        cell: { width: '32px' },
        header: { width: '32px' }
      },
      cell: ({ row }) => {
        if (!canUpdateGroup) return null;

        const member = row.original;
        const currentRoles =
          (member.id && memberRoles[member.id]) || [];
        const currentRoleIds = new Set(currentRoles.map(r => r.id));
        const excludedRoles = roles.filter(r => !currentRoleIds.has(r.id));

        return (
          <Flex align="center" justify="center">
            <Menu.Trigger
              handle={menuHandle}
              payload={{
                memberId: member.id || '',
                excludedRoles
              }}
              render={
                <IconButton
                  size={3}
                  aria-label="Member actions"
                  data-test-id="frontier-sdk-team-member-actions-btn"
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
