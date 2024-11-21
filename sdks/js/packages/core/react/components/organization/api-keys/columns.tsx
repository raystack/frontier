import { ApsaraColumnDef } from '@raystack/apsara';
import { Flex, Text } from '@raystack/apsara/v1';
import { V1Beta1ServiceUser } from '~/api-client';

export const getColumns = (): ApsaraColumnDef<V1Beta1ServiceUser>[] => {
  return [
    {
      header: 'Name',
      accessorKey: 'title',
      cell: ({ row, getValue }) => {
        return (
          <Flex direction="column">
            <Text>{getValue()}</Text>
          </Flex>
        );
      }
    },
    {
      header: 'Created on',
      accessorKey: 'created_at',
      cell: ({ row, getValue }) => {
        return (
          <Flex direction="column">
            <Text>{getValue()}</Text>
          </Flex>
        );
      }
    },
    {
      header: '',
      accessorKey: 'id',
      enableSorting: false,
      cell: ({ row, getValue }) => {
        return (
          <Flex direction="column">
            <Text>Action</Text>
          </Flex>
        );
      }
    }
  ];
};
