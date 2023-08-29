import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Avatar, Flex, Label, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { User } from '~/src/types';
import { getInitials } from '~/utils';

export const columns: ColumnDef<User, any>[] = [
  {
    header: '',
    accessorKey: 'profile_picture',
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
    cell: ({ row, getValue }) => {
      return <Text>{getValue() || row.original?.user_id}</Text>;
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
    cell: info => <DotsHorizontalIcon />
  }
];
