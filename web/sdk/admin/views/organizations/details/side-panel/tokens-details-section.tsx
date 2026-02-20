import { List, Text, Flex } from "@raystack/apsara";
import styles from "./side-panel.module.css";
import { MixIcon } from "@radix-ui/react-icons";
import { useContext, useEffect } from "react";
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

  const { data: tokensUsed = "0", isLoading: isTokensLoading, error } = useQuery(
    FrontierServiceQueries.totalDebitedTransactions,
    create(TotalDebitedTransactionsRequestSchema, {
      orgId: organizationId,
      billingId: billingAccountId,
    }),
    {
      enabled: !!organizationId && !!billingAccountId,
      select: (data) => String(data?.debited?.amount || "0"),
    }
  );

  useEffect(() => {
    if (error) {
      console.error("Error fetching debited transactions:", error);
    }
  }, [error]);

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
              <MixIcon />
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
              <MixIcon style={{ color: "var(--rs-color-foreground-base-tertiary)" }} />
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
              {Number(billingAccountDetails?.creditMin) < 0
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
              {Math.abs(Number(billingAccountDetails?.creditMin))}
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
            <Flex gap={3}>{String(billingAccountDetails?.dueInDays || 0)} days</Flex>
          )}
        </List.Value>
      </List.Item>
    </List>
  );
};
