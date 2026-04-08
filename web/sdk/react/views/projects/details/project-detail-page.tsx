'use client';

import { Tabs, Image, toast, Flex } from '@raystack/apsara';
import { useCallback, useEffect, useMemo, useState } from 'react';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useQuery } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    ListProjectGroupsRequestSchema,
    ListProjectUsersRequestSchema,
    GetProjectRequestSchema,
    ListRolesRequestSchema,
    type Role as ProtoRole,
    type Organization
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { PERMISSIONS } from '~/utils';
import { ProjectGeneral } from './project-general';
import { ProjectMembers } from './project-members';
import { DeleteProjectDialog } from './delete-project-dialog';
import { RemoveProjectMemberDialog } from './remove-project-member-dialog';
import { PageHeader } from '~/react/components/common/page-header';
import sharedStyles from '../../../components/organization/styles.module.css';
import styles from './project-detail.module.css';

interface ProjectGroupRolePair {
    groupId?: string;
    roles: ProtoRole[];
}

interface ProjectUserRolePair {
    userId?: string;
    roles: ProtoRole[];
}

export interface ProjectDetailPageProps {
    projectId: string;
    onBack?: () => void;
}

export const ProjectDetailPage = ({
    projectId,
    onBack
}: ProjectDetailPageProps) => {
    const { activeOrganization: organization } = useFrontier();

    const [deleteProjectState, setDeleteProjectState] = useState({ open: false });
    const [removeMemberState, setRemoveMemberState] = useState({
        open: false,
        memberId: '',
        memberType: 'user' as 'user' | 'group'
    });

    const {
        data: projectGroupsData,
        isLoading: isTeamsLoading,
        error: projectGroupsError,
        refetch: refetchProjectGroups
    } = useQuery(
        FrontierServiceQueries.listProjectGroups,
        create(ListProjectGroupsRequestSchema, {
            id: projectId || '',
            withRoles: true
        }),
        {
            enabled: !!organization?.id && !!projectId
        }
    );

    const projectGroups = useMemo(
        () => ({
            groups: projectGroupsData?.groups ?? [],
            groupRoles: (projectGroupsData?.rolePairs ?? []).reduce(
                (acc: Record<string, ProtoRole[]>, gr: ProjectGroupRolePair) => {
                    const key = gr.groupId;
                    if (key) acc[key] = gr.roles;
                    return acc;
                },
                {}
            )
        }),
        [projectGroupsData]
    );

    useEffect(() => {
        if (projectGroupsError) {
            toast.error('Something went wrong', {
                description: projectGroupsError.message
            });
        }
    }, [projectGroupsError]);

    const {
        data: projectUsersData,
        isLoading: isMembersLoadingQuery,
        refetch: refetchProjectUsers
    } = useQuery(
        FrontierServiceQueries.listProjectUsers,
        create(ListProjectUsersRequestSchema, {
            id: projectId || '',
            withRoles: true
        }),
        {
            enabled: !!organization?.id && !!projectId
        }
    );

    const projectUsers = useMemo(
        () => ({
            users: projectUsersData?.users ?? [],
            memberRoles: (projectUsersData?.rolePairs ?? []).reduce(
                (acc: Record<string, ProtoRole[]>, mr: ProjectUserRolePair) => {
                    const key = mr.userId;
                    if (key) acc[key] = mr.roles;
                    return acc;
                },
                {}
            )
        }),
        [projectUsersData]
    );

    const {
        data: project,
        isLoading: isProjectLoadingQuery,
        error: projectError
    } = useQuery(
        FrontierServiceQueries.getProject,
        create(GetProjectRequestSchema, { id: projectId || '' }),
        {
            enabled: !!organization?.id && !!projectId,
            select: d => d?.project
        }
    );

    useEffect(() => {
        if (projectError) {
            toast.error('Something went wrong', {
                description: projectError.message
            });
        }
    }, [projectError]);

    const {
        data: rolesData,
        isLoading: isProjectRoleLoadingQuery,
        error: rolesError
    } = useQuery(
        FrontierServiceQueries.listRoles,
        create(ListRolesRequestSchema, {
            state: 'enabled',
            scopes: [PERMISSIONS.ProjectNamespace]
        }),
        {
            enabled: !!organization?.id && !!projectId
        }
    );

    const roles = useMemo(() => rolesData?.roles ?? [], [rolesData]);

    useEffect(() => {
        if (rolesError) {
            toast.error('Something went wrong', {
                description: rolesError.message
            });
        }
    }, [rolesError]);

    const isLoading =
        isProjectLoadingQuery ||
        isTeamsLoading ||
        isMembersLoadingQuery ||
        isProjectRoleLoadingQuery;

    const refetchTeamAndMembers = useCallback(() => {
        refetchProjectUsers();
        refetchProjectGroups();
    }, [refetchProjectUsers, refetchProjectGroups]);

    const handleDeleteProjectOpenChange = (value: boolean) => {
        setDeleteProjectState({ open: value });
    };

    const handleRemoveMemberOpenChange = (value: boolean) => {
        if (!value) {
            setRemoveMemberState({ open: false, memberId: '' });
            refetchTeamAndMembers();
        } else {
            setRemoveMemberState(prev => ({ ...prev, open: value }));
        }
    };

    const onRemoveMember = (memberId: string, memberType: 'user' | 'group' = 'user') => {
        setRemoveMemberState({ open: true, memberId, memberType });
    };

    return (
        <Flex direction="column" style={{ width: '100%' }}>
            <Flex direction="column" className={sharedStyles.container}>
                <Flex
                    direction="row"
                    justify="between"
                    align="center"
                    className={sharedStyles.header}
                >
                    <Flex gap={3} align="center">
                        <Image
                            alt="back-icon"
                            style={{ cursor: 'pointer' }}
                            src={backIcon as unknown as string}
                            onClick={onBack}
                            data-test-id="frontier-sdk-projects-page-back-link"
                        />
                        <PageHeader
                            title="Project"
                            description="Manage project settings and members."
                        />
                    </Flex>
                </Flex>
                <Tabs defaultValue="general" className={styles.container}>
                    <Tabs.List>
                        <Tabs.Trigger value="general">General</Tabs.Trigger>
                        <Tabs.Trigger value="members">Members</Tabs.Trigger>
                    </Tabs.List>
                    <Tabs.Content value="general">
                        <ProjectGeneral
                            projectId={projectId}
                            organization={organization as Organization}
                            project={project}
                            isLoading={isProjectLoadingQuery}
                            onDeleteClick={() => handleDeleteProjectOpenChange(true)}
                        />
                    </Tabs.Content>
                    <Tabs.Content value="members" className={styles.tabContent}>
                        <ProjectMembers
                            projectId={projectId}
                            members={projectUsers.users}
                            memberRoles={projectUsers.memberRoles}
                            groupRoles={projectGroups.groupRoles}
                            isLoading={isLoading}
                            teams={projectGroups.groups}
                            roles={roles}
                            refetch={refetchTeamAndMembers}
                            onRemoveMember={onRemoveMember}
                        />
                    </Tabs.Content>
                </Tabs>
            </Flex>
            <DeleteProjectDialog
                open={deleteProjectState.open}
                onOpenChange={handleDeleteProjectOpenChange}
                projectId={projectId}
                onDeleteSuccess={onBack}
            />
            <RemoveProjectMemberDialog
                open={removeMemberState.open}
                onOpenChange={handleRemoveMemberOpenChange}
                projectId={projectId}
                memberId={removeMemberState.memberId}
                memberType={removeMemberState.memberType}
            />
        </Flex>
    );
};

