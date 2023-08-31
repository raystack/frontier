'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { useCallback, useEffect, useState } from 'react';
import { Outlet, useRouterState, useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Project } from '~/src';
import { styles } from '../styles';
import { columns } from './projects.columns';

export default function WorkspaceProjects() {
  const { client, activeOrganization: organization } = useFrontier();
  const routerState = useRouterState();
  const [projects, setProjects] = useState([]);

  const getProjects = useCallback(async () => {
    const {
      // @ts-ignore
      data: { projects = [] }
    } = await client?.adminServiceListProjects({ orgId: organization?.id });
    setProjects(projects);
  }, [client, organization?.id]);

  useEffect(() => {
    getProjects();
  }, [getProjects, routerState.location.key]);

  useEffect(() => {
    getProjects();
  }, [client, getProjects, organization?.id]);

  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Projects</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <ProjectsTable projects={projects} />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

interface WorkspaceProjectsProps {
  projects: V1Beta1Project[];
}

const ProjectsTable = ({ projects }: WorkspaceProjectsProps) => {
  let navigate = useNavigate({ from: '/projects' });

  const tableStyle = projects?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  return (
    <Flex direction="row">
      <DataTable
        data={projects ?? []}
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
              onClick={() => navigate({ to: '/projects/modal' })}
            >
              Add project
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
    <h3>0 projects in your organization</h3>
    <div className="pera">Try adding new project.</div>
  </EmptyState>
);
