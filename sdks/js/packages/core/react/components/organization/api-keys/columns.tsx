import { TrashIcon } from '@radix-ui/react-icons';
import {
  Button,
  type DataTableColumnDef,
  Flex,
  Text
} from '@raystack/apsara/v1';
import { Link, useNavigate } from '@tanstack/react-router';
import dayjs from 'dayjs';
import type { V1Beta1ServiceUser } from '~/api-client';

export const getColumns = ({
  dateFormat
}: {
  dateFormat: string;
}): DataTableColumnDef<V1Beta1ServiceUser, unknown>[] => {
  return [
    {
      header: 'Name',
      accessorKey: 'title',
      cell: ({ row, getValue }) => {
        return (
          <Link
            to={`/api-keys/$id`}
            params={{
              id: row.original.id || ''
            }}
            style={{
              textDecoration: 'none',
              color: 'var(--rs-color-foreground-base-primary)'
            }}
          >
            {getValue()}
          </Link>
        );
      }
    },
    {
      header: 'Created on',
      accessorKey: 'created_at',
      cell: ({ row, getValue }) => {
        const value = getValue();
        return (
          <Flex direction="column">
            <Text>{dayjs(value).format(dateFormat)}</Text>
          </Flex>
        );
      }
    },
    {
      header: '',
      accessorKey: 'id',
      enableSorting: false,
      cell: ({ row, getValue }) => {
        return <ServiceAccountDeleteAction id={getValue()} />;
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
      data-test-id="frontier-sdk-delete-service-account-btn"
      onClick={onDeleteClick}
    >
      <TrashIcon />
    </Button>
  );
}
