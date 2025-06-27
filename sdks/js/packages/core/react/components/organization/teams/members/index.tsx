import type React from 'react';
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
import { Link, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';

import {
  ExclamationTriangleIcon,
  PaperPlaneIcon
} from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import type { V1Beta1Role, V1Beta1User } from '~/src';
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
  const { teamId } = useParams({ from: '/teams/$teamId' });

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
    <Flex direction="column" className={styles.container}>
      <DataTable
        isLoading={isLoading}
        data={members}
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
                disabled={canUpdateGroup}
              >
                <AddMemberDropdown
                  canUpdateGroup={canUpdateGroup}
                  refetchMembers={refetchMembers}
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
          { with_roles: true }
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
          user_ids: [userId]
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
    <Popover>
      <Popover.Trigger asChild>
        <Button
          size="normal"
          style={{ width: 'fit-content', display: 'flex' }}
          data-test-id="frontier-sdk-add-member-btn"
          disabled={!canUpdateGroup}
        >
          Add a member
        </Button>
      </Popover.Trigger>
      <Popover.Content align="end" className={styles.popoverContent}>
        <Search
          data-test-id="frontier-sdk-add-project-member-textfield"
          value={query}
          variant='borderless'
          placeholder="Add team member"
          onChange={onTextChange}
          showClearButton
          disabled={isUserLoading}
          onClear={() => setQuery('')}
        />
        <Separator />

        {isUserLoading ? (
          <Skeleton height={'32px'} />
        ) : topUsers.length ? (
          <div style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}>
            {topUsers.map(user => {
              const initials = getInitials(user?.title || user.email);
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
                    fallback={initials}
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
            <Text size="small">No Users found</Text>
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
    heading={'0 members in your team'}
    subHeading={'Try adding new members.'}
  />
);
