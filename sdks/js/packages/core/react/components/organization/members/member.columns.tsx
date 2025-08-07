import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import { useNavigate } from '@tanstack/react-router';
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
import { useFrontier } from '~/react/contexts/FrontierContext';
import type { V1Beta1Policy, V1Beta1Role } from '~/src';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import styles from '../organization.module.css';
import type { MemberWithInvite } from '~/react/hooks/useOrganizationMembers';

export const getColumns = (
  organizationId: string,
  memberRoles: Record<string, V1Beta1Role[]> = {},
  roles: V1Beta1Role[] = [],
  canDeleteUser = false,
  refetch = () => {}
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
        row.original?.title || row.original?.email || row.original?.user_id; // user_id will be email in invitations
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
      const email = row.original.invited
        ? row.original.user_id
        : row.original.email;
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
                  .map((r: V1Beta1Role) => r.title || r.name)
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
        excludedRoles={differenceWith<V1Beta1Role>(
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
  canUpdateGroup,
  excludedRoles = [],
  refetch = () => null
}: {
  member: MemberWithInvite;
  canUpdateGroup?: boolean;
  organizationId: string;
  excludedRoles: V1Beta1Role[];
  refetch: () => void;
}) => {
  const { client } = useFrontier();
  const navigate = useNavigate({ from: '/members' });

  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/organization:${organizationId}`;
      const principal = `app/user:${member?.id}`;
      const {
        // @ts-ignore
        data: { policies = [] }
      } = await client?.frontierServiceListPolicies({
        org_id: organizationId,
        user_id: member.id
      });
      const deletePromises = policies.map((p: V1Beta1Policy) =>
        client?.frontierServiceDeletePolicy(p.id as string)
      );

      await Promise.all(deletePromises);
      await client?.frontierServiceCreatePolicy({
        role_id: role.id as string,
        title: role.name as string,
        resource: resource,
        principal: principal
      });
      refetch();
      toast.success('Member role updated');
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    }
  }

  return canUpdateGroup ? (
    <>
      <DropdownMenu placement="bottom-end">
        <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
          <DotsHorizontalIcon />
        </DropdownMenu.Trigger>
        {/* @ts-ignore */}
        <DropdownMenu.Content portal={false}>
          <DropdownMenu.Group style={{ padding: 0 }}>
            {excludedRoles.map((role: V1Beta1Role) => (
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
                navigate({
                  to: `/members/remove-member/$memberId/$invited`,
                  params: {
                    memberId: member?.id || '',
                    invited: (member?.invited || false).toString()
                  }
                })
              }
              data-test-id="remove-member-dropdown-item"
            >
              <TrashIcon />
              Remove
            </DropdownMenu.Item>
          </DropdownMenu.Group>
        </DropdownMenu.Content>
      </DropdownMenu>
    </>
  ) : null;
};
