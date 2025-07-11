import {
  CardStackPlusIcon,
  PlusIcon,
  ExclamationTriangleIcon
} from '@radix-ui/react-icons';
import type React from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  EmptyState,
  Tooltip,
  toast,
  Separator,
  Avatar,
  Skeleton,
  Text,
  Flex,
  DataTable,
  Popover,
  Search
} from '@raystack/apsara/v1';
import { useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationTeams } from '~/react/hooks/useOrganizationTeams';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import type {
  V1Beta1CreatePolicyForProjectBody,
  V1Beta1Group,
  V1Beta1Role,
  V1Beta1User
} from '~/src';
import type { Role } from '~/src/types';
import {
  PERMISSIONS,
  filterUsersfromUsers,
  getInitials,
  shouldShowComponent
} from '~/utils';
import { getColumns } from './member.columns';
import styles from './members.module.css';

export type MembersProps = {
  teams?: V1Beta1Group[];
  members?: V1Beta1User[];
  roles?: V1Beta1Role[];
  memberRoles?: Record<string, Role[]>;
  groupRoles?: Record<string, Role[]>;
  isLoading?: boolean;
  refetch: () => void;
};

export const Members = ({
  teams = [],
  members = [],
  roles = [],
  memberRoles,
  groupRoles,
  isLoading: isMemberLoading,
  refetch
}: MembersProps) => {
  const { projectId } = useParams({ from: '/projects/$projectId' });

  const resource = `app/project:${projectId}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!projectId
  );

  const { canUpdateProject } = useMemo(() => {
    return {
      canUpdateProject: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading = isMemberLoading || isPermissionsFetching;

  const columns = useMemo(
    () =>
      getColumns(
        memberRoles,
        groupRoles,
        roles,
        canUpdateProject,
        projectId,
        refetch
      ),
    [memberRoles, groupRoles, roles, canUpdateProject, projectId, refetch]
  );

  const updatedUsers = useMemo(() => {
    const updatedTeams = teams.map(t => ({ ...t, isTeam: true }));
    return members?.length || updatedTeams?.length
      ? [...updatedTeams, ...members]
      : [];
  }, [members, teams]);

  return (
    <Flex direction="column" className={styles.container}>
      <DataTable
        isLoading={isLoading}
        data={updatedUsers}
        columns={columns}
        defaultSort={{ name: 'name', order: 'asc' }}
        mode="client"
      >
        <Flex direction="column" gap={7} className={styles.tableWrapper}>
          <Flex justify="between" gap={3}>
            <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
              <DataTable.Search
                placeholder="Search by name or email"
                size="medium"
              />
            </Flex>
            {isLoading ? (
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Tooltip
                message={AuthTooltipMessage}
                side="left"
                disabled={canUpdateProject}
              >
                <AddMemberDropdown
                  canUpdateProject={canUpdateProject}
                  refetch={refetch}
                  members={members}
                />
              </Tooltip>
            )}
          </Flex>
          <DataTable.Content
            emptyState={noDataChildren}
            classNames={{
              root: styles.tableRoot,
              header: styles.tableHeader
            }}
          />
        </Flex>
      </DataTable>
    </Flex>
  );
};

interface AddMemberDropdownProps {
  canUpdateProject: boolean;
  members?: V1Beta1User[];
  refetch?: () => void;
}

const AddMemberDropdown = ({
  canUpdateProject,
  members,
  refetch
}: AddMemberDropdownProps) => {
  const { projectId } = useParams({ from: '/projects/$projectId' });
  const [orgMembers, setOrgMembers] = useState<V1Beta1User[]>([]);
  const [isOrgMembersLoading, setIsOrgMembersLoading] = useState(false);
  const [query, setQuery] = useState('');
  const [showTeam, setShowTeam] = useState(false);

  const { client, activeOrganization: organization } = useFrontier();
  const { isFetching: isTeamsLoading, teams } = useOrganizationTeams({});

  const toggleShowTeam = (e: React.MouseEvent<HTMLElement>) => {
    e.preventDefault();
    setQuery('');
    setShowTeam(prev => !prev);
  };

  useEffect(() => {
    async function getOrganizationMembers() {
      if (!organization?.id) return;
      try {
        setIsOrgMembersLoading(true);
        const resp = await client?.frontierServiceListOrganizationUsers(
          organization?.id
        );
        const users = resp?.data?.users || [];
        setOrgMembers(users);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      } finally {
        setIsOrgMembersLoading(false);
      }
    }
    if (canUpdateProject) {
      getOrganizationMembers();
    }
  }, [client, organization?.id, canUpdateProject]);

  const invitableUser = useMemo(
    () => filterUsersfromUsers(orgMembers, members) || [],
    [orgMembers, members]
  );

  const isUserLoading = isOrgMembersLoading;

  const topUsers = useMemo(
    () =>
      invitableUser
        .filter(user =>
          query
            ? user.title?.toLowerCase().includes(query.toLowerCase()) ||
              user.email?.includes(query)
            : true
        )
        .slice(0, 7),
    [invitableUser, query]
  );

  const topTeams: V1Beta1Group[] = useMemo(
    () =>
      teams
        .filter((team: V1Beta1Group) =>
          query
            ? team.title &&
              team.title.toLowerCase().includes(query.toLowerCase())
            : true
        )
        .slice(0, 7),
    [query, teams]
  );

  function onTextChange(e: React.ChangeEvent<HTMLInputElement>) {
    setQuery(e.target.value);
  }

  const addMember = useCallback(
    async (userId: string) => {
      if (!userId || !organization?.id || !projectId) return;
      try {
        const principal = `${PERMISSIONS.UserNamespace}:${userId}`;

        const policy: V1Beta1CreatePolicyForProjectBody = {
          role_id: PERMISSIONS.RoleProjectViewer,
          principal
        };
        await client?.frontierServiceCreatePolicyForProject(projectId, policy);
        toast.success('Member added');
        if (refetch) {
          refetch();
        }
      } catch ({ error }: any) {
        console.error(error);
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, organization?.id, projectId, refetch]
  );

  const addTeam = useCallback(
    async (teamId: string) => {
      if (!teamId || !organization?.id || !projectId) return;
      try {
        const principal = `${PERMISSIONS.GroupNamespace}:${teamId}`;

        const policy: V1Beta1CreatePolicyForProjectBody = {
          role_id: PERMISSIONS.RoleProjectViewer,
          principal
        };
        await client?.frontierServiceCreatePolicyForProject(projectId, policy);
        toast.success('Team added');
        if (refetch) {
          refetch();
        }
      } catch ({ error }: any) {
        console.error(error);
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, organization?.id, projectId, refetch]
  );

  return (
    <Popover>
      <Popover.Trigger asChild>
        <Button
          size="normal"
          style={{ width: 'fit-content', display: 'flex' }}
          data-test-id="frontier-sdk-add-project-member-btn"
          disabled={!canUpdateProject}
        >
          Add a member
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className={styles.popoverContent}>
        <Search
          data-test-id="frontier-sdk-add-project-project-textfield"
          value={query}
          placeholder={showTeam ? 'Add team to project' : 'Add project member'}
          onChange={onTextChange}
          variant='borderless'
          showClearButton
          disabled={isTeamsLoading || isUserLoading}
          onClear={() => setQuery('')}
        />
        <Separator />

        {showTeam ? (
          isTeamsLoading ? (
            <Skeleton height={'32px'} />
          ) : topTeams.length ? (
            <div style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}>
              {topTeams.map((team, i) => {
                const initals = getInitials(team?.title || team.name);
                return (
                  <Flex
                    gap="small"
                    key={team.id}
                    onClick={() => addTeam(team?.id || '')}
                    className={styles.inviteDropdownItem}
                    data-test-id={`frontier-sdk-add-team-to-project-dropdown-item-${i}`}
                  >
                    <Avatar
                      fallback={initals}
                      size={1}
                      radius="small"
                      imageProps={{
                        fontSize: '10px'
                      }}
                    />
                    <Text>{team?.title || team?.name}</Text>
                  </Flex>
                );
              })}
            </div>
          ) : (
            <Flex
              style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}
              justify={'center'}
              align={'center'}
            >
              <Text size="small">No Teams found</Text>
            </Flex>
          )
        ) : isUserLoading ? (
          <Skeleton height={'32px'} />
        ) : topUsers.length ? (
          <div style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}>
            {topUsers.map((user, i) => {
              const initals = getInitials(user?.title || user.email);
              return (
                <Flex
                  gap="small"
                  key={user.id}
                  className={styles.inviteDropdownItem}
                  onClick={() => addMember(user?.id || '')}
                  data-test-id={`frontier-sdk-add-user-to-project-dropdown-item-${i}`}
                >
                  <Avatar
                    src={user?.avatar}
                    fallback={initals}
                    size={1}
                    radius="small"
                    imageProps={{
                      fontSize: '10px'
                    }}
                  />
                  <Text>{user?.title || user?.email}</Text>
                </Flex>
              );
            })}
          </div>
        ) : (
          <Flex
            style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}
            justify={'center'}
            align={'center'}
          >
            <Text size="small">No Users found</Text>
          </Flex>
        )}
        <Separator style={{ margin: 0 }} />

        <div style={{ padding: 'var(--rs-space-2)' }}>
          <Flex
            onClick={toggleShowTeam}
            gap="small"
            className={styles.inviteDropdownItem}
            data-test-id={`frontier-sdk-add-project-member-toggle`}
          >
            {showTeam ? (
              <>
                <PlusIcon color="var(--rs-color-foreground-base-primary)" />{' '}
                <Text>Add project member</Text>
              </>
            ) : (
              <>
                <CardStackPlusIcon color="var(--rs-color-foreground-base-primary)" />{' '}
                <Text>Add team to project</Text>
              </>
            )}
          </Flex>
        </div>
      </Popover.Content>
    </Popover>
  );
};

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading='No members found'
    subHeading='Get started by adding your first member.'
  />
);
