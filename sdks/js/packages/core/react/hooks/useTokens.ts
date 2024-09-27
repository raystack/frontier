import { useEffect, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import { toast } from 'sonner';

export const useTokens = () => {
  const { client, activeOrganization, billingAccount } = useFrontier();

  const [tokenBalance, setTokenBalance] = useState(0);
  const [isTokensLoading, setIsTokensLoading] = useState(true);

  useEffect(() => {
    async function getBalance(orgId: string, billingAccountId: string) {
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
    }

    if (client && activeOrganization?.id && billingAccount?.id) {
      getBalance(activeOrganization.id, billingAccount.id);
    }
  }, [billingAccount?.id, client, activeOrganization?.id]);

  return { tokenBalance, isTokensLoading };
};
