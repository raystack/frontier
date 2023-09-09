'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1User } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './member.columns';
import type { MembersTableType } from './member.types';

export default function WorkspaceMembers() {
  const [users, setUsers] = useState([]);
  const [isUsersLoading, setIsUsersLoading] = useState(false);
  const { client, activeOrganization: organization } = useFrontier();
  const routerState = useRouterState();

  const fetchOrganizationUser = useCallback(async () => {
    if (!organization?.id) return;
    try {
      setIsUsersLoading(true);
      const {
        // @ts-ignore
        data: { users }
      } = await client?.frontierServiceListOrganizationUsers(organization?.id);
      setUsers(users);

      const {
        // @ts-ignore
        data: { invitations }
      } = await client?.frontierServiceListOrganizationInvitations(
        organization?.id
      );

      const invitedUsers = invitations.map((user: V1Beta1User) => ({
        ...user,
        invited: true
      }));
      // @ts-ignore
      setUsers([...users, ...invitedUsers]);
    } catch (err) {
      console.error(err);
    } finally {
      setIsUsersLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    fetchOrganizationUser();
  }, [fetchOrganizationUser]);

  useEffect(() => {
    fetchOrganizationUser();
  }, [fetchOrganizationUser, routerState.location.key]);

  const updatedUsers = useMemo(
    () =>
      isUsersLoading
        ? [{ id: 1 }, { id: 2 }, { id: 3 }]
        : users.length
        ? users
        : [],
    [isUsersLoading, users]
  );

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = [
    {
      permission: PERMISSIONS.InvitationCreatePermission,
      resource
    },
    {
      permission: PERMISSIONS.UpdatePermission,
      resource
    }
  ];

  const { permissions } = usePermissions(
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

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Members</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <ManageMembers />
          {organization?.id ? (
            <MembersTable
              // @ts-ignore
              users={updatedUsers}
              organizationId={organization?.id}
              isLoading={isUsersLoading}
              canCreateInvite={canCreateInvite}
              canDeleteUser={canDeleteUser}
            />
          ) : null}
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

const ManageMembers = () => (
  <Flex direction="row" justify="between" align="center">
    <Flex direction="column" gap="small">
      <Text size={6}>Manage members</Text>
      <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
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
  organizationId
}: MembersTableType) => {
  let navigate = useNavigate({ from: '/members' });

  const tableStyle = useMemo(
    () =>
      users?.length ? { width: '100%' } : { width: '100%', height: '100%' },
    [users?.length]
  );

  const columns = useMemo(
    () => getColumns(organizationId, canDeleteUser, isLoading),
    [organizationId, canDeleteUser, isLoading]
  );

  return (
    <Flex direction="row">
      <DataTable
        // @ts-ignore
        data={users}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 222px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar style={{ padding: 0, border: 0 }}>
          <Flex justify="between" gap="small">
            <Flex style={{ maxWidth: '360px' }}>
              <DataTable.GloabalSearch
                placeholder="Search by name or email"
                size="medium"
              />
            </Flex>

            {canCreateInvite ? (
              <Button
                variant="primary"
                style={{ width: 'fit-content' }}
                onClick={() => navigate({ to: '/members/modal' })}
              >
                Invite people
              </Button>
            ) : null}
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
};

const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 members in your workspace</h3>
    <div className="pera">Try adding new members.</div>
  </EmptyState>
);
