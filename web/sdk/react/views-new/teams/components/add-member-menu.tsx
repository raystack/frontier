'use client';

import { useCallback, useEffect, useMemo } from 'react';
import {
  Avatar,
  Button,
  Flex,
  Menu,
  Skeleton,
  Tooltip
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListOrganizationUsersRequestSchema,
  SetGroupMemberRoleRequestSchema,
  type Role as ProtoRole,
  type User
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '../../../contexts/FrontierContext';
import { AuthTooltipMessage } from '../../../utils';
import { PERMISSIONS, filterUsersfromUsers, getInitials } from '../../../../utils';
import styles from '../team-details-view.module.css';

interface AddMemberMenuProps {
  teamId: string;
  canUpdateGroup: boolean;
  members: User[];
  roles: ProtoRole[];
  refetch: () => void;
}

export function AddMemberMenu({
  teamId,
  canUpdateGroup,
  members,
  roles,
  refetch
}: AddMemberMenuProps) {
  const { activeOrganization: organization } = useFrontier();

  const {
    data: orgUsersData,
    isLoading: isOrgUsersLoading,
    error: orgUsersError
  } = useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    create(ListOrganizationUsersRequestSchema, {
      id: organization?.id || ''
    }),
    { enabled: !!organization?.id && canUpdateGroup }
  );

  const orgUsers = useMemo(
    () => orgUsersData?.users ?? [],
    [orgUsersData]
  );

  useEffect(() => {
    if (orgUsersError) {
      toastManager.add({
        title: 'Something went wrong',
        description: orgUsersError.message,
        type: 'error'
      });
    }
  }, [orgUsersError]);

  const invitableUsers = useMemo(
    () => filterUsersfromUsers(orgUsers, members),
    [orgUsers, members]
  );

  const memberRoleId = useMemo(
    () => roles.find(r => r.name === PERMISSIONS.RoleGroupMember)?.id ?? '',
    [roles]
  );

  const { mutate: setGroupMemberRole } = useMutation(
    FrontierServiceQueries.setGroupMemberRole,
    {
      onSuccess: () => {
        toastManager.add({
          title: 'Member added',
          type: 'success'
        });
        refetch();
      },
      onError: (err: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: err.message,
          type: 'error'
        });
      }
    }
  );

  const addMember = useCallback(
    (userId: string) => {
      if (!userId || !organization?.id || !teamId || !memberRoleId) return;
      setGroupMemberRole(
        create(SetGroupMemberRoleRequestSchema, {
          groupId: teamId,
          orgId: organization.id,
          principalId: userId,
          principalType: PERMISSIONS.UserPrincipal,
          roleId: memberRoleId
        })
      );
    },
    [setGroupMemberRole, organization?.id, teamId, memberRoleId]
  );

  if (!canUpdateGroup) {
    return (
      <Tooltip>
        <Tooltip.Trigger render={<span />}>
          <Button
            variant="solid"
            color="accent"
            disabled
            data-test-id="frontier-sdk-add-team-member-btn"
          >
            Add a member
          </Button>
        </Tooltip.Trigger>
        <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
      </Tooltip>
    );
  }

  return (
    <Menu autocomplete>
      <Menu.Trigger
        render={
          <Button
            variant="solid"
            color="accent"
            data-test-id="frontier-sdk-add-team-member-btn"
          />
        }
      >
        Add a member
      </Menu.Trigger>
      <Menu.Content
        align="end"
        className={styles.addMemberContent}
        searchPlaceholder="Search..."
      >
        <div className={styles.addMemberMenuList}>
          {isOrgUsersLoading ? (
            <Flex gap={1} direction="column">
              {Array.from({ length: 6 }, (_, i) => (
                <Skeleton key={i} height="30px" width="100%" />
              ))}
            </Flex>
          ) : (
            invitableUsers.map(user => (
              <Menu.Item
                key={user.id}
                value={user.title || user.email || ''}
                className={styles.addMemberMenuItem}
                leadingIcon={
                  <Avatar
                    src={user.avatar}
                    fallback={getInitials(
                      user.title || user.email || ''
                    )}
                    size={1}
                  />
                }
                onClick={() => addMember(user.id || '')}
                data-test-id={`frontier-sdk-add-user-to-team-item-${user.id}`}
              >
                <span className={styles.addMemberMenuItemText}>
                  {user.title || user.email}
                </span>
              </Menu.Item>
            ))
          )}

          {!isOrgUsersLoading && !invitableUsers.length && (
            <Menu.EmptyState className={styles.addMemberEmptyState}>
              No users found
            </Menu.EmptyState>
          )}
        </div>
      </Menu.Content>
    </Menu>
  );
}
