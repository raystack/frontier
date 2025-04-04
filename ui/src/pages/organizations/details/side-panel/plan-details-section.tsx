import { V1Beta1Subscription, V1Beta1Plan } from "~/api/frontier";
import { useContext, useEffect, useState } from "react";
import { api } from "~/api";
import { List, Text, Flex } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import dayjs from "dayjs";
import { CalendarIcon } from "@radix-ui/react-icons";
import Skeleton from "react-loading-skeleton";
import { OrganizationContext } from "../contexts/organization-context";

export const PlanDetailsSection = () => {
  const { billingAccount, organization } = useContext(OrganizationContext);
  const [isSubscriptionLoading, setIsSubscriptionLoading] = useState(false);
  const [subscription, setSubscription] = useState<V1Beta1Subscription>();
  const [plan, setPlan] = useState<V1Beta1Plan>();

  const billingAccountId = billingAccount?.id || "";
  const organizationId = organization?.id || "";

  useEffect(() => {
    async function fetchBillingDetails(id: string, billingAccountId: string) {
      try {
        setIsSubscriptionLoading(true);
        const subResponse = await api?.frontierServiceListSubscriptions(
          id,
          billingAccountId,
        );
        const subscriptions = subResponse?.data?.subscriptions || [];
        const sub = subscriptions.find(
          (sub) => sub.state === "active" || sub.state === "trialing",
        );
        if (sub && sub.plan_id) {
          setSubscription(sub);
          const planResponse = await api?.frontierServiceGetPlan(sub.plan_id);
          const plan = planResponse?.data?.plan;
          setPlan(plan);
        }
      } catch (error) {
        console.error("Error fetching billing details:", error);
      } finally {
        setIsSubscriptionLoading(false);
      }
    }
    if (organizationId && billingAccountId) {
      fetchBillingDetails(organizationId, billingAccountId);
    }
  }, [organizationId, billingAccountId]);

  const isLoading = isSubscriptionLoading;

  return (
    <List.Root>
      <List.Header>Plan details</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Name
        </List.Label>
        <List.Value>
          {isLoading ? <Skeleton /> : <Text>{plan?.title || "Standard"}</Text>}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Started from
        </List.Label>
        <List.Value>
          {isLoading ? (
            <Skeleton />
          ) : subscription?.current_period_start_at ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>
                {dayjs(subscription?.current_period_start_at).format(
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
          ) : subscription?.current_period_end_at ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>
                {dayjs(subscription?.current_period_end_at).format(
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
          Payment mode
        </List.Label>
        <List.Value>
          <Text>Prepaid</Text>
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
