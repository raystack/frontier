import { List, Text, Flex } from "@raystack/apsara";
import styles from "./side-panel.module.css";
import CoinIcon from "~/assets/icons/coin.svg?react";
import CoinColoredIcon from "~/assets/icons/coin-colored.svg?react";
import { useContext } from "react";
import Skeleton from "react-loading-skeleton";
import { OrganizationContext } from "../contexts/organization-context";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, TotalDebitedTransactionsRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

export const TokensDetailsSection = () => {
  const {
    tokenBalance,
    billingAccount,
    organization,
    isTokenBalanceLoading,
    billingAccountDetails,
  } = useContext(OrganizationContext);

  const organizationId = organization?.id || "";
  const billingAccountId = billingAccount?.id || "";

  const { data: debitedData, isLoading: isTokensLoading, error } = useQuery(
    FrontierServiceQueries.totalDebitedTransactions,
    create(TotalDebitedTransactionsRequestSchema, {
      orgId: organizationId,
      billingId: billingAccountId,
    }),
    {
      enabled: !!organizationId && !!billingAccountId,
    }
  );

  if (error) {
    console.error("Error fetching debited transactions:", error);
  }

  const tokensUsed = String(debitedData?.debited?.amount || "0");
  const isLoading = isTokensLoading;

  return (
    <List>
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
              {Number(billingAccountDetails?.credit_min) < 0
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
            <Flex gap={3}>
              {Math.abs(Number(billingAccountDetails?.credit_min))}
            </Flex>
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
    </List>
  );
};
