import { DotsHorizontalIcon, UpdateIcon } from '@radix-ui/react-icons';
import { Avatar, DropdownMenu, Flex, Label, Text } from '@raystack/apsara';
import { useNavigate } from '@tanstack/react-router';
import { ColumnDef } from '@tanstack/react-table';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import teamIcon from '~/react/assets/users.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Policy, V1Beta1Role, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { differenceWith, getInitials, isEqualById } from '~/utils';
import styles from '../../organization.module.css';

type ColumnType = V1Beta1User & (V1Beta1Group & { isTeam: boolean });

const teamAvatarStyles: React.CSSProperties = {
  height: '32px',
  width: '32px',
  padding: '6px',
  boxSizing: 'border-box',
  color: 'var(--foreground-base)'
};

export const getColumns = (
  memberRoles: Record<string, Role[]> = {},
  groupRoles: Record<string, Role[]> = {},
  roles: V1Beta1Role[] = [],
  canUpdateProject: boolean,
  isLoading: boolean,
  projectId: string,
  refetch: () => void
): ColumnDef<ColumnType, any>[] => [
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const avatarSrc = row.original?.isTeam ? teamIcon : getValue();
          const fallback = row.original?.isTeam
            ? ''
            : getInitials(row.original?.title || row.original?.email);
          const imageProps = row.original?.isTeam ? teamAvatarStyles : {};
          return (
            <Avatar
              src={avatarSrc}
              fallback={fallback}
              shape={'square'}
              imageProps={imageProps}
              // @ts-ignore
              style={{ marginRight: 'var(--mr-12)' }}
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          const label = row.original?.isTeam ? row.original.title : getValue();
          const subLabel = row.original?.isTeam
            ? row.original.name
            : row.original.email;

          return (
            <Flex direction="column" gap="extra-small">
              <Label style={{ fontWeight: '$500' }}>{label}</Label>
              <Text>{subLabel}</Text>
            </Flex>
          );
        }
  },
  {
    header: 'Roles',
    accessorKey: 'email',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
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
    cell: isLoading
      ? () => <Skeleton />
      : ({ row }) => (
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
        projectId: projectId,
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
                onClick={() => updateRole(role)}
                className={styles.dropdownActionItem}
              >
                <UpdateIcon />
                Make {role.title}
              </div>
            </DropdownMenu.Item>
          ))}
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
