import {
  DataTable,
  Flex,
  Popover,
  Text,
  TextField
} from '@raystack/apsara';
import { Button, EmptyState, Tooltip, toast, Separator, Avatar, Skeleton } from '@raystack/apsara/v1';
import { Link, useNavigate, useParams } from '@tanstack/react-router';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

import { MagnifyingGlassIcon, PaperPlaneIcon, ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import { V1Beta1Role, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import {
  PERMISSIONS,
  filterUsersfromUsers,
  getInitials,
  shouldShowComponent
} from '~/utils';
import { getColumns } from './member.columns';
import styles from './members.module.css';

export type MembersProps = {
  members: V1Beta1User[];
  roles: V1Beta1Role[];
  organizationId: string;
  memberRoles?: Record<string, Role[]>;
  isLoading?: boolean;
  refetchMembers: () => void;
};

export const Members = ({
  members,
  roles = [],
  organizationId,
  memberRoles = {},
  isLoading: isMemberLoading,
  refetchMembers
}: MembersProps) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const navigate = useNavigate({ from: '/teams/$teamId' });

  const membersCount = members?.length || 0;

  const tableStyle = useMemo(
    () =>
      membersCount ? { width: '100%' } : { width: '100%', height: '100%' },
    [membersCount]
  );

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
        roles,
        organizationId,
        canUpdateGroup,
        memberRoles,
        refetchMembers
      }),
    [roles, organizationId, canUpdateGroup, memberRoles, refetchMembers]
  );

  return (
    <Flex direction="column" style={{ paddingTop: '32px' }}>
      <DataTable
        isLoading={isLoading}
        data={members}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 212px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar
          style={{ padding: 0, border: 0, marginBottom: 'var(--rs-space-5)' }}
        >
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
                <AddMemberDropdown
                  canUpdateGroup={canUpdateGroup}
                  refetchMembers={refetchMembers}
                />
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
  refetchMembers: () => void;
}

const AddMemberDropdown = ({
  canUpdateGroup,
  refetchMembers
}: AddMemberDropdownProps) => {
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
        if (refetchMembers) {
          refetchMembers();
        }
      } catch ({ error }: any) {
        console.error(error);
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, organization?.id, refetchMembers, teamId]
  );

  return (
    <Popover style={{ height: '100%' }}>
      <Popover.Trigger
        asChild
        style={{ cursor: 'pointer' }}
        disabled={!canUpdateGroup}
      >
        <Button
          size="small"
          style={{ width: 'fit-content', display: 'flex' }}
          data-test-id="frontier-sdk-add-member-btn"
        >
          Add a member
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" style={{ padding: 0, minWidth: '300px' }}>
        <TextField
          // @ts-ignore
          leading={
            <MagnifyingGlassIcon style={{ color: 'var(--rs-color-foreground-base-primary)' }} />
          }
          value={query}
          placeholder="Add team member"
          className={styles.inviteDropdownSearch}
          onChange={onTextChange}
          data-test-id="frontier-sdk-add-member-search"
        />
        <Separator />

        {isUserLoading ? (
          <Skeleton height={'32px'} />
        ) : topUsers.length ? (
          <div style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}>
            {topUsers.map(user => {
              const initals = getInitials(user?.title || user.email);
              return (
                <Flex
                  gap="small"
                  key={user.id}
                  onClick={() => addMember(user?.id || '')}
                  className={styles.inviteDropdownItem}
                  data-test-id={`frontier-sdk-add-member-${user.id}`}
                >
                  <Avatar
                    src={user?.avatar}
                    fallback={initals}
                    size={1}
                    radius="small"
                    imageProps={{ fontSize: '10px' }}
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
            <Text size={2}>No Users found</Text>
          </Flex>
        )}
        <Separator style={{ margin: 0 }} />
        <div style={{ padding: 'var(--rs-space-2)' }}>
          <Link
            to={'/teams/$teamId/invite'}
            params={{ teamId: teamId }}
            className={styles.inviteDropdownItem}
          >
            <PaperPlaneIcon color="var(--rs-color-foreground-base-primary)" />{' '}
            <Text>Invite People</Text>
          </Link>
        </div>
      </Popover.Content>
    </Popover>
  );
};

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading={"0 members in your team"}
    subHeading={"Try adding new members."}
  />
);
