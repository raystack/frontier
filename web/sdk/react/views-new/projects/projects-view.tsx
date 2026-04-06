'use client';

import { useEffect, useMemo } from 'react';
import { ExclamationTriangleIcon, Pencil1Icon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  Skeleton,
  Flex,
  EmptyState,
  DataTable,
  Dialog,
  AlertDialog,
  Image,
  Menu
} from '@raystack/apsara-v1';
import deleteIcon from '../../assets/delete.svg';
import { toastManager } from '@raystack/apsara-v1';
import { useFrontier } from '../../contexts/FrontierContext';
import { useOrganizationProjects } from '../../hooks/useOrganizationProjects';
import { usePermissions } from '../../hooks/usePermissions';
import { AuthTooltipMessage } from '../../utils';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { getColumns, type ProjectMenuPayload } from './components/project-columns';
import { AddProjectDialog } from './components/add-project-dialog';
import { EditProjectDialog, type EditProjectPayload } from './components/edit-project-dialog';
import { DeleteProjectDialog, type DeleteProjectPayload } from './components/delete-project-dialog';
import styles from './projects-view.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';

const projectMenuHandle = Menu.createHandle<ProjectMenuPayload>();
const addProjectDialogHandle = Dialog.createHandle();
const editProjectDialogHandle = Dialog.createHandle<EditProjectPayload>();
const deleteProjectDialogHandle = AlertDialog.createHandle<DeleteProjectPayload>();

export interface ProjectsViewProps {
  title?: string;
  description?: string;
  onProjectClick?: (projectId: string) => void;
}

export function ProjectsView({
  title = 'Projects',
  description,
  onProjectClick
}: ProjectsViewProps) {
  const {
    isFetching: isProjectsLoading,
    projects,
    userAccessOnProject,
    refetch,
    error: projectsError
  } = useOrganizationProjects({
    withMemberCount: true
  });

  const { activeOrganization: organization } = useFrontier();
  const t = useTerminology();

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

  const { canCreateProject } = useMemo(() => {
    return {
      canCreateProject: shouldShowComponent(
        permissions,
        `${PERMISSIONS.ProjectCreatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  useEffect(() => {
    if (projectsError) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          projectsError instanceof Error
            ? projectsError.message
            : 'Failed to load projects',
        type: 'error'
      });
    }
  }, [projectsError]);

  const isLoading = !organization?.id || isPermissionsFetching || isProjectsLoading;

  const columns = useMemo(
    () =>
      getColumns({
        userAccessOnProject,
        menuHandle: projectMenuHandle
      }),
    [userAccessOnProject]
  );

  return (
    <ViewContainer>
      <ViewHeader title={title} description={description ?? `Manage projects for this ${t.organization({ case: 'lower' })}`} />

      <DataTable
        data={projects ?? []}
        columns={columns}
        isLoading={isLoading}
        defaultSort={{ name: 'title', order: 'asc' }}
        mode="client"
        onRowClick={row => onProjectClick?.(row.id)}
      >
        <Flex direction="column" gap={7}>
          <Flex justify="between" gap={3}>
            <Flex gap={3} align="center">
              {isLoading ? (
                <Skeleton height="34px" width="360px" />
              ) : (
                <DataTable.Search
                  placeholder="Search by name"
                  size="large"
                  width={360}
                />
              )}
            </Flex>
            {isLoading ? (
              <Skeleton height="34px" width="120px" />
            ) : (
              <Tooltip>
                <Tooltip.Trigger
                  disabled={canCreateProject}
                  render={<span />}
                >
                  <Button
                    variant="solid"
                    color="accent"
                    onClick={() => addProjectDialogHandle.open(null)}
                    disabled={!canCreateProject}
                    data-test-id="frontier-sdk-add-project-button"
                  >
                    Add project
                  </Button>
                </Tooltip.Trigger>
                {!canCreateProject && (
                  <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
                )}
              </Tooltip>
            )}
          </Flex>
          <DataTable.Content
            emptyState={
              <EmptyState
                icon={<ExclamationTriangleIcon />}
                heading="No projects found"
                subHeading="Get started by creating your first project."
              />
            }
            classNames={{
              root: styles.tableRoot
            }}
          />
        </Flex>
      </DataTable>

      <Menu handle={projectMenuHandle} modal={false}>
        {({ payload: rawPayload }) => {
          const payload = rawPayload as ProjectMenuPayload | undefined;
          return (
            <Menu.Content align="end" className={styles.menuContent}>
              {payload?.canUpdate && (
                <Menu.Item
                  leadingIcon={<Pencil1Icon />}
                  onClick={() =>
                    editProjectDialogHandle.openWithPayload({
                      projectId: payload.projectId,
                      title: payload.title
                    })
                  }
                  data-test-id="edit-project-dropdown-item"
                >
                  Edit
                </Menu.Item>
              )}
              {payload?.canDelete && (
                <Menu.Item
                  leadingIcon={<Image src={deleteIcon as unknown as string} alt="Delete" width={16} height={16} />}
                  onClick={() =>
                    deleteProjectDialogHandle.openWithPayload({
                      projectId: payload.projectId
                    })
                  }
                  data-test-id="delete-project-dropdown-item"
                  style={{ color: 'var(--rs-color-foreground-danger-primary)' }}
                >
                  Delete project
                </Menu.Item>
              )}
            </Menu.Content>
          );
        }}
      </Menu>

      <AddProjectDialog handle={addProjectDialogHandle} refetch={refetch} />
      <EditProjectDialog handle={editProjectDialogHandle} refetch={refetch} />
      <DeleteProjectDialog handle={deleteProjectDialogHandle} refetch={refetch} />
    </ViewContainer>
  );
}
