'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import {
    Button,
    toast,
    Skeleton,
    Image,
    Text,
    Flex,
    Dialog,
    Select,
    Label
} from '@raystack/apsara';
import { useCallback, useEffect, useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS, filterUsersfromUsers } from '~/utils';
import cross from '~/react/assets/cross.svg';
import { handleSelectValueChange } from '~/react/utils';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    CreatePolicyRequestSchema,
    AddGroupUsersRequestSchema,
    ListOrganizationUsersRequestSchema,
    ListGroupUsersRequestSchema,
    ListOrganizationRolesRequestSchema,
    ListRolesRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import orgStyles from '../../../components/organization/organization.module.css';

const inviteSchema = yup.object({
    userId: yup.string().required('Member is required'),
    role: yup.string().required('Role is required')
});

type InviteSchemaType = yup.InferType<typeof inviteSchema>;

export interface InviteTeamMemberDialogProps {
    open: boolean;
    onOpenChange: (value: boolean) => void;
    teamId: string;
}

export const InviteTeamMemberDialog = ({
    open,
    onOpenChange,
    teamId
}: InviteTeamMemberDialogProps) => {
    const { activeOrganization: organization } = useFrontier();

    const {
        control,
        handleSubmit,
        formState: { errors, isSubmitting },
        reset
    } = useForm({
        resolver: yupResolver(inviteSchema)
    });

    // Reset form when dialog closes
    const handleOpenChange = (value: boolean) => {
        if (!value) {
            reset?.();
        }
        onOpenChange?.(value);
    };

    // Get organization members using Connect RPC
    const {
        data: orgMembersData,
        isLoading: isOrgMembersLoading,
        error: orgMembersError
    } = useQuery(
        FrontierServiceQueries.listOrganizationUsers,
        create(ListOrganizationUsersRequestSchema, {
            id: organization?.id || ''
        }),
        { enabled: !!organization?.id && open }
    );

    // Handle organization members error
    useEffect(() => {
        if (orgMembersError) {
            toast.error('Something went wrong', {
                description: orgMembersError.message
            });
        }
    }, [orgMembersError]);

    // Get team members using Connect RPC
    const {
        data: teamMembersData,
        isLoading: isTeamMembersLoading,
        error: teamMembersError
    } = useQuery(
        FrontierServiceQueries.listGroupUsers,
        create(ListGroupUsersRequestSchema, {
            id: teamId || '',
            orgId: organization?.id || '',
            withRoles: true
        }),
        { enabled: !!organization?.id && !!teamId && open }
    );

    // Handle team members error
    useEffect(() => {
        if (teamMembersError) {
            toast.error('Something went wrong', {
                description: teamMembersError.message
            });
        }
    }, [teamMembersError]);

    // Get organization roles using Connect RPC
    const {
        data: orgRolesData,
        isLoading: isOrgRolesLoading,
        error: orgRolesError
    } = useQuery(
        FrontierServiceQueries.listOrganizationRoles,
        create(ListOrganizationRolesRequestSchema, {
            orgId: organization?.id || '',
            scopes: [PERMISSIONS.GroupNamespace]
        }),
        { enabled: !!organization?.id && open }
    );

    // Get roles using Connect RPC
    const {
        data: rolesData,
        isLoading: isRolesLoading,
        error: rolesError
    } = useQuery(
        FrontierServiceQueries.listRoles,
        create(ListRolesRequestSchema, {
            scopes: [PERMISSIONS.GroupNamespace]
        }),
        { enabled: !!organization?.id && open }
    );

    const roles = useMemo(() => {
        const orgRoles = orgRolesData?.roles || [];
        const systemRoles = rolesData?.roles || [];
        return [...systemRoles, ...orgRoles];
    }, [orgRolesData?.roles, rolesData?.roles]);

    const isRolesLoadingCombined = isOrgRolesLoading || isRolesLoading;

    // Handle roles errors
    useEffect(() => {
        if (orgRolesError) {
            toast.error('Something went wrong', {
                description: orgRolesError.message
            });
        }
    }, [orgRolesError]);

    useEffect(() => {
        if (rolesError) {
            toast.error('Something went wrong', {
                description: rolesError.message
            });
        }
    }, [rolesError]);

    // Create policy using Connect RPC
    const createPolicyMutation = useMutation(
        FrontierServiceQueries.createPolicy,
        {
            onError: error => {
                toast.error('Something went wrong', {
                    description: error.message
                });
            }
        }
    );

    const addGroupTeamPolicy = useCallback(
        async (roleId: string, userId: string) => {
            const role = roles.find(r => r.id === roleId);
            if (role?.name && role.name !== PERMISSIONS.RoleGroupMember) {
                const resource = `${PERMISSIONS.GroupPrincipal}:${teamId}`;
                const principal = `${PERMISSIONS.UserPrincipal}:${userId}`;

                const request = create(CreatePolicyRequestSchema, {
                    body: {
                        roleId: roleId,
                        resource,
                        principal
                    }
                });

                await createPolicyMutation.mutateAsync(request);
            }
        },
        [roles, teamId, createPolicyMutation]
    );

    // Add group users using Connect RPC
    const addGroupUsersMutation = useMutation(
        FrontierServiceQueries.addGroupUsers,
        {
            onError: error => {
                toast.error('Something went wrong', {
                    description: error.message
                });
            }
        }
    );

    async function onSubmit({ role, userId }: InviteSchemaType) {
        if (!userId || !role || !organization?.id) return;

        const request = create(AddGroupUsersRequestSchema, {
            id: teamId as string,
            orgId: organization.id,
            userIds: [userId]
        });

        await addGroupUsersMutation.mutateAsync(request);
        await addGroupTeamPolicy(role, userId);
        toast.success('member added');
        handleOpenChange(false);
    }

    const invitableUser = useMemo(
        () =>
            filterUsersfromUsers(
                orgMembersData?.users || [],
                teamMembersData?.users || []
            ) || [],
        [orgMembersData?.users, teamMembersData?.users]
    );

    const isUserLoading = isOrgMembersLoading || isTeamMembersLoading;

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <Dialog.Content
                style={{ padding: 0, maxWidth: '600px', width: '100%' }}
                overlayClassName={orgStyles.overlay}
            >
                <Dialog.Header>
                    <Flex justify="between" align="center" style={{ width: '100%' }}>
                        <Text size="large" weight="medium">
                            Add Member
                        </Text>

                        <Image
                            alt="cross"
                            style={{ cursor: 'pointer' }}
                            src={cross as unknown as string}
                            onClick={() => handleOpenChange(false)}
                            data-test-id="frontier-sdk-invite-team-members-close-btn"
                        />
                    </Flex>
                </Dialog.Header>
                <Dialog.Body>
                    <form onSubmit={handleSubmit(onSubmit)}>
                        <Flex direction="column" gap={5}>
                            <Flex direction="column" gap={2}>
                                <Label>Members</Label>
                                {isUserLoading ? (
                                    <Skeleton height={'25px'} />
                                ) : (
                                    <Controller
                                        render={({ field: { onChange, ref, ...rest } }) => (
                                            <Select
                                                {...rest}
                                                onValueChange={handleSelectValueChange(onChange)}
                                            >
                                                <Select.Trigger ref={ref}>
                                                    <Select.Value placeholder="Select members" />
                                                </Select.Trigger>
                                                <Select.Content
                                                    style={{ width: '100% !important' }}
                                                >
                                                    <Select.Viewport style={{ maxHeight: '300px' }}>
                                                        <Select.Group>
                                                            {!invitableUser.length && (
                                                                <Text className={orgStyles.noSelectItem}>
                                                                    No member to invite
                                                                </Text>
                                                            )}
                                                            {invitableUser.map(user => (
                                                                <Select.Item
                                                                    value={user.id || ''}
                                                                    key={user.id}
                                                                >
                                                                    {user.title || user.email}
                                                                </Select.Item>
                                                            ))}
                                                        </Select.Group>
                                                    </Select.Viewport>
                                                </Select.Content>
                                            </Select>
                                        )}
                                        control={control}
                                        name="userId"
                                    />
                                )}
                                <Text size="mini" variant="danger">
                                    {errors.userId && String(errors.userId?.message)}
                                </Text>
                            </Flex>
                            <Flex direction="column" gap={2}>
                                <Label>Invite as</Label>
                                {isRolesLoadingCombined ? (
                                    <Skeleton height={'25px'} />
                                ) : (
                                    <Controller
                                        render={({ field: { onChange, ref, ...rest } }) => (
                                            <Select
                                                {...rest}
                                                onValueChange={handleSelectValueChange(onChange)}
                                            >
                                                <Select.Trigger ref={ref}>
                                                    <Select.Value placeholder="Select a role" />
                                                </Select.Trigger>
                                                <Select.Content
                                                    style={{ width: '100% !important' }}
                                                >
                                                    <Select.Group>
                                                        {!roles.length && (
                                                            <Text className={orgStyles.noSelectItem}>
                                                                No roles available
                                                            </Text>
                                                        )}
                                                        {roles.map(role => (
                                                            <Select.Item value={role.id} key={role.id}>
                                                                {role.title || role.name}
                                                            </Select.Item>
                                                        ))}
                                                    </Select.Group>
                                                </Select.Content>
                                            </Select>
                                        )}
                                        control={control}
                                        name="role"
                                    />
                                )}
                                <Text size="mini" variant="danger">
                                    {errors.role && String(errors.role?.message)}
                                </Text>
                            </Flex>
                            <Flex justify="end">
                                <Button
                                    type="submit"
                                    data-test-id="frontier-sdk-add-team-members-btn"
                                    disabled={isUserLoading || isRolesLoadingCombined}
                                    loading={isSubmitting}
                                    loaderText="Adding..."
                                >
                                    Add Member
                                </Button>
                            </Flex>
                        </Flex>
                    </form>
                </Dialog.Body>
            </Dialog.Content>
        </Dialog>
    );
};

