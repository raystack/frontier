'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
    Tooltip,
    EmptyState,
    Skeleton,
    Flex,
    Button,
    Select,
    DataTable,
    toast
} from '@raystack/apsara';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationProjects } from '~/react/hooks/useOrganizationProjects';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './projects-columns';
import { Project } from '@raystack/proton/frontier';
import { PageHeader } from '~/react/components/common/page-header';
import { AddProjectDialog } from './add-project-dialog';
import { DeleteProjectDialog } from '../details/delete-project-dialog';
import sharedStyles from '../../../components/organization/styles.module.css';
import styles from './projects.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';

const projectsSelectOptions = [
    { value: 'my-projects', label: 'My Projects' },
    { value: 'all-projects', label: 'All Projects' }
];

interface ProjectsTableProps {
    projects: Project[];
    isLoading?: boolean;
    canCreateProject?: boolean;
    userAccessOnProject: Record<string, string[]>;
    canListOrgProjects?: boolean;
    onOrgProjectsFilterChange?: (value: string) => void;
    onProjectClick?: (projectId: string) => void;
    onDeleteProjectClick?: (projectId: string) => void;
    onAddProjectClick?: () => void;
}

export interface ProjectsListPageProps {
    title?: string;
    description?: string;
    onProjectClick?: (projectId: string) => void;
}

export function ProjectsListPage({
    title,
    description,
    onProjectClick
}: ProjectsListPageProps = {}) {
    const [showOrgProjects, setShowOrgProjects] = useState(false);
    const t = useTerminology();

    const {
        isFetching: isProjectsLoading,
        projects,
        userAccessOnProject,
        refetch,
        error: projectsError
    } = useOrganizationProjects({
        allProjects: showOrgProjects,
        withMemberCount: true
    });
    const { activeOrganization: organization } = useFrontier();

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
        if (projectsError) {
            toast.error('Something went wrong', {
                description: projectsError.message
            });
        }
    }, [projectsError]);

    const isLoading = isPermissionsFetching || isProjectsLoading;

    const [showAddProjectDialog, setShowAddProjectDialog] = useState(false);
    const [deleteProjectState, setDeleteProjectState] = useState({
        open: false,
        projectId: ''
    });

    const handleAddProjectOpenChange = (value: boolean) => {
        setShowAddProjectDialog(value);
        refetch();
    };

    const handleDeleteProjectOpenChange = (value: boolean) => {
        if (!value) {
            setDeleteProjectState({ open: false, projectId: '' });
            refetch();
        } else {
            setDeleteProjectState(prev => ({ ...prev, open: value }));
        }
    };

    return (
        <Flex direction="column" className={sharedStyles.pageWrapper}>
            <Flex
                direction="column"
                className={`${sharedStyles.container} ${sharedStyles.containerFlex}`}
            >
                <Flex
                    direction="row"
                    justify="between"
                    align="center"
                    className={sharedStyles.header}
                >
                    <PageHeader
                        title={title || 'Projects'}
                        description={description || `Manage projects in this ${t.organization({
                            case: 'lower'
                        })}.`}
                    />
                </Flex>
                <Flex
                    direction="column"
                    gap={9}
                    className={sharedStyles.contentWrapper}
                >
                    <ProjectsTable
                        projects={projects}
                        isLoading={isLoading}
                        canCreateProject={canCreateProject}
                        userAccessOnProject={userAccessOnProject}
                        canListOrgProjects={canUpdateOrganization}
                        onOrgProjectsFilterChange={onOrgProjectsFilterChange}
                        onProjectClick={onProjectClick}
                        onDeleteProjectClick={(projectId) =>
                            setDeleteProjectState({ open: true, projectId })
                        }
                        onAddProjectClick={() => setShowAddProjectDialog(true)}
                    />
                </Flex>
            </Flex>
            <AddProjectDialog
                open={showAddProjectDialog}
                onOpenChange={handleAddProjectOpenChange}
            />
            <DeleteProjectDialog
                open={deleteProjectState.open}
                onOpenChange={handleDeleteProjectOpenChange}
                projectId={deleteProjectState.projectId}
            />
        </Flex>
    );
}

const ProjectsTable = ({
    projects,
    isLoading,
    canCreateProject,
    userAccessOnProject,
    canListOrgProjects,
    onOrgProjectsFilterChange,
    onProjectClick,
    onDeleteProjectClick,
    onAddProjectClick
}: ProjectsTableProps) => {
    const columns = useMemo(
        () => getColumns(userAccessOnProject, onProjectClick, onDeleteProjectClick),
        [userAccessOnProject, onProjectClick, onDeleteProjectClick]
    );

    return (
        <DataTable
            data={projects ?? []}
            isLoading={isLoading}
            defaultSort={{ name: 'title', order: 'asc' }}
            columns={columns}
            mode="client"
        >
            <Flex direction="column" gap={7} className={styles.tableWrapper}>
                <Flex justify="between" gap={3}>
                    <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
                        {isLoading ? (
                            <Skeleton height="34px" width="500px" />
                        ) : (
                            <DataTable.Search placeholder="Search by title " size="large" />
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
                                onClick={onAddProjectClick}
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
        heading="No projects found"
        subHeading="Get started by creating your first project."
    />
);

