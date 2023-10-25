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
import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';

import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import {
  PERMISSIONS,
  filterUsersfromUsers,
  getInitials,
  shouldShowComponent
} from '~/utils';
import { getColumns } from './member.columns';
import Skeleton from 'react-loading-skeleton';
import { AuthTooltipMessage } from '~/react/utils';
import { MagnifyingGlassIcon, PaperPlaneIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import styles from './members.module.css';

export type MembersProps = {
  members: V1Beta1User[];
  organizationId: string;
  memberRoles?: Record<string, Role[]>;
  isLoading?: boolean;
};

export const Members = ({
  members,
  organizationId,
  memberRoles = {},
  isLoading: isMemberLoading
}: MembersProps) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const navigate = useNavigate({ from: '/teams/$teamId' });

  const membersCount = members?.length || 0;

  const tableStyle = membersCount
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const resource = `app/group:${teamId}`;
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
    !!teamId
  );
  const { canUpdateGroup } = useMemo(() => {
    return {
      canUpdateGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading = isPermissionsFetching || isMemberLoading;

  const columns = useMemo(
    () =>
      getColumns({
        organizationId,
        canUpdateGroup,
        memberRoles,
        isLoading,
        membersCount
      }),
    [organizationId, canUpdateGroup, memberRoles, isLoading, membersCount]
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
        // @ts-ignore
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
                disabled={canUpdateGroup}
              >
                <AddMemberDropdown canUpdateGroup={canUpdateGroup} />
              </Tooltip>
            )}
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
};

interface AddMemberDropdownProps {
  canUpdateGroup: boolean;
}

const AddMemberDropdown = ({ canUpdateGroup }: AddMemberDropdownProps) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const [orgMembers, setOrgMembers] = useState<V1Beta1User[]>([]);
  const [isOrgMembersLoading, setIsOrgMembersLoading] = useState(false);
  const [query, setQuery] = useState('');

  const [members, setMembers] = useState<V1Beta1User[]>([]);

  const [isTeamMembersLoading, setIsTeamMembersLoading] = useState(false);

  const { client, activeOrganization: organization } = useFrontier();

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
    if (canUpdateGroup) {
      getOrganizationMembers();
    }
  }, [client, organization?.id, canUpdateGroup]);

  useEffect(() => {
    async function getTeamMembers() {
      if (!organization?.id || !teamId) return;
      try {
        setIsTeamMembersLoading(true);
        const {
          // @ts-ignore
          data: { users, role_pairs = [] }
        } = await client?.frontierServiceListGroupUsers(
          organization?.id,
          teamId,
          { withRoles: true }
        );

        setMembers(users);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      } finally {
        setIsTeamMembersLoading(false);
      }
    }
    if (canUpdateGroup) {
      getTeamMembers();
    }
  }, [canUpdateGroup, client, organization?.id, teamId]);

  const invitableUser = useMemo(
    () => filterUsersfromUsers(orgMembers, members) || [],
    [orgMembers, members]
  );

  const isUserLoading = isOrgMembersLoading || isTeamMembersLoading;

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

  function onTextChange(e: React.ChangeEvent<HTMLInputElement>) {
    setQuery(e.target.value);
  }

  const addMember = useCallback(
    async (userId: string) => {
      if (!userId || !organization?.id) return;
      try {
        await client?.frontierServiceAddGroupUsers(organization?.id, teamId, {
          userIds: [userId]
        });
        toast.success('member added');
      } catch ({ error }: any) {
        console.error(error);
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, organization?.id, teamId]
  );

  return (
    <DropdownMenu style={{ height: '100%' }}>
      <DropdownMenu.Trigger
        asChild
        style={{ cursor: 'pointer' }}
        disabled={!canUpdateGroup}
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
              placeholder="Add team member"
              className={styles.inviteDropdownSearch}
              onChange={onTextChange}
            />
          </DropdownMenu.Item>
        </DropdownMenu.Group>
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
        <DropdownMenu.Separator style={{ margin: 0 }} />
        <DropdownMenu.Group>
          <DropdownMenu.Item style={{ padding: 0 }}>
            <Link
              to={'/teams/$teamId/invite'}
              params={{ teamId: teamId }}
              style={{
                display: 'flex',
                gap: 'var(--pd-8)',
                padding: 'var(--pd-8)'
              }}
            >
              <PaperPlaneIcon color="var(--foreground-base)" />{' '}
              <Text>Invite People</Text>
            </Link>
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
