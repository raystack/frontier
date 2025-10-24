import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import { useNavigate, useParams } from '@tanstack/react-router';
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
import type { V1Beta1Policy, V1Beta1Role, V1Beta1User } from '~/src';
import type { Role } from '~/src/types';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, RemoveGroupUserRequestSchema, DeletePolicyRequestSchema, CreatePolicyRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

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
  const navigate = useNavigate({ from: '/teams/$teamId' });

  // Remove group user using Connect RPC
  const removeGroupUserMutation = useMutation(FrontierServiceQueries.removeGroupUser, {
    onSuccess: () => {
      navigate({
        to: '/teams/$teamId',
        params: {
          teamId
        }
      });
      toast.success('Member deleted');
    },
    onError: (error) => {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  });

  function deleteMember() {
    const request = create(RemoveGroupUserRequestSchema, {
      id: teamId as string,
      orgId: organizationId,
      userId: member?.id as string
    });

    removeGroupUserMutation.mutate(request);
  }

  // Get policies using Connect RPC
  const { refetch: refetchPolicies } = useQuery(
    FrontierServiceQueries.listPolicies,
    { groupId: teamId as string, userId: member?.id as string },
    { enabled: false } // Only fetch when needed
  );

  // Delete policy using Connect RPC
  const deletePolicyMutation = useMutation(FrontierServiceQueries.deletePolicy, {
    onError: (error) => {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  });

  // Create policy using Connect RPC
  const createPolicyMutation = useMutation(FrontierServiceQueries.createPolicy, {
    onSuccess: () => {
      refetch();
      toast.success('Team member role updated');
    },
    onError: (error) => {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  });

  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/group:${teamId}`;
      const principal = `app/user:${member?.id}`;
      
      // Get policies using Connect RPC
      const policiesResponse = await refetchPolicies();
      const policies = policiesResponse?.data?.policies || [];

      // Delete existing policies
      const deletePromises = policies.map((p: V1Beta1Policy) => {
        const deleteRequest = create(DeletePolicyRequestSchema, {
          id: p.id as string
        });
        return deletePolicyMutation.mutateAsync(deleteRequest);
      });

      await Promise.all(deletePromises);
      
      // Create new policy
      const createRequest = create(CreatePolicyRequestSchema, {
        body: {
          roleId: role.id as string,
          title: role.name as string,
          resource: resource,
          principal: principal
        }
      });
      
      await createPolicyMutation.mutateAsync(createRequest);
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
          {excludedRoles.map((role: V1Beta1Role) => (
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
