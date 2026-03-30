'use client';

import { DotsVerticalIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Avatar,
  getAvatarColor,
  Menu,
  IconButton,
  DataTableColumnDef
} from '@raystack/apsara-v1';
import type { Role } from '@raystack/proton/frontier';
import type { Menu as BaseMenu } from '@base-ui/react';
import type { MemberWithInvite } from '~/react/hooks/useOrganizationMembers';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import styles from './member-columns.module.css';

export interface MemberMenuPayload {
  memberId: string;
  invited: boolean;
  excludedRoles: Role[];
  canUpdateRole: boolean;
  canRemove: boolean;
}

export const getColumns = ({
  memberRoles,
  roles,
  canDeleteUser,
  menuHandle
}: {
  memberRoles: Record<string, Role[]>;
  roles: Role[];
  canDeleteUser: boolean;
  menuHandle: BaseMenu.Handle<MemberMenuPayload>;
}): DataTableColumnDef<MemberWithInvite, MemberWithInvite>[] => [
  {
    header: '',
    accessorKey: 'avatar',
    enableSorting: false,
    styles: {
      cell: { width: 'var(--rs-space-5)' }
    },
    cell: ({ row, getValue }) => {
      const id = row.original?.id || '';
      const fallback =
        row.original?.title || row.original?.email || row.original?.userId;
      return (
        <Avatar
          src={getValue() as string}
          fallback={getInitials(fallback)}
          color={getAvatarColor(id)}
          size={5}
          radius="small"
        />
      );
    }
  },
  {
    header: 'Title',
    accessorKey: 'title',
    cell: ({ row, getValue }) => {
      const title = getValue() as string;
      const email = row.original.email || row.original.userId;
      return (
        <Flex direction="column" gap={1}>
          {title && (
            <Text size="regular" weight="medium">
              {title}
            </Text>
          )}
          <Text size="small" variant="secondary">
            {email}
          </Text>
        </Flex>
      );
    }
  },
  {
    header: 'Role',
    accessorKey: 'email',
    cell: ({ row }) => {
      const member = row.original;
      let roleDisplay: string;

      if (member.invited) {
        const inviteRoleIds = (member as { roleIds?: string[] }).roleIds;
        if (inviteRoleIds?.length) {
          roleDisplay =
            inviteRoleIds
              .map(id => roles.find(r => r.id === id))
              .filter(Boolean)
              .map(r => r?.title || r?.name)
              .join(', ') || 'Member';
        } else {
          roleDisplay = 'Member';
        }
      } else if (member.id && memberRoles[member.id]) {
        roleDisplay = memberRoles[member.id]
          .map((r: Role) => r.title || r.name)
          .join(', ');
      } else {
        roleDisplay = 'Inherited role';
      }

      return (
        <Text size="regular" variant="secondary">
          {roleDisplay}
          {member.invited && (
            <Text as="span" size="regular" className={styles.pendingText}>
              {' '}
              (Pending invite)
            </Text>
          )}
        </Text>
      );
    }
  },
  {
    header: '',
    accessorKey: 'id',
    enableSorting: false,
    styles: {
      cell: { width: '48px' }
    },
    cell: ({ row }) => {
      const member = row.original;
      const memberId = member.id || '';
      const userRoles = memberId ? memberRoles[memberId] : [];
      const excludedRoles = differenceWith<Role>(
        isEqualById,
        roles,
        userRoles || []
      );
      const canUpdateRole = canDeleteUser && !member.invited;
      const canRemove = canDeleteUser;

      if (!canUpdateRole && !canRemove) return null;

      return (
        <Flex align="center" justify="center" className={styles.actionsCell}>
          <Menu.Trigger
            handle={menuHandle}
            payload={{
              memberId,
              invited: !!member.invited,
              excludedRoles,
              canUpdateRole,
              canRemove
            }}
            render={
              <IconButton
                size={3}
                aria-label="Member actions"
                data-test-id="member-actions-btn"
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
