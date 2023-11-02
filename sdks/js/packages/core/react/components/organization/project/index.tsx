'use client';

import {
  Button,
  DataTable,
  EmptyState,
  Flex,
  Text,
  Tooltip
} from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationProjects } from '~/react/hooks/useOrganizationProjects';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Project } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './projects.columns';
import Skeleton from 'react-loading-skeleton';
import { AuthTooltipMessage } from '~/react/utils';

export default function WorkspaceProjects() {
  const {
    isFetching: isProjectsLoading,
    projects,
    userAccessOnProject,
    refetch
  } = useOrganizationProjects();
  const { activeOrganization: organization } = useFrontier();

  const routerState = useRouterState();

  const isListRoute = useMemo(() => {
    return routerState.matches.some(route => route.routeId === '/projects');
  }, [routerState.matches]);

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.ProjectCreatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
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

  useEffect(() => {
    if (isListRoute) {
      refetch();
    }
  }, [isListRoute, refetch, routerState.location.state.key]);

  const isLoading = isPermissionsFetching || isProjectsLoading;

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
            isLoading={isLoading}
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
            {isLoading ? (
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Tooltip
                message={AuthTooltipMessage}
                side="left"
                disabled={canCreateProject}
              >
                <Button
                  variant="primary"
                  disabled={!canCreateProject}
                  style={{ width: 'fit-content' }}
                  onClick={() => navigate({ to: '/projects/modal' })}
                >
                  Add project
                </Button>
              </Tooltip>
            )}
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
