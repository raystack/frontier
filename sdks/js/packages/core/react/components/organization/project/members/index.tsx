import {
  Avatar,
  Button,
  DataTable,
  DropdownMenu,
  EmptyState,
  Flex,
  Text,
  TextField,
  Tooltip
} from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Group, V1Beta1PolicyRequestBody, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import {
  PERMISSIONS,
  filterUsersfromUsers,
  getInitials,
  shouldShowComponent
} from '~/utils';
import { getColumns } from './member.columns';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { toast } from 'sonner';
import {
  CardStackPlusIcon,
  MagnifyingGlassIcon,
  PlusIcon
} from '@radix-ui/react-icons';
import styles from './members.module.css';
import Skeleton from 'react-loading-skeleton';
import { AuthTooltipMessage } from '~/react/utils';
import { useOrganizationTeams } from '~/react/hooks/useOrganizationTeams';

export type MembersProps = {
  members?: V1Beta1User[];
  memberRoles?: Record<string, Role[]>;
  isLoading?: boolean;
};

export const Members = ({
  members,
  memberRoles,
  isLoading: isMemberLoading
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

  const tableStyle = members?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const isLoading = isMemberLoading || isPermissionsFetching;

  const columns = useMemo(
    () => getColumns(memberRoles, isLoading),
    [memberRoles, isLoading]
  );

  const updatedUsers = useMemo(() => {
    return isLoading
      ? ([{ id: 1 }, { id: 2 }, { id: 3 }] as any)
      : members?.length
      ? members
      : [];
  }, [members, isLoading]);

  return (
    <Flex direction="column" style={{ paddingTop: '32px' }}>
      <DataTable
        data={updatedUsers}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 212px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar style={{ padding: 0, border: 0 }}>
          <Flex justify="between" gap="small">
            <Flex style={{ maxWidth: '360px', width: '100%' }}>
              <DataTable.GloabalSearch
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
                disabled={!canUpdateProject}
              >
                <AddMemberDropdown canUpdateProject={!canUpdateProject} />
              </Tooltip>
            )}
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
};

interface AddMemberDropdownProps {
  canUpdateProject: boolean;
  members?: V1Beta1User[];
}

const AddMemberDropdown = ({
  canUpdateProject,
  members
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
        const {
          // @ts-ignore
          data: { users }
        } = await client?.frontierServiceListOrganizationUsers(
          organization?.id
        );
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
            ? user.title &&
              user.title.toLowerCase().includes(query.toLowerCase())
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
        const resource = `${PERMISSIONS.ProjectNamespace}:${projectId}`;
        const principal = `${PERMISSIONS.UserNamespace}:${userId}`;

        const policy: V1Beta1PolicyRequestBody = {
          roleId: PERMISSIONS.RoleProjectViewer,
          resource,
          principal
        };
        await client?.frontierServiceCreatePolicy(policy);
        toast.success('member added');
      } catch ({ error }: any) {
        console.error(error);
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, organization?.id, projectId]
  );

  const addTeam = useCallback(
    async (teamId: string) => {
      if (!teamId || !organization?.id || !projectId) return;
      try {
        const resource = `${PERMISSIONS.ProjectNamespace}:${projectId}`;
        const principal = `${PERMISSIONS.GroupNamespace}:${teamId}`;

        const policy: V1Beta1PolicyRequestBody = {
          roleId: PERMISSIONS.RoleProjectViewer,
          resource,
          principal
        };
        await client?.frontierServiceCreatePolicy(policy);
        toast.success('team added');
      } catch ({ error }: any) {
        console.error(error);
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, organization?.id, projectId]
  );

  return (
    <DropdownMenu style={{ height: '100%' }}>
      <DropdownMenu.Trigger
        asChild
        style={{ cursor: 'pointer' }}
        disabled={!canUpdateProject}
      >
        <Button
          variant="primary"
          style={{ width: 'fit-content', display: 'flex' }}
        >
          Add a member
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        align="end"
        style={{ padding: 0, minWidth: '300px' }}
      >
        <DropdownMenu.Group style={{ padding: 0 }}>
          <DropdownMenu.Item
            style={{ padding: 0 }}
            // prevent dropdown to close on clicking the search box
            onClick={(e: React.MouseEvent<HTMLElement>) => e.preventDefault()}
          >
            <TextField
              // @ts-ignore
              leading={
                <MagnifyingGlassIcon
                  style={{ color: 'var(--foreground-base)' }}
                />
              }
              value={query}
              placeholder={
                showTeam ? 'Add team to project' : 'Add project member'
              }
              className={styles.inviteDropdownSearch}
              onChange={onTextChange}
            />
          </DropdownMenu.Item>
        </DropdownMenu.Group>
        {showTeam ? (
          <DropdownMenu.Group>
            {isTeamsLoading ? (
              <Skeleton height={'32px'} />
            ) : topTeams.length ? (
              topTeams.map(team => {
                const initals = getInitials(team?.title || team.name);
                return (
                  <DropdownMenu.Item
                    key={team.id}
                    asChild
                    onClick={() => addTeam(team?.id || '')}
                  >
                    <Flex
                      gap="small"
                      style={{ padding: 'var(--pd-8)', userSelect: 'none' }}
                    >
                      <Avatar
                        fallback={initals}
                        imageProps={{
                          width: '16px',
                          height: '16px',
                          fontSize: '10px'
                        }}
                      />
                      <Text>{team?.title || team?.name}</Text>
                    </Flex>
                  </DropdownMenu.Item>
                );
              })
            ) : (
              <Flex style={{ padding: '0 var(--pd-8)' }}>
                <Text size={2}>No Teams found</Text>
              </Flex>
            )}
          </DropdownMenu.Group>
        ) : (
          <DropdownMenu.Group>
            {isUserLoading ? (
              <Skeleton height={'32px'} />
            ) : topUsers.length ? (
              topUsers.map(user => {
                const initals = getInitials(user?.title || user.email);
                return (
                  <DropdownMenu.Item
                    key={user.id}
                    asChild
                    onClick={() => addMember(user?.id || '')}
                  >
                    <Flex
                      gap="small"
                      style={{ padding: 'var(--pd-8)', userSelect: 'none' }}
                    >
                      <Avatar
                        fallback={initals}
                        imageProps={{
                          width: '16px',
                          height: '16px',
                          fontSize: '10px'
                        }}
                      />
                      <Text>{user?.title || user?.email}</Text>
                    </Flex>
                  </DropdownMenu.Item>
                );
              })
            ) : (
              <Flex style={{ padding: '0 var(--pd-8)' }}>
                <Text size={2}>No Users found</Text>
              </Flex>
            )}
          </DropdownMenu.Group>
        )}
        <DropdownMenu.Separator style={{ margin: 0 }} />
        <DropdownMenu.Group>
          <DropdownMenu.Item style={{ padding: 0 }}>
            <Flex
              onClick={toggleShowTeam}
              gap="small"
              style={{
                padding: 'var(--pd-8)',
                userSelect: 'none'
              }}
            >
              {showTeam ? (
                <>
                  <PlusIcon color="var(--foreground-base)" />{' '}
                  <Text>Add project member</Text>
                </>
              ) : (
                <>
                  <CardStackPlusIcon color="var(--foreground-base)" />{' '}
                  <Text>Add team to project</Text>
                </>
              )}
            </Flex>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
};

const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 members in your team</h3>
    <div className="pera">Try adding new members.</div>
  </EmptyState>
);
