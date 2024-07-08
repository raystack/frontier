import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import {
  ApsaraColumnDef,
  Avatar,
  DropdownMenu,
  Flex,
  Label,
  Text
} from '@raystack/apsara';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import {
  V1Beta1Invitation,
  V1Beta1Policy,
  V1Beta1Role,
  V1Beta1User
} from '~/src';
import { Role } from '~/src/types';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import styles from '../organization.module.css';

export const getColumns: (
  id: string,
  memberRoles: Record<string, Role[]>,
  roles: Role[],
  canDeleteUser?: boolean,
  refetch?: () => void
) => ApsaraColumnDef<
  V1Beta1User & V1Beta1Invitation & { invited?: boolean }
>[] = (
  organizationId,
  memberRoles = {},
  roles = [],
  canDeleteUser = false,
  refetch = () => null
) => [
  {
    header: '',
    accessorKey: 'avatar',
    size: 44,
    meta: {
      style: {
        width: '30px',
        padding: 0
      }
    },
    enableSorting: false,
    cell: ({ row, getValue }) => {
      return (
        <Avatar
          src={getValue()}
          fallback={getInitials(
            row.original?.title ||
              row.original?.email ||
              // @ts-ignore
              row.original?.user_id
          )}
          // @ts-ignore
          style={{ marginRight: 'var(--mr-12)', zIndex: -1 }}
        />
      );
    }
  },
  {
    header: 'Title',
    accessorKey: 'title',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: ({ row, getValue }) => {
      return (
        <Flex direction="column" gap="extra-small">
          <Label style={{ fontWeight: '$500' }}>{getValue()}</Label>
          <Text>
            {row.original.invited
              ? // @ts-ignore
                row.original.user_id
              : row.original.email}
          </Text>
        </Flex>
      );
    }
  },
  {
    header: 'Roles',
    accessorKey: 'email',
    cell: ({ row, getValue }) => {
      return row.original.invited
        ? 'Pending Invite'
        : (row.original?.id &&
            memberRoles[row.original?.id] &&
            memberRoles[row.original?.id]
              .map((r: any) => r.title || r.name)
              .join(', ')) ??
            'Inherited role';
    }
  },
  {
    header: '',
    accessorKey: 'id',
    meta: {
      style: {
        textAlign: 'end'
      }
    },
    enableSorting: false,
    cell: ({ row }) => (
      <MembersActions
        refetch={refetch}
        member={row.original as V1Beta1User}
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
  member: V1Beta1User;
  canUpdateGroup?: boolean;
  organizationId: string;
  excludedRoles: V1Beta1Role[];
  refetch: () => void;
}) => {
  const { client } = useFrontier();
  const navigate = useNavigate({ from: '/members' });

  async function deleteMember() {
    try {
      // @ts-ignore
      if (member?.invited) {
        await client?.frontierServiceDeleteOrganizationInvitation(
          // @ts-ignore
          member.org_id,
          member?.id as string
        );
      } else {
        await client?.frontierServiceRemoveOrganizationUser(
          organizationId,
          member?.id as string
        );
      }
      navigate({ to: '/members' });
      toast.success('Member deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }
  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/organization:${organizationId}`;
      const principal = `app/user:${member?.id}`;
      const {
        // @ts-ignore
        data: { policies = [] }
      } = await client?.frontierServiceListPolicies({
        orgId: organizationId,
        userId: member.id
      });
      const deletePromises = policies.map((p: V1Beta1Policy) =>
        client?.frontierServiceDeletePolicy(p.id as string)
      );

      await Promise.all(deletePromises);
      await client?.frontierServiceCreatePolicy({
        roleId: role.id as string,
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
    <DropdownMenu style={{ padding: '0 !important' }}>
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group style={{ padding: 0 }}>
          {excludedRoles.map((role: V1Beta1Role) => (
            <DropdownMenu.Item style={{ padding: 0 }} key={role.id}>
              <div
                onClick={() => updateRole(role)}
                className={styles.dropdownActionItem}
              >
                <UpdateIcon />
                Make {role.title}
              </div>
            </DropdownMenu.Item>
          ))}

          <DropdownMenu.Item style={{ padding: 0 }}>
            <div onClick={deleteMember} className={styles.dropdownActionItem}>
              <TrashIcon />
              Remove
            </div>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
