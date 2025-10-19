import { useMemo, useEffect } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { FrontierServiceQueries, GetBillingBalanceRequestSchema } from '@raystack/proton/frontier';
import { useFrontier } from '../contexts/FrontierContext';
import { toast } from '@raystack/apsara';

export const useTokens = () => {
  const { billingAccount } = useFrontier();

  const {
    data,
    isLoading: isTokensLoading,
    error,
    refetch
  } = useQuery(
    FrontierServiceQueries.getBillingBalance,
    create(GetBillingBalanceRequestSchema, {
      orgId: billingAccount?.orgId ?? '',
      id: billingAccount?.id ?? ''
    }),
    {
      enabled: !!billingAccount?.orgId && !!billingAccount?.id,
      retry: false
    }
  );

  // Handle errors
  useEffect(() => {
    if (error) {
      console.error(error);
      toast.error('Unable to fetch balance');
    }
  }, [error]);

  const tokenBalance = useMemo(
    () => Number(data?.balance?.amount || '0'),
    [data?.balance?.amount]
  );

  const fetchTokenBalance = () => {
    refetch();
  };

  return { tokenBalance, isTokensLoading, fetchTokenBalance };
};
