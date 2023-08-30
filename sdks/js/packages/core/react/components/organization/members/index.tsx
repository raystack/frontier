'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { useCallback, useEffect, useState } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Organization, V1Beta1User } from '~/src';
import { styles } from '../styles';
import { getColumns } from './member.columns';
import type { MembersTableType } from './member.types';

export default function WorkspaceMembers({
  organization
}: {
  organization?: V1Beta1Organization;
}) {
  const [users, setUsers] = useState([]);
  const { client } = useFrontier();
  const location = useLocation();

  const fetchOrganizationUser = useCallback(async () => {
    if (!organization?.id) return;
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
  }, [client, organization?.id]);

  useEffect(() => {
    fetchOrganizationUser();
  }, [organization?.id, client, fetchOrganizationUser]);

  useEffect(() => {
    fetchOrganizationUser();
  }, [fetchOrganizationUser, location.key]);

  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Members</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <ManageMembers />
          <MembersTable users={users} organizationId={organization?.id} />
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

const MembersTable = ({ users, organizationId }: MembersTableType) => {
  let navigate = useNavigate();

  const tableStyle = users?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  return (
    <Flex direction="row">
      <DataTable
        // @ts-ignore
        data={users ?? []}
        // @ts-ignore
        columns={getColumns(organizationId)}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 190px)' }}
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

            <Button
              variant="primary"
              style={{ width: 'fit-content' }}
              onClick={() => navigate('/members/modal')}
            >
              Invite people
            </Button>
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
