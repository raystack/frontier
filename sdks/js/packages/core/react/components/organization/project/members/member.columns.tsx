import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Avatar, Flex, Label, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { User } from '~/src/types';

export const columns: ColumnDef<User, any>[] = [
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
          src="https://assets.codepen.io/285131/github.svg"
          fallback="P"
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
        <Flex direction="column">
          <Label style={{ fontWeight: '$500' }}>{getValue()}</Label>
          <Text>{row.original.email}</Text>
        </Flex>
      );
    }
  },
  {
    accessorKey: 'email',
    cell: info => <Text>{info.getValue()}</Text>
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
