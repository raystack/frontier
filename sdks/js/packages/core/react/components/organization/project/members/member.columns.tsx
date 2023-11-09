import { Avatar, Flex, Label, Text } from '@raystack/apsara';
import { ColumnDef } from '@tanstack/react-table';
import Skeleton from 'react-loading-skeleton';
import { V1Beta1Group, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { getInitials } from '~/utils';
import teamIcon from '~/react/assets/users.svg';

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
  isLoading: boolean
): ColumnDef<ColumnType, any>[] => [
  {
    header: '',
    accessorKey: 'image',
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
              'Project Viewer'
            : (row.original?.id &&
                memberRoles[row.original?.id] &&
                memberRoles[row.original?.id]
                  .map((r: any) => r.title || r.name)
                  .join(', ')) ??
                'Inherited role';
        }
  }
];
