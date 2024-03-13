import { DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { V1Beta1BillingTransaction } from '~/src';
import { getColumns } from './columns';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { useMemo } from 'react';
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
  const columns = useMemo(
    () =>
      getColumns({
        isLoading: isLoading,
        dateFormat: config.dateFormat || DEFAULT_DATE_FORMAT
      }),
    [config.dateFormat, isLoading]
  );
  const data = useMemo(() => {
    return isLoading
      ? [...new Array(3)].map<V1Beta1BillingTransaction>((_, i) => ({
          id: i.toString()
        }))
      : transactions;
  }, [isLoading, transactions]);
  return (
    <Flex>
      <DataTable columns={columns} data={data} emptyState={noDataChildren}>
        <DataTable.Toolbar>
          <Flex className={tokenStyles.txnTableHeader}>
            <Text size={2} weight={500}>
              Token transactions
            </Text>
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>No Transactions</h3>
  </EmptyState>
);
