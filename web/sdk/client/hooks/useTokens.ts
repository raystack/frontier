import { useMemo, useEffect } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { FrontierServiceQueries, GetBillingBalanceRequestSchema } from '@raystack/proton/frontier';
import { useFrontier } from '../contexts/FrontierContext';
import { toastManager } from '@raystack/apsara';

interface UseTokensReturn {
  tokenBalance: bigint;
  isTokensLoading: boolean;
  fetchTokenBalance: () => Promise<any>;
}

export const useTokens = (): UseTokensReturn => {
  const { billingAccount } = useFrontier();

  const {
    data,
    isLoading: isTokensLoading,
    error,
    refetch
  } = useQuery(
    FrontierServiceQueries.getBillingBalance,
    create(GetBillingBalanceRequestSchema, {
      id: billingAccount?.id ?? ''
    }),
    {
      enabled: !!billingAccount?.id,
      retry: false
    }
  );

  // Handle errors
  useEffect(() => {
    if (error) {
      console.error(error);
      toastManager.add({
        title: 'Unable to fetch balance',
        type: 'error'
      });
    }
  }, [error]);

  const tokenBalance = useMemo(
    () => BigInt(data?.balance?.amount || '0'),
    [data?.balance?.amount]
  );

  return {
    tokenBalance,
    isTokensLoading,
    fetchTokenBalance: refetch
  };
};
