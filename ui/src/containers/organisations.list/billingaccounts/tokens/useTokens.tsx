import { useEffect, useState } from "react";
import { useFrontier } from "@raystack/frontier/react";
import { toast } from "sonner";

interface useTokensProps {
  organisationId?: string;
  billingaccountId?: string;
}

export const useTokens = function ({
  organisationId,
  billingaccountId,
}: useTokensProps) {
  const { client } = useFrontier();
  const [tokenBalance, setTokenBalance] = useState(0);
  const [isTokensLoading, setIsTokensLoading] = useState(false);

  useEffect(() => {
    async function getBalance(orgId: string, billingAccountId: string) {
      try {
        setIsTokensLoading(true);
        const resp = await client?.frontierServiceGetBillingBalance(
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
  }, [billingaccountId, client, organisationId]);
  return {
    tokenBalance,
    isTokensLoading,
  };
};
