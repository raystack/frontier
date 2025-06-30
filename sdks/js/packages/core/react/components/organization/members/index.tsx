'use client';

import { useEffect, useMemo } from 'react';

import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  Skeleton,
  Text,
  EmptyState,
  Flex,
  DataTable
} from '@raystack/apsara/v1';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationMembers } from '~/react/hooks/useOrganizationMembers';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './member.columns';
import type { MembersTableType } from './member.types';
import styles from './members.module.css';

export default function WorkspaceMembers() {
  const { activeOrganization: organization } = useFrontier();

  const routerState = useRouterState();

  const isListRoute = useMemo(() => {
    return routerState.location.pathname === '/members';
  }, [routerState.location.pathname]);

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.InvitationCreatePermission,
        resource
      },
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canCreateInvite, canDeleteUser } = useMemo(() => {
    return {
      canCreateInvite: shouldShowComponent(
        permissions,
        `${PERMISSIONS.InvitationCreatePermission}::${resource}`
      ),
      canDeleteUser: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const {
    roles,
    members,
    memberRoles,
    refetch,
    isFetching: isOrgMembersLoading
  } = useOrganizationMembers({
    showInvitations: canCreateInvite
  });

  const isLoading = isOrgMembersLoading || isPermissionsFetching;

  useEffect(() => {
    if (isListRoute) {
      refetch();
    }
  }, [isListRoute, refetch, routerState.location.state.key]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Text size="large">Members</Text>
      </Flex>
      <Flex direction="column" gap={9} className={styles.container}>
        <ManageMembers />
        {organization?.id ? (
          <MembersTable
            roles={roles}
            users={members}
            organizationId={organization?.id}
            isLoading={isLoading}
            canCreateInvite={canCreateInvite}
            canDeleteUser={canDeleteUser}
            memberRoles={memberRoles}
            refetch={refetch}
          />
        ) : null}
      </Flex>
      <Outlet />
    </Flex>
  );
}

const ManageMembers = () => (
  <Flex direction="row" justify="between" align="center">
    <Flex direction="column" gap={3}>
      <Text size="large" weight="medium">
        Manage members
      </Text>
      <Text size="regular" variant="secondary">
        Manage members for this domain.
      </Text>
    </Flex>
  </Flex>
);

const MembersTable = ({
  isLoading,
  users,
  canCreateInvite,
  canDeleteUser,
  organizationId,
  memberRoles,
  roles,
  refetch
}: MembersTableType) => {
  const navigate = useNavigate({ from: '/members' });

  const columns = useMemo(
    () =>
      getColumns(organizationId, memberRoles, roles, canDeleteUser, refetch),
    [organizationId, memberRoles, canDeleteUser, roles, refetch]
  );

  return (
    <DataTable
      data={users}
      isLoading={isLoading}
      defaultSort={{ name: 'name', order: 'asc' }}
      columns={columns}
      mode="client"
    >
      <Flex direction="column" gap={7} className={styles.tableWrapper}>
        <Flex justify="between" gap={3}>
          <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
          {isLoading ? (
            <Skeleton height='34px' width='500px' />
          ) : (
            <DataTable.Search
              placeholder="Search by name or email"
              size="medium"
            />
          )}
          </Flex>
          {isLoading ? (
            <Skeleton height='34px' width='64px' />
          ) : (
            <Tooltip
              message={AuthTooltipMessage}
              side="left"
              disabled={canCreateInvite}
            >
              <Button
                size="small"
                style={{ width: 'fit-content', height: '100%' }}
                onClick={() =>
                  navigate({
                    to: '/members/modal',
                    state: { from: '/members' }
                  })
                }
                disabled={!canCreateInvite}
                data-test-id="frontier-sdk-remove-member-link"
              >
                Invite people
              </Button>
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
  );
};

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading='No members found'
    subHeading='Get started by adding your first member'
  />
);
