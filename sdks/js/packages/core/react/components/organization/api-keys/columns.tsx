import { TrashIcon } from '@radix-ui/react-icons';
import { ApsaraColumnDef } from '@raystack/apsara';
import { Button, Flex, Text } from '@raystack/apsara/v1';
import { Link, useNavigate } from '@tanstack/react-router';
import dayjs from 'dayjs';
import { V1Beta1ServiceUser } from '~/api-client';

export const getColumns = ({
  dateFormat
}: {
  dateFormat: string;
}): ApsaraColumnDef<V1Beta1ServiceUser>[] => {
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
              color: 'var(--rs-color-text-base-primary)'
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
      meta: {
        style: {
          padding: 0
        }
      },
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
