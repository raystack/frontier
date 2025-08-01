import { CheckCircledIcon, TrashIcon } from '@radix-ui/react-icons';
import { Button, Text, Flex } from '@raystack/apsara/v1';
import { useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import type { V1Beta1Domain } from '~/src';
import dayjs from 'dayjs';
import type { DataTableColumnDef } from '@raystack/apsara/v1';

interface getColumnsOptions {
  canCreateDomain?: boolean;
  dateFormat: string;
}

export const getColumns: (
  options: getColumnsOptions
) => DataTableColumnDef<V1Beta1Domain, unknown>[] = ({
  canCreateDomain,
  dateFormat
}) => [
  {
    header: 'Name',
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
    header: 'Created at',
    accessorKey: 'created_at',
    cell: info => (
      <Text>
        {dayjs(info.getValue() as string).format(`${dateFormat}, hh:mmA`)}
      </Text>
    )
  },
  {
    header: '',
    accessorKey: 'id',
    meta: {
      style: {
        textAlign: 'end'
      }
    },
    enableSorting: false,
    cell: ({ row, getValue }) => (
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

  return canCreateDomain ? (
    <Flex align="center" justify="end" gap={9}>
      {domain.state === 'pending' ? (
        <Button
          variant="outline"
          color="neutral"
          size="small"
          data-test-id="frontier-sdk-verify-domain-btn-verify"
          onClick={() =>
            navigate({
              to: `/domains/$domainId/verify`,
              params: {
                domainId: domain?.id || ''
              }
            })
          }
        >
          Verify domain
        </Button>
      ) : (
        <Flex
          gap={2}
          align="center"
          style={{ color: 'var(--rs-color-foreground-success-primary)' }}
        >
          <CheckCircledIcon style={{ cursor: 'pointer' }} />
          <Text variant="success">Verified</Text>
        </Flex>
      )}

      <TrashIcon
        data-test-id="frontier-sdk-delete-domain-btn"
        onClick={() =>
          navigate({
            to: `/domains/$domainId/delete`,
            params: {
              domainId: domain?.id || ''
            }
          })
        }
        color="var(--rs-color-foreground-danger-primary)"
        style={{ cursor: 'pointer' }}
      />
    </Flex>
  ) : null;
};
