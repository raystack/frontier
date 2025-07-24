import { useCallback, useEffect, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import { toast } from '@raystack/apsara/v1';

export const useTokens = () => {
  const { client, billingAccount } = useFrontier();

  const [tokenBalance, setTokenBalance] = useState(0);
  const [isTokensLoading, setIsTokensLoading] = useState(true);

  const getBalance = useCallback(
    async (orgId: string, billingAccountId: string) => {
      try {
        setIsTokensLoading(true);
        const resp = await client?.frontierServiceGetBillingBalance(
          orgId,
          billingAccountId
        );
        const tokens = resp?.data?.balance?.amount || '0';
        setTokenBalance(Number(tokens));
      } catch (err: any) {
        console.error(err);
        toast.error('Unable to fetch balance');
      } finally {
        setIsTokensLoading(false);
      }
    },
    [client]
  );

  const fetchTokenBalance = useCallback(() => {
    if (client && billingAccount?.org_id && billingAccount?.id) {
      getBalance(billingAccount?.org_id, billingAccount.id);
    }
  }, [billingAccount?.org_id, billingAccount?.id, client, getBalance]);

  useEffect(() => {
    fetchTokenBalance();
  }, [fetchTokenBalance]);

  return { tokenBalance, isTokensLoading, fetchTokenBalance };
};
