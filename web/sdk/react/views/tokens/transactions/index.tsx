import {
  EmptyState,
  Text,
  Flex,
  DataTable,
  DataTableSort
} from '@raystack/apsara';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import type { BillingTransaction } from '~/src';
import { getColumns } from './columns';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import tokenStyles from '../token.module.css';

interface TransactionsTableProps {
  transactions: BillingTransaction[];
  isLoading: boolean;
}

const DEFAULT_SORT: DataTableSort = { name: 'createdAt', order: 'desc' };

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
      defaultSort={DEFAULT_SORT}
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
