'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';

import { useOrganizationTeams } from '~/react/hooks/useOrganizationTeams';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Group } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './teams.columns';

interface WorkspaceTeamProps {
  teams: V1Beta1Group[];
  isLoading?: boolean;
  canCreateGroup?: boolean;
  userAccessOnTeam: Record<string, string[]>;
}

export default function WorkspaceTeams() {
  const { isFetching, teams, userAccessOnTeam } = useOrganizationTeams({
    withPermissions: ['update', 'delete']
  });
  const { activeOrganization: organization } = useFrontier();

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = [
    {
      permission: PERMISSIONS.GroupCreatePermission,
      resource
    }
  ];

  const { permissions } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canCreateGroup } = useMemo(() => {
    return {
      canCreateGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.GroupCreatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Teams</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <TeamsTable
            teams={teams}
            isLoading={isFetching}
            canCreateGroup={canCreateGroup}
            userAccessOnTeam={userAccessOnTeam}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

const TeamsTable = ({
  teams,
  isLoading,
  canCreateGroup,
  userAccessOnTeam
}: WorkspaceTeamProps) => {
  let navigate = useNavigate({ from: '/members' });

  const tableStyle = teams?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(
    () => getColumns(userAccessOnTeam, isLoading),
    [isLoading, userAccessOnTeam]
  );

  return (
    <Flex direction="row">
      <DataTable
        data={teams ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 180px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar style={{ padding: 0, border: 0 }}>
          <Flex justify="between" gap="small">
            <Flex style={{ maxWidth: '360px', width: '100%' }}>
              <DataTable.GloabalSearch
                placeholder="Search by name"
                size="medium"
              />
            </Flex>

            {canCreateGroup ? (
              <Button
                variant="primary"
                style={{ width: 'fit-content' }}
                onClick={() => navigate({ to: '/teams/modal' })}
              >
                Add team
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
    <h3>0 teams in your organization</h3>
    <div className="pera">Try adding new team.</div>
  </EmptyState>
);
