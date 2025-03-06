import { useEffect, useState } from "react";
import { toast } from "sonner";
import { api } from "~/api";

interface useTokensProps {
  organisationId?: string;
  billingaccountId?: string;
}

export const useTokens = function ({
  organisationId,
  billingaccountId,
}: useTokensProps) {
  const [tokenBalance, setTokenBalance] = useState(0);
  const [isTokensLoading, setIsTokensLoading] = useState(false);

  useEffect(() => {
    async function getBalance(orgId: string, billingAccountId: string) {
      try {
        setIsTokensLoading(true);
        const resp = await api?.frontierServiceGetBillingBalance(
          orgId,
          billingAccountId
        );
        const tokens = resp?.data?.balance?.amount || "0";
        setTokenBalance(Number(tokens));
      } catch (err: any) {
        console.error(err);
        toast.error("Unable to fetch balance");
      } finally {
        setIsTokensLoading(false);
      }
    }

    if (organisationId && billingaccountId) {
      getBalance(organisationId, billingaccountId);
    }
  }, [billingaccountId, organisationId]);
  return {
    tokenBalance,
    isTokensLoading,
  };
};
