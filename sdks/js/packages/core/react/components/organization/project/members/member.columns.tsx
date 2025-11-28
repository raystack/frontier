import { useEffect, useMemo } from 'react';
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
import { useNavigate } from '@tanstack/react-router';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListPoliciesRequestSchema,
  DeletePolicyRequestSchema,
  CreatePolicyRequestSchema,
  type Role,
  type User,
  type Group
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { differenceWith, getInitials, isEqualById } from '~/utils';

import teamIcon from '~/react/assets/users.svg';

type RowMember = (User & { isTeam?: false }) | (Group & { isTeam: true });

export const getColumns = (
  memberRoles: Record<string, Role[]> = {},
  groupRoles: Record<string, Role[]> = {},
  roles: Role[] = [],
  canUpdateProject: boolean,
  projectId: string,
  refetch: () => void
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
            ? // hardcoding roles as we dont have team roles and team are invited as viewer and we dont allow role change
              (row.original?.id &&
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
      />
    )
  }
];

const MembersActions = ({
  projectId,
  member,
  canUpdateProject,
  excludedRoles = [],
  refetch = () => null
}: {
  projectId: string;
  member: RowMember;
  canUpdateProject?: boolean;
  excludedRoles: Role[];
  refetch: () => void;
}) => {
  const navigate = useNavigate({ from: '/projects' });

  function removeMember() {
    navigate({
      to: '/projects/$projectId/$membertype/$memberId/remove',
      params: {
        projectId: projectId,
        membertype: member?.isTeam ? 'team' : 'user',
        memberId: member?.id as string
      }
    });
  }

  const { data: policiesData, refetch: refetchPolicies, error: policiesError } = useQuery(
    FrontierServiceQueries.listPolicies,
    create(ListPoliciesRequestSchema, {
      projectId: projectId,
      userId: member.isTeam ? undefined : (member.id as string),
      groupId: member.isTeam ? (member.id as string) : undefined
    }),
    { enabled: !!projectId && !!member?.id }
  );

  const policies = useMemo(() => policiesData?.policies ?? [], [policiesData]);

  const { mutateAsync: deletePolicy } = useMutation(FrontierServiceQueries.deletePolicy, {
    onError: (err: Error) =>
      toast.error('Something went wrong', { description: err.message })
  });
  const { mutateAsync: createPolicy } = useMutation(FrontierServiceQueries.createPolicy, {
    onError: (err: Error) =>
      toast.error('Something went wrong', { description: err.message })
  });

  useEffect(() => {
    if (policiesError) {
      toast.error('Something went wrong', { description: (policiesError as Error).message });
    }
  }, [policiesError]);

  async function updateRole(role: Role) {
    try {
      const resource = `app/project:${projectId}`;
      const principal = member.isTeam
        ? `app/group:${member?.id}`
        : `app/user:${member?.id}`;

      await Promise.all(
        (policies || []).map(p =>
          deletePolicy(create(DeletePolicyRequestSchema, { id: p.id || '' }))
        )
      );

      await createPolicy(
        create(CreatePolicyRequestSchema, {
          body: {
            roleId: role.id as string,
            title: (role.title || role.name) as string,
            resource,
            principal
          }
        })
      );
      await refetchPolicies();
      refetch();
      toast.success('Project member role updated');
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
