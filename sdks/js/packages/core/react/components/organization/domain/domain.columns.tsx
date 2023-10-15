import { CheckCircledIcon, TrashIcon } from '@radix-ui/react-icons';
import { Button, Flex, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Domain } from '~/src';
import Skeleton from 'react-loading-skeleton';

export const getColumns: (
  canCreateDomain?: boolean,
  isLoading?: boolean
) => ColumnDef<V1Beta1Domain, any>[] = (canCreateDomain, isLoading) => [
  {
    header: 'Name',
    accessorKey: 'name',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          return (
            <Flex direction="column">
              <Text>{row.original.name}</Text>
            </Flex>
          );
        }
  },
  {
    header: 'Created at',
    accessorKey: 'created_at',
    cell: isLoading
      ? () => <Skeleton />
      : info => <Text>{info.getValue()}</Text>
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
      : ({ row, getValue }) => (
          <DomainActions
            domain={row.original as V1Beta1Domain}
            canCreateDomain={canCreateDomain}
          />
        )
  }
];

const DomainActions = ({
  domain,
  canCreateDomain
}: {
  domain: V1Beta1Domain;
  canCreateDomain?: boolean;
}) => {
  const { client } = useFrontier();
  const navigate = useNavigate({ from: '/domains' });

  async function deleteDomain() {
    if (!domain.id) return;
    // @ts-ignore. TODO: fix buf openapi plugin
    if (!domain.org_id) return;

    try {
      await client?.frontierServiceDeleteOrganizationDomain(
        // @ts-ignore
        domain.org_id,
        domain.id
      );
      navigate({ to: '/domains' });
      toast.success('Domain deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return canCreateDomain ? (
    <Flex align="center" justify="end" gap="large">
      {domain.state === 'pending' ? (
        <Button
          style={{
            color: 'var(--background-base)',
            background: 'var(--foreground-base)'
          }}
          onClick={() =>
            navigate({
              to: `/domains/$domainId/verify`,
              params: {
                domainId: domain.id
              }
            })
          }
        >
          verify domain
        </Button>
      ) : (
        <Flex
          onClick={deleteDomain}
          gap="extra-small"
          style={{ color: 'var(--foreground-success)' }}
        >
          <CheckCircledIcon style={{ cursor: 'pointer' }} />
          Verified
        </Flex>
      )}

      <TrashIcon
        onClick={deleteDomain}
        color="var(--foreground-danger)"
        style={{ cursor: 'pointer' }}
      />
    </Flex>
  ) : null;
};
