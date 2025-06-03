import React from 'react';
import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import {
  ApsaraColumnDef,
  DropdownMenu,
} from '@raystack/apsara';
import { Avatar, Label, Text, Flex, toast } from '@raystack/apsara/v1';
import { useNavigate } from '@tanstack/react-router';
import teamIcon from '~/react/assets/users.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Policy, V1Beta1Role, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import styles from '../../organization.module.css';

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
): ApsaraColumnDef<ColumnType>[] => [
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
      const avatarSrc = row.original?.isTeam ? teamIcon : getValue();
      const fallback = row.original?.isTeam
        ? ''
        : getInitials(row.original?.title || row.original?.email);
      const imageProps = row.original?.isTeam ? teamAvatarStyles : {};
      return (
        <Avatar
          src={avatarSrc}
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
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: ({ row, getValue }) => {
      const label = row.original?.isTeam ? row.original.title : getValue();
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
      return row.original?.isTeam
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
    <DropdownMenu style={{ padding: '0 !important' }}>
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group style={{ padding: 0 }}>
          {excludedRoles.map((role: V1Beta1Role) => (
            <DropdownMenu.Item style={{ padding: 0 }} key={role.id}>
              <div
                data-test-id="frontier-sdk-update-project-member-role-btn"
                onClick={() => updateRole(role)}
                className={styles.dropdownActionItem}
              >
                <UpdateIcon />
                Make {role.title}
              </div>
            </DropdownMenu.Item>
          ))}
          <DropdownMenu.Item style={{ padding: 0 }}>
            <div
              data-test-id="frontier-sdk-remove-project-member-btn"
              className={styles.dropdownActionItem}
              onClick={() => removeMember()}
            >
              <TrashIcon />
              Remove from project
            </div>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
