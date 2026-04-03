'use client';

import {
    CardStackPlusIcon,
    PlusIcon,
    ExclamationTriangleIcon
} from '@radix-ui/react-icons';
import type React from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
    Button,
    EmptyState,
    Tooltip,
    toast,
    Separator,
    Avatar,
    Skeleton,
    Text,
    Flex,
    DataTable,
    Popover,
    Search
} from '@raystack/apsara';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationTeams } from '~/react/hooks/useOrganizationTeams';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import {
    PERMISSIONS,
    filterUsersfromUsers,
    getInitials,
    shouldShowComponent
} from '~/utils';
import { getColumns } from './project-member-columns';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    ListOrganizationUsersRequestSchema,
    SetProjectMemberRoleRequestSchema,
    type Group,
    type User,
    type Role
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import styles from './project-members.module.css';

export type ProjectMembersProps = {
    projectId: string;
    teams?: Group[];
    members?: User[];
    roles?: Role[];
    memberRoles?: Record<string, Role[]>;
    groupRoles?: Record<string, Role[]>;
    isLoading?: boolean;
    refetch: () => void;
    onRemoveMember?: (memberId: string) => void;
};

export const ProjectMembers = ({
    projectId,
    teams = [],
    members = [],
    roles = [],
    memberRoles,
    groupRoles,
    isLoading: isMemberLoading,
    refetch,
    onRemoveMember
}: ProjectMembersProps) => {
    const resource = `app/project:${projectId}`;
    const listOfPermissionsToCheck = useMemo(
        () => [
            {
                permission: PERMISSIONS.UpdatePermission,
                resource
            }
        ],
        [resource]
    );

    const { permissions, isFetching: isPermissionsFetching } = usePermissions(
        listOfPermissionsToCheck,
        !!projectId
    );

    const { canUpdateProject } = useMemo(() => {
        return {
            canUpdateProject: shouldShowComponent(
                permissions,
                `${PERMISSIONS.UpdatePermission}::${resource}`
            )
        };
    }, [permissions, resource]);

    const isLoading = isMemberLoading || isPermissionsFetching;

    const columns = useMemo(
        () =>
            getColumns(
                memberRoles,
                groupRoles,
                roles,
                canUpdateProject,
                projectId,
                refetch,
                onRemoveMember
            ),
        [memberRoles, groupRoles, roles, canUpdateProject, projectId, refetch, onRemoveMember]
    );

    const updatedUsers = useMemo(() => {
        const updatedTeams = teams.map(t => ({ ...t, isTeam: true }));
        return members?.length || updatedTeams?.length
            ? [...updatedTeams, ...members]
            : [];
    }, [members, teams]);

    return (
        <Flex direction="column" className={styles.container}>
            <DataTable
                isLoading={isLoading}
                data={updatedUsers}
                columns={columns}
                defaultSort={{ name: 'name', order: 'asc' }}
                mode="client"
            >
                <Flex direction="column" gap={7} className={styles.tableWrapper}>
                    <Flex justify="between" gap={3}>
                        <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
                            <DataTable.Search
                                placeholder="Search by name or email"
                                size="large"
                            />
                        </Flex>
                        {isLoading ? (
                            <Skeleton height={'32px'} width={'64px'} />
                        ) : (
                            <Tooltip
                                message={AuthTooltipMessage}
                                side="left"
                                disabled={canUpdateProject}
                            >
                                <AddMemberDropdown
                                    projectId={projectId}
                                    canUpdateProject={canUpdateProject}
                                    refetch={refetch}
                                    members={members}
                                />
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
        </Flex>
    );
};

interface AddMemberDropdownProps {
    projectId: string;
    canUpdateProject: boolean;
    members?: User[];
    refetch?: () => void;
}

const AddMemberDropdown = ({
    projectId,
    canUpdateProject,
    members,
    refetch
}: AddMemberDropdownProps) => {
    const [query, setQuery] = useState('');
    const [showTeam, setShowTeam] = useState(false);

    const { activeOrganization: organization } = useFrontier();
    const { isFetching: isTeamsLoading, teams } = useOrganizationTeams({});

    const toggleShowTeam = (e: React.MouseEvent<HTMLElement>) => {
        e.preventDefault();
        setQuery('');
        setShowTeam(prev => !prev);
    };

    const { data: orgUsersData, isLoading: isOrgUsersLoading, error: orgUsersError } = useQuery(
        FrontierServiceQueries.listOrganizationUsers,
        create(ListOrganizationUsersRequestSchema, { id: organization?.id || '' }),
        { enabled: !!organization?.id && canUpdateProject }
    );

    const orgUsersResp = useMemo(() => orgUsersData?.users ?? [], [orgUsersData]);

    useEffect(() => {
        if (orgUsersError) {
            toast.error('Something went wrong', { description: orgUsersError.message });
        }
    }, [orgUsersError]);

    const invitableUser = useMemo(
        () => filterUsersfromUsers(orgUsersResp || [], members) || [],
        [orgUsersResp, members]
    );

    const topUsers = useMemo(
        () =>
            invitableUser
                .filter(user =>
                    query
                        ? user.title?.toLowerCase().includes(query.toLowerCase()) ||
                        user.email?.includes(query)
                        : true
                )
                .slice(0, 7),
        [invitableUser, query]
    );

    const topTeams = useMemo(() =>
        teams
            .filter(team =>
                query
                    ? team.title && team.title.toLowerCase().includes(query.toLowerCase())
                    : true
            )
            .slice(0, 7),
        [query, teams]);

    function onTextChange(e: React.ChangeEvent<HTMLInputElement>) {
        setQuery(e.target.value);
    }

    const viewerRole = useMemo(
        () => (roles as Role[]).find((r: Role) => r.name === PERMISSIONS.RoleProjectViewer),
        [roles]
    );

    const { mutate: setProjectMemberRole, isPending: isCreatingPolicy } = useMutation(
        FrontierServiceQueries.setProjectMemberRole,
        {
            onSuccess: () => {
                toast.success('Member added');
                if (refetch) refetch();
            },
            onError: (err: Error) => {
                toast.error('Something went wrong', { description: err.message });
            }
        }
    );

    const addMember = useCallback(
        (userId: string) => {
            if (!userId || !organization?.id || !projectId || !viewerRole?.id) return;
            setProjectMemberRole(
                create(SetProjectMemberRoleRequestSchema, {
                    projectId: projectId,
                    principalId: userId,
                    principalType: PERMISSIONS.UserNamespace,
                    roleId: viewerRole.id
                })
            );
        },
        [setProjectMemberRole, organization?.id, projectId, viewerRole]
    );

    const addTeam = useCallback(
        (teamId: string) => {
            if (!teamId || !organization?.id || !projectId || !viewerRole?.id) return;
            setProjectMemberRole(
                create(SetProjectMemberRoleRequestSchema, {
                    projectId: projectId,
                    principalId: teamId,
                    principalType: PERMISSIONS.GroupNamespace,
                    roleId: viewerRole.id
                })
            );
        },
        [setProjectMemberRole, organization?.id, projectId, viewerRole]
    );

    return (
        <Popover>
            <Popover.Trigger asChild>
                <Button
                    size="normal"
                    style={{ width: 'fit-content', display: 'flex' }}
                    data-test-id="frontier-sdk-add-project-member-btn"
                    disabled={!canUpdateProject}
                >
                    Add a member
                </Button>
            </Popover.Trigger>
            <Popover.Content align="end" className={styles.popoverContent}>
                <Search
                    data-test-id="frontier-sdk-add-project-project-textfield"
                    value={query}
                    placeholder={showTeam ? 'Add team to project' : 'Add project member'}
                    onChange={onTextChange}
                    variant="borderless"
                    showClearButton
                    disabled={isTeamsLoading || isOrgUsersLoading}
                    onClear={() => setQuery('')}
                />
                <Separator />

                {showTeam ? (
                    isTeamsLoading ? (
                        <Skeleton height={'32px'} />
                    ) : topTeams.length ? (
                        <div style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}>
                            {topTeams.map((team, i) => {
                                const initals = getInitials(team?.title || team.name);
                                return (
                                    <Flex
                                        gap="small"
                                        key={team.id}
                                        onClick={() => addTeam(team?.id || '')}
                                        className={styles.inviteDropdownItem}
                                        data-test-id={`frontier-sdk-add-team-to-project-dropdown-item-${i}`}
                                    >
                                        <Avatar
                                            fallback={initals}
                                            size={1}
                                            radius="small"
                                        />
                                        <Text>{team?.title || team?.name}</Text>
                                    </Flex>
                                );
                            })}
                        </div>
                    ) : (
                        <Flex
                            style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}
                            justify={'center'}
                            align={'center'}
                        >
                            <Text size="small">No Teams found</Text>
                        </Flex>
                    )
                ) : isOrgUsersLoading ? (
                    <Skeleton height={'32px'} />
                ) : topUsers.length ? (
                    <div style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}>
                        {topUsers.map((user, i) => {
                            const initals = getInitials(user?.title || user.email);
                            return (
                                <Flex
                                    gap="small"
                                    key={user.id}
                                    className={styles.inviteDropdownItem}
                                    onClick={() => addMember(user?.id || '')}
                                    data-test-id={`frontier-sdk-add-user-to-project-dropdown-item-${i}`}
                                >
                                    <Avatar
                                        src={user?.avatar}
                                        fallback={initals}
                                        size={1}
                                        radius="small"
                                    />
                                    <Text>{user?.title || user?.email}</Text>
                                </Flex>
                            );
                        })}
                    </div>
                ) : (
                    <Flex
                        style={{ padding: 'var(--rs-space-2)', minHeight: '246px' }}
                        justify={'center'}
                        align={'center'}
                    >
                        <Text size="small">No Users found</Text>
                    </Flex>
                )}
                <Separator />

                <div style={{ padding: 'var(--rs-space-2)' }}>
                    <Flex
                        onClick={toggleShowTeam}
                        gap="small"
                        className={styles.inviteDropdownItem}
                        data-test-id={`frontier-sdk-add-project-member-toggle`}
                    >
                        {showTeam ? (
                            <>
                                <PlusIcon color="var(--rs-color-foreground-base-primary)" />{' '}
                                <Text>Add project member</Text>
                            </>
                        ) : (
                            <>
                                <CardStackPlusIcon color="var(--rs-color-foreground-base-primary)" />{' '}
                                <Text>Add team to project</Text>
                            </>
                        )}
                    </Flex>
                </div>
            </Popover.Content>
        </Popover>
    );
};

const noDataChildren = (
    <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="No members found"
        subHeading="Get started by adding your first member."
    />
);

