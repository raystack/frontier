'use client';

import { useMemo } from 'react';
import {
    DotsHorizontalIcon,
    TrashIcon,
    UpdateIcon
} from '@radix-ui/react-icons';
import {
    Avatar,
    Label,
    Text,
    Flex,
    toast,
    DropdownMenu,
    type DataTableColumnDef,
    getAvatarColor
} from '@raystack/apsara';
import { useMutation } from '@connectrpc/connect-query';
import {
    FrontierServiceQueries,
    SetProjectMemberRoleRequestSchema,
    type Role,
    type User,
    type Group
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

const PRINCIPAL_TYPES = {
    USER: 'app/user',
    GROUP: 'app/group',
} as const;
import { differenceWith, getInitials, isEqualById } from '~/utils';

import teamIcon from '~/react/assets/users.svg';

type RowMember = (User & { isTeam?: false }) | (Group & { isTeam: true });

export const getColumns = (
    memberRoles: Record<string, Role[]> = {},
    groupRoles: Record<string, Role[]> = {},
    roles: Role[] = [],
    canUpdateProject: boolean,
    projectId: string,
    refetch: () => void,
    onRemoveMember?: (memberId: string, memberType: 'user' | 'group') => void
): DataTableColumnDef<RowMember, unknown>[] => [
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
            const avatarSrc = row.original?.isTeam ? teamIcon : getValue();
            const fallback = row.original?.isTeam
                ? ''
                : getInitials(row.original?.title || row.original?.email);
            const color = getAvatarColor(row?.original?.id || '');
            return (
                <Avatar
                    src={avatarSrc as string}
                    color={color}
                    fallback={fallback}
                    size={5}
                    radius="small"
                    style={{ marginRight: 'var(--rs-space-4)' }}
                />
            );
        }
    },
    {
        header: 'Title',
        accessorKey: 'title',
        cell: ({ row, getValue }) => {
            const label = row.original?.isTeam
                ? row.original.title
                : (getValue() as string);
            const subLabel = row.original?.isTeam
                ? row.original.name
                : row.original.email;

            return (
                <Flex direction="column" gap={2}>
                    <Label style={{ fontWeight: '$500' }}>{label}</Label>
                    <Text>{subLabel}</Text>
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
                    {row.original?.isTeam
                        ? (row.original?.id &&
                            groupRoles[row.original?.id] &&
                            groupRoles[row.original?.id]
                                .map((r: Role) => r.title || r.name)
                                .join(', ')) ??
                        'Project Viewer'
                        : (row.original?.id &&
                            memberRoles[row.original?.id] &&
                            memberRoles[row.original?.id]
                                .map((r: Role) => r.title || r.name)
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
                refetch={refetch}
                projectId={projectId}
                member={row.original as RowMember}
                canUpdateProject={canUpdateProject}
                excludedRoles={differenceWith<Role>(
                    isEqualById,
                    roles,
                    row.original.isTeam
                        ? row.original?.id && groupRoles[row.original?.id]
                            ? groupRoles[row.original?.id]
                            : []
                        : row.original?.id && memberRoles[row.original?.id]
                            ? memberRoles[row.original?.id]
                            : []
                )}
                onRemoveMember={onRemoveMember}
            />
        )
    }
];

const MembersActions = ({
    projectId,
    member,
    canUpdateProject,
    excludedRoles = [],
    refetch = () => null,
    onRemoveMember
}: {
    projectId: string;
    member: RowMember;
    canUpdateProject?: boolean;
    excludedRoles: Role[];
    refetch: () => void;
    onRemoveMember?: (memberId: string, memberType: 'user' | 'group') => void;
}) => {
    function removeMember() {
        onRemoveMember?.(member?.id as string, member?.isTeam ? 'group' : 'user');
    }

    const { mutateAsync: setMemberRole } = useMutation(
        FrontierServiceQueries.setProjectMemberRole,
        {
            onSuccess: () => {
                refetch();
                toast.success('Project member role updated');
            },
            onError: (err: Error) =>
                toast.error('Something went wrong', { description: err.message })
        }
    );

    async function updateRole(role: Role) {
        try {
            await setMemberRole(
                create(SetProjectMemberRoleRequestSchema, {
                    projectId: projectId,
                    principalId: member?.id as string,
                    principalType: member.isTeam ? PRINCIPAL_TYPES.GROUP : PRINCIPAL_TYPES.USER,
                    roleId: role.id as string
                })
            );
        } catch (err) {
            const message = (err as Error)?.message || 'Failed to update role';
            toast.error('Something went wrong', { description: message });
        }
    }

    return canUpdateProject ? (
        <DropdownMenu placement="bottom-end">
            <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
                <DotsHorizontalIcon />
            </DropdownMenu.Trigger>
            {/* @ts-ignore */}
            <DropdownMenu.Content portal={false}>
                <DropdownMenu.Group style={{ padding: 0 }}>
                    {excludedRoles.map((role: Role) => (
                        <DropdownMenu.Item
                            key={role.id}
                            onClick={() => updateRole(role)}
                            data-test-id="frontier-sdk-update-project-member-role-btn"
                        >
                            <UpdateIcon />
                            Make {role.title}
                        </DropdownMenu.Item>
                    ))}
                    <DropdownMenu.Item
                        data-test-id="frontier-sdk-remove-project-member-btn"
                        onClick={() => removeMember()}
                    >
                        <TrashIcon />
                        Remove from project
                    </DropdownMenu.Item>
                </DropdownMenu.Group>
            </DropdownMenu.Content>
        </DropdownMenu>
    ) : null;
};
