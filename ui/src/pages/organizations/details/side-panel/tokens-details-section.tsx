import { List, Text, Flex } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import CoinIcon from "~/assets/icons/coin.svg?react";
import CoinColoredIcon from "~/assets/icons/coin-colored.svg?react";
import { useContext, useEffect, useState } from "react";
import Skeleton from "react-loading-skeleton";
import { api } from "~/api";
import { OrganizationContext } from "../contexts/organization-context";

export const TokensDetailsSection = () => {
  const {
    tokenBalance,
    billingAccount,
    organization,
    isTokenBalanceLoading,
    billingAccountDetails,
  } = useContext(OrganizationContext);
  const [tokensUsed, setTokensUsed] = useState("0");
  const [isTokensLoading, setIsTokensLoading] = useState(false);

  const organizationId = organization?.id || "";
  const billingAccountId = billingAccount?.id || "";

  useEffect(() => {
    async function fetchTokenUsed(id: string, billingAccountId: string) {
      try {
        setIsTokensLoading(true);
        const resp = await api.frontierServiceTotalDebitedTransactions(
          id,
          billingAccountId,
        );

        const newTokensUsed = resp.data.debited?.amount || "0";
        setTokensUsed(newTokensUsed);
      } catch (error) {
        console.error(error);
      } finally {
        setIsTokensLoading(false);
      }
    }
    if (organizationId && billingAccountId) {
      fetchTokenUsed(organizationId, billingAccountId);
    }
  }, [organizationId, billingAccountId]);

  const isLoading = isTokensLoading;

  return (
    <List.Root>
      <List.Header>Tokens</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Available tokens
        </List.Label>
        <List.Value>
          {isTokenBalanceLoading ? (
            <Skeleton />
          ) : (
            <Flex gap={3}>
              <CoinColoredIcon />
              <Text>{tokenBalance}</Text>
            </Flex>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Used till date
        </List.Label>
        <List.Value>
          {isLoading ? (
            <Skeleton />
          ) : (
            <Flex gap={3}>
              <CoinIcon color="var(--rs-color-foreground-base-tertiary)" />
              <Text>{tokensUsed}</Text>
            </Flex>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Payment Mode
        </List.Label>
        <List.Value>
          {isLoading ? (
            <Skeleton />
          ) : (
            <Flex gap={3}>
              {Number(billingAccountDetails?.credit_min) > 0
                ? "Postpaid"
                : "Prepaid"}
            </Flex>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Credit Limit
        </List.Label>
        <List.Value>
          {isLoading ? (
            <Skeleton />
          ) : (
            <Flex gap={3}>{billingAccountDetails?.credit_min}</Flex>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Invoice due date
        </List.Label>
        <List.Value>
          {isLoading ? (
            <Skeleton />
          ) : (
            <Flex gap={3}>{billingAccountDetails?.due_in_days} days</Flex>
          )}
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
