'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { useCallback, useEffect, useState } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Organization } from '~/src';
import { styles } from '../styles';
import { columns } from './teams.columns';

interface WorkspaceTeamProps {
  teams: V1Beta1Group[];
}

export default function WorkspaceTeams({
  organization
}: {
  organization?: V1Beta1Organization;
}) {
  const [teams, setTeams] = useState([]);

  const { client } = useFrontier();
  const location = useLocation();

  const getTeams = useCallback(async () => {
    const {
      // @ts-ignore
      data: { groups = [] }
    } = await client?.adminServiceListGroups({ orgId: organization?.id });
    setTeams(groups);
  }, [client, organization?.id]);

  useEffect(() => {
    getTeams();
  }, [getTeams, location.key]);

  useEffect(() => {
    getTeams();
  }, [client, getTeams, organization?.id]);

  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Teams</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <TeamsTable teams={teams} />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

const TeamsTable = ({ teams }: WorkspaceTeamProps) => {
  let navigate = useNavigate();

  const tableStyle = teams?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  return (
    <Flex direction="row">
      <DataTable
        data={teams ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 120px)' }}
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
              onClick={() => navigate('/teams/modal')}
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
