'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group } from '~/src';
import { styles } from '../styles';
import { getColumns } from './teams.columns';

interface WorkspaceTeamProps {
  teams: V1Beta1Group[];
  isLoading?: boolean;
}

export default function WorkspaceTeams() {
  const [teams, setTeams] = useState([]);
  const [isTeamsLoading, setIsTeamsLoading] = useState(false);
  const { client, activeOrganization: organization } = useFrontier();
  const routerState = useRouterState();

  const getTeams = useCallback(async () => {
    try {
      setIsTeamsLoading(true);
      const {
        // @ts-ignore
        data: { groups = [] }
      } = await client?.frontierServiceListCurrentUserGroups();
      const orgGroups = groups.filter(
        // @ts-ignore TODO: Fix proto ts config
        (group: V1Beta1Group) => group.org_id === organization?.id
      );
      setTeams(orgGroups);
    } catch (err) {
      console.error(err);
    } finally {
      setIsTeamsLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    getTeams();
  }, [getTeams, routerState.location.key]);

  useEffect(() => {
    getTeams();
  }, [client, getTeams, organization?.id]);

  const updatedTeams = useMemo(
    () =>
      isTeamsLoading
        ? [{ id: 1 }, { id: 2 }, { id: 3 }]
        : teams.length
        ? teams
        : [],
    [isTeamsLoading, teams]
  );

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Teams</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          {/* @ts-ignore */}
          <TeamsTable teams={updatedTeams} isLoading={isTeamsLoading} />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

const TeamsTable = ({ teams, isLoading }: WorkspaceTeamProps) => {
  let navigate = useNavigate({ from: '/members' });

  const tableStyle = teams?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(() => getColumns(isLoading), [isLoading]);

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

            <Button
              variant="primary"
              style={{ width: 'fit-content' }}
              onClick={() => navigate({ to: '/teams/modal' })}
            >
              Add team
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
    <h3>0 teams in your organization</h3>
    <div className="pera">Try adding new team.</div>
  </EmptyState>
);
