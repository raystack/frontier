'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  DataTable,
  Flex,
  Select,
  Text
} from '@raystack/apsara';
import { Tooltip, EmptyState, Skeleton } from '@raystack/apsara/v1';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationProjects } from '~/react/hooks/useOrganizationProjects';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import { V1Beta1Project } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './projects.columns';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { styles } from '../styles';

const projectsSelectOptions = [
  { value: 'my-projects', label: 'My Projects' },
  { value: 'all-projects', label: 'All Projects' }
];

export default function WorkspaceProjects() {
  const [showOrgProjects, setShowOrgProjects] = useState(false);
  const {
    isFetching: isProjectsLoading,
    projects,
    userAccessOnProject,
    refetch
  } = useOrganizationProjects({
    allProjects: showOrgProjects,
    withMemberCount: true
  });
  const { activeOrganization: organization } = useFrontier();

  const routerState = useRouterState();

  const isListRoute = useMemo(() => {
    return routerState.location.pathname === '/projects';
  }, [routerState.location.pathname]);

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.ProjectCreatePermission,
        resource
      },
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canCreateProject, canUpdateOrganization } = useMemo(() => {
    return {
      canCreateProject: shouldShowComponent(
        permissions,
        `${PERMISSIONS.ProjectCreatePermission}::${resource}`
      ),
      canUpdateOrganization: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const onOrgProjectsFilterChange = useCallback((value: string) => {
    if (value === 'all-projects') {
      setShowOrgProjects(true);
    } else {
      setShowOrgProjects(false);
    }
  }, []);

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
            onOrgProjectsFilterChange={onOrgProjectsFilterChange}
            canListOrgProjects={canUpdateOrganization}
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
  canListOrgProjects?: boolean;
  onOrgProjectsFilterChange?: (value: string) => void;
}

const ProjectsTable = ({
  projects,
  isLoading,
  canCreateProject,
  userAccessOnProject,
  canListOrgProjects,
  onOrgProjectsFilterChange
}: WorkspaceProjectsProps) => {
  let navigate = useNavigate({ from: '/projects' });

  const tableStyle = projects?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(
    () => getColumns(userAccessOnProject),
    [userAccessOnProject]
  );
  return (
    <Flex direction="row">
      <DataTable
        data={projects ?? []}
        isLoading={isLoading}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 150px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar
          style={{ padding: 0, border: 0, marginBottom: 'var(--rs-space-5)' }}
        >
          <Flex justify="between" gap="small">
            <Flex
              style={{
                maxWidth: canListOrgProjects ? '500px' : '360px',
                width: '100%'
              }}
              gap={'medium'}
            >
              <DataTable.GloabalSearch
                placeholder="Search by name"
                size="medium"
              />
              {canListOrgProjects ? (
                <Select
                  defaultValue={projectsSelectOptions[0].value}
                  onValueChange={onOrgProjectsFilterChange}
                >
                  <Select.Trigger style={{ minWidth: '140px' }}>
                    <Select.Value />
                  </Select.Trigger>
                  <Select.Content>
                    {projectsSelectOptions.map(opt => (
                      <Select.Item value={opt.value} key={opt.value}>
                        {opt.label}
                      </Select.Item>
                    ))}
                  </Select.Content>
                </Select>
              ) : null}
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
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading={"0 projects in your organization"}
    subHeading={"Try adding new project."}
  />
);
