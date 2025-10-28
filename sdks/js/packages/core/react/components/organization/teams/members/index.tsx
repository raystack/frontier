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
} from '@raystack/apsara';
import { Link, useParams } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { ExclamationTriangleIcon, PaperPlaneIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import type { Role, User } from '@raystack/proton/frontier';
import {
  PERMISSIONS,
  filterUsersfromUsers,
  getInitials,
  shouldShowComponent
} from '~/utils';
import { getColumns } from './member.columns';

import { useQuery, useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries, AddGroupUsersRequestSchema, ListOrganizationUsersRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import styles from './members.module.css';

export type MembersProps = {
  members: User[];
  roles: Role[];
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
                size="large"
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
  canUpdateGroup: boolean;
  refetchMembers: () => void;
  members: User[];
}

const AddMemberDropdown = ({
  canUpdateGroup,
  refetchMembers,
  members
}: AddMemberDropdownProps) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const [query, setQuery] = useState('');

  const { activeOrganization: organization } = useFrontier();

  // Get organization members using Connect RPC
  const { data: orgMembersData, isLoading: isOrgMembersLoading, error: orgMembersError } = useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    create(ListOrganizationUsersRequestSchema, { id: organization?.id || '' }),
    { enabled: !!organization?.id && canUpdateGroup }
  );

  // Handle organization members error
  useEffect(() => {
    if (orgMembersError) {
      toast.error('Something went wrong', {
        description: orgMembersError.message
      });
    }
  }, [orgMembersError]);


  const invitableUser = useMemo(
    () => filterUsersfromUsers(orgMembersData?.users || [], members) || [],
    [orgMembersData?.users, members]
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

  function onTextChange(e: React.ChangeEvent<HTMLInputElement>) {
    setQuery(e.target.value);
  }

  // Add group user using Connect RPC
  const addGroupUserMutation = useMutation(FrontierServiceQueries.addGroupUsers, {
    onSuccess: () => {
      toast.success('member added');
      if (refetchMembers) {
        refetchMembers();
      }
    },
    onError: (error) => {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  });

  const addMember = useCallback(
    (userId: string) => {
      if (!userId || !organization?.id) return;

      const request = create(AddGroupUsersRequestSchema, {
        id: teamId as string,
        orgId: organization.id,
        userIds: [userId]
      });

      addGroupUserMutation.mutate(request);
    },
    [organization?.id, teamId, addGroupUserMutation]
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
          variant="borderless"
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
        <Separator />
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
    heading="No members found"
    subHeading="Get started by adding your first team member."
  />
);
