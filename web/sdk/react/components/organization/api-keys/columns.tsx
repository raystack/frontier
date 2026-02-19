import { TrashIcon } from '@radix-ui/react-icons';
import { Button, type DataTableColumnDef, Flex, Text } from '@raystack/apsara';
import { Link, useNavigate } from '@tanstack/react-router';
import type { ServiceUser } from '~/src';
import { timestampToDayjs } from '~/utils/timestamp';
import type { Timestamp } from '@bufbuild/protobuf/wkt';

export const getColumns = ({
  dateFormat
}: {
  dateFormat: string;
}): DataTableColumnDef<ServiceUser, unknown>[] => {
  return [
    {
      header: 'Name',
      accessorKey: 'title',
      cell: ({ row, getValue }) => {
        const value = getValue() as string;
        return (
          <Link
            to={`/api-keys/$id`}
            params={{
              id: row.original.id || ''
            }}
            state={{
              enableServiceUserTokensListFetch: true
            }}
            style={{
              textDecoration: 'none',
              color: 'var(--rs-color-foreground-base-primary)',
              fontSize: 'var(--rs-font-size-small)'
            }}
          >
            {value}
          </Link>
        );
      }
    },
    {
      header: 'Created on',
      accessorKey: 'createdAt',
      cell: ({ row, getValue }) => {
        const value = getValue() as Timestamp | undefined;
        return (
          <Flex direction="column">
            <Text>{timestampToDayjs(value)?.format(dateFormat) ?? '-'}</Text>
          </Flex>
        );
      }
    },
    {
      header: '',
      accessorKey: 'id',
      enableSorting: false,
      cell: ({ row, getValue }) => {
        const value = getValue() as string;
        return <ServiceAccountDeleteAction id={value} />;
      }
    }
  ];
};

function ServiceAccountDeleteAction({ id }: { id: string }) {
  const navigate = useNavigate({ from: '/api-keys' });

  function onDeleteClick() {
    return navigate({ to: '/api-keys/$id/delete', params: { id: id } });
  }
  return (
    <Button
      variant="text"
      size="small"
      color="danger"
      data-test-id="frontier-sdk-delete-service-account-btn"
      onClick={onDeleteClick}
    >
      <TrashIcon />
    </Button>
  );
}
