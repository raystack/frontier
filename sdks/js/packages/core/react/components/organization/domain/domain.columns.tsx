import {
  DotsHorizontalIcon,
  Pencil1Icon,
  TrashIcon
} from '@radix-ui/react-icons';
import { DropdownMenu, Flex, Link, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
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
      <DomainActions domain={row.original as V1Beta1Domain} />
    )
  }
];

const DomainActions = ({ domain }: { domain: V1Beta1Domain }) => {
  const { client } = useFrontier();
  const navigate = useNavigate();

  async function deleteDomain() {
    if (!domain.id) return;
    if (!domain.org_id) return;

    try {
      await client?.frontierServiceDeleteOrganizationDomain(
        domain.org_id,
        domain.id
      );
      navigate('/domains');
      toast.success('Domain deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
          <DropdownMenu.Item style={{ padding: 0 }}>
            <div
              onClick={() => navigate(`/domains/${domain.id}/verify`)}
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
            </div>
          </DropdownMenu.Item>
          <DropdownMenu.Item style={{ padding: 0 }}>
            <Link
              onClick={deleteDomain}
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
