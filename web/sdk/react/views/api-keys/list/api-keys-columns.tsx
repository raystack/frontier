import { TrashIcon } from '@radix-ui/react-icons';
import { Button, type DataTableColumnDef, Flex, Text } from '@raystack/apsara';
import type { ServiceUser } from '~/src';
import { timestampToDayjs } from '~/utils/timestamp';
import type { Timestamp } from '@bufbuild/protobuf/wkt';

interface GetColumnsOptions {
  dateFormat: string;
  onServiceAccountClick?: (id: string) => void;
  onDeleteClick?: (id: string) => void;
}

export const getColumns = ({
  dateFormat,
  onServiceAccountClick,
  onDeleteClick
}: GetColumnsOptions): DataTableColumnDef<ServiceUser, unknown>[] => {
  return [
    {
      header: 'Name',
      accessorKey: 'title',
      cell: ({ row, getValue }) => {
        const value = getValue() as string;
        return (
          <Text
            size="small"
            style={{
              textDecoration: 'none',
              color: 'var(--rs-color-foreground-base-primary)',
              cursor: 'pointer'
            }}
            onClick={() => onServiceAccountClick?.(row.original.id || '')}
          >
            {value}
          </Text>
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
        return (
          <Button
            variant="text"
            size="small"
            color="danger"
            data-test-id="frontier-sdk-delete-service-account-btn"
            onClick={() => onDeleteClick?.(value)}
          >
            <TrashIcon />
          </Button>
        );
      }
    }
  ];
};
