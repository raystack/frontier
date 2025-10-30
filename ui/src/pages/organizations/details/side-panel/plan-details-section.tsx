import { useContext } from "react";
import { List, Text, Flex } from "@raystack/apsara";
import styles from "./side-panel.module.css";
import { CalendarIcon } from "@radix-ui/react-icons";
import Skeleton from "react-loading-skeleton";
import { OrganizationContext } from "../contexts/organization-context";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, ListSubscriptionsRequestSchema, GetPlanRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { timestampToDayjs } from "~/utils/connect-timestamp";

export const PlanDetailsSection = () => {
  const { billingAccount, organization } = useContext(OrganizationContext);

  const billingAccountId = billingAccount?.id || "";
  const organizationId = organization?.id || "";

  const { data: activeSubscription, isLoading: isSubscriptionLoading, error: subscriptionsError } = useQuery(
    FrontierServiceQueries.listSubscriptions,
    create(ListSubscriptionsRequestSchema, {
      orgId: organizationId,
      billingId: billingAccountId,
    }),
    {
      enabled: !!organizationId && !!billingAccountId,
      select: (data) => {
        const subscriptions = data?.subscriptions ?? [];
        return subscriptions.find(
          (sub) => sub.state === "active" || sub.state === "trialing"
        );
      },
    }
  );

  if (subscriptionsError) {
    console.error("Error fetching subscriptions:", subscriptionsError);
  }

  const { data: planData, isLoading: isPlanLoading, error: planError } = useQuery(
    FrontierServiceQueries.getPlan,
    create(GetPlanRequestSchema, {
      id: activeSubscription?.planId ?? "",
    }),
    {
      enabled: !!activeSubscription?.planId,
    }
  );

  if (planError) {
    console.error("Error fetching plan:", planError);
  }

  const isLoading = isSubscriptionLoading || isPlanLoading;

  return (
    <List>
      <List.Header>Plan details</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Name
        </List.Label>
        <List.Value>
          {isLoading ? <Skeleton /> : <Text>{planData?.plan?.title || "Standard"}</Text>}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Started from
        </List.Label>
        <List.Value>
          {isLoading ? (
            <Skeleton />
          ) : timestampToDayjs(activeSubscription?.currentPeriodStartAt) ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>
                {timestampToDayjs(activeSubscription?.currentPeriodStartAt)?.format(
                  "DD MMM YYYY",
                )}
              </Text>
            </Flex>
          ) : (
            <Text>-</Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Ends on
        </List.Label>
        <List.Value>
          {isLoading ? (
            <Skeleton />
          ) : timestampToDayjs(activeSubscription?.currentPeriodEndAt) ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>
                {timestampToDayjs(activeSubscription?.currentPeriodEndAt)?.format(
                  "DD MMM YYYY",
                )}
              </Text>
            </Flex>
          ) : (
            <Text>-</Text>
          )}
        </List.Value>
      </List.Item>
    </List>
  );
};
