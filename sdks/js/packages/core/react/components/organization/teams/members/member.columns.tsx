import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import { DropdownMenu } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import {
  toast,
  Label,
  Text,
  Flex,
  Avatar,
  type DataTableColumnDef,
  getAvatarColor
} from '@raystack/apsara/v1';
import { useFrontier } from '~/react/contexts/FrontierContext';
import type { V1Beta1Policy, V1Beta1Role, V1Beta1User } from '~/src';
import type { Role } from '~/src/types';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import styles from '../../organization.module.css';

interface getColumnsOptions {
  roles: V1Beta1Role[];
  organizationId: string;
  canUpdateGroup?: boolean;
  memberRoles?: Record<string, Role[]>;
  refetchMembers: () => void;
}

export const getColumns: (
  options: getColumnsOptions
) => DataTableColumnDef<V1Beta1User, unknown>[] = ({
  roles = [],
  organizationId,
  canUpdateGroup = false,
  memberRoles = {},
  refetchMembers
}) => [
  {
    header: '',
    accessorKey: 'avatar',
    enableSorting: false,
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
    cell: ({ row, getValue }) => {
      return (
        (row.original?.id &&
          memberRoles[row.original?.id] &&
          memberRoles[row.original?.id]
            .map((r: any) => r.title || r.name)
            .join(', ')) ??
        'Inherited role'
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
        member={row.original as V1Beta1User}
        organizationId={organizationId}
        canUpdateGroup={canUpdateGroup}
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
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const { client } = useFrontier();
  const navigate = useNavigate({ from: '/teams/$teamId' });

  async function deleteMember() {
    try {
      await client?.frontierServiceRemoveGroupUser(
        organizationId,
        teamId as string,
        member?.id as string
      );
      navigate({
        to: '/teams/$teamId',
        params: {
          teamId
        }
      });
      toast.success('Member deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/group:${teamId}`;
      const principal = `app/user:${member?.id}`;
      const {
        // @ts-ignore
        data: { policies = [] }
      } = await client?.frontierServiceListPolicies({
        groupId: teamId,
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
      toast.success('Team member role updated');
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    }
  }
  return canUpdateGroup ? (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
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
              Remove from team
            </div>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
