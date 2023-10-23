'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationMembers } from '~/react/hooks/useOrganizationMembers';
import { usePermissions } from '~/react/hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './member.columns';
import type { MembersTableType } from './member.types';

export default function WorkspaceMembers() {
  const { activeOrganization: organization } = useFrontier();

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
    isFetching: isOrgMembersLoading,
    members,
    memberRoles
  } = useOrganizationMembers({
    showInvitations: canCreateInvite
  });

  const isLoading = isOrgMembersLoading || isPermissionsFetching;

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
              users={members}
              organizationId={organization?.id}
              isLoading={isLoading}
              canCreateInvite={canCreateInvite}
              canDeleteUser={canDeleteUser}
              memberRoles={memberRoles}
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
  organizationId,
  memberRoles
}: MembersTableType) => {
  let navigate = useNavigate({ from: '/members' });

  const tableStyle = useMemo(
    () =>
      users?.length ? { width: '100%' } : { width: '100%', height: '100%' },
    [users?.length]
  );

  const columns = useMemo(
    () => getColumns(organizationId, memberRoles, canDeleteUser, isLoading),
    [organizationId, memberRoles, canDeleteUser, isLoading]
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

            {canCreateInvite && !isLoading ? (
              <Button
                variant="primary"
                style={{ width: 'fit-content' }}
                onClick={() =>
                  navigate({
                    to: '/members/modal',
                    state: { from: '/members' }
                  })
                }
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
