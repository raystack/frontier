import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import { useState } from 'react';
import {
  toast,
  Label,
  Flex,
  Avatar,
  Text,
  getAvatarColor,
  DropdownMenu,
  DataTableColumnDef
} from '@raystack/apsara';
import type { Role } from '@raystack/proton/frontier';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  SetOrganizationMemberRoleRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import type { MemberWithInvite } from '~/react/hooks/useOrganizationMembers';
import { MembersTableType } from './member-types';

export const getColumns = (
  organizationId: string,
  memberRoles: Record<string, Role[]> = {},
  roles: Role[] = [],
  canDeleteUser = false,
  refetch = () => {},
  onRemoveMember?: MembersTableType['onRemoveMember']
): DataTableColumnDef<MemberWithInvite, MemberWithInvite>[] => [
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
      const id = row.original?.id || '';
      const fallback =
        row.original?.title || row.original?.email || row.original?.userId; // userId will be email in invitations
      return (
        <Avatar
          src={getValue() as string}
          fallback={getInitials(fallback)}
          color={getAvatarColor(id)}
          size={5}
          radius="full"
        />
      );
    }
  },
  {
    header: 'Title',
    accessorKey: 'title',
    cell: ({ row, getValue }) => {
      const title = getValue() as string;
      const email = row.original.email;
      return (
        <Flex direction="column" gap={2}>
          <Label style={{ fontWeight: '$500' }}>{title}</Label>
          <Text>{email}</Text>
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
          {row.original.invited
            ? 'Pending Invite'
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
        member={row.original}
        organizationId={organizationId}
        canUpdateGroup={canDeleteUser}
        excludedRoles={differenceWith<Role>(
          isEqualById,
          roles,
          row.original?.id && memberRoles[row.original?.id]
            ? memberRoles[row.original?.id]
            : []
        )}
        onRemoveMember={onRemoveMember}
      />
    )
  }
];

const MembersActions = ({
  member,
  organizationId,
  canUpdateGroup,
  excludedRoles = [],
  refetch = () => null,
  onRemoveMember
}: {
  member: MemberWithInvite;
  canUpdateGroup?: boolean;
  organizationId: string;
  excludedRoles: Role[];
  refetch: () => void;
  onRemoveMember?: MembersTableType['onRemoveMember'];
}) => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const { mutateAsync: setMemberRole } = useMutation(
    FrontierServiceQueries.setOrganizationMemberRole,
    {
      onSuccess: () => {
        refetch();
        toast.success('Member role updated');
      },
      onError: (error: any) => {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to update member role'
        });
      }
    }
  );

  async function updateRole(role: Role) {
    try {
      const req = create(SetOrganizationMemberRoleRequestSchema, {
        orgId: organizationId,
        userId: member?.id,
        roleId: role.id as string
      });
      await setMemberRole(req);
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message || 'Failed to update member role'
      });
    }
  }

  return canUpdateGroup ? (
    <DropdownMenu
      placement="bottom-end"
      open={isMenuOpen}
      onOpenChange={setIsMenuOpen}
    >
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      {/* @ts-ignore */}
      <DropdownMenu.Content portal={false}>
        <DropdownMenu.Group style={{ padding: 0 }}>
          {!member.invited &&
            excludedRoles.map((role: Role) => (
              <DropdownMenu.Item
                key={role.id}
                onClick={() => updateRole(role)}
                data-test-id={`update-role-${role?.name}-dropdown-item`}
              >
                <UpdateIcon />
                Make {role.title}
              </DropdownMenu.Item>
            ))}
          <DropdownMenu.Item
            onClick={() =>
              onRemoveMember?.(
                member?.id || '',
                String(member?.invited || false)
              )
            }
            data-test-id="remove-member-dropdown-item"
          >
            <TrashIcon />
            Remove
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
