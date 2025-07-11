import React from 'react';
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
} from '@raystack/apsara/v1';
import { useNavigate } from '@tanstack/react-router';
import teamIcon from '~/react/assets/users.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import type {
  V1Beta1Group,
  V1Beta1Policy,
  V1Beta1Role,
  V1Beta1User
} from '~/src';
import type { Role } from '~/src/types';
import { differenceWith, getInitials, isEqualById } from '~/utils';

type ColumnType = V1Beta1User & (V1Beta1Group & { isTeam?: boolean });

const teamAvatarStyles: React.CSSProperties = {
  height: '32px',
  width: '32px',
  padding: '6px',
  boxSizing: 'border-box',
  color: 'var(--rs-color-foreground-base-primary)'
};

export const getColumns = (
  memberRoles: Record<string, Role[]> = {},
  groupRoles: Record<string, Role[]> = {},
  roles: V1Beta1Role[] = [],
  canUpdateProject: boolean,
  projectId: string,
  refetch: () => void
): DataTableColumnDef<ColumnType, unknown>[] => [
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
      const imageProps = row.original?.isTeam ? teamAvatarStyles : {};
      const color = getAvatarColor(row?.original?.id || '');
      return (
        <Avatar
          src={avatarSrc as string}
          color={color}
          fallback={fallback}
          size={5}
          radius="small"
          imageProps={imageProps}
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
    cell: ({ row, getValue }) => {
      return (
        <Text>
          {row.original?.isTeam
            ? // hardcoding roles as we dont have team roles and team are invited as viewer and we dont allow role change
              (row.original?.id &&
                groupRoles[row.original?.id] &&
                groupRoles[row.original?.id]
                  .map((r: any) => r.title || r.name)
                  .join(', ')) ??
                'Project Viewer'
            : (row.original?.id &&
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
        refetch={refetch}
        projectId={projectId}
        member={row.original as V1Beta1User & { isTeam: boolean }}
        canUpdateProject={canUpdateProject}
        excludedRoles={differenceWith<V1Beta1Role>(
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
  member: V1Beta1User & { isTeam: boolean };
  canUpdateProject?: boolean;
  excludedRoles: V1Beta1Role[];
  refetch: () => void;
}) => {
  const { client } = useFrontier();
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

  async function updateRole(role: V1Beta1Role) {
    try {
      const resource = `app/project:${projectId}`;
      const principal = member.isTeam
        ? `app/group:${member?.id}`
        : `app/user:${member?.id}`;
      const {
        // @ts-ignore
        data: { policies = [] }
      } = await client?.frontierServiceListPolicies({
        project_id: projectId,
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
      toast.success('Project member role updated');
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
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
          {excludedRoles.map((role: V1Beta1Role) => (
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
