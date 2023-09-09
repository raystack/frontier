import { Avatar, Flex, Label, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { V1Beta1User } from '~/src';
import { getInitials } from '~/utils';

export const columns: ColumnDef<V1Beta1User, any>[] = [
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
    accessorKey: 'email',
    cell: info => <Text>{info.getValue()}</Text>
  }
];
