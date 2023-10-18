import { Avatar, Flex, Label, Text } from '@raystack/apsara';
import { ColumnDef } from '@tanstack/react-table';
import { V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { getInitials } from '~/utils';

export const getColumns = (
  memberRoles?: Record<string, Role[]>
): ColumnDef<V1Beta1User, any>[] => [
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
    cell: ({ row, getValue }) => {
      return (
        <Avatar
          src={getValue()}
          fallback={getInitials(row.original?.name)}
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
    cell: ({ row, getValue }) => {
      return (
        <Flex direction="column" gap="extra-small">
          <Label style={{ fontWeight: '$500' }}>{getValue()}</Label>
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
        (memberRoles[row.original?.id] &&
          memberRoles[row.original?.id].map((r: any) => r.name).join(', ')) ??
        'Inherited role'
      );
    }
  }
];
