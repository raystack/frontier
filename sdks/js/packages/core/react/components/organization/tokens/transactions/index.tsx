import { DataTable, Flex,  } from '@raystack/apsara';
import { EmptyState, Text } from '@raystack/apsara/v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { V1Beta1BillingTransaction } from '~/src';
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
    <Flex>
      <DataTable
        columns={columns}
        data={transactions}
        emptyState={noDataChildren}
        isLoading={isLoading}
      >
        <DataTable.Toolbar>
          <Flex className={tokenStyles.txnTableHeader}>
            <Text size="small" weight="medium">
              Token transactions
            </Text>
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading={"No Transactions"}
  />
);
