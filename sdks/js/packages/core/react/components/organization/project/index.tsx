'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationProjects } from '~/react/hooks/useOrganizationProjects';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Project } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './projects.columns';

export default function WorkspaceProjects() {
  const { isFetching, projects, userAccessOnProject } =
    useOrganizationProjects();
  const { activeOrganization: organization } = useFrontier();

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = [
    {
      permission: PERMISSIONS.ProjectCreatePermission,
      resource
    }
  ];

  const { permissions } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canCreateProject } = useMemo(() => {
    return {
      canCreateProject: shouldShowComponent(
        permissions,
        `${PERMISSIONS.ProjectCreatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Projects</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <ProjectsTable
            // @ts-ignore
            projects={projects}
            isLoading={isFetching}
            canCreateProject={canCreateProject}
            userAccessOnProject={userAccessOnProject}
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
  canCreateProject?: boolean;
  userAccessOnProject: Record<string, string[]>;
}

const ProjectsTable = ({
  projects,
  isLoading,
  canCreateProject,
  userAccessOnProject
}: WorkspaceProjectsProps) => {
  let navigate = useNavigate({ from: '/projects' });

  const tableStyle = projects?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(
    () => getColumns(userAccessOnProject, isLoading),
    [isLoading, userAccessOnProject]
  );
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

            {canCreateProject ? (
              <Button
                variant="primary"
                style={{ width: 'fit-content' }}
                onClick={() => navigate({ to: '/projects/modal' })}
              >
                Add project
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
    <h3>0 projects in your organization</h3>
    <div className="pera">Try adding new project.</div>
  </EmptyState>
);
