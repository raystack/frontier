import { useMemo, useEffect } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { FrontierServiceQueries, GetBillingBalanceRequestSchema } from '@raystack/proton/frontier';
import { useFrontier } from '../contexts/FrontierContext';
import { toastManager } from '@raystack/apsara';

interface UseTokensReturn {
  tokenBalance: bigint;
  isTokensLoading: boolean;
  // unlike isTokensLoading, this is also true while an already cached
  // balance is being refetched
  isTokensFetching: boolean;
  fetchTokenBalance: () => Promise<any>;
}

export interface UseTokensOptions {
  // Set this to false to skip fetching the balance. The delete dialog uses
  // it so the balance is only fetched while the dialog is open.
  enabled?: boolean;
}

export const useTokens = (options: UseTokensOptions = {}): UseTokensReturn => {
  const { billingAccount } = useFrontier();

  const {
    data,
    isLoading: isTokensLoading,
    isFetching: isTokensFetching,
    error,
    refetch
  } = useQuery(
    FrontierServiceQueries.getBillingBalance,
    create(GetBillingBalanceRequestSchema, {
      id: billingAccount?.id ?? ''
    }),
    {
      enabled: !!billingAccount?.id && (options.enabled ?? true),
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
    isTokensFetching,
    fetchTokenBalance: refetch
  };
};
