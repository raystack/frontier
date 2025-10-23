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
import type { V1Beta1Policy, V1Beta1Role } from '~/src';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, DeletePolicyRequestSchema, CreatePolicyRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
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
      const email = row.original.invited
        ? row.original.userId
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
  const navigate = useNavigate({ from: '/members' });
  
  // Query to fetch policies for the current member
  const { data: policiesData, refetch: refetchPolicies } = useQuery(
    FrontierServiceQueries.listPolicies,
    { orgId: organizationId, userId: member.id },
    { enabled: !!member.id }
  );
  
  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy,
    {
      onError: (error: any) => {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to delete policy'
        });
      },
    }
  );
  
  const { mutateAsync: createPolicy } = useMutation(
    FrontierServiceQueries.createPolicy,
    {
      onSuccess: () => {
        refetch();
        toast.success('Member role updated');
      },
      onError: (error: any) => {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to create policy'
        });
      },
    }
  );

  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/organization:${organizationId}`;
      const principal = `app/user:${member?.id}`;
      
      // Use policies from Connect RPC query
      const policies = policiesData?.policies || [];
      
      // Delete existing policies with individual error handling
      const deleteResults = await Promise.allSettled(
        policies.map((p: V1Beta1Policy) => {
          const req = create(DeletePolicyRequestSchema, {
            id: p.id as string
          });
          return deletePolicy(req);
        })
      );
      
      // Check for delete errors
      const deleteErrors = deleteResults
        .filter((result): result is PromiseRejectedResult => result.status === 'rejected')
        .map(result => result.reason);
      
      if (deleteErrors.length > 0) {
        console.warn('Some policy deletions failed:', deleteErrors);
        // Continue with creation even if some deletes failed
      }
      
      // Create new policy
      const createReq = create(CreatePolicyRequestSchema, {
        body: {
          roleId: role.id as string,
          title: role.name as string,
          resource: resource,
          principal: principal
        }
      });
      await createPolicy(createReq);
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message || 'Failed to update member role'
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
