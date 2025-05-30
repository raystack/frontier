import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import {
  ApsaraColumnDef,
  DropdownMenu,
} from '@raystack/apsara';
import { useNavigate } from '@tanstack/react-router';
import { toast, Label, Flex, Avatar, Text } from '@raystack/apsara/v1';
import { useFrontier } from '~/react/contexts/FrontierContext';
import {
  V1Beta1Policy,
  V1Beta1Role,
} from '~/src';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import { MemberWithInvite } from '~/react/hooks/useOrganizationMembers';
import styles from '../organization.module.css';



export const getColumns = (
  organizationId: string,
  memberRoles: Record<string, V1Beta1Role[]> = {},
  roles: V1Beta1Role[] = [],
  canDeleteUser = false,
  refetch = () => {},
): ApsaraColumnDef<MemberWithInvite>[] => [
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
            row.original?.title || row.original?.email || row.original?.user_id
          )}
          size={5}
          radius="full"
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
        <Flex direction="column" gap={2}>
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
                  data-test-id={`update-role-${role?.name}-dropdown-item`}
                >
                  <UpdateIcon />
                  Make {role.title}
                </div>
              </DropdownMenu.Item>
            ))}

            <DropdownMenu.Item style={{ padding: 0 }}>
              <div
                onClick={() => navigate({ to: `/members/remove-member/$memberId/$invited`, params: {
                  memberId: member?.id || "",
                  invited: (member?.invited || false).toString()
                } 
              })}
                className={styles.dropdownActionItem}
                data-test-id="remove-member-dropdown-item"
              >
                <TrashIcon />
                Remove
              </div>
            </DropdownMenu.Item>
          </DropdownMenu.Group>
        </DropdownMenu.Content>
      </DropdownMenu>
    </>
  ) : null;
};
