'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Project } from '~/src';
import { styles } from '../styles';
import { getColumns } from './projects.columns';

export default function WorkspaceProjects() {
  const { client, activeOrganization: organization } = useFrontier();
  const routerState = useRouterState();
  const [projects, setProjects] = useState([]);
  const [isProjectsLoading, setIsProjectsLoading] = useState(false);

  const getProjects = useCallback(async () => {
    try {
      setIsProjectsLoading(true);
      const {
        // @ts-ignore
        data: { projects = [] }
      } = await client?.frontierServiceListProjectsByCurrentUser({
        // @ts-ignore
        org_id: organization?.id
      });
      setProjects(projects);
    } catch (err) {
      console.error(err);
    } finally {
      setIsProjectsLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    getProjects();
  }, [getProjects, routerState.location.key]);

  useEffect(() => {
    getProjects();
  }, [client, getProjects, organization?.id]);

  const updatedProjects = useMemo(
    () =>
      isProjectsLoading
        ? [{ id: 1 }, { id: 2 }, { id: 3 }]
        : projects.length
        ? projects
        : [],
    [isProjectsLoading, projects]
  );
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Projects</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <ProjectsTable
            // @ts-ignore
            projects={updatedProjects}
            isLoading={isProjectsLoading}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

interface WorkspaceProjectsProps {
  projects: V1Beta1Project[];
  isLoading?: boolean;
}

const ProjectsTable = ({ projects, isLoading }: WorkspaceProjectsProps) => {
  let navigate = useNavigate({ from: '/projects' });

  const tableStyle = projects?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(() => getColumns(isLoading), [isLoading]);
  return (
    <Flex direction="row">
      <DataTable
        data={projects ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 150px)' }}
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
