'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Tooltip,
  EmptyState,
  Skeleton,
  Flex,
  Button,
  Text,
  Select,
  DataTable
} from '@raystack/apsara/v1';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationProjects } from '~/react/hooks/useOrganizationProjects';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import type { V1Beta1Project } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './projects.columns';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import styles from './project.module.css';

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
      <Flex className={styles.header}>
        <Text size={6}>Projects</Text>
      </Flex>
      <Flex direction="column" gap={9} className={styles.container}>
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
  const navigate = useNavigate({ from: '/projects' });

  const columns = useMemo(
    () => getColumns(userAccessOnProject),
    [userAccessOnProject]
  );
  return (
    <DataTable
      data={projects ?? []}
      isLoading={isLoading}
      defaultSort={{ name: 'name', order: 'asc' }}
      columns={columns}
      mode="client"
    >
      <Flex direction="column" gap={7} className={styles.tableWrapper}>
        <Flex justify="between" gap={3}>
          <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
            {isLoading ? (
                <Skeleton height='34px' width='500px' />
            ) : (
              <DataTable.Search placeholder="Search by title " size="medium" />
            )}
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
            <Skeleton height={'34px'} width={'64px'} />
          ) : (
            <Tooltip
              message={AuthTooltipMessage}
              side="left"
              disabled={canCreateProject}
            >
              <Button
                disabled={!canCreateProject}
                style={{ width: 'fit-content' }}
                onClick={() => navigate({ to: '/projects/modal' })}
                data-test-id="frontier-sdk-add-project-button"
              >
                Add project
              </Button>
            </Tooltip>
          )}
        </Flex>
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{
            root: styles.tableRoot,
            header: styles.tableHeader
          }}
        />
      </Flex>
    </DataTable>
  );
};

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading='No projects found'
    subHeading='Get started by creating your first project.'
  />
);
