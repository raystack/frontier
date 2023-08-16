import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Flex, Link, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { V1Beta1Domain } from '~/src';

export const columns: ColumnDef<V1Beta1Domain, any>[] = [
  {
    accessorKey: 'name',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: ({ row, getValue }) => {
      return (
        <Flex direction="column">
          <Text>{row.original.name}</Text>
        </Flex>
      );
    }
  },
  {
    accessorKey: 'created_at',
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
    cell: ({ row, getValue }) => (
      <ProjectActions domain={row.original as V1Beta1Domain} />
    )
  }
];

const ProjectActions = ({ domain }: { domain: V1Beta1Domain }) => {
  return (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
          <DropdownMenu.Item style={{ padding: 0 }}>
            <Link
              to={`/domains`}
              style={{
                gap: 'var(--pd-8)',
                display: 'flex',
                alignItems: 'center',
                textDecoration: 'none',
                color: 'var(--foreground-base)',
                padding: 'var(--pd-8)'
              }}
            >
              <Pencil1Icon /> verify domain
            </Link>
          </DropdownMenu.Item>
          <DropdownMenu.Item style={{ padding: 0 }}>
            <Link
              to={`/domains`}
              style={{
                gap: 'var(--pd-8)',
                display: 'flex',
                alignItems: 'center',
                textDecoration: 'none',
                color: 'var(--foreground-base)',
                padding: 'var(--pd-8)'
              }}
            >
              <TrashIcon /> Delete domain
            </Link>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
};
