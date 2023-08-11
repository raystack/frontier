import {
  DotsHorizontalIcon,
  Pencil1Icon,
  PlusIcon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Flex, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { User } from '~/src/types';

export const columns: ColumnDef<User, any>[] = [
  {
    header: 'Title',
    accessorKey: 'name',
    cell: info => <Text>{info.getValue()}</Text>
  },
  {
    accessorKey: 'members',
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
    cell: info => (
      <DropdownMenu>
        <DropdownMenu.Trigger asChild>
          <DotsHorizontalIcon />
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="end">
          <DropdownMenu.Group>
            <DropdownMenu.Item>
              <Flex align="center" gap="small">
                <Pencil1Icon /> Rename
              </Flex>
            </DropdownMenu.Item>
            <DropdownMenu.Item>
              <Flex align="center" gap="small">
                <PlusIcon /> Add a member
              </Flex>
            </DropdownMenu.Item>
            <DropdownMenu.Item>
              <Flex align="center" gap="small">
                <TrashIcon /> Delete team
              </Flex>
            </DropdownMenu.Item>
          </DropdownMenu.Group>
        </DropdownMenu.Content>
      </DropdownMenu>
    )
  }
];
