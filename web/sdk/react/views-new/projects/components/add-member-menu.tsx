'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { CardStackPlusIcon, PlusIcon } from '@radix-ui/react-icons';
import {
  Avatar,
  Button,
  Flex,
  Menu,
  Separator,
  Skeleton,
  Tooltip
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListOrganizationUsersRequestSchema,
  CreatePolicyForProjectRequestSchema,
  type User
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useOrganizationTeams } from '../../../hooks/useOrganizationTeams';
import { AuthTooltipMessage } from '../../../utils';
import {
  PERMISSIONS,
  filterUsersfromUsers,
  getInitials
} from '../../../../utils';
import styles from '../project-details-view.module.css';

interface AddMemberMenuProps {
  projectId: string;
  canUpdateProject: boolean;
  members: User[];
  refetch: () => void;
}

export function AddMemberMenu({
  projectId,
  canUpdateProject,
  members,
  refetch
}: AddMemberMenuProps) {
  const [showTeam, setShowTeam] = useState(false);

  const { activeOrganization: organization } = useFrontier();
  const { isFetching: isTeamsLoading, teams } = useOrganizationTeams({});

  const {
    data: orgUsersData,
    isLoading: isOrgUsersLoading,
    error: orgUsersError
  } = useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    create(ListOrganizationUsersRequestSchema, {
      id: organization?.id || ''
    }),
    { enabled: !!organization?.id && canUpdateProject }
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

  const { mutate: createPolicyForProject } = useMutation(
    FrontierServiceQueries.createPolicyForProject,
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
      if (!userId || !organization?.id || !projectId) return;
      const principal = `${PERMISSIONS.UserNamespace}:${userId}`;
      createPolicyForProject(
        create(CreatePolicyForProjectRequestSchema, {
          projectId,
          body: { roleId: PERMISSIONS.RoleProjectViewer, principal }
        })
      );
    },
    [createPolicyForProject, organization?.id, projectId]
  );

  const addTeam = useCallback(
    (teamId: string) => {
      if (!teamId || !organization?.id || !projectId) return;
      const principal = `${PERMISSIONS.GroupNamespace}:${teamId}`;
      createPolicyForProject(
        create(CreatePolicyForProjectRequestSchema, {
          projectId,
          body: { roleId: PERMISSIONS.RoleProjectViewer, principal }
        })
      );
    },
    [createPolicyForProject, organization?.id, projectId]
  );

  const toggleShowTeam = useCallback(() => {
    setShowTeam(prev => !prev);
  }, []);

  const isLoading = showTeam ? isTeamsLoading : isOrgUsersLoading;

  if (!canUpdateProject) {
    return (
      <Tooltip>
        <Tooltip.Trigger render={<span />}>
          <Button
            variant="solid"
            color="accent"
            disabled
            data-test-id="frontier-sdk-add-project-member-btn"
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
            data-test-id="frontier-sdk-add-project-member-btn"
          />
        }
      >
        Add a member
      </Menu.Trigger>
      <Menu.Content
        align="end"
        className={styles.addMemberContent}
        searchPlaceholder={showTeam ? 'Search teams...' : 'Search...'}
      >
        <div className={styles.addMemberMenuList}>
          {isLoading
            ? <Flex gap={1} direction="column">
              {
                Array.from({ length: 6 }, (_, i) => (
                  <Skeleton key={i} height="30px" width="100%" />
                ))
              }
            </Flex>
            : showTeam
              ? teams.map(team => (
                <Menu.Item
                  key={team.id}
                  value={team.title || team.name || ''}
                  className={styles.addMemberMenuItem}
                  leadingIcon={
                    <Avatar
                      fallback={getInitials(team.title || team.name || '')}
                      size={1}
                    />
                  }
                  onClick={() => addTeam(team.id || '')}
                  data-test-id={`frontier-sdk-add-team-to-project-item-${team.id}`}
                >
                  <span className={styles.addMemberMenuItemText}>{team.title || team.name}</span>
                </Menu.Item>
              ))
              : invitableUsers.map(user => (
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
                  data-test-id={`frontier-sdk-add-user-to-project-item-${user.id}`}
                >
                  <span className={styles.addMemberMenuItemText}>{user.title || user.email}</span>
                </Menu.Item>
              ))}

          {!isLoading &&
            (showTeam ? !teams.length : !invitableUsers.length) && (
              <Menu.EmptyState className={styles.addMemberEmptyState}>
                {showTeam ? 'No teams found' : 'No users found'}
              </Menu.EmptyState>
            )}
        </div>
        <Flex direction="column" align="center" className={styles.addMemberFooter}>
          <Separator className={styles.addMemberSeparator} />
          <Button
            width="100%"
            variant="text"
            color="neutral"
            className={styles.addMemberToggleBtn}
            leadingIcon={
              showTeam ? (
                <PlusIcon />
              ) : (
                <CardStackPlusIcon />
              )
            }
            onClick={toggleShowTeam}
            data-test-id="frontier-sdk-add-project-member-toggle"
          >
            {showTeam ? 'Add project member' : 'Add team to project'}
          </Button>
        </Flex>
      </Menu.Content>
    </Menu>
  );
}
