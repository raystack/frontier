import { EmptyState, Text, Flex, DataTable } from '@raystack/apsara/v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import type { V1Beta1BillingTransaction } from '~/src';
import { getColumns } from './columns';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import tokenStyles from '../token.module.css';

interface TransactionsTableProps {
  transactions: V1Beta1BillingTransaction[];
  isLoading: boolean;
}

export function TransactionsTable({
  isLoading,
  transactions
}: TransactionsTableProps) {
  const { config } = useFrontier();
  const columns = getColumns({
    dateFormat: config.dateFormat || DEFAULT_DATE_FORMAT
  });

  return (
    <DataTable
      columns={columns}
      data={transactions}
      isLoading={isLoading}
      mode="client"
      defaultSort={{ name: 'created_at', order: 'desc' }}
    >
      <Flex gap={7} direction="column">
        <Text size="small" weight="medium">
          Token transactions
        </Text>
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{ header: tokenStyles.txnTableHeader }}
        />
      </Flex>
    </DataTable>
  );
}

const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading={'No Transactions'} />
);
