import { Flex, Label, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { User } from '~/src/types';

export const columns: ColumnDef<User, any>[] = [
  {
    accessorKey: 'name',
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
    accessorKey: 'privacy',
    cell: info => <Text>{info.getValue()}</Text>
  },
  {
    accessorKey: 'members',
    cell: info => <Text>{info.getValue()}</Text>
  }
];
