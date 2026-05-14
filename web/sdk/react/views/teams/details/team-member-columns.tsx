'use client';

import {
    DotsHorizontalIcon,
    TrashIcon,
    UpdateIcon
} from '@radix-ui/react-icons';
import {
    toast,
    Label,
    Text,
    Flex,
    Avatar,
    DropdownMenu,
    type DataTableColumnDef,
    getAvatarColor
} from '@raystack/apsara';
import { PERMISSIONS, differenceWith, getInitials, isEqualById } from '~/utils';
import { useMutation } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    RemoveGroupUserRequestSchema,
    SetGroupMemberRoleRequestSchema,
    Role,
    User
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

interface getColumnsOptions {
    roles: Role[];
    organizationId: string;
    teamId: string;
    canUpdateGroup?: boolean;
    memberRoles?: Record<string, Role[]>;
    refetchMembers: () => void;
}

export const getColumns = ({
    roles = [],
    organizationId,
    teamId,
    canUpdateGroup = false,
    memberRoles = {},
    refetchMembers
}: getColumnsOptions): DataTableColumnDef<User, unknown>[] => [
        {
            header: '',
            accessorKey: 'avatar',
            enableSorting: false,
            styles: {
                cell: {
                    width: 'var(--rs-space-5)'
                }
            },
            cell: ({ row, getValue }) => {
                const color = getAvatarColor(row?.original?.id || '');
                return (
                    <Avatar
                        src={getValue() as string}
                        color={color}
                        fallback={getInitials(row.original?.title || row.original?.email)}
                        size={5}
                        radius="full"
                        style={{ marginRight: 'var(--rs-space-4)' }}
                    />
                );
            }
        },
        {
            header: 'Title',
            accessorKey: 'title',
            cell: ({ row, getValue }) => {
                return (
                    <Flex direction="column" gap={2}>
                        <Label style={{ fontWeight: '$500' }}>{getValue() as string}</Label>
                        <Text>{row.original.email}</Text>
                    </Flex>
                );
            }
        },
        {
            header: 'Roles',
            accessorKey: 'email',
            cell: ({ row }) => {
                return (
                    <Text>
                        {(row.original?.id &&
                            memberRoles[row.original?.id] &&
                            memberRoles[row.original?.id]
                                .map((r: any) => r.title || r.name)
                                .join(', ')) ??
                            'Inherited role'}
                    </Text>
                );
            }
        },
        {
            header: '',
            accessorKey: 'id',
            enableSorting: false,
            cell: ({ row }) => (
                <MembersActions
                    refetch={refetchMembers}
                    member={row.original as User}
                    organizationId={organizationId}
                    teamId={teamId}
                    canUpdateGroup={canUpdateGroup}
                    excludedRoles={differenceWith<Role>(
                        isEqualById,
                        roles,
                        row.original?.id && memberRoles[row.original?.id]
                            ? memberRoles[row.original?.id]
                            : []
                    )}
                />
            )
        }
    ];

const MembersActions = ({
    member,
    organizationId,
    teamId,
    canUpdateGroup,
    excludedRoles = [],
    refetch = () => null
}: {
    member: User;
    canUpdateGroup?: boolean;
    organizationId: string;
    teamId: string;
    excludedRoles: Role[];
    refetch: () => void;
}) => {
    // Remove group user using Connect RPC
    const removeGroupUserMutation = useMutation(
        FrontierServiceQueries.removeGroupUser,
        {
            onSuccess: () => {
                refetch();
                toast.success('Member deleted');
            },
            onError: error => {
                toast.error('Something went wrong', {
                    description: error.message
                });
            }
        }
    );

    function deleteMember() {
        const request = create(RemoveGroupUserRequestSchema, {
            id: teamId,
            orgId: organizationId,
            userId: member?.id as string
        });

        removeGroupUserMutation.mutate(request);
    }

    // Upsert the member role via SetGroupMemberRole RPC.
    const setGroupMemberRoleMutation = useMutation(
        FrontierServiceQueries.setGroupMemberRole,
        {
            onSuccess: () => {
                refetch();
                toast.success('Team member role updated');
            },
            onError: error => {
                toast.error('Something went wrong', {
                    description: error.message
                });
            }
        }
    );

    async function updateRole(role: Role) {
        try {
            const request = create(SetGroupMemberRoleRequestSchema, {
                groupId: teamId,
                orgId: organizationId,
                principalId: member?.id as string,
                principalType: PERMISSIONS.UserPrincipal,
                roleId: role.id as string
            });

            await setGroupMemberRoleMutation.mutateAsync(request);
        } catch (error: any) {
            toast.error('Something went wrong', {
                description: error?.message
            });
        }
    }

    return canUpdateGroup ? (
        <DropdownMenu placement="bottom-end">
            <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
                <DotsHorizontalIcon />
            </DropdownMenu.Trigger>
            {/* @ts-ignore */}
            <DropdownMenu.Content portal={false}>
                <DropdownMenu.Group>
                    {excludedRoles.map((role: Role) => (
                        <DropdownMenu.Item
                            key={role.id}
                            onClick={() => updateRole(role)}
                            data-test-id="frontier-sdk-update-team-member-role-btn"
                        >
                            <UpdateIcon />
                            Make {role.title}
                        </DropdownMenu.Item>
                    ))}
                    <DropdownMenu.Item
                        onClick={deleteMember}
                        data-test-id="frontier-sdk-remove-team-member-btn"
                    >
                        <TrashIcon />
                        Remove from team
                    </DropdownMenu.Item>
                </DropdownMenu.Group>
            </DropdownMenu.Content>
        </DropdownMenu>
    ) : null;
};

