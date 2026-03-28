'use client';

import { useState } from 'react';
import { DotsVerticalIcon, TrashIcon, UpdateIcon } from '@radix-ui/react-icons';
import {
  Flex,
  Text,
  Avatar,
  getAvatarColor,
  Popover,
  IconButton
} from '@raystack/apsara-v1';
import type { Role } from '@raystack/proton/frontier';
import type { MemberWithInvite } from '~/react/hooks/useOrganizationMembers';
import { getInitials } from '~/utils';
import styles from './member-list-item.module.css';

export interface MemberListItemProps {
  member: MemberWithInvite;
  roleDisplay: string;
  excludedRoles: Role[];
  canUpdateRole: boolean;
  canRemove: boolean;
  onUpdateRole: (role: Role) => void;
  onRemove: () => void;
}

export function MemberListItem({
  member,
  roleDisplay,
  excludedRoles,
  canUpdateRole,
  canRemove,
  onUpdateRole,
  onRemove
}: MemberListItemProps) {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const id = member.id || '';
  const fallback = member.title || member.email || member.userId;
  const name = member.title;
  const email = member.email || member.userId;
  const showActions = canUpdateRole || canRemove;

  return (
    <Flex align="center" className={styles.row}>
      <Flex align="center" gap={4} className={styles.infoColumn}>
        <Avatar
          src={member.avatar as string}
          fallback={getInitials(fallback)}
          color={getAvatarColor(id)}
          size={5}
          radius="small"
        />
        <Flex direction="column" gap={1} className={styles.nameWrapper}>
          {name && (
            <Text size="regular" weight="medium" className={styles.truncate}>
              {name}
            </Text>
          )}
          <Text size="small" variant="secondary" className={styles.truncate}>
            {email}
          </Text>
        </Flex>
      </Flex>

      <Flex align="center" className={styles.roleColumn}>
        <Text size="regular" variant="secondary">
          {roleDisplay}
          {member.invited && (
            <Text as="span" size="regular" className={styles.pendingText}>
              {' '}
              (Pending invite)
            </Text>
          )}
        </Text>
      </Flex>

      <Flex align="center" justify="center" className={styles.actionsColumn} data-open={isMenuOpen || undefined}>
        {showActions && (
          <Popover open={isMenuOpen} onOpenChange={setIsMenuOpen}>
            <Popover.Trigger
              render={
                <IconButton size={3} aria-label="Member actions" />
              }
            >
              <DotsVerticalIcon />
            </Popover.Trigger>
            <Popover.Content
              side="bottom"
              align="end"
              className={styles.menuContent}
            >
              {canUpdateRole &&
                excludedRoles.map(role => (
                  <button
                    key={role.id}
                    className={styles.menuItem}
                    onClick={() => {
                      setIsMenuOpen(false);
                      onUpdateRole(role);
                    }}
                    data-test-id={`update-role-${role.name}-dropdown-item`}
                  >
                    <UpdateIcon />
                    Make {role.title}
                  </button>
                ))}
              {canRemove && (
                <button
                  className={styles.menuItem}
                  onClick={() => {
                    setIsMenuOpen(false);
                    onRemove();
                  }}
                  data-test-id="remove-member-dropdown-item"
                >
                  <TrashIcon />
                  Remove
                </button>
              )}
            </Popover.Content>
          </Popover>
        )}
      </Flex>
    </Flex>
  );
}
