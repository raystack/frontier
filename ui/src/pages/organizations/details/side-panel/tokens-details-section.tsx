import { List, Text, Flex } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import CoinIcon from "~/assets/icons/coin.svg?react";
import CoinColoredIcon from "~/assets/icons/coin-colored.svg?react";
import { useEffect, useState } from "react";
import Skeleton from "react-loading-skeleton";
import { api } from "~/api";

interface TokensDetailsSectionProps {
  organizationId: string;
  isLoading: boolean;
  billingAccountId: string;
}

export const TokensDetailsSection = ({
  organizationId,
  billingAccountId,
  isLoading,
}: TokensDetailsSectionProps) => {
  const [balance, setBalance] = useState("0");
  const [tokensUsed, setTokensUsed] = useState(0);
  const [isTokensLoading, setIsTokensLoading] = useState(false);

  useEffect(() => {
    async function fetchTokenDetails(id: string, billingAccountId: string) {
      try {
        setIsTokensLoading(true);
        const [balanceResp] = await Promise.all([
          api.frontierServiceGetBillingBalance(id, billingAccountId),
        ]);
        const newBalance = balanceResp.data.balance?.amount || "0";
        setBalance(newBalance);
      } catch (error) {
        console.error(error);
      } finally {
        setIsTokensLoading(false);
      }
    }
    if (organizationId && billingAccountId) {
      fetchTokenDetails(organizationId, billingAccountId);
    }
  }, [organizationId, billingAccountId]);

  const isDataLoading = isLoading || isTokensLoading;

  return (
    <List.Root>
      <List.Header>Tokens</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Available tokens
        </List.Label>
        <List.Value>
          {isDataLoading ? (
            <Skeleton />
          ) : (
            <Flex gap={3}>
              <CoinColoredIcon />
              <Text>{balance}</Text>
            </Flex>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Used till date
        </List.Label>
        <List.Value>
          {isDataLoading ? (
            <Skeleton />
          ) : (
            <Flex gap={3}>
              <CoinIcon color="var(--rs-color-foreground-base-tertiary)" />
              <Text>{tokensUsed}</Text>
            </Flex>
          )}
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
