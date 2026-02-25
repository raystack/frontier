'use client';

import { useEffect, useMemo, useState } from 'react';
import { Tabs, Image, Flex, toast } from '@raystack/apsara';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS } from '~/utils';
import { TeamGeneral } from './team-general';
import { TeamMembers } from './team-members';
import { DeleteTeamDialog } from './delete-team-dialog';
import { InviteTeamMemberDialog } from './invite-team-member-dialog';
import { useQuery } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    GetGroupRequestSchema,
    ListGroupUsersRequestSchema,
    ListRolesRequestSchema,
    Organization,
    type Role
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { PageHeader } from '~/react/components/common/page-header';
import sharedStyles from '../../../components/organization/styles.module.css';
import styles from './team-detail.module.css';

export interface TeamDetailPageProps {
    teamId: string;
    onBack?: () => void;
}

export const TeamDetailPage = ({ teamId, onBack }: TeamDetailPageProps) => {
    const { activeOrganization: organization } = useFrontier();

    const [showDeleteDialog, setShowDeleteDialog] = useState(false);
    const [showInviteDialog, setShowInviteDialog] = useState(false);

    // Get team details using Connect RPC
    const {
        data: teamData,
        isLoading: isTeamLoading,
        error: teamError
    } = useQuery(
        FrontierServiceQueries.getGroup,
        create(GetGroupRequestSchema, {
            id: teamId || '',
            orgId: organization?.id || '',
            withMembers: true
        }),
        { enabled: !!organization?.id && !!teamId }
    );

    const team = teamData?.group;

    // Handle team error
    useEffect(() => {
        if (teamError) {
            toast.error('Something went wrong', {
                description: teamError.message
            });
        }
    }, [teamError]);

    // Get team members using Connect RPC
    const {
        data: membersData,
        isLoading: isMembersLoading,
        error: membersError,
        refetch: refetchMembers
    } = useQuery(
        FrontierServiceQueries.listGroupUsers,
        create(ListGroupUsersRequestSchema, {
            id: teamId || '',
            orgId: organization?.id || '',
            withRoles: true
        }),
        { enabled: !!organization?.id && !!teamId }
    );

    const members = membersData?.users || [];
    const memberRoles = useMemo(() => {
        if (!membersData?.rolePairs) return {};
        return membersData.rolePairs.reduce(
            (
                previous: Record<string, Role[]>,
                mr: { userId: string; roles: Role[] }
            ) => {
                return { ...previous, [mr.userId]: mr.roles };
            },
            {}
        );
    }, [membersData?.rolePairs]);

    // Handle members error
    useEffect(() => {
        if (membersError) {
            toast.error('Something went wrong', {
                description: membersError.message
            });
        }
    }, [membersError]);

    // Get team roles using Connect RPC
    const { data: rolesData, error: rolesError } = useQuery(
        FrontierServiceQueries.listRoles,
        create(ListRolesRequestSchema, {
            state: 'enabled',
            scopes: [PERMISSIONS.GroupNamespace]
        }),
        { enabled: !!organization?.id && !!teamId }
    );

    const roles = rolesData?.roles || [];

    // Handle roles error
    useEffect(() => {
        if (rolesError) {
            toast.error('Something went wrong', {
                description: rolesError.message
            });
        }
    }, [rolesError]);

    const handleDeleteOpenChange = (value: boolean) => {
        setShowDeleteDialog(value);
    };

    const handleInviteOpenChange = (value: boolean) => {
        setShowInviteDialog(value);
        if (!value) refetchMembers();
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
                            data-test-id="frontier-sdk-team-back-btn"
                        />
                        <PageHeader
                            title="Team"
                            description="Manage team settings and members."
                        />
                    </Flex>
                </Flex>
                <Tabs defaultValue="general" className={styles.container}>
                    <Tabs.List>
                        <Tabs.Trigger value="general">General</Tabs.Trigger>
                        <Tabs.Trigger value="members">Members</Tabs.Trigger>
                    </Tabs.List>
                    <Tabs.Content value="general">
                        <TeamGeneral
                            organization={organization as Organization}
                            team={team}
                            teamId={teamId}
                            isLoading={isTeamLoading}
                            onDeleteClick={() => setShowDeleteDialog(true)}
                        />
                    </Tabs.Content>
                    <Tabs.Content value="members" className={styles.tabContent}>
                        <TeamMembers
                            members={members}
                            roles={roles}
                            memberRoles={memberRoles}
                            organizationId={organization?.id || ''}
                            teamId={teamId}
                            isLoading={isMembersLoading}
                            refetchMembers={refetchMembers}
                            onInviteClick={() => setShowInviteDialog(true)}
                        />
                    </Tabs.Content>
                </Tabs>
            </Flex>
            <DeleteTeamDialog
                open={showDeleteDialog}
                onOpenChange={handleDeleteOpenChange}
                teamId={teamId}
                onDeleteSuccess={onBack}
            />
            <InviteTeamMemberDialog
                open={showInviteDialog}
                onOpenChange={handleInviteOpenChange}
                teamId={teamId}
            />
        </Flex>
    );
};

